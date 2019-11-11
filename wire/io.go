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

// errorResponse represents a response indicating whether the operation was okay
// and errors, if any.
type errorResponse struct {
	OK    int    `bson:"ok"`
	Error string `bson:"errmsg,omitempty"`
}

// makeErrorResponse returns an errorResponse with the given ok status and error
// message, if any.
func makeErrorResponse(ok bool, err error) errorResponse {
	resp := errorResponse{OK: intOK(ok)}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp
}

// makeSuccessResponse returns an errorResponse that is ok and has no error.
func makeSuccessResponse() errorResponse {
	return errorResponse{OK: intOK(true)}
}

// message returns the errorResponse as an equivalent mongowire.Message. The
// inverse operation is extractErrorResponse.
func (r errorResponse) message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// extractErrorResponse extracts an errorResponse from the given
// mongowire.Message. The inverse operation is (errorResponse).message.
func extractErrorResponse(msg mongowire.Message) (errorResponse, error) {
	r := errorResponse{}
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

// infoRequest represents a request for runtime information regarding the
// process given by ID.
type infoRequest struct {
	ID string `bson:"info"`
}

func makeInfoRequest(id string) infoRequest { //nolint
	return infoRequest{ID: id}
}

// message returns the infoRequest as an equivalent mongowire.Message. The
// inverse operation is extractInfoRequest.
func (r infoRequest) message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// extractInfoRequest extracts an infoRequest from the given mongowire.Message.
// The inverse operation is (infoRequest).message.
func extractInfoRequest(msg mongowire.Message) (infoRequest, error) {
	r := infoRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// infoResponse represents a response indicating runtime information for a
// process.
type infoResponse struct {
	ErrorResponse errorResponse      `bson:"error_response,inline"`
	Info          jasper.ProcessInfo `bson:"info"`
}

func makeInfoResponse(info jasper.ProcessInfo) infoResponse {
	return infoResponse{Info: info, ErrorResponse: makeSuccessResponse()}
}

// message returns the infoResponse as an equivalent mongowire.Message. The
// inverse operation is extractInfoResponse.
func (r infoResponse) message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// extractInfoResponse extracts an infoResponse from the given
// mongowire.Message. The inverse operation is (infoResponse).message.
func extractInfoResponse(msg mongowire.Message) (infoResponse, error) {
	r := infoResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.ErrorResponse.Error != "" {
		return r, errors.New(r.ErrorResponse.Error)
	}
	if r.ErrorResponse.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// runningRequest represents a request for the running state of the process
// given by ID.
type runningRequest struct {
	ID string `bson:"running"`
}

func makeRunningRequest(id string) runningRequest { //nolint
	return runningRequest{ID: id}
}

// message returns the runningRequest as an equivalent mongowire.Message. The
// inverse operation is extractRunningRequest.
func (r runningRequest) message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// extractRunningRequest extracts a runningRequest from the given
// mongowire.Message. The inverse operation is (runningRequest).message.
func extractRunningRequest(msg mongowire.Message) (runningRequest, error) {
	r := runningRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// RunningResponse represents a response indicating the running state of a
// process.
type RunningResponse struct {
	ErrorResponse errorResponse `bson:"error_response,inline"`
	Running       bool          `bson:"running"`
}

func makeRunningResponse(running bool) RunningResponse {
	return RunningResponse{Running: running, ErrorResponse: makeSuccessResponse()}
}

// message returns the RunningResponse as an equivalent mongowire.Message. The
// inverse operation is extractRunningResponse.
func (r RunningResponse) message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// extractRunningResponse extracts a RunningResponse from the given
// mongowire.Message. The inverse operation is (RunningResponse).message.
func extractRunningResponse(msg mongowire.Message) (RunningResponse, error) {
	r := RunningResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.ErrorResponse.Error != "" {
		return r, errors.New(r.ErrorResponse.Error)
	}
	if r.ErrorResponse.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// completeRequest represents a request for the completion status of the process
// given by ID.
type completeRequest struct {
	ID string `bson:"complete"`
}

func makeCompleteRequest(id string) completeRequest { //nolint
	return completeRequest{ID: id}
}

// message returns the completeRequest as an equivalent mongowire.Message. The
// inverse operation is extractCompleteRequest.
func (r completeRequest) message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// extractCompleteRequest extracts a completeRequest from the given
// mongowire.Message. The inverse operation is (completeRequest).message.
func extractCompleteRequest(msg mongowire.Message) (completeRequest, error) {
	r := completeRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// completeResponse represents a response indicating the completion status of a
// process.
type completeResponse struct {
	ErrorResponse errorResponse `bson:"error_response,inline"`
	Complete      bool          `bson:"complete"`
}

func makeCompleteResponse(complete bool) completeResponse {
	return completeResponse{Complete: complete, ErrorResponse: makeSuccessResponse()}
}

// message returns the completeResponse as an equivalent mongowire.Message. The
// inverse operation is extractCompleteResponse.
func (r completeResponse) message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// extractCompleteResponse extracts a completeResponse from the given
// mongowire.Message. The inverse operation is (completeResponse).message.
func extractCompleteResponse(msg mongowire.Message) (completeResponse, error) {
	r := completeResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.ErrorResponse.Error != "" {
		return r, errors.New(r.ErrorResponse.Error)
	}
	if r.ErrorResponse.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// waitRequest represents a request for the wait status of the process given  by
// ID.
type waitRequest struct {
	ID string `bson:"wait"`
}

func makeWaitRequest(id string) waitRequest { //nolint
	return waitRequest{ID: id}
}

// message returns the waitRequest as an equivalent mongowire.Message. The
// inverse operation is extractWaitRequest.
func (r waitRequest) message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// extractWaitRequest extracts a waitRequest from the given mongowire.Message.
// The inverse operation is (waitRequest).message.
func extractWaitRequest(msg mongowire.Message) (waitRequest, error) {
	r := waitRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// waitResponse represents a response indicating the exit code and error of
// a waited process.
type waitResponse struct {
	ErrorResponse errorResponse `bson:"error_response,inline"`
	ExitCode      int           `bson:"exit_code"`
}

func makeWaitResponse(exitCode int, err error) waitResponse {
	return waitResponse{ExitCode: exitCode, ErrorResponse: makeErrorResponse(true, err)}
}

// message returns the waitResponse as an equivalent mongowire.Message. The
// inverse operation is extractWaitResponse.
func (r waitResponse) message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// extractWaitResponse extracts a waitResponse from the given
// mongowire.Message. The inverse operation is (waitResponse).message.
func extractWaitResponse(msg mongowire.Message) (waitResponse, error) {
	r := waitResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.ErrorResponse.Error != "" {
		return r, errors.New(r.ErrorResponse.Error)
	}
	if r.ErrorResponse.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// respawnRequest represents a request to respawn the process given by ID.
type respawnRequest struct {
	ID string `bson:"respawn"`
}

func makeRespawnRequest(id string) respawnRequest { //nolint
	return respawnRequest{ID: id}
}

// message returns the respawnRequest as an equivalent mongowire.Message. The
// inverse operation is extractRespawnRequest.
func (r respawnRequest) message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// extractRespawnRequest extracts a respawnRequest from the given
// mongowire.Message. The inverse operation is (respawnRequest).message.
func extractRespawnRequest(msg mongowire.Message) (respawnRequest, error) {
	r := respawnRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// signalRequest represents a request to send a signal to the process given by
// ID.
type signalRequest struct {
	Params struct {
		ID     string `bson:"id"`
		Signal int    `bson:"signal"`
	} `bson:"signal"`
}

func makeSignalRequest(id string, signal int) signalRequest { //nolint
	req := signalRequest{}
	req.Params.ID = id
	req.Params.Signal = signal
	return req
}

// message returns the signalRequest as an equivalent mongowire.Message. The
// inverse operation is extractSignalRequest.
func (r signalRequest) message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// extractSignalRequest extracts a signalRequest from the given
// mongowire.Message. The inverse operation is (signalRequest).message.
func extractSignalRequest(msg mongowire.Message) (signalRequest, error) {
	r := signalRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// registerSignalTriggerIDRequest represents a request to register the signal
// trigger ID on the process given by ID.
type registerSignalTriggerIDRequest struct {
	Params struct {
		ID              string                 `bson:"id"`
		SignalTriggerID jasper.SignalTriggerID `bson:"signal_trigger_id"`
	} `bson:"register_signal_trigger_id"`
}

func makeRegisterSignalTriggerIDRequest(id string, sigID jasper.SignalTriggerID) registerSignalTriggerIDRequest { //nolint
	req := registerSignalTriggerIDRequest{}
	req.Params.ID = id
	req.Params.SignalTriggerID = sigID
	return req
}

// message returns the registerSignalTriggerIDRequest as an equivalent
// mongowire.Message. The inverse operation is
// extractRegisterSignalTriggerIDRequest.
func (r registerSignalTriggerIDRequest) message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// extractRegisterSignalTriggerIDRequest extracts a
// registerSignalTriggerIDRequest from the given mongowire.Message. The inverse
// operation is (registerSignalTriggerIDRequest).message.
func extractRegisterSignalTriggerIDRequest(msg mongowire.Message) (registerSignalTriggerIDRequest, error) {
	r := registerSignalTriggerIDRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// tagRequest represents a request to associate the process given by ID with the
// tag.
type tagRequest struct {
	Params struct {
		ID  string `bson:"id"`
		Tag string `bson:"tag"`
	} `bson:"tag"`
}

func makeTagRequest(id, tag string) tagRequest { //nolint
	req := tagRequest{}
	req.Params.ID = id
	req.Params.Tag = tag
	return req
}

// message returns the tagRequest as an equivalent mongowire.Message. The
// inverse operation is extractTagRequest.
func (r tagRequest) message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// extractTagRequest extracts a tagRequest from the given mongowire.Message. The
// inverse operation is (tagRequest).message.
func extractTagRequest(msg mongowire.Message) (tagRequest, error) {
	r := tagRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// getTagsRequest represents a request to get all the tags for the process given
// by ID.
type getTagsRequest struct {
	ID string `bson:"get_tags"`
}

func makeGetTagsRequest(id string) getTagsRequest { //nolint
	return getTagsRequest{ID: id}
}

// message returns the getTagsRequest as an equivalent mongowire.Message. The
// inverse operation is extractGetTagsRequest.
func (r getTagsRequest) message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// extractGetTagsRequest extracts a getTagsRequest from the given
// mongowire.Message. The inverse operation is (getTagsRequest).message.
func extractGetTagsRequest(msg mongowire.Message) (getTagsRequest, error) {
	r := getTagsRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// getTagsResponse represents a response indicating the tags of a process.
type getTagsResponse struct {
	ErrorResponse errorResponse `bson:"error_response,inline"`
	Tags          []string      `bson:"tags"`
}

func makeGetTagsResponse(tags []string) getTagsResponse {
	return getTagsResponse{Tags: tags, ErrorResponse: makeSuccessResponse()}
}

// message returns the getTagsResponse as an equivalent mongowire.Message. The
// inverse operation is extractGetTagsResponse.
func (r getTagsResponse) message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// extractGetTagsResponse extracts a getTagsResponse from the given
// mongowire.Message. The inverse operation is (getTagsResponse).message.
func extractGetTagsResponse(msg mongowire.Message) (getTagsResponse, error) {
	r := getTagsResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.ErrorResponse.Error != "" {
		return r, errors.New(r.ErrorResponse.Error)
	}
	if r.ErrorResponse.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// resetTagsRequest represents a request to clear all the tags for the process
// given by ID.
type resetTagsRequest struct {
	ID string `bson:"reset_tags"`
}

func makeResetTagsRequest(id string) resetTagsRequest { //nolint
	return resetTagsRequest{ID: id}
}

// message returns the resetTagsRequest as an equivalent mongowire.Message. The
// inverse operation is extractResetTagsRequest.
func (r resetTagsRequest) message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// extractResetTagsRequest extracts a resetTagsRequest from the given
// mongowire.Message. The inverse operation is (resetTagsRequest).message.
func extractResetTagsRequest(msg mongowire.Message) (resetTagsRequest, error) {
	r := resetTagsRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// idRequest represents a request to get the ID associated with the service
// manager.
type idRequest struct {
	ID int `bson:"id"`
}

func makeIDRequest() idRequest { //nolint
	return idRequest{ID: 1}
}

// message returns the idRequest as an equivalent mongowire.Message. The
// inverse operation is extractIDRequest.
func (r idRequest) message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// extractIDRequest extracts a idRequest from the given
// mongowire.Message. The inverse operation is (idRequest).message.
func extractIDRequest(msg mongowire.Message) (idRequest, error) {
	r := idRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// idResponse requests a response indicating the service manager's ID.
type idResponse struct {
	ErrorResponse errorResponse `bson:"error_response,inline"`
	ID            string        `bson:"id"`
}

func makeIDResponse(id string) idResponse {
	return idResponse{ID: id, ErrorResponse: makeSuccessResponse()}
}

// message returns the idResponse as an equivalent mongowire.Message. The
// inverse operation is extractIDResponse.
func (r idResponse) message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// extractIDResponse extracts a idResponse from the given
// mongowire.Message. The inverse operation is (idResponse).message.
func extractIDResponse(msg mongowire.Message) (idResponse, error) {
	r := idResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.ErrorResponse.Error != "" {
		return r, errors.New(r.ErrorResponse.Error)
	}
	if r.ErrorResponse.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// createProcessRequest represents a request to create a process with the given
// options.
type createProcessRequest struct {
	Options options.Create `bson:"create_process"`
}

func makeCreateProcessRequest(opts options.Create) createProcessRequest { //nolint
	return createProcessRequest{Options: opts}
}

// message returns the createProcessRequest as an equivalent mongowire.Message.
// The inverse operation is extractCreateProcessRequest.
func (r createProcessRequest) message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// extractCreateProcessRequest extracts a createProcessRequest from the given
// mongowire.Message. The inverse operation is (createProcessRequest).message.
func extractCreateProcessRequest(msg mongowire.Message) (createProcessRequest, error) {
	r := createProcessRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// listRequest represents a request to get information regarding the processes
// matching the given filter.
type listRequest struct {
	Filter options.Filter `bson:"list"`
}

func makeListRequest(filter options.Filter) listRequest { //nolint
	return listRequest{Filter: filter}
}

// message returns the listRequest as an equivalent mongowire.Message. The
// inverse operation is extractListRequest.
func (r listRequest) message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// extractListRequest extracts a listRequest from the given mongowire.Message.
// The inverse operation is (listRequest).message.
func extractListRequest(msg mongowire.Message) (listRequest, error) {
	r := listRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// groupRequest represents a request to get information regarding the processes
// matching the given tag.
type groupRequest struct {
	Tag string `bson:"group"`
}

func makeGroupRequest(tag string) groupRequest { //nolint
	return groupRequest{Tag: tag}
}

// message returns the groupRequest as an equivalent mongowire.Message. The
// inverse operation is extractGroupRequest.
func (r groupRequest) message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// extractGroupRequest extracts a groupRequest from the given mongowire.Message.
// The inverse operation is (groupRequest).message.
func extractGroupRequest(msg mongowire.Message) (groupRequest, error) {
	r := groupRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// getRequest represents a request to get information regarding the process
// given by ID.
type getRequest struct {
	ID string `bson:"get"`
}

func makeGetRequest(id string) getRequest { //nolint
	return getRequest{ID: id}
}

// message returns the getRequest as an equivalent mongowire.Message. The
// inverse operation is extractGetRequest.
func (r getRequest) message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// extractGetRequest extracts a getRequest from the given mongowire.Message.
// The inverse operation is (getRequest).message.
func extractGetRequest(msg mongowire.Message) (getRequest, error) {
	r := getRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// infosResponse represents a response indicating the runtime information for
// multiple processes.
type infosResponse struct {
	ErrorResponse errorResponse        `bson:"error_response,inline"`
	Infos         []jasper.ProcessInfo `bson:"infos"`
}

func makeInfosResponse(infos []jasper.ProcessInfo) infosResponse {
	return infosResponse{Infos: infos, ErrorResponse: makeSuccessResponse()}
}

// message returns the infosResponse as an equivalent mongowire.Message. The
// inverse operation is extractInfosResponse.
func (r infosResponse) message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// extractInfosResponse extracts a infosResponse from the given mongowire.Message.
// The inverse operation is (infosResponse).message.
func extractInfosResponse(msg mongowire.Message) (infosResponse, error) {
	r := infosResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.ErrorResponse.Error != "" {
		return r, errors.New(r.ErrorResponse.Error)
	}
	if r.ErrorResponse.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// clearRequest represents a request to clear the current processes that have
// completed.
type clearRequest struct {
	Clear int `bson:"clear"`
}

func makeClearRequest() clearRequest { //nolint
	return clearRequest{Clear: 1}
}

// message returns the clearRequest as an equivalent mongowire.Message. The
// inverse operation is extractClearRequest.
func (r clearRequest) message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// extractClearRequest extracts a clearRequest from the given mongowire.Message.
// The inverse operation is (clearRequest).message.
func extractClearRequest(msg mongowire.Message) (clearRequest, error) {
	r := clearRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// closeRequest represents a request to terminate all processes.
type closeRequest struct {
	Close int `bson:"close"`
}

func makeCloseRequest() closeRequest { //nolint
	return closeRequest{Close: 1}
}

// message returns the closeRequest as an equivalent mongowire.Message. The
// inverse operation is extractCloseRequest.
func (r closeRequest) message() (mongowire.Message, error) {
	return requestToMessage(r)
}

// extractCloseRequest extracts a closeRequest from the given mongowire.Message.
// The inverse operation is (closeRequest).message.
func extractCloseRequest(msg mongowire.Message) (closeRequest, error) {
	r := closeRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

// whatsMyURIResponse represents a response indicating the service's URI.
type whatsMyURIResponse struct {
	ErrorResponse errorResponse `bson:"error_response,inline"`
	You           string        `bson:"you"`
}

func makeWhatsMyURIResponse(uri string) whatsMyURIResponse {
	return whatsMyURIResponse{You: uri, ErrorResponse: makeSuccessResponse()}
}

// message returns the whatsMyURIResponse as an equivalent mongowire.Message. The
// inverse operation is extractWhatsMyURIResponse.
func (r whatsMyURIResponse) message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// extractWhatsMyURIResponse extracts a whatsMyURIResponse from the given
// mongowire.Message. The inverse operation is (whatsMyURIResponse).message.
func extractWhatsMyURIResponse(msg mongowire.Message) (whatsMyURIResponse, error) {
	r := whatsMyURIResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.ErrorResponse.Error != "" {
		return r, errors.New(r.ErrorResponse.Error)
	}
	if r.ErrorResponse.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// buildInfoResponse represents a response indicating the service's build
// information.
type buildInfoResponse struct {
	ErrorResponse errorResponse `bson:"error_response,inline"`
	Version       string        `bson:"version"`
}

func makeBuildInfoResponse(version string) buildInfoResponse {
	return buildInfoResponse{Version: version, ErrorResponse: makeSuccessResponse()}
}

// message returns the buildInfoResponse as an equivalent mongowire.Message. The
// inverse operation is extractBuildInfoResponse.
func (r buildInfoResponse) message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// extractBuildInfoResponse extracts a buildInfoResponse from the given
// mongowire.Message. The inverse operation is (buildInfoResponse).message.
func extractBuildInfoResponse(msg mongowire.Message) (buildInfoResponse, error) {
	r := buildInfoResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.ErrorResponse.Error != "" {
		return r, errors.New(r.ErrorResponse.Error)
	}
	if r.ErrorResponse.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}

// getLogResponse represents a response indicating the service's currently
// available logs.
type getLogResponse struct {
	ErrorResponse errorResponse `bson:"error_response,inline"`
	Log           []string      `bson:"log"`
}

func makeGetLogResponse(log []string) getLogResponse {
	return getLogResponse{Log: log, ErrorResponse: makeSuccessResponse()}
}

// message returns the getLogResponse as an equivalent mongowire.Message. The
// inverse operation is extractGetLogResponse.
func (r getLogResponse) message() (mongowire.Message, error) {
	return responseToMessage(r)
}

// extractGetLogResponse extracts a getLogResponse from the given
// mongowire.Message. The inverse operation is (getLogResponse).message.
func extractGetLogResponse(msg mongowire.Message) (getLogResponse, error) {
	r := getLogResponse{}
	if err := messageToResponse(msg, &r); err != nil {
		return r, err
	}
	if r.ErrorResponse.Error != "" {
		return r, errors.New(r.ErrorResponse.Error)
	}
	if r.ErrorResponse.OK == 0 {
		return r, errors.New("response was not ok")
	}
	return r, nil
}
