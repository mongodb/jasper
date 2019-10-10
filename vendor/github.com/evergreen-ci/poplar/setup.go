package poplar

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/evergreen-ci/aviation"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	yaml "gopkg.in/yaml.v2"
)

const (
	ProjectEnv      = "project"
	VersionEnv      = "version_id"
	OrderEnv        = "revision_order_id"
	VariantEnv      = "build_variant"
	TaskNameEnv     = "task_name"
	ExecutionEnv    = "execution"
	MainlineEnv     = "is_patch"
	APIKeyEnv       = "API_KEY"
	APISecretEnv    = "API_SECRET"
	APITokenEnv     = "API_TOKEN"
	BucketNameEnv   = "BUCKET_NAME"
	BucketPrefixEnv = "BUCKET_PREFIX"
	BucketRegionEnv = "BUCKET_REGION"
)

// ReportType describes the marshalled report type.
type ReportType string

const (
	ReportTypeJSON ReportType = "JSON"
	ReportTypeBSON ReportType = "BSON"
	ReportTypeYAML ReportType = "YAML"
	ReportTypeEnv  ReportType = "ENV"
)

// ReportSetup sets up a Report struct with the given ReportType and filename.
// Note that not all ReportTypes require a filename (such as ReportTypeEnv), if
// this is the case, pass in an empty string for the filename.
func ReportSetup(reportType ReportType, filename string) (*Report, error) {
	switch reportType {
	case ReportTypeJSON:
		return reportSetupUnmarshal(reportType, filename, json.Unmarshal)
	case ReportTypeBSON:
		return reportSetupUnmarshal(reportType, filename, bson.Unmarshal)
	case ReportTypeYAML:
		return reportSetupUnmarshal(reportType, filename, yaml.Unmarshal)
	case ReportTypeEnv:
		return reportSetupEnv()
	default:
		return nil, errors.Errorf("invalid report type %s", reportType)
	}
}

func reportSetupUnmarshal(reportType ReportType, filename string, unmarshal func([]byte, interface{}) error) (*Report, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "problem opening file %s", filename)
	}

	report := &Report{}
	if err := unmarshal(data, report); err != nil {
		return nil, errors.Wrapf(err, "problem unmarshalling %s from %s", reportType, filename)
	}

	return report, nil
}

func reportSetupEnv() (*Report, error) {
	var order int
	var execution int
	var mainline bool
	var err error
	if os.Getenv(MainlineEnv) != "" {
		mainline, err = strconv.ParseBool(os.Getenv(MainlineEnv))
		if err != nil {
			return nil, errors.Wrapf(err, "env var %s should be a bool", MainlineEnv)
		}
	}
	if mainline && os.Getenv(OrderEnv) != "" {
		order, err = strconv.Atoi(os.Getenv(OrderEnv))
		if err != nil {
			return nil, errors.Wrapf(err, "env var %s should be an int", OrderEnv)
		}
	}
	if os.Getenv(ExecutionEnv) != "" {
		execution, err = strconv.Atoi(os.Getenv(ExecutionEnv))
		if err != nil {
			return nil, errors.Wrapf(err, "env var %s should be an int", ExecutionEnv)
		}
	}

	return &Report{
		Project:   os.Getenv(ProjectEnv),
		Version:   os.Getenv(VersionEnv),
		Order:     order,
		Variant:   os.Getenv(VariantEnv),
		TaskName:  os.Getenv(TaskNameEnv),
		Execution: execution,
		Mainline:  mainline,
		BucketConf: BucketConfiguration{
			APIKey:    os.Getenv(APIKeyEnv),
			APISecret: os.Getenv(APISecretEnv),
			APIToken:  os.Getenv(APITokenEnv),
			Name:      os.Getenv(BucketNameEnv),
			Prefix:    os.Getenv(BucketPrefixEnv),
			Region:    os.Getenv(BucketRegionEnv),
		},
		Tests: []Test{},
	}, nil
}

type userCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// DialCedar is a convenience function for creating a RPC client connection
// with cedar via gRPC. The username and password are LDAP credentials for the
// cedar service.
func DialCedar(ctx context.Context, username, password string, retries int) (*grpc.ClientConn, error) {
	cedarRPCAddress := "cedar.mongodb.com:7070"

	creds := &userCredentials{
		Username: username,
		Password: password,
	}
	credsPayload, err := json.Marshal(creds)
	if err != nil {
		return nil, errors.Wrap(err, "problem building credentials payload")
	}

	ca, err := makeCedarCertRequest(ctx, "/ca", nil)
	if err != nil {
		return nil, errors.Wrap(err, "problem getting cedar root cert")
	}
	crt, err := makeCedarCertRequest(ctx, "/users/certificate", bytes.NewBuffer(credsPayload))
	if err != nil {
		return nil, errors.Wrap(err, "problem getting cedar user cert")
	}
	key, err := makeCedarCertRequest(ctx, "/users/certificate/key", bytes.NewBuffer(credsPayload))
	if err != nil {
		return nil, errors.Wrap(err, "problem getting cedar user key")
	}

	tlsConf, err := aviation.GetClientTLSConfig(ca, crt, key)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating TLS config")
	}

	return aviation.Dial(ctx, aviation.DialOptions{
		Address: cedarRPCAddress,
		Retries: retries,
		TLSConf: tlsConf,
	})
}

func makeCedarCertRequest(ctx context.Context, url string, body io.Reader) ([]byte, error) {
	cedarHTTPAddress := "https://cedar.mongodb.com/rest/v1/admin"
	client := &http.Client{Timeout: 5 * time.Minute}

	req, err := http.NewRequest(http.MethodGet, cedarHTTPAddress+url, body)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating http request")
	}
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "problem with request")
	}
	defer resp.Body.Close()

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "problem reading response")
	}

	if resp.StatusCode != http.StatusOK {
		return out, errors.Errorf("failed request with status code %d", resp.StatusCode)
	}

	return out, nil
}
