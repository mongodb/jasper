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
// source: google/devtools/containeranalysis/v1beta1/provenance/provenance.proto

package provenance

import (
	reflect "reflect"
	sync "sync"

	proto "github.com/golang/protobuf/proto"
	source "google.golang.org/genproto/googleapis/devtools/containeranalysis/v1beta1/source"
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

// Specifies the hash algorithm.
type Hash_HashType int32

const (
	// Unknown.
	Hash_HASH_TYPE_UNSPECIFIED Hash_HashType = 0
	// A SHA-256 hash.
	Hash_SHA256 Hash_HashType = 1
)

// Enum value maps for Hash_HashType.
var (
	Hash_HashType_name = map[int32]string{
		0: "HASH_TYPE_UNSPECIFIED",
		1: "SHA256",
	}
	Hash_HashType_value = map[string]int32{
		"HASH_TYPE_UNSPECIFIED": 0,
		"SHA256":                1,
	}
)

func (x Hash_HashType) Enum() *Hash_HashType {
	p := new(Hash_HashType)
	*p = x
	return p
}

func (x Hash_HashType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Hash_HashType) Descriptor() protoreflect.EnumDescriptor {
	return file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_enumTypes[0].Descriptor()
}

func (Hash_HashType) Type() protoreflect.EnumType {
	return &file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_enumTypes[0]
}

func (x Hash_HashType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Hash_HashType.Descriptor instead.
func (Hash_HashType) EnumDescriptor() ([]byte, []int) {
	return file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_rawDescGZIP(), []int{3, 0}
}

// Provenance of a build. Contains all information needed to verify the full
// details about the build from source to completion.
type BuildProvenance struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Required. Unique identifier of the build.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// ID of the project.
	ProjectId string `protobuf:"bytes,2,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
	// Commands requested by the build.
	Commands []*Command `protobuf:"bytes,3,rep,name=commands,proto3" json:"commands,omitempty"`
	// Output of the build.
	BuiltArtifacts []*Artifact `protobuf:"bytes,4,rep,name=built_artifacts,json=builtArtifacts,proto3" json:"built_artifacts,omitempty"`
	// Time at which the build was created.
	CreateTime *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty"`
	// Time at which execution of the build was started.
	StartTime *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=start_time,json=startTime,proto3" json:"start_time,omitempty"`
	// Time at which execution of the build was finished.
	EndTime *timestamppb.Timestamp `protobuf:"bytes,7,opt,name=end_time,json=endTime,proto3" json:"end_time,omitempty"`
	// E-mail address of the user who initiated this build. Note that this was the
	// user's e-mail address at the time the build was initiated; this address may
	// not represent the same end-user for all time.
	Creator string `protobuf:"bytes,8,opt,name=creator,proto3" json:"creator,omitempty"`
	// URI where any logs for this provenance were written.
	LogsUri string `protobuf:"bytes,9,opt,name=logs_uri,json=logsUri,proto3" json:"logs_uri,omitempty"`
	// Details of the Source input to the build.
	SourceProvenance *Source `protobuf:"bytes,10,opt,name=source_provenance,json=sourceProvenance,proto3" json:"source_provenance,omitempty"`
	// Trigger identifier if the build was triggered automatically; empty if not.
	TriggerId string `protobuf:"bytes,11,opt,name=trigger_id,json=triggerId,proto3" json:"trigger_id,omitempty"`
	// Special options applied to this build. This is a catch-all field where
	// build providers can enter any desired additional details.
	BuildOptions map[string]string `protobuf:"bytes,12,rep,name=build_options,json=buildOptions,proto3" json:"build_options,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Version string of the builder at the time this build was executed.
	BuilderVersion string `protobuf:"bytes,13,opt,name=builder_version,json=builderVersion,proto3" json:"builder_version,omitempty"`
}

func (x *BuildProvenance) Reset() {
	*x = BuildProvenance{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BuildProvenance) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BuildProvenance) ProtoMessage() {}

func (x *BuildProvenance) ProtoReflect() protoreflect.Message {
	mi := &file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BuildProvenance.ProtoReflect.Descriptor instead.
func (*BuildProvenance) Descriptor() ([]byte, []int) {
	return file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_rawDescGZIP(), []int{0}
}

func (x *BuildProvenance) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *BuildProvenance) GetProjectId() string {
	if x != nil {
		return x.ProjectId
	}
	return ""
}

func (x *BuildProvenance) GetCommands() []*Command {
	if x != nil {
		return x.Commands
	}
	return nil
}

func (x *BuildProvenance) GetBuiltArtifacts() []*Artifact {
	if x != nil {
		return x.BuiltArtifacts
	}
	return nil
}

func (x *BuildProvenance) GetCreateTime() *timestamppb.Timestamp {
	if x != nil {
		return x.CreateTime
	}
	return nil
}

func (x *BuildProvenance) GetStartTime() *timestamppb.Timestamp {
	if x != nil {
		return x.StartTime
	}
	return nil
}

func (x *BuildProvenance) GetEndTime() *timestamppb.Timestamp {
	if x != nil {
		return x.EndTime
	}
	return nil
}

func (x *BuildProvenance) GetCreator() string {
	if x != nil {
		return x.Creator
	}
	return ""
}

func (x *BuildProvenance) GetLogsUri() string {
	if x != nil {
		return x.LogsUri
	}
	return ""
}

func (x *BuildProvenance) GetSourceProvenance() *Source {
	if x != nil {
		return x.SourceProvenance
	}
	return nil
}

func (x *BuildProvenance) GetTriggerId() string {
	if x != nil {
		return x.TriggerId
	}
	return ""
}

func (x *BuildProvenance) GetBuildOptions() map[string]string {
	if x != nil {
		return x.BuildOptions
	}
	return nil
}

func (x *BuildProvenance) GetBuilderVersion() string {
	if x != nil {
		return x.BuilderVersion
	}
	return ""
}

// Source describes the location of the source used for the build.
type Source struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// If provided, the input binary artifacts for the build came from this
	// location.
	ArtifactStorageSourceUri string `protobuf:"bytes,1,opt,name=artifact_storage_source_uri,json=artifactStorageSourceUri,proto3" json:"artifact_storage_source_uri,omitempty"`
	// Hash(es) of the build source, which can be used to verify that the original
	// source integrity was maintained in the build.
	//
	// The keys to this map are file paths used as build source and the values
	// contain the hash values for those files.
	//
	// If the build source came in a single package such as a gzipped tarfile
	// (.tar.gz), the FileHash will be for the single path to that file.
	FileHashes map[string]*FileHashes `protobuf:"bytes,2,rep,name=file_hashes,json=fileHashes,proto3" json:"file_hashes,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// If provided, the source code used for the build came from this location.
	Context *source.SourceContext `protobuf:"bytes,3,opt,name=context,proto3" json:"context,omitempty"`
	// If provided, some of the source code used for the build may be found in
	// these locations, in the case where the source repository had multiple
	// remotes or submodules. This list will not include the context specified in
	// the context field.
	AdditionalContexts []*source.SourceContext `protobuf:"bytes,4,rep,name=additional_contexts,json=additionalContexts,proto3" json:"additional_contexts,omitempty"`
}

func (x *Source) Reset() {
	*x = Source{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Source) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Source) ProtoMessage() {}

func (x *Source) ProtoReflect() protoreflect.Message {
	mi := &file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Source.ProtoReflect.Descriptor instead.
func (*Source) Descriptor() ([]byte, []int) {
	return file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_rawDescGZIP(), []int{1}
}

func (x *Source) GetArtifactStorageSourceUri() string {
	if x != nil {
		return x.ArtifactStorageSourceUri
	}
	return ""
}

func (x *Source) GetFileHashes() map[string]*FileHashes {
	if x != nil {
		return x.FileHashes
	}
	return nil
}

func (x *Source) GetContext() *source.SourceContext {
	if x != nil {
		return x.Context
	}
	return nil
}

func (x *Source) GetAdditionalContexts() []*source.SourceContext {
	if x != nil {
		return x.AdditionalContexts
	}
	return nil
}

// Container message for hashes of byte content of files, used in source
// messages to verify integrity of source input to the build.
type FileHashes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Required. Collection of file hashes.
	FileHash []*Hash `protobuf:"bytes,1,rep,name=file_hash,json=fileHash,proto3" json:"file_hash,omitempty"`
}

func (x *FileHashes) Reset() {
	*x = FileHashes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileHashes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileHashes) ProtoMessage() {}

func (x *FileHashes) ProtoReflect() protoreflect.Message {
	mi := &file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileHashes.ProtoReflect.Descriptor instead.
func (*FileHashes) Descriptor() ([]byte, []int) {
	return file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_rawDescGZIP(), []int{2}
}

func (x *FileHashes) GetFileHash() []*Hash {
	if x != nil {
		return x.FileHash
	}
	return nil
}

// Container message for hash values.
type Hash struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Required. The type of hash that was performed.
	Type Hash_HashType `protobuf:"varint,1,opt,name=type,proto3,enum=grafeas.v1beta1.provenance.Hash_HashType" json:"type,omitempty"`
	// Required. The hash value.
	Value []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Hash) Reset() {
	*x = Hash{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Hash) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Hash) ProtoMessage() {}

func (x *Hash) ProtoReflect() protoreflect.Message {
	mi := &file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Hash.ProtoReflect.Descriptor instead.
func (*Hash) Descriptor() ([]byte, []int) {
	return file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_rawDescGZIP(), []int{3}
}

func (x *Hash) GetType() Hash_HashType {
	if x != nil {
		return x.Type
	}
	return Hash_HASH_TYPE_UNSPECIFIED
}

func (x *Hash) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

// Command describes a step performed as part of the build pipeline.
type Command struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Required. Name of the command, as presented on the command line, or if the
	// command is packaged as a Docker container, as presented to `docker pull`.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Environment variables set before running this command.
	Env []string `protobuf:"bytes,2,rep,name=env,proto3" json:"env,omitempty"`
	// Command-line arguments used when executing this command.
	Args []string `protobuf:"bytes,3,rep,name=args,proto3" json:"args,omitempty"`
	// Working directory (relative to project source root) used when running this
	// command.
	Dir string `protobuf:"bytes,4,opt,name=dir,proto3" json:"dir,omitempty"`
	// Optional unique identifier for this command, used in wait_for to reference
	// this command as a dependency.
	Id string `protobuf:"bytes,5,opt,name=id,proto3" json:"id,omitempty"`
	// The ID(s) of the command(s) that this command depends on.
	WaitFor []string `protobuf:"bytes,6,rep,name=wait_for,json=waitFor,proto3" json:"wait_for,omitempty"`
}

func (x *Command) Reset() {
	*x = Command{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Command) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Command) ProtoMessage() {}

func (x *Command) ProtoReflect() protoreflect.Message {
	mi := &file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Command.ProtoReflect.Descriptor instead.
func (*Command) Descriptor() ([]byte, []int) {
	return file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_rawDescGZIP(), []int{4}
}

func (x *Command) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Command) GetEnv() []string {
	if x != nil {
		return x.Env
	}
	return nil
}

func (x *Command) GetArgs() []string {
	if x != nil {
		return x.Args
	}
	return nil
}

func (x *Command) GetDir() string {
	if x != nil {
		return x.Dir
	}
	return ""
}

func (x *Command) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Command) GetWaitFor() []string {
	if x != nil {
		return x.WaitFor
	}
	return nil
}

