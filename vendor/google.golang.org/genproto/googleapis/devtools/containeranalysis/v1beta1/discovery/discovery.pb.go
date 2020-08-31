// Copyright 2018 The Grafeas Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.12.3
// source: google/devtools/containeranalysis/v1beta1/discovery/discovery.proto

package discovery

import (
	reflect "reflect"
	sync "sync"

	proto "github.com/golang/protobuf/proto"
	common "google.golang.org/genproto/googleapis/devtools/containeranalysis/v1beta1/common"
	status "google.golang.org/genproto/googleapis/rpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// Whether the resource is continuously analyzed.
type Discovered_ContinuousAnalysis int32

const (
	// Unknown.
	Discovered_CONTINUOUS_ANALYSIS_UNSPECIFIED Discovered_ContinuousAnalysis = 0
	// The resource is continuously analyzed.
	Discovered_ACTIVE Discovered_ContinuousAnalysis = 1
	// The resource is ignored for continuous analysis.
	Discovered_INACTIVE Discovered_ContinuousAnalysis = 2
)

// Enum value maps for Discovered_ContinuousAnalysis.
var (
	Discovered_ContinuousAnalysis_name = map[int32]string{
		0: "CONTINUOUS_ANALYSIS_UNSPECIFIED",
		1: "ACTIVE",
		2: "INACTIVE",
	}
	Discovered_ContinuousAnalysis_value = map[string]int32{
		"CONTINUOUS_ANALYSIS_UNSPECIFIED": 0,
		"ACTIVE":                          1,
		"INACTIVE":                        2,
	}
)

func (x Discovered_ContinuousAnalysis) Enum() *Discovered_ContinuousAnalysis {
	p := new(Discovered_ContinuousAnalysis)
	*p = x
	return p
}

func (x Discovered_ContinuousAnalysis) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Discovered_ContinuousAnalysis) Descriptor() protoreflect.EnumDescriptor {
	return file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_enumTypes[0].Descriptor()
}

func (Discovered_ContinuousAnalysis) Type() protoreflect.EnumType {
	return &file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_enumTypes[0]
}

func (x Discovered_ContinuousAnalysis) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Discovered_ContinuousAnalysis.Descriptor instead.
func (Discovered_ContinuousAnalysis) EnumDescriptor() ([]byte, []int) {
	return file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_rawDescGZIP(), []int{2, 0}
}

// Analysis status for a resource. Currently for initial analysis only (not
// updated in continuous analysis).
type Discovered_AnalysisStatus int32

const (
	// Unknown.
	Discovered_ANALYSIS_STATUS_UNSPECIFIED Discovered_AnalysisStatus = 0
	// Resource is known but no action has been taken yet.
	Discovered_PENDING Discovered_AnalysisStatus = 1
	// Resource is being analyzed.
	Discovered_SCANNING Discovered_AnalysisStatus = 2
	// Analysis has finished successfully.
	Discovered_FINISHED_SUCCESS Discovered_AnalysisStatus = 3
	// Analysis has finished unsuccessfully, the analysis itself is in a bad
	// state.
	Discovered_FINISHED_FAILED Discovered_AnalysisStatus = 4
	// The resource is known not to be supported
	Discovered_FINISHED_UNSUPPORTED Discovered_AnalysisStatus = 5
)

// Enum value maps for Discovered_AnalysisStatus.
var (
	Discovered_AnalysisStatus_name = map[int32]string{
		0: "ANALYSIS_STATUS_UNSPECIFIED",
		1: "PENDING",
		2: "SCANNING",
		3: "FINISHED_SUCCESS",
		4: "FINISHED_FAILED",
		5: "FINISHED_UNSUPPORTED",
	}
	Discovered_AnalysisStatus_value = map[string]int32{
		"ANALYSIS_STATUS_UNSPECIFIED": 0,
		"PENDING":                     1,
		"SCANNING":                    2,
		"FINISHED_SUCCESS":            3,
		"FINISHED_FAILED":             4,
		"FINISHED_UNSUPPORTED":        5,
	}
)

