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

func makeErrorResponse(ok bool, err error) ErrorResponse {
	resp := ErrorResponse{OK: intOK(ok)}
	if err != nil {
		resp.Error = err.Error()
	}
	return resp
}

func makeSuccessResponse() ErrorResponse {
	return ErrorResponse{OK: intOK(true)}
}

func (r ErrorResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

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

type InfoRequest struct {
	ID string `bson:"info"`
}

func makeInfoRequest(id string) InfoRequest {
	return InfoRequest{ID: id}
}

func (r InfoRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

func ExtractInfoRequest(msg mongowire.Message) (InfoRequest, error) {
	r := InfoRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

type RunningRequest struct {
	ID string `bson:"running"`
}

func makeRunningRequest(id string) RunningRequest {
	return RunningRequest{ID: id}
}

func (r RunningRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

func ExtractRunningRequest(msg mongowire.Message) (RunningRequest, error) {
	r := RunningRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

type RunningResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Running       bool `bson:"running"`
}

func makeRunningResponse(running bool) RunningResponse {
	return RunningResponse{Running: running, ErrorResponse: makeSuccessResponse()}
}

func (r RunningResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

func ExtractRunningResponse(msg mongowire.Message) (RunningResponse, error) {
	r := RunningResponse{}
	if err := messageToRequest(msg, &r); err != nil {
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

type CompleteRequest struct {
	ID string `bson:"complete"`
}

func makeCompleteRequest(id string) CompleteRequest {
	return CompleteRequest{ID: id}
}

func (r CompleteRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

func ExtractCompleteRequest(msg mongowire.Message) (CompleteRequest, error) {
	r := CompleteRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

type CompleteResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Complete      bool `bson:"complete"`
}

func makeCompleteResponse(complete bool) CompleteResponse {
	return CompleteResponse{Complete: complete, ErrorResponse: makeSuccessResponse()}
}

func (r CompleteResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

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

type WaitRequest struct {
	ID string `bson:"wait"`
}

func makeWaitRequest(id string) WaitRequest {
	return WaitRequest{ID: id}
}

func (r WaitRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

func ExtractWaitRequest(msg mongowire.Message) (WaitRequest, error) {
	r := WaitRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

type WaitResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	ExitCode      int `bson:"exit_code"`
}

func makeWaitResponse(exitCode int, err error) WaitResponse {
	return WaitResponse{ExitCode: exitCode, ErrorResponse: makeErrorResponse(true, err)}
}

func (r WaitResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

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

type RespawnRequest struct {
	ID string `bson:"respawn"`
}

func makeRespawnRequest(id string) RespawnRequest {
	return RespawnRequest{ID: id}
}

func (r RespawnRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

func ExtractRespawnRequest(msg mongowire.Message) (RespawnRequest, error) {
	r := RespawnRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

type SignalRequest struct {
	Params struct {
		ID     string  `bson:"id"`
		Signal float64 `bson:"signal"` // The mongo shell sends integers as doubles by default
	} `bson:"signal"`
}

func makeSignalRequest(id string, signal int) SignalRequest {
	req := SignalRequest{}
	req.Params.ID = id
	req.Params.Signal = float64(signal)
	return req
}

func (r SignalRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

func ExtractSignalRequest(msg mongowire.Message) (SignalRequest, error) {
	r := SignalRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

type RegisterSignalTriggerIDRequest struct {
	Params struct {
		ID              string                 `bson:"id"`
		SignalTriggerID jasper.SignalTriggerID `bson:"signal_trigger_id"`
	} `bson:"register_signal_trigger_id"`
}

func makeRegisterSignalTriggerIDRequest(id string, sigID jasper.SignalTriggerID) RegisterSignalTriggerIDRequest {
	req := RegisterSignalTriggerIDRequest{}
	req.Params.ID = id
	req.Params.SignalTriggerID = sigID
	return req
}

func (r RegisterSignalTriggerIDRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

func ExtractRegisterSignalTriggerIDRequest(msg mongowire.Message) (RegisterSignalTriggerIDRequest, error) {
	r := RegisterSignalTriggerIDRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

type TagRequest struct {
	Params struct {
		ID  string `bson:"id"`
		Tag string `bson:"tag"`
	} `bson:"tag"`
}

func makeTagRequest(id, tag string) TagRequest {
	req := TagRequest{}
	req.Params.ID = id
	req.Params.Tag = tag
	return req
}

func (r TagRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

func ExtractTagRequest(msg mongowire.Message) (TagRequest, error) {
	r := TagRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

type GetTagsRequest struct {
	ID string `bson:"get_tags"`
}

func makeGetTagsRequest(id string) GetTagsRequest {
	return GetTagsRequest{ID: id}
}

func (r GetTagsRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

func ExtractGetTagsRequest(msg mongowire.Message) (GetTagsRequest, error) {
	r := GetTagsRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

type GetTagsResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Tags          []string `bson:"tags"`
}

func makeGetTagsResponse(tags []string) GetTagsResponse {
	return GetTagsResponse{Tags: tags, ErrorResponse: makeSuccessResponse()}
}

func (r GetTagsResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

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

type ResetTagsRequest struct {
	ID string `bson:"reset_tags"`
}

func makeResetTagsRequest(id string) ResetTagsRequest {
	return ResetTagsRequest{ID: id}
}

func (r ResetTagsRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

func ExtractResetTagsRequest(msg mongowire.Message) (ResetTagsRequest, error) {
	r := ResetTagsRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

type IDRequest struct {
	ID int `bson:"id"`
}

func makeIDRequest() IDRequest {
	return IDRequest{ID: 1}
}

func (r IDRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

func ExtractIDRequest(msg mongowire.Message) (IDRequest, error) {
	r := IDRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

type IDResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	ID            string `bson:"id"`
}

func makeIDResponse(id string) IDResponse {
	return IDResponse{ID: id, ErrorResponse: makeSuccessResponse()}
}

func (r IDResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

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

type WhatsMyURIResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	You           string `bson:"you"`
}

func makeWhatsMyURIResponse(uri string) WhatsMyURIResponse {
	return WhatsMyURIResponse{You: uri, ErrorResponse: makeSuccessResponse()}
}

func (r WhatsMyURIResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

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

type BuildInfoResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Version       string `bson:"version"`
}

func makeBuildInfoResponse(version string) BuildInfoResponse {
	return BuildInfoResponse{Version: version, ErrorResponse: makeSuccessResponse()}
}

func (r BuildInfoResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

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

type GetLogResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Log           []string `bson:"log"`
}

func makeGetLogResponse(log []string) GetLogResponse {
	return GetLogResponse{Log: log, ErrorResponse: makeSuccessResponse()}
}

func (r GetLogResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

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

type CreateProcessRequest struct {
	Options options.Create `bson:"create_process"`
}

func makeCreateProcessRequest(opts options.Create) CreateProcessRequest {
	return CreateProcessRequest{Options: opts}
}

func (r CreateProcessRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

func ExtractCreateProcessRequest(msg mongowire.Message) (CreateProcessRequest, error) {
	r := CreateProcessRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

type ListRequest struct {
	Filter options.Filter `bson:"list"`
}

func makeListRequest(filter options.Filter) ListRequest {
	return ListRequest{Filter: filter}
}

func (r ListRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

func ExtractListRequest(msg mongowire.Message) (ListRequest, error) {
	r := ListRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

type GroupRequest struct {
	Tag string `bson:"group"`
}

func makeGroupRequest(tag string) GroupRequest {
	return GroupRequest{Tag: tag}
}

func (r GroupRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

func ExtractGroupRequest(msg mongowire.Message) (GroupRequest, error) {
	r := GroupRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

type GetRequest struct {
	ID string `bson:"get"`
}

func makeGetRequest(id string) GetRequest {
	return GetRequest{ID: id}
}

func (r GetRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

func ExtractGetRequest(msg mongowire.Message) (GetRequest, error) {
	r := GetRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

type InfoResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Info          jasper.ProcessInfo `bson:"info"`
}

func makeInfoResponse(info jasper.ProcessInfo) InfoResponse {
	return InfoResponse{Info: info, ErrorResponse: makeSuccessResponse()}
}

func (r InfoResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

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

type InfosResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Infos         []jasper.ProcessInfo `bson:"infos"`
}

func makeInfosResponse(infos []jasper.ProcessInfo) InfosResponse {
	return InfosResponse{Infos: infos, ErrorResponse: makeSuccessResponse()}
}

func (r InfosResponse) Message() (mongowire.Message, error) {
	return responseToMessage(r)
}

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

type ClearRequest struct {
	Clear int `bson:"clear"`
}

func makeClearRequest() ClearRequest {
	return ClearRequest{Clear: 1}
}

func (r ClearRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

func ExtractClearRequest(msg mongowire.Message) (ClearRequest, error) {
	r := ClearRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}

type CloseRequest struct {
	Close int `bson:"close"`
}

func makeCloseRequest() CloseRequest {
	return CloseRequest{Close: 1}
}

func (r CloseRequest) Message() (mongowire.Message, error) {
	return requestToMessage(r)
}

func ExtractCloseRequest(msg mongowire.Message) (CloseRequest, error) {
	r := CloseRequest{}
	if err := messageToRequest(msg, &r); err != nil {
		return r, err
	}
	return r, nil
}