// Artifact describes a build product.
type Artifact struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Hash or checksum value of a binary, or Docker Registry 2.0 digest of a
	// container.
	Checksum string `protobuf:"bytes,1,opt,name=checksum,proto3" json:"checksum,omitempty"`
	// Artifact ID, if any; for container images, this will be a URL by digest
	// like `gcr.io/projectID/imagename@sha256:123456`.
	Id string `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	// Related artifact names. This may be the path to a binary or jar file, or in
	// the case of a container build, the name used to push the container image to
	// Google Container Registry, as presented to `docker push`. Note that a
	// single Artifact ID can have multiple names, for example if two tags are
	// applied to one image.
	Names []string `protobuf:"bytes,3,rep,name=names,proto3" json:"names,omitempty"`
}

func (x *Artifact) Reset() {
	*x = Artifact{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Artifact) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Artifact) ProtoMessage() {}

func (x *Artifact) ProtoReflect() protoreflect.Message {
	mi := &file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Artifact.ProtoReflect.Descriptor instead.
func (*Artifact) Descriptor() ([]byte, []int) {
	return file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_rawDescGZIP(), []int{5}
}

func (x *Artifact) GetChecksum() string {
	if x != nil {
		return x.Checksum
	}
	return ""
}

func (x *Artifact) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Artifact) GetNames() []string {
	if x != nil {
		return x.Names
	}
	return nil
}

var File_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto protoreflect.FileDescriptor

var file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_rawDesc = []byte{
	0x0a, 0x45, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x64, 0x65, 0x76, 0x74, 0x6f, 0x6f, 0x6c,
	0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x61, 0x6e, 0x61, 0x6c, 0x79,
	0x73, 0x69, 0x73, 0x2f, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2f, 0x70, 0x72, 0x6f, 0x76,
	0x65, 0x6e, 0x61, 0x6e, 0x63, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x6e, 0x61, 0x6e, 0x63,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1a, 0x67, 0x72, 0x61, 0x66, 0x65, 0x61, 0x73,
	0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x6e, 0x61,
	0x6e, 0x63, 0x65, 0x1a, 0x3d, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x64, 0x65, 0x76, 0x74,
	0x6f, 0x6f, 0x6c, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x61, 0x6e,
	0x61, 0x6c, 0x79, 0x73, 0x69, 0x73, 0x2f, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2f, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x2f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0xf2, 0x05, 0x0a, 0x0f, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x50, 0x72, 0x6f,
	0x76, 0x65, 0x6e, 0x61, 0x6e, 0x63, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x6a, 0x65,
	0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f,
	0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x12, 0x3f, 0x0a, 0x08, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e,
	0x64, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x67, 0x72, 0x61, 0x66, 0x65,
	0x61, 0x73, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x65,
	0x6e, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x52, 0x08, 0x63,
	0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x73, 0x12, 0x4d, 0x0a, 0x0f, 0x62, 0x75, 0x69, 0x6c, 0x74,
	0x5f, 0x61, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x24, 0x2e, 0x67, 0x72, 0x61, 0x66, 0x65, 0x61, 0x73, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74,
	0x61, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x6e, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x41, 0x72,
	0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x52, 0x0e, 0x62, 0x75, 0x69, 0x6c, 0x74, 0x41, 0x72, 0x74,
	0x69, 0x66, 0x61, 0x63, 0x74, 0x73, 0x12, 0x3b, 0x0a, 0x0b, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54,
	0x69, 0x6d, 0x65, 0x12, 0x39, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x74, 0x69, 0x6d,
	0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x52, 0x09, 0x73, 0x74, 0x61, 0x72, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x35,
	0x0a, 0x08, 0x65, 0x6e, 0x64, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x07, 0x65, 0x6e,
	0x64, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x6f, 0x72,
	0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x6f, 0x72, 0x12,
	0x19, 0x0a, 0x08, 0x6c, 0x6f, 0x67, 0x73, 0x5f, 0x75, 0x72, 0x69, 0x18, 0x09, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x6c, 0x6f, 0x67, 0x73, 0x55, 0x72, 0x69, 0x12, 0x4f, 0x0a, 0x11, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x5f, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x6e, 0x61, 0x6e, 0x63, 0x65, 0x18,
	0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x67, 0x72, 0x61, 0x66, 0x65, 0x61, 0x73, 0x2e,
	0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x6e, 0x61, 0x6e,
	0x63, 0x65, 0x2e, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52, 0x10, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x50, 0x72, 0x6f, 0x76, 0x65, 0x6e, 0x61, 0x6e, 0x63, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x74,
	0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x74, 0x72, 0x69, 0x67, 0x67, 0x65, 0x72, 0x49, 0x64, 0x12, 0x62, 0x0a, 0x0d, 0x62, 0x75,
	0x69, 0x6c, 0x64, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x0c, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x3d, 0x2e, 0x67, 0x72, 0x61, 0x66, 0x65, 0x61, 0x73, 0x2e, 0x76, 0x31, 0x62, 0x65,
	0x74, 0x61, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x6e, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x42,
	0x75, 0x69, 0x6c, 0x64, 0x50, 0x72, 0x6f, 0x76, 0x65, 0x6e, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x42,
	0x75, 0x69, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x52, 0x0c, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x27,
	0x0a, 0x0f, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72,
	0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x1a, 0x3f, 0x0a, 0x11, 0x42, 0x75, 0x69, 0x6c, 0x64,
	0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x9c, 0x03, 0x0a, 0x06, 0x53, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x12, 0x3d, 0x0a, 0x1b, 0x61, 0x72, 0x74, 0x69, 0x66, 0x61, 0x63, 0x74, 0x5f,
	0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x75,
	0x72, 0x69, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x18, 0x61, 0x72, 0x74, 0x69, 0x66, 0x61,
	0x63, 0x74, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x55,
	0x72, 0x69, 0x12, 0x53, 0x0a, 0x0b, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x65,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x32, 0x2e, 0x67, 0x72, 0x61, 0x66, 0x65, 0x61,
	0x73, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x6e,
	0x61, 0x6e, 0x63, 0x65, 0x2e, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x46, 0x69, 0x6c, 0x65,
	0x48, 0x61, 0x73, 0x68, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0a, 0x66, 0x69, 0x6c,
	0x65, 0x48, 0x61, 0x73, 0x68, 0x65, 0x73, 0x12, 0x3f, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x78, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x67, 0x72, 0x61, 0x66, 0x65,
	0x61, 0x73, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x2e, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52,
	0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x12, 0x56, 0x0a, 0x13, 0x61, 0x64, 0x64, 0x69,
	0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x73, 0x18,
	0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x67, 0x72, 0x61, 0x66, 0x65, 0x61, 0x73, 0x2e,
	0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x53,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52, 0x12, 0x61, 0x64,
	0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x73,
	0x1a, 0x65, 0x0a, 0x0f, 0x46, 0x69, 0x6c, 0x65, 0x48, 0x61, 0x73, 0x68, 0x65, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x3c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x67, 0x72, 0x61, 0x66, 0x65, 0x61, 0x73, 0x2e, 0x76,
	0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x6e, 0x61, 0x6e, 0x63,
	0x65, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x48, 0x61, 0x73, 0x68, 0x65, 0x73, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x4b, 0x0a, 0x0a, 0x46, 0x69, 0x6c, 0x65, 0x48,
	0x61, 0x73, 0x68, 0x65, 0x73, 0x12, 0x3d, 0x0a, 0x09, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x68, 0x61,
	0x73, 0x68, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x67, 0x72, 0x61, 0x66, 0x65,
	0x61, 0x73, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x65,
	0x6e, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x48, 0x61, 0x73, 0x68, 0x52, 0x08, 0x66, 0x69, 0x6c, 0x65,
	0x48, 0x61, 0x73, 0x68, 0x22, 0x8e, 0x01, 0x0a, 0x04, 0x48, 0x61, 0x73, 0x68, 0x12, 0x3d, 0x0a,
	0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x29, 0x2e, 0x67, 0x72,
	0x61, 0x66, 0x65, 0x61, 0x73, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x70, 0x72,
	0x6f, 0x76, 0x65, 0x6e, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x48, 0x61, 0x73, 0x68, 0x2e, 0x48, 0x61,
	0x73, 0x68, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x22, 0x31, 0x0a, 0x08, 0x48, 0x61, 0x73, 0x68, 0x54, 0x79, 0x70, 0x65, 0x12, 0x19,
	0x0a, 0x15, 0x48, 0x41, 0x53, 0x48, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x4e, 0x53, 0x50,
	0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x53, 0x48, 0x41,
	0x32, 0x35, 0x36, 0x10, 0x01, 0x22, 0x80, 0x01, 0x0a, 0x07, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e,
	0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x65, 0x6e, 0x76, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x09, 0x52, 0x03, 0x65, 0x6e, 0x76, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x72, 0x67, 0x73, 0x18,
	0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x61, 0x72, 0x67, 0x73, 0x12, 0x10, 0x0a, 0x03, 0x64,
	0x69, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x64, 0x69, 0x72, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x19, 0x0a,
	0x08, 0x77, 0x61, 0x69, 0x74, 0x5f, 0x66, 0x6f, 0x72, 0x18, 0x06, 0x20, 0x03, 0x28, 0x09, 0x52,
	0x07, 0x77, 0x61, 0x69, 0x74, 0x46, 0x6f, 0x72, 0x22, 0x4c, 0x0a, 0x08, 0x41, 0x72, 0x74, 0x69,
	0x66, 0x61, 0x63, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x73, 0x75, 0x6d,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x73, 0x75, 0x6d,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x14, 0x0a, 0x05, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52,
	0x05, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x42, 0x87, 0x01, 0x0a, 0x1d, 0x69, 0x6f, 0x2e, 0x67, 0x72,
	0x61, 0x66, 0x65, 0x61, 0x73, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x70, 0x72,
	0x6f, 0x76, 0x65, 0x6e, 0x61, 0x6e, 0x63, 0x65, 0x50, 0x01, 0x5a, 0x5e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x67, 0x65,
	0x6e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x61, 0x70, 0x69,
	0x73, 0x2f, 0x64, 0x65, 0x76, 0x74, 0x6f, 0x6f, 0x6c, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61,
	0x69, 0x6e, 0x65, 0x72, 0x61, 0x6e, 0x61, 0x6c, 0x79, 0x73, 0x69, 0x73, 0x2f, 0x76, 0x31, 0x62,
	0x65, 0x74, 0x61, 0x31, 0x2f, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x6e, 0x61, 0x6e, 0x63, 0x65, 0x3b,
	0x70, 0x72, 0x6f, 0x76, 0x65, 0x6e, 0x61, 0x6e, 0x63, 0x65, 0xa2, 0x02, 0x03, 0x47, 0x52, 0x41,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_rawDescOnce sync.Once
	file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_rawDescData = file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_rawDesc
)

func file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_rawDescGZIP() []byte {
	file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_rawDescOnce.Do(func() {
		file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_rawDescData)
	})
	return file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_rawDescData
}

var file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_goTypes = []interface{}{
	(Hash_HashType)(0),            // 0: grafeas.v1beta1.provenance.Hash.HashType
	(*BuildProvenance)(nil),       // 1: grafeas.v1beta1.provenance.BuildProvenance
	(*Source)(nil),                // 2: grafeas.v1beta1.provenance.Source
	(*FileHashes)(nil),            // 3: grafeas.v1beta1.provenance.FileHashes
	(*Hash)(nil),                  // 4: grafeas.v1beta1.provenance.Hash
	(*Command)(nil),               // 5: grafeas.v1beta1.provenance.Command
	(*Artifact)(nil),              // 6: grafeas.v1beta1.provenance.Artifact
	nil,                           // 7: grafeas.v1beta1.provenance.BuildProvenance.BuildOptionsEntry
	nil,                           // 8: grafeas.v1beta1.provenance.Source.FileHashesEntry
	(*timestamppb.Timestamp)(nil), // 9: google.protobuf.Timestamp
	(*source.SourceContext)(nil),  // 10: grafeas.v1beta1.source.SourceContext
}
var file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_depIdxs = []int32{
	5,  // 0: grafeas.v1beta1.provenance.BuildProvenance.commands:type_name -> grafeas.v1beta1.provenance.Command
	6,  // 1: grafeas.v1beta1.provenance.BuildProvenance.built_artifacts:type_name -> grafeas.v1beta1.provenance.Artifact
	9,  // 2: grafeas.v1beta1.provenance.BuildProvenance.create_time:type_name -> google.protobuf.Timestamp
	9,  // 3: grafeas.v1beta1.provenance.BuildProvenance.start_time:type_name -> google.protobuf.Timestamp
	9,  // 4: grafeas.v1beta1.provenance.BuildProvenance.end_time:type_name -> google.protobuf.Timestamp
	2,  // 5: grafeas.v1beta1.provenance.BuildProvenance.source_provenance:type_name -> grafeas.v1beta1.provenance.Source
	7,  // 6: grafeas.v1beta1.provenance.BuildProvenance.build_options:type_name -> grafeas.v1beta1.provenance.BuildProvenance.BuildOptionsEntry
	8,  // 7: grafeas.v1beta1.provenance.Source.file_hashes:type_name -> grafeas.v1beta1.provenance.Source.FileHashesEntry
	10, // 8: grafeas.v1beta1.provenance.Source.context:type_name -> grafeas.v1beta1.source.SourceContext
	10, // 9: grafeas.v1beta1.provenance.Source.additional_contexts:type_name -> grafeas.v1beta1.source.SourceContext
	4,  // 10: grafeas.v1beta1.provenance.FileHashes.file_hash:type_name -> grafeas.v1beta1.provenance.Hash
	0,  // 11: grafeas.v1beta1.provenance.Hash.type:type_name -> grafeas.v1beta1.provenance.Hash.HashType
	3,  // 12: grafeas.v1beta1.provenance.Source.FileHashesEntry.value:type_name -> grafeas.v1beta1.provenance.FileHashes
	13, // [13:13] is the sub-list for method output_type
	13, // [13:13] is the sub-list for method input_type
	13, // [13:13] is the sub-list for extension type_name
	13, // [13:13] is the sub-list for extension extendee
	0,  // [0:13] is the sub-list for field type_name
}

func init() { file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_init() }
func file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_init() {
	if File_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BuildProvenance); i {
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
		file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Source); i {
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
		file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileHashes); i {
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
		file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Hash); i {
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
		file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Command); i {
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
		file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Artifact); i {
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
			RawDescriptor: file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_goTypes,
		DependencyIndexes: file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_depIdxs,
		EnumInfos:         file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_enumTypes,
		MessageInfos:      file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_msgTypes,
	}.Build()
	File_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto = out.File
	file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_rawDesc = nil
	file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_goTypes = nil
	file_google_devtools_containeranalysis_v1beta1_provenance_provenance_proto_depIdxs = nil
}