func (x Discovered_AnalysisStatus) Enum() *Discovered_AnalysisStatus {
	p := new(Discovered_AnalysisStatus)
	*p = x
	return p
}

func (x Discovered_AnalysisStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Discovered_AnalysisStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_enumTypes[1].Descriptor()
}

func (Discovered_AnalysisStatus) Type() protoreflect.EnumType {
	return &file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_enumTypes[1]
}

func (x Discovered_AnalysisStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Discovered_AnalysisStatus.Descriptor instead.
func (Discovered_AnalysisStatus) EnumDescriptor() ([]byte, []int) {
	return file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_rawDescGZIP(), []int{2, 1}
}

// A note that indicates a type of analysis a provider would perform. This note
// exists in a provider's project. A `Discovery` occurrence is created in a
// consumer's project at the start of analysis.
type Discovery struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Required. Immutable. The kind of analysis that is handled by this
	// discovery.
	AnalysisKind common.NoteKind `protobuf:"varint,1,opt,name=analysis_kind,json=analysisKind,proto3,enum=grafeas.v1beta1.NoteKind" json:"analysis_kind,omitempty"`
}

func (x *Discovery) Reset() {
	*x = Discovery{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Discovery) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Discovery) ProtoMessage() {}

func (x *Discovery) ProtoReflect() protoreflect.Message {
	mi := &file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Discovery.ProtoReflect.Descriptor instead.
func (*Discovery) Descriptor() ([]byte, []int) {
	return file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_rawDescGZIP(), []int{0}
}

func (x *Discovery) GetAnalysisKind() common.NoteKind {
	if x != nil {
		return x.AnalysisKind
	}
	return common.NoteKind_NOTE_KIND_UNSPECIFIED
}

// Details of a discovery occurrence.
type Details struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Required. Analysis status for the discovered resource.
	Discovered *Discovered `protobuf:"bytes,1,opt,name=discovered,proto3" json:"discovered,omitempty"`
}

func (x *Details) Reset() {
	*x = Details{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Details) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Details) ProtoMessage() {}

func (x *Details) ProtoReflect() protoreflect.Message {
	mi := &file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Details.ProtoReflect.Descriptor instead.
func (*Details) Descriptor() ([]byte, []int) {
	return file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_rawDescGZIP(), []int{1}
}

func (x *Details) GetDiscovered() *Discovered {
	if x != nil {
		return x.Discovered
	}
	return nil
}

// Provides information about the analysis status of a discovered resource.
type Discovered struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Whether the resource is continuously analyzed.
	ContinuousAnalysis Discovered_ContinuousAnalysis `protobuf:"varint,1,opt,name=continuous_analysis,json=continuousAnalysis,proto3,enum=grafeas.v1beta1.discovery.Discovered_ContinuousAnalysis" json:"continuous_analysis,omitempty"`
	// The last time continuous analysis was done for this resource.
	LastAnalysisTime *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=last_analysis_time,json=lastAnalysisTime,proto3" json:"last_analysis_time,omitempty"`
	// The status of discovery for the resource.
	AnalysisStatus Discovered_AnalysisStatus `protobuf:"varint,3,opt,name=analysis_status,json=analysisStatus,proto3,enum=grafeas.v1beta1.discovery.Discovered_AnalysisStatus" json:"analysis_status,omitempty"`
	// When an error is encountered this will contain a LocalizedMessage under
	// details to show to the user. The LocalizedMessage is output only and
	// populated by the API.
	AnalysisStatusError *status.Status `protobuf:"bytes,4,opt,name=analysis_status_error,json=analysisStatusError,proto3" json:"analysis_status_error,omitempty"`
}

func (x *Discovered) Reset() {
	*x = Discovered{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Discovered) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Discovered) ProtoMessage() {}

