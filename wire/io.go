package wire

import (
	"github.com/evergreen-ci/mrpc/mongowire"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
)

func intOK(ok bool) int {
	if ok {
		return 1
	}
	return 0
}

// ErrorResponse represents a response indicating whether the operation was okay
// and errors, if any.
type ErrorResponse struct {
	OK    int    `bson:"ok"`
	Error string `bson:"errmsg,omitempty"`
}

// makeErrorResponse returns an ErrorResponse with the given ok status and error
// message, if any.
func makeErrorResponse(ok bool, err error) ErrorResponse {
	resp := ErrorResponse{OK: intOK(ok)}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp
}

// makeSuccessResponse returns an ErrorResponse that is ok and has no error.
func makeSuccessResponse() ErrorResponse {
	return ErrorResponse{OK: intOK(true)}
}

// Message returns the ErrorResponse as an equivalent mongowire.Message. The
// inverse operation is ExtractErrorResponse.
func (r ErrorResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// ExtractErrorResponse extracts an ErrorResponse from the given
// mongowire.Message. The inverse operation is (ErrorResponse).Message.
func ExtractErrorResponse(msg mongowire.Message) (ErrorResponse, error) {
	r := ErrorResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.Error != "" {
		return r, errors.New(r.Error)
	}
	if r.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// InfoRequest represents a request for runtime information regarding the
// process given by ID.
type InfoRequest struct {
	ID string `bson:"info"`
}

func makeInfoRequest(id string) InfoRequest { //nolint
	return InfoRequest{ID: id}
}

// Message returns the InfoRequest as an equivalent mongowire.Message. The
// inverse operation is ExtractInfoRequest.
func (r InfoRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// ExtractInfoRequest extracts an InfoRequest from the given mongowire.Message.
// The inverse operation is (InfoRequest).Message.
func ExtractInfoRequest(msg mongowire.Message) (InfoRequest, error) {
	r := InfoRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// InfoResponse represents a response indicating runtime information for a
// process.
type InfoResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Info          jasper.ProcessInfo `bson:"info"`
}

func makeInfoResponse(info jasper.ProcessInfo) InfoResponse {
	return InfoResponse{Info: info, ErrorResponse: makeSuccessResponse()}
}

// Message returns the InfoResponse as an equivalent mongowire.Message. The
// inverse operation is ExtractInfoResponse.
func (r InfoResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// ExtractInfoResponse extracts an InfoResponse from the given
// mongowire.Message. The inverse operation is (InfoResponse).Message.
func ExtractInfoResponse(msg mongowire.Message) (InfoResponse, error) {
	r := InfoResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.Error != "" {
		return r, errors.New(r.Error)
	}
	if r.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// RunningRequest represents a request for the running state of the process
// given by ID.
type RunningRequest struct {
	ID string `bson:"running"`
}

func makeRunningRequest(id string) RunningRequest { //nolint
	return RunningRequest{ID: id}
}

// Message returns the RunningRequest as an equivalent mongowire.Message. The
// inverse operation is ExtractRunningRequest.
func (r RunningRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// ExtractRunningRequest extracts a RunningRequest from the given
// mongowire.Message. The inverse operation is (RunningRequest).Message.
func ExtractRunningRequest(msg mongowire.Message) (RunningRequest, error) {
	r := RunningRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// RunningResponse represents a response indicating the running state of a
// process.
type RunningResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Running       bool `bson:"running"`
}

func makeRunningResponse(running bool) RunningResponse {
	return RunningResponse{Running: running, ErrorResponse: makeSuccessResponse()}
}

// Message returns the RunningResponse as an equivalent mongowire.Message. The
// inverse operation is ExtractRunningResponse.
func (r RunningResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// ExtractRunningResponse extracts a RunningResponse from the given
// mongowire.Message. The inverse operation is (RunningResponse).Message.
func ExtractRunningResponse(msg mongowire.Message) (RunningResponse, error) {
	r := RunningResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.Error != "" {
		return r, errors.New(r.Error)
	}
	if r.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// CompleteRequest represents a request for the completion status of the process
// given by ID.
type CompleteRequest struct {
	ID string `bson:"complete"`
}

func makeCompleteRequest(id string) CompleteRequest { //nolint
	return CompleteRequest{ID: id}
}

// Message returns the CompleteRequest as an equivalent mongowire.Message. The
// inverse operation is ExtractCompleteRequest.
func (r CompleteRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// ExtractCompleteRequest extracts a CompleteRequest from the given
// mongowire.Message. The inverse operation is (CompleteRequest).Message.
func ExtractCompleteRequest(msg mongowire.Message) (CompleteRequest, error) {
	r := CompleteRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// CompleteResponse represents a response indicating the completion status of a
// process.
type CompleteResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Complete      bool `bson:"complete"`
}

func makeCompleteResponse(complete bool) CompleteResponse {
	return CompleteResponse{Complete: complete, ErrorResponse: makeSuccessResponse()}
}

// Message returns the CompleteResponse as an equivalent mongowire.Message. The
// inverse operation is ExtractCompleteResponse.
func (r CompleteResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// ExtractCompleteResponse extracts a CompleteResponse from the given
// mongowire.Message. The inverse operation is (CompleteResponse).Message.
func ExtractCompleteResponse(msg mongowire.Message) (CompleteResponse, error) {
	r := CompleteResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.Error != "" {
		return r, errors.New(r.Error)
	}
	if r.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// WaitRequest represents a request for the wait status of the process given  by
// ID.
type WaitRequest struct {
	ID string `bson:"wait"`
}

func makeWaitRequest(id string) WaitRequest { //nolint
	return WaitRequest{ID: id}
}

// Message returns the WaitRequest as an equivalent mongowire.Message. The
// inverse operation is ExtractWaitRequest.
func (r WaitRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// ExtractWaitRequest extracts a WaitRequest from the given mongowire.Message.
// The inverse operation is (WaitRequest).Message.
func ExtractWaitRequest(msg mongowire.Message) (WaitRequest, error) {
	r := WaitRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// WaitResponse represents a response indicating the exit code and error of
// a waited process.
type WaitResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	ExitCode      int `bson:"exit_code"`
}

func makeWaitResponse(exitCode int, err error) WaitResponse {
	return WaitResponse{ExitCode: exitCode, ErrorResponse: makeErrorResponse(true, err)}
}

// Message returns the WaitResponse as an equivalent mongowire.Message. The
// inverse operation is ExtractWaitResponse.
func (r WaitResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// ExtractWaitResponse extracts a WaitResponse from the given
// mongowire.Message. The inverse operation is (WaitResponse).Message.
func ExtractWaitResponse(msg mongowire.Message) (WaitResponse, error) {
	r := WaitResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.Error != "" {
		return r, errors.New(r.Error)
	}
	if r.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// RespawnRequest represents a request to respawn the process given by ID.
type RespawnRequest struct {
	ID string `bson:"respawn"`
}

func makeRespawnRequest(id string) RespawnRequest { //nolint
	return RespawnRequest{ID: id}
}

// Message returns the RespawnRequest as an equivalent mongowire.Message. The
// inverse operation is ExtractRespawnRequest.
func (r RespawnRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// ExtractRespawnRequest extracts a RespawnRequest from the given
// mongowire.Message. The inverse operation is (RespawnRequest).Message.
func ExtractRespawnRequest(msg mongowire.Message) (RespawnRequest, error) {
	r := RespawnRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// SignalRequest represents a request to send a signal to the process given by
// ID.
type SignalRequest struct {
	Params struct {
		ID     string `bson:"id"`
		Signal int    `bson:"signal"`
	} `bson:"signal"`
}

func makeSignalRequest(id string, signal int) SignalRequest { //nolint
	req := SignalRequest{}
	req.Params.ID = id
	req.Params.Signal = signal
	return req
}

// Message returns the SignalRequest as an equivalent mongowire.Message. The
// inverse operation is ExtractSignalRequest.
func (r SignalRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// ExtractSignalRequest extracts a SignalRequest from the given
// mongowire.Message. The inverse operation is (SignalRequest).Message.
func ExtractSignalRequest(msg mongowire.Message) (SignalRequest, error) {
	r := SignalRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// RegisterSignalTriggerIDRequest represents a request to register the signal
// trigger ID on the process given by ID.
type RegisterSignalTriggerIDRequest struct {
	Params struct {
		ID              string                 `bson:"id"`
		SignalTriggerID jasper.SignalTriggerID `bson:"signal_trigger_id"`
	} `bson:"register_signal_trigger_id"`
}

func makeRegisterSignalTriggerIDRequest(id string, sigID jasper.SignalTriggerID) RegisterSignalTriggerIDRequest { //nolint
	req := RegisterSignalTriggerIDRequest{}
	req.Params.ID = id
	req.Params.SignalTriggerID = sigID
	return req
}

// Message returns the RegisterSignalTriggerIDRequest as an equivalent
// mongowire.Message. The inverse operation is
// ExtractRegisterSignalTriggerIDRequest.
func (r RegisterSignalTriggerIDRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// ExtractRegisterSignalTriggerIDRequest extracts a
// RegisterSignalTriggerIDRequest from the given mongowire.Message. The inverse
// operation is (RegisterSignalTriggerIDRequest).Message.
func ExtractRegisterSignalTriggerIDRequest(msg mongowire.Message) (RegisterSignalTriggerIDRequest, error) {
	r := RegisterSignalTriggerIDRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// TagRequest represents a request to associate the process given by ID with the
// tag.
type TagRequest struct {
	Params struct {
		ID  string `bson:"id"`
		Tag string `bson:"tag"`
	} `bson:"tag"`
}

func makeTagRequest(id, tag string) TagRequest { //nolint
	req := TagRequest{}
	req.Params.ID = id
	req.Params.Tag = tag
	return req
}

// Message returns the TagRequest as an equivalent mongowire.Message. The
// inverse operation is ExtractTagRequest.
func (r TagRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// ExtractTagRequest extracts a TagRequest from the given mongowire.Message. The
// inverse operation is (TagRequest).Message.
func ExtractTagRequest(msg mongowire.Message) (TagRequest, error) {
	r := TagRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// GetTagsRequest represents a request to get all the tags for the process given
// by ID.
type GetTagsRequest struct {
	ID string `bson:"get_tags"`
}

func makeGetTagsRequest(id string) GetTagsRequest { //nolint
	return GetTagsRequest{ID: id}
}

// Message returns the GetTagsRequest as an equivalent mongowire.Message. The
// inverse operation is ExtractGetTagsRequest.
func (r GetTagsRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// ExtractGetTagsRequest extracts a GetTagsRequest from the given
// mongowire.Message. The inverse operation is (GetTagsRequest).Message.
func ExtractGetTagsRequest(msg mongowire.Message) (GetTagsRequest, error) {
	r := GetTagsRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// GetTagsResponse represents a response indicating the tags of a process.
type GetTagsResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Tags          []string `bson:"tags"`
}

func makeGetTagsResponse(tags []string) GetTagsResponse {
	return GetTagsResponse{Tags: tags, ErrorResponse: makeSuccessResponse()}
}

// Message returns the GetTagsResponse as an equivalent mongowire.Message. The
// inverse operation is ExtractGetTagsResponse.
func (r GetTagsResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// ExtractGetTagsResponse extracts a GetTagsResponse from the given
// mongowire.Message. The inverse operation is (GetTagsResponse).Message.
func ExtractGetTagsResponse(msg mongowire.Message) (GetTagsResponse, error) {
	r := GetTagsResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.Error != "" {
		return r, errors.New(r.Error)
	}
	if r.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// ResetTagsRequest represents a request to clear all the tags for the process
// given by ID.
type ResetTagsRequest struct {
	ID string `bson:"reset_tags"`
}

func makeResetTagsRequest(id string) ResetTagsRequest { //nolint
	return ResetTagsRequest{ID: id}
}

// Message returns the ResetTagsRequest as an equivalent mongowire.Message. The
// inverse operation is ExtractResetTagsRequest.
func (r ResetTagsRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// ExtractResetTagsRequest extracts a ResetTagsRequest from the given
// mongowire.Message. The inverse operation is (ResetTagsRequest).Message.
func ExtractResetTagsRequest(msg mongowire.Message) (ResetTagsRequest, error) {
	r := ResetTagsRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// IDRequest represents a request to get the ID associated with the service
// manager.
type IDRequest struct {
	ID int `bson:"id"`
}

func makeIDRequest() IDRequest { //nolint
	return IDRequest{ID: 1}
}

// Message returns the IDRequest as an equivalent mongowire.Message. The
// inverse operation is ExtractIDRequest.
func (r IDRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// ExtractIDRequest extracts a IDRequest from the given
// mongowire.Message. The inverse operation is (IDRequest).Message.
func ExtractIDRequest(msg mongowire.Message) (IDRequest, error) {
	r := IDRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// IDResponse requests a response indicating the service manager's ID.
type IDResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	ID            string `bson:"id"`
}

func makeIDResponse(id string) IDResponse {
	return IDResponse{ID: id, ErrorResponse: makeSuccessResponse()}
}

// Message returns the IDResponse as an equivalent mongowire.Message. The
// inverse operation is ExtractIDResponse.
func (r IDResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// ExtractIDResponse extracts a IDResponse from the given
// mongowire.Message. The inverse operation is (IDResponse).Message.
func ExtractIDResponse(msg mongowire.Message) (IDResponse, error) {
	r := IDResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.Error != "" {
		return r, errors.New(r.Error)
	}
	if r.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// CreateProcessRequest represents a request to create a process with the given
// options.
type CreateProcessRequest struct {
	Options options.Create `bson:"create_process"`
}

func makeCreateProcessRequest(opts options.Create) CreateProcessRequest { //nolint
	return CreateProcessRequest{Options: opts}
}

// Message returns the CreateProcessRequest as an equivalent mongowire.Message.
// The inverse operation is ExtractCreateProcessRequest.
func (r CreateProcessRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// ExtractCreateProcessRequest extracts a CreateProcessRequest from the given
// mongowire.Message. The inverse operation is (CreateProcessRequest).Message.
func ExtractCreateProcessRequest(msg mongowire.Message) (CreateProcessRequest, error) {
	r := CreateProcessRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// ListRequest represents a request to get information regarding the processes
// matching the given filter.
type ListRequest struct {
	Filter options.Filter `bson:"list"`
}

func makeListRequest(filter options.Filter) ListRequest { //nolint
	return ListRequest{Filter: filter}
}

// Message returns the ListRequest as an equivalent mongowire.Message. The
// inverse operation is ExtractListRequest.
func (r ListRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// ExtractListRequest extracts a ListRequest from the given mongowire.Message.
// The inverse operation is (ListRequest).Message.
func ExtractListRequest(msg mongowire.Message) (ListRequest, error) {
	r := ListRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// GroupRequest represents a request to get information regarding the processes
// matching the given tag.
type GroupRequest struct {
	Tag string `bson:"group"`
}

func makeGroupRequest(tag string) GroupRequest { //nolint
	return GroupRequest{Tag: tag}
}

// Message returns the GroupRequest as an equivalent mongowire.Message. The
// inverse operation is ExtractGroupRequest.
func (r GroupRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// ExtractGroupRequest extracts a GroupRequest from the given mongowire.Message.
// The inverse operation is (GroupRequest).Message.
func ExtractGroupRequest(msg mongowire.Message) (GroupRequest, error) {
	r := GroupRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// GetRequest represents a request to get information regarding the process
// given by ID.
type GetRequest struct {
	ID string `bson:"get"`
}

func makeGetRequest(id string) GetRequest { //nolint
	return GetRequest{ID: id}
}

// Message returns the GetRequest as an equivalent mongowire.Message. The
// inverse operation is ExtractGetRequest.
func (r GetRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// ExtractGetRequest extracts a GetRequest from the given mongowire.Message.
// The inverse operation is (GetRequest).Message.
func ExtractGetRequest(msg mongowire.Message) (GetRequest, error) {
	r := GetRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// InfosResponse represents a response indicating the runtime information for
// multiple processes.
type InfosResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Infos         []jasper.ProcessInfo `bson:"infos"`
}

func makeInfosResponse(infos []jasper.ProcessInfo) InfosResponse {
	return InfosResponse{Infos: infos, ErrorResponse: makeSuccessResponse()}
}

// Message returns the InfosResponse as an equivalent mongowire.Message. The
// inverse operation is ExtractInfosResponse.
func (r InfosResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// ExtractInfosResponse extracts a InfosResponse from the given mongowire.Message.
// The inverse operation is (InfosResponse).Message.
func ExtractInfosResponse(msg mongowire.Message) (InfosResponse, error) {
	r := InfosResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.Error != "" {
		return r, errors.New(r.Error)
	}
	if r.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// ClearRequest represents a request to clear the current processes that have
// completed.
type ClearRequest struct {
	Clear int `bson:"clear"`
}

func makeClearRequest() ClearRequest { //nolint
	return ClearRequest{Clear: 1}
}

// Message returns the ClearRequest as an equivalent mongowire.Message. The
// inverse operation is ExtractClearRequest.
func (r ClearRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// ExtractClearRequest extracts a ClearRequest from the given mongowire.Message.
// The inverse operation is (ClearRequest).Message.
func ExtractClearRequest(msg mongowire.Message) (ClearRequest, error) {
	r := ClearRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// CloseRequest represents a request to terminate all processes.
type CloseRequest struct {
	Close int `bson:"close"`
}

func makeCloseRequest() CloseRequest { //nolint
	return CloseRequest{Close: 1}
}

// Message returns the CloseRequest as an equivalent mongowire.Message. The
// inverse operation is ExtractCloseRequest.
func (r CloseRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// ExtractCloseRequest extracts a CloseRequest from the given mongowire.Message.
// The inverse operation is (CloseRequest).Message.
func ExtractCloseRequest(msg mongowire.Message) (CloseRequest, error) {
	r := CloseRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// WhatsMyURIResponse represents a response indicating the service's URI.
type WhatsMyURIResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	You           string `bson:"you"`
}

func makeWhatsMyURIResponse(uri string) WhatsMyURIResponse {
	return WhatsMyURIResponse{You: uri, ErrorResponse: makeSuccessResponse()}
}

// Message returns the WhatsMyURIResponse as an equivalent mongowire.Message. The
// inverse operation is ExtractWhatsMyURIResponse.
func (r WhatsMyURIResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// ExtractWhatsMyURIResponse extracts a WhatsMyURIResponse from the given
// mongowire.Message. The inverse operation is (WhatsMyURIResponse).Message.
func ExtractWhatsMyURIResponse(msg mongowire.Message) (WhatsMyURIResponse, error) {
	r := WhatsMyURIResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.Error != "" {
		return r, errors.New(r.Error)
	}
	if r.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// BuildInfoResponse represents a response indicating the service's build
// information.
type BuildInfoResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Version       string `bson:"version"`
}

func makeBuildInfoResponse(version string) BuildInfoResponse {
	return BuildInfoResponse{Version: version, ErrorResponse: makeSuccessResponse()}
}

// Message returns the BuildInfoResponse as an equivalent mongowire.Message. The
// inverse operation is ExtractBuildInfoResponse.
func (r BuildInfoResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// ExtractBuildInfoResponse extracts a BuildInfoResponse from the given
// mongowire.Message. The inverse operation is (BuildInfoResponse).Message.
func ExtractBuildInfoResponse(msg mongowire.Message) (BuildInfoResponse, error) {
	r := BuildInfoResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.Error != "" {
		return r, errors.New(r.Error)
	}
	if r.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// GetLogResponse represents a response indicating the service's currently
// available logs.
type GetLogResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Log           []string `bson:"log"`
}

func makeGetLogResponse(log []string) GetLogResponse {
	return GetLogResponse{Log: log, ErrorResponse: makeSuccessResponse()}
}

// Message returns the GetLogResponse as an equivalent mongowire.Message. The
// inverse operation is ExtractGetLogResponse.
func (r GetLogResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// ExtractGetLogResponse extracts a GetLogResponse from the given
// mongowire.Message. The inverse operation is (GetLogResponse).Message.
func ExtractGetLogResponse(msg mongowire.Message) (GetLogResponse, error) {
	r := GetLogResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.Error != "" {
		return r, errors.New(r.Error)
	}
	if r.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}
