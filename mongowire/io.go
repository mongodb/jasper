package mongowire

import (
	"github.com/evergreen-ci/birch"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/options"
	"github.com/pkg/errors"
	"github.com/tychoish/mongorpc/mongowire"
	"gopkg.in/mgo.v2/bson"
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
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractErrorResponse(msg mongowire.Message) (ErrorResponse, error) {
	resp := ErrorResponse{}
	doc, err := responseToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type ProcessInfoRequest struct {
	ID string `bson:"info"`
}

func makeProcessInfoRequest(id string) ProcessInfoRequest {
	return ProcessInfoRequest{ID: id}
}

func (r ProcessInfoRequest) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractProcessInfoRequest(msg mongowire.Message) (ProcessInfoRequest, error) {
	resp := ProcessInfoRequest{}
	doc, err := requestToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type RunningRequest struct {
	ID string `bson:"running"`
}

func makeRunningRequest(id string) RunningRequest {
	return RunningRequest{ID: id}
}

func (r RunningRequest) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractRunningRequest(msg mongowire.Message) (RunningRequest, error) {
	resp := RunningRequest{}
	doc, err := requestToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type RunningResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Running       bool `bson:"running"`
}

func makeRunningResponse(running bool) RunningResponse {
	return RunningResponse{Running: running, ErrorResponse: makeSuccessResponse()}
}

func (r RunningResponse) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractRunningResponse(msg mongowire.Message) (RunningResponse, error) {
	resp := RunningResponse{}
	doc, err := responseToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type CompleteRequest struct {
	ID string `bson:"complete"`
}

func makeCompleteRequest(id string) CompleteRequest {
	return CompleteRequest{ID: id}
}

func (r CompleteRequest) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractCompleteRequest(msg mongowire.Message) (CompleteRequest, error) {
	resp := CompleteRequest{}
	doc, err := requestToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type CompleteResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Complete      bool `bson:"complete"`
}

func makeCompleteResponse(complete bool) CompleteResponse {
	return CompleteResponse{Complete: complete, ErrorResponse: makeSuccessResponse()}
}

func (r CompleteResponse) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractCompleteResponse(msg mongowire.Message) (CompleteResponse, error) {
	resp := CompleteResponse{}
	doc, err := responseToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type WaitRequest struct {
	ID string `bson:"wait"`
}

func makeWaitRequest(id string) WaitRequest {
	return WaitRequest{ID: id}
}

func (r WaitRequest) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractWaitRequest(msg mongowire.Message) (WaitRequest, error) {
	resp := WaitRequest{}
	doc, err := requestToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type WaitResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	ExitCode      int `bson:"exit_code"`
}

func makeWaitResponse(exitCode int, err error) WaitResponse {
	return WaitResponse{ExitCode: exitCode, ErrorResponse: makeErrorResponse(true, err)}
}

func (r WaitResponse) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractWaitResponse(msg mongowire.Message) (WaitResponse, error) {
	resp := WaitResponse{}
	doc, err := responseToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type RespawnRequest struct {
	ID string `bson:"respawn"`
}

func makeRespawnRequest(id string) RespawnRequest {
	return RespawnRequest{ID: id}
}

func (r RespawnRequest) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractRespawnRequest(msg mongowire.Message) (RespawnRequest, error) {
	resp := RespawnRequest{}
	doc, err := requestToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
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
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractSignalRequest(msg mongowire.Message) (SignalRequest, error) {
	resp := SignalRequest{}
	doc, err := requestToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type RegisterSignalTriggerIDRequest struct {
	Params struct {
		ID              string  `bson:"id"`
		SignalTriggerID float64 `bson:"signal_trigger_id"` // The mongo shell sends integers as doubles by default
	} `bson:"register_signal_trigger_id"`
}

func makeRegisterSignalTriggerIDRequest(id string, sigID int) RegisterSignalTriggerIDRequest {
	req := RegisterSignalTriggerIDRequest{}
	req.Params.ID = id
	req.Params.SignalTriggerID = float64(sigID)
	return req
}

func (r RegisterSignalTriggerIDRequest) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractRegisterSignalTriggerIDRequest(msg mongowire.Message) (RegisterSignalTriggerIDRequest, error) {
	resp := RegisterSignalTriggerIDRequest{}
	doc, err := requestToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
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
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractTagRequest(msg mongowire.Message) (TagRequest, error) {
	resp := TagRequest{}
	doc, err := requestToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type GetTagsRequest struct {
	ID string `bson:"get_tags"`
}

func makeGetTagsRequest(id string) GetTagsRequest {
	return GetTagsRequest{ID: id}
}

func (r GetTagsRequest) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractGetTagsRequest(msg mongowire.Message) (GetTagsRequest, error) {
	resp := GetTagsRequest{}
	doc, err := requestToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type GetTagsResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Tags          []string `bson:"tags"`
}

func makeGetTagsResponse(tags []string) GetTagsResponse {
	return GetTagsResponse{Tags: tags, ErrorResponse: makeSuccessResponse()}
}

func (r GetTagsResponse) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractGetTagsResponse(msg mongowire.Message) (GetTagsResponse, error) {
	resp := GetTagsResponse{}
	doc, err := responseToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type ResetTagsRequest struct {
	ID string `bson:"reset_tags"`
}

func makeResetTagsRequest(id string) ResetTagsRequest {
	return ResetTagsRequest{ID: id}
}

func (r ResetTagsRequest) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractResetTagsRequest(msg mongowire.Message) (ResetTagsRequest, error) {
	resp := ResetTagsRequest{}
	doc, err := requestToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type IDResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	ID            string `bson:"id"`
}

func makeIDResponse(id string) IDResponse {
	return IDResponse{ID: id, ErrorResponse: makeSuccessResponse()}
}

func (r IDResponse) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractIDResponse(msg mongowire.Message) (IDResponse, error) {
	resp := IDResponse{}
	doc, err := responseToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type WhatsMyURIResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	You           string `bson:"you"`
}

func makeWhatsMyURIResponse(uri string) WhatsMyURIResponse {
	return WhatsMyURIResponse{You: uri, ErrorResponse: makeSuccessResponse()}
}

func (r WhatsMyURIResponse) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractWhatsMyURIResponse(msg mongowire.Message) (WhatsMyURIResponse, error) {
	resp := WhatsMyURIResponse{}
	doc, err := responseToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type BuildInfoResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Version       string `bson:"version"`
}

func makeBuildInfoResponse(version string) BuildInfoResponse {
	return BuildInfoResponse{Version: version, ErrorResponse: makeSuccessResponse()}
}

func (r BuildInfoResponse) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractBuildInfoResponse(msg mongowire.Message) (BuildInfoResponse, error) {
	resp := BuildInfoResponse{}
	doc, err := responseToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type GetLogResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Log           []string `bson:"log"`
}

func makeGetLogResponse(log []string) GetLogResponse {
	return GetLogResponse{Log: log, ErrorResponse: makeSuccessResponse()}
}

func (r GetLogResponse) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractGetLogResponse(msg mongowire.Message) (GetLogResponse, error) {
	resp := GetLogResponse{}
	doc, err := responseToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type CreateProcessRequest struct {
	Options options.Create `bson:"create_process"`
}

func makeCreateProcessRequest(opts options.Create) CreateProcessRequest {
	return CreateProcessRequest{Options: opts}
}

func (r CreateProcessRequest) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	// return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractCreateProcessRequest(msg mongowire.Message) (CreateProcessRequest, error) {
	resp := CreateProcessRequest{}
	doc, err := requestToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not vonert")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type ListRequest struct {
	Filter options.Filter `bson:"list"`
}

func makeListRequest(filter options.Filter) ListRequest {
	return ListRequest{Filter: filter}
}

func (r ListRequest) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	// return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractListRequest(msg mongowire.Message) (ListRequest, error) {
	resp := ListRequest{}
	doc, err := requestToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type GroupRequest struct {
	Tag string `bson:"group"`
}

func makeGroupRequest(tag string) GroupRequest {
	return GroupRequest{Tag: tag}
}

func (r GroupRequest) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
}

func ExtractGroupRequest(msg mongowire.Message) (GroupRequest, error) {
	resp := GroupRequest{}
	doc, err := requestToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type GetRequest struct {
	ID string `bson:"get"`
}

func makeGetRequest(id string) GetRequest {
	return GetRequest{ID: id}
}

func (r GetRequest) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	// return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractGetRequest(msg mongowire.Message) (GetRequest, error) {
	resp := GetRequest{}
	doc, err := requestToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type InfoResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Info          jasper.ProcessInfo `bson:"info"`
}

func makeInfoResponse(info jasper.ProcessInfo) InfoResponse {
	return InfoResponse{Info: info, ErrorResponse: makeSuccessResponse()}
}

func (r InfoResponse) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractInfoResponse(msg mongowire.Message) (InfoResponse, error) {
	resp := InfoResponse{}
	doc, err := responseToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}

type InfosResponse struct {
	ErrorResponse `bson:"error_response,inline"`
	Infos         []jasper.ProcessInfo `bson:"infos"`
}

func makeInfosResponse(infos []jasper.ProcessInfo) InfosResponse {
	return InfosResponse{Infos: infos, ErrorResponse: makeSuccessResponse()}
}

func (r InfosResponse) Message() (mongowire.Message, error) {
	b, err := bson.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert response to BSON")
	}
	doc, err := birch.ReadDocument(b)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert BSON response to document")
	}
	// return mongowire.NewCommandReply(doc, birch.NewDocument(), []birch.Document{}), nil
	return mongowire.NewReply(0, 0, 0, 1, []*birch.Document{doc}), nil
}

func ExtractInfosResponse(msg mongowire.Message) (InfosResponse, error) {
	resp := InfosResponse{}
	doc, err := responseToDocument(msg)
	if err != nil {
		return resp, errors.Wrap(err, "could not read response")
	}
	b, err := doc.MarshalBSON()
	if err != nil {
		return resp, errors.Wrap(err, "could not convert document to BSON")
	}
	if err := bson.Unmarshal(b, &resp); err != nil {
		return resp, errors.Wrap(err, "could not convert BSON to response")
	}
	return resp, nil
}