func (x *Discovered) ProtoReflect() protoreflect.Message {
	mi := &file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Discovered.ProtoReflect.Descriptor instead.
func (*Discovered) Descriptor() ([]byte, []int) {
	return file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_rawDescGZIP(), []int{2}
}

func (x *Discovered) GetContinuousAnalysis() Discovered_ContinuousAnalysis {
	if x != nil {
		return x.ContinuousAnalysis
	}
	return Discovered_CONTINUOUS_ANALYSIS_UNSPECIFIED
}

func (x *Discovered) GetLastAnalysisTime() *timestamppb.Timestamp {
	if x != nil {
		return x.LastAnalysisTime
	}
	return nil
}

func (x *Discovered) GetAnalysisStatus() Discovered_AnalysisStatus {
	if x != nil {
		return x.AnalysisStatus
	}
	return Discovered_ANALYSIS_STATUS_UNSPECIFIED
}

func (x *Discovered) GetAnalysisStatusError() *status.Status {
	if x != nil {
		return x.AnalysisStatusError
	}
	return nil
}

var File_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto protoreflect.FileDescriptor

var file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_rawDesc = []byte{
	0x0a, 0x43, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x64, 0x65, 0x76, 0x74, 0x6f, 0x6f, 0x6c,
	0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x61, 0x6e, 0x61, 0x6c, 0x79,
	0x73, 0x69, 0x73, 0x2f, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2f, 0x64, 0x69, 0x73, 0x63,
	0x6f, 0x76, 0x65, 0x72, 0x79, 0x2f, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x19, 0x67, 0x72, 0x61, 0x66, 0x65, 0x61, 0x73, 0x2e, 0x76,
	0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79,
	0x1a, 0x3d, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x64, 0x65, 0x76, 0x74, 0x6f, 0x6f, 0x6c,
	0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x61, 0x6e, 0x61, 0x6c, 0x79,
	0x73, 0x69, 0x73, 0x2f, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2f, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x17, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x72, 0x70, 0x63, 0x2f, 0x73, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x4b, 0x0a, 0x09, 0x44, 0x69, 0x73,
	0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x12, 0x3e, 0x0a, 0x0d, 0x61, 0x6e, 0x61, 0x6c, 0x79, 0x73,
	0x69, 0x73, 0x5f, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x19, 0x2e,
	0x67, 0x72, 0x61, 0x66, 0x65, 0x61, 0x73, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e,
	0x4e, 0x6f, 0x74, 0x65, 0x4b, 0x69, 0x6e, 0x64, 0x52, 0x0c, 0x61, 0x6e, 0x61, 0x6c, 0x79, 0x73,
	0x69, 0x73, 0x4b, 0x69, 0x6e, 0x64, 0x22, 0x50, 0x0a, 0x07, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c,
	0x73, 0x12, 0x45, 0x0a, 0x0a, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x65, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x67, 0x72, 0x61, 0x66, 0x65, 0x61, 0x73, 0x2e,
	0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72,
	0x79, 0x2e, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x65, 0x64, 0x52, 0x0a, 0x64, 0x69,
	0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x65, 0x64, 0x22, 0xd1, 0x04, 0x0a, 0x0a, 0x44, 0x69, 0x73,
	0x63, 0x6f, 0x76, 0x65, 0x72, 0x65, 0x64, 0x12, 0x69, 0x0a, 0x13, 0x63, 0x6f, 0x6e, 0x74, 0x69,
	0x6e, 0x75, 0x6f, 0x75, 0x73, 0x5f, 0x61, 0x6e, 0x61, 0x6c, 0x79, 0x73, 0x69, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x38, 0x2e, 0x67, 0x72, 0x61, 0x66, 0x65, 0x61, 0x73, 0x2e, 0x76,
	0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79,
	0x2e, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x65, 0x64, 0x2e, 0x43, 0x6f, 0x6e, 0x74,
	0x69, 0x6e, 0x75, 0x6f, 0x75, 0x73, 0x41, 0x6e, 0x61, 0x6c, 0x79, 0x73, 0x69, 0x73, 0x52, 0x12,
	0x63, 0x6f, 0x6e, 0x74, 0x69, 0x6e, 0x75, 0x6f, 0x75, 0x73, 0x41, 0x6e, 0x61, 0x6c, 0x79, 0x73,
	0x69, 0x73, 0x12, 0x48, 0x0a, 0x12, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x61, 0x6e, 0x61, 0x6c, 0x79,
	0x73, 0x69, 0x73, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x10, 0x6c, 0x61, 0x73, 0x74,
	0x41, 0x6e, 0x61, 0x6c, 0x79, 0x73, 0x69, 0x73, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x5d, 0x0a, 0x0f,
	0x61, 0x6e, 0x61, 0x6c, 0x79, 0x73, 0x69, 0x73, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x34, 0x2e, 0x67, 0x72, 0x61, 0x66, 0x65, 0x61, 0x73, 0x2e,
	0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72,
	0x79, 0x2e, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x65, 0x64, 0x2e, 0x41, 0x6e, 0x61,
	0x6c, 0x79, 0x73, 0x69, 0x73, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x0e, 0x61, 0x6e, 0x61,
	0x6c, 0x79, 0x73, 0x69, 0x73, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x46, 0x0a, 0x15, 0x61,
	0x6e, 0x61, 0x6c, 0x79, 0x73, 0x69, 0x73, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x5f, 0x65,
	0x72, 0x72, 0x6f, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x13,
	0x61, 0x6e, 0x61, 0x6c, 0x79, 0x73, 0x69, 0x73, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x45, 0x72,
	0x72, 0x6f, 0x72, 0x22, 0x53, 0x0a, 0x12, 0x43, 0x6f, 0x6e, 0x74, 0x69, 0x6e, 0x75, 0x6f, 0x75,
	0x73, 0x41, 0x6e, 0x61, 0x6c, 0x79, 0x73, 0x69, 0x73, 0x12, 0x23, 0x0a, 0x1f, 0x43, 0x4f, 0x4e,
	0x54, 0x49, 0x4e, 0x55, 0x4f, 0x55, 0x53, 0x5f, 0x41, 0x4e, 0x41, 0x4c, 0x59, 0x53, 0x49, 0x53,
	0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0a,
	0x0a, 0x06, 0x41, 0x43, 0x54, 0x49, 0x56, 0x45, 0x10, 0x01, 0x12, 0x0c, 0x0a, 0x08, 0x49, 0x4e,
	0x41, 0x43, 0x54, 0x49, 0x56, 0x45, 0x10, 0x02, 0x22, 0x91, 0x01, 0x0a, 0x0e, 0x41, 0x6e, 0x61,
	0x6c, 0x79, 0x73, 0x69, 0x73, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1f, 0x0a, 0x1b, 0x41,
	0x4e, 0x41, 0x4c, 0x59, 0x53, 0x49, 0x53, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x55,
	0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07,
	0x50, 0x45, 0x4e, 0x44, 0x49, 0x4e, 0x47, 0x10, 0x01, 0x12, 0x0c, 0x0a, 0x08, 0x53, 0x43, 0x41,
	0x4e, 0x4e, 0x49, 0x4e, 0x47, 0x10, 0x02, 0x12, 0x14, 0x0a, 0x10, 0x46, 0x49, 0x4e, 0x49, 0x53,
	0x48, 0x45, 0x44, 0x5f, 0x53, 0x55, 0x43, 0x43, 0x45, 0x53, 0x53, 0x10, 0x03, 0x12, 0x13, 0x0a,
	0x0f, 0x46, 0x49, 0x4e, 0x49, 0x53, 0x48, 0x45, 0x44, 0x5f, 0x46, 0x41, 0x49, 0x4c, 0x45, 0x44,
	0x10, 0x04, 0x12, 0x18, 0x0a, 0x14, 0x46, 0x49, 0x4e, 0x49, 0x53, 0x48, 0x45, 0x44, 0x5f, 0x55,
	0x4e, 0x53, 0x55, 0x50, 0x50, 0x4f, 0x52, 0x54, 0x45, 0x44, 0x10, 0x05, 0x42, 0x84, 0x01, 0x0a,
	0x1c, 0x69, 0x6f, 0x2e, 0x67, 0x72, 0x61, 0x66, 0x65, 0x61, 0x73, 0x2e, 0x76, 0x31, 0x62, 0x65,
	0x74, 0x61, 0x31, 0x2e, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0x50, 0x01, 0x5a,
	0x5c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x2e, 0x6f,
	0x72, 0x67, 0x2f, 0x67, 0x65, 0x6e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x61, 0x70, 0x69, 0x73, 0x2f, 0x64, 0x65, 0x76, 0x74, 0x6f, 0x6f, 0x6c, 0x73, 0x2f,
	0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x61, 0x6e, 0x61, 0x6c, 0x79, 0x73, 0x69,
	0x73, 0x2f, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2f, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x76,
	0x65, 0x72, 0x79, 0x3b, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79, 0xa2, 0x02, 0x03,
	0x47, 0x52, 0x41, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_rawDescOnce sync.Once
	file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_rawDescData = file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_rawDesc
)

func file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_rawDescGZIP() []byte {
	file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_rawDescOnce.Do(func() {
		file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_rawDescData)
	})
	return file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_rawDescData
}

