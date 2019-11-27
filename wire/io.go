package wire

import (
	"github.com/evergreen-ci/mrpc/shell"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
)

// infoRequest represents a request for runtime information regarding the
// process given by ID.
type infoRequest struct {
	ID string `bson:"info"`
}

func makeInfoRequest(id string) infoRequest { //nolint
	return infoRequest{ID: id}
}

// infoResponse represents a response indicating runtime information for a
// process.
type infoResponse struct {
	ErrorResponse shell.ErrorResponse `bson:"error_response,inline"`
	Info          jasper.ProcessInfo  `bson:"info"`
}

func makeInfoResponse(info jasper.ProcessInfo) infoResponse {
	return infoResponse{Info: info, ErrorResponse: shell.MakeSuccessResponse()}
}

// runningRequest represents a request for the running state of the process
// given by ID.
type runningRequest struct {
	ID string `bson:"running"`
}

func makeRunningRequest(id string) runningRequest { //nolint
	return runningRequest{ID: id}
}

// RunningResponse represents a response indicating the running state of a
// process.
type RunningResponse struct {
	ErrorResponse shell.ErrorResponse `bson:"error_response,inline"`
	Running       bool                `bson:"running"`
}

func makeRunningResponse(running bool) RunningResponse {
	return RunningResponse{Running: running, ErrorResponse: shell.MakeSuccessResponse()}
}

// completeRequest represents a request for the completion status of the process
// given by ID.
type completeRequest struct {
	ID string `bson:"complete"`
}

func makeCompleteRequest(id string) completeRequest { //nolint
	return completeRequest{ID: id}
}

// completeResponse represents a response indicating the completion status of a
// process.
type completeResponse struct {
	ErrorResponse shell.ErrorResponse `bson:"error_response,inline"`
	Complete      bool                `bson:"complete"`
}

func makeCompleteResponse(complete bool) completeResponse {
	return completeResponse{Complete: complete, ErrorResponse: shell.MakeSuccessResponse()}
}

// waitRequest represents a request for the wait status of the process given  by
// ID.
type waitRequest struct {
	ID string `bson:"wait"`
}

func makeWaitRequest(id string) waitRequest { //nolint
	return waitRequest{ID: id}
}

// waitResponse represents a response indicating the exit code and error of
// a waited process.
type waitResponse struct {
	ErrorResponse shell.ErrorResponse `bson:"error_response,inline"`
	ExitCode      int                 `bson:"exit_code"`
}

func makeWaitResponse(exitCode int, err error) waitResponse {
	return waitResponse{ExitCode: exitCode, ErrorResponse: shell.MakeErrorResponse(true, err)}
}

// respawnRequest represents a request to respawn the process given by ID.
type respawnRequest struct {
	ID string `bson:"respawn"`
}

func makeRespawnRequest(id string) respawnRequest { //nolint
	return respawnRequest{ID: id}
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

// getTagsRequest represents a request to get all the tags for the process given
// by ID.
type getTagsRequest struct {
	ID string `bson:"get_tags"`
}

func makeGetTagsRequest(id string) getTagsRequest { //nolint
	return getTagsRequest{ID: id}
}

// getTagsResponse represents a response indicating the tags of a process.
type getTagsResponse struct {
	ErrorResponse shell.ErrorResponse `bson:"error_response,inline"`
	Tags          []string            `bson:"tags"`
}

func makeGetTagsResponse(tags []string) getTagsResponse {
	return getTagsResponse{Tags: tags, ErrorResponse: shell.MakeSuccessResponse()}
}

// resetTagsRequest represents a request to clear all the tags for the process
// given by ID.
type resetTagsRequest struct {
	ID string `bson:"reset_tags"`
}

func makeResetTagsRequest(id string) resetTagsRequest { //nolint
	return resetTagsRequest{ID: id}
}

// idRequest represents a request to get the ID associated with the service
// manager.
type idRequest struct { //nolint
	ID int `bson:"id"`
}

func makeIDRequest() idRequest { //nolint
	return idRequest{ID: 1}
}

// idResponse requests a response indicating the service manager's ID.
type idResponse struct {
	ErrorResponse shell.ErrorResponse `bson:"error_response,inline"`
	ID            string              `bson:"id"`
}

func makeIDResponse(id string) idResponse {
	return idResponse{ID: id, ErrorResponse: shell.MakeSuccessResponse()}
}

// createProcessRequest represents a request to create a process with the given
// options.
type createProcessRequest struct {
	Options options.Create `bson:"create_process"`
}

func makeCreateProcessRequest(opts options.Create) createProcessRequest { //nolint
	return createProcessRequest{Options: opts}
}

// listRequest represents a request to get information regarding the processes
// matching the given filter.
type listRequest struct {
	Filter options.Filter `bson:"list"`
}

func makeListRequest(filter options.Filter) listRequest { //nolint
	return listRequest{Filter: filter}
}

// groupRequest represents a request to get information regarding the processes
// matching the given tag.
type groupRequest struct {
	Tag string `bson:"group"`
}

func makeGroupRequest(tag string) groupRequest { //nolint
	return groupRequest{Tag: tag}
}

// getRequest represents a request to get information regarding the process
// given by ID.
type getRequest struct {
	ID string `bson:"get"`
}

func makeGetRequest(id string) getRequest { //nolint
	return getRequest{ID: id}
}

// infosResponse represents a response indicating the runtime information for
// multiple processes.
type infosResponse struct {
	ErrorResponse shell.ErrorResponse  `bson:"error_response,inline"`
	Infos         []jasper.ProcessInfo `bson:"infos"`
}

func makeInfosResponse(infos []jasper.ProcessInfo) infosResponse {
	return infosResponse{Infos: infos, ErrorResponse: shell.MakeSuccessResponse()}
}

// clearRequest represents a request to clear the current processes that have
// completed.
type clearRequest struct { //nolint
	Clear int `bson:"clear"`
}

func makeClearRequest() clearRequest { //nolint
	return clearRequest{Clear: 1}
}

// closeRequest represents a request to terminate all processes.
type closeRequest struct { //nolint
	Close int `bson:"close"`
}

func makeCloseRequest() closeRequest { //nolint
	return closeRequest{Close: 1}
}