var file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_goTypes = []interface{}{
	(Discovered_ContinuousAnalysis)(0), // 0: grafeas.v1beta1.discovery.Discovered.ContinuousAnalysis
	(Discovered_AnalysisStatus)(0),     // 1: grafeas.v1beta1.discovery.Discovered.AnalysisStatus
	(*Discovery)(nil),                  // 2: grafeas.v1beta1.discovery.Discovery
	(*Details)(nil),                    // 3: grafeas.v1beta1.discovery.Details
	(*Discovered)(nil),                 // 4: grafeas.v1beta1.discovery.Discovered
	(common.NoteKind)(0),               // 5: grafeas.v1beta1.NoteKind
	(*timestamppb.Timestamp)(nil),      // 6: google.protobuf.Timestamp
	(*status.Status)(nil),              // 7: google.rpc.Status
}
var file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_depIdxs = []int32{
	5, // 0: grafeas.v1beta1.discovery.Discovery.analysis_kind:type_name -> grafeas.v1beta1.NoteKind
	4, // 1: grafeas.v1beta1.discovery.Details.discovered:type_name -> grafeas.v1beta1.discovery.Discovered
	0, // 2: grafeas.v1beta1.discovery.Discovered.continuous_analysis:type_name -> grafeas.v1beta1.discovery.Discovered.ContinuousAnalysis
	6, // 3: grafeas.v1beta1.discovery.Discovered.last_analysis_time:type_name -> google.protobuf.Timestamp
	1, // 4: grafeas.v1beta1.discovery.Discovered.analysis_status:type_name -> grafeas.v1beta1.discovery.Discovered.AnalysisStatus
	7, // 5: grafeas.v1beta1.discovery.Discovered.analysis_status_error:type_name -> google.rpc.Status
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_init() }
func file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_init() {
	if File_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Discovery); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Details); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Discovered); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_goTypes,
		DependencyIndexes: file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_depIdxs,
		EnumInfos:         file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_enumTypes,
		MessageInfos:      file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_msgTypes,
	}.Build()
	File_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto = out.File
	file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_rawDesc = nil
	file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_goTypes = nil
	file_google_devtools_containeranalysis_v1beta1_discovery_discovery_proto_depIdxs = nil
}
