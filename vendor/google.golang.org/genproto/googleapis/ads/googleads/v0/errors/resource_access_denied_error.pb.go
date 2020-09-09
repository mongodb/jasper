// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/ads/googleads/v0/errors/resource_access_denied_error.proto

package errors // import "google.golang.org/genproto/googleapis/ads/googleads/v0/errors"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Enum describing possible resource access denied errors.
type ResourceAccessDeniedErrorEnum_ResourceAccessDeniedError int32

const (
	// Enum unspecified.
	ResourceAccessDeniedErrorEnum_UNSPECIFIED ResourceAccessDeniedErrorEnum_ResourceAccessDeniedError = 0
	// The received error code is not known in this version.
	ResourceAccessDeniedErrorEnum_UNKNOWN ResourceAccessDeniedErrorEnum_ResourceAccessDeniedError = 1
	// User did not have write access.
	ResourceAccessDeniedErrorEnum_WRITE_ACCESS_DENIED ResourceAccessDeniedErrorEnum_ResourceAccessDeniedError = 3
)

var ResourceAccessDeniedErrorEnum_ResourceAccessDeniedError_name = map[int32]string{
	0: "UNSPECIFIED",
	1: "UNKNOWN",
	3: "WRITE_ACCESS_DENIED",
}
var ResourceAccessDeniedErrorEnum_ResourceAccessDeniedError_value = map[string]int32{
	"UNSPECIFIED":         0,
	"UNKNOWN":             1,
	"WRITE_ACCESS_DENIED": 3,
}

func (x ResourceAccessDeniedErrorEnum_ResourceAccessDeniedError) String() string {
	return proto.EnumName(ResourceAccessDeniedErrorEnum_ResourceAccessDeniedError_name, int32(x))
}
func (ResourceAccessDeniedErrorEnum_ResourceAccessDeniedError) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_resource_access_denied_error_7ec8d6c2b40945ba, []int{0, 0}
}

// Container for enum describing possible resource access denied errors.
type ResourceAccessDeniedErrorEnum struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ResourceAccessDeniedErrorEnum) Reset()         { *m = ResourceAccessDeniedErrorEnum{} }
func (m *ResourceAccessDeniedErrorEnum) String() string { return proto.CompactTextString(m) }
func (*ResourceAccessDeniedErrorEnum) ProtoMessage()    {}
func (*ResourceAccessDeniedErrorEnum) Descriptor() ([]byte, []int) {
	return fileDescriptor_resource_access_denied_error_7ec8d6c2b40945ba, []int{0}
}
func (m *ResourceAccessDeniedErrorEnum) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ResourceAccessDeniedErrorEnum.Unmarshal(m, b)
}
func (m *ResourceAccessDeniedErrorEnum) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ResourceAccessDeniedErrorEnum.Marshal(b, m, deterministic)
}
func (dst *ResourceAccessDeniedErrorEnum) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ResourceAccessDeniedErrorEnum.Merge(dst, src)
}
func (m *ResourceAccessDeniedErrorEnum) XXX_Size() int {
	return xxx_messageInfo_ResourceAccessDeniedErrorEnum.Size(m)
}
func (m *ResourceAccessDeniedErrorEnum) XXX_DiscardUnknown() {
	xxx_messageInfo_ResourceAccessDeniedErrorEnum.DiscardUnknown(m)
}

var xxx_messageInfo_ResourceAccessDeniedErrorEnum proto.InternalMessageInfo

func init() {
	proto.RegisterType((*ResourceAccessDeniedErrorEnum)(nil), "google.ads.googleads.v0.errors.ResourceAccessDeniedErrorEnum")
	proto.RegisterEnum("google.ads.googleads.v0.errors.ResourceAccessDeniedErrorEnum_ResourceAccessDeniedError", ResourceAccessDeniedErrorEnum_ResourceAccessDeniedError_name, ResourceAccessDeniedErrorEnum_ResourceAccessDeniedError_value)
}

func init() {
	proto.RegisterFile("google/ads/googleads/v0/errors/resource_access_denied_error.proto", fileDescriptor_resource_access_denied_error_7ec8d6c2b40945ba)
}

var fileDescriptor_resource_access_denied_error_7ec8d6c2b40945ba = []byte{
	// 294 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x90, 0xcf, 0x4a, 0xc3, 0x30,
	0x1c, 0xc7, 0x5d, 0x07, 0x0a, 0xd9, 0xc1, 0x51, 0x0f, 0xe2, 0xc1, 0x1e, 0xfa, 0x00, 0x69, 0xc1,
	0x5b, 0x3c, 0x65, 0x6d, 0x1c, 0x45, 0xa8, 0xa5, 0x75, 0x1d, 0x48, 0xa1, 0xd4, 0x26, 0x84, 0xc1,
	0xd6, 0x8c, 0xfc, 0xdc, 0x1e, 0xc8, 0xa3, 0x8f, 0xe2, 0xa3, 0xf8, 0x06, 0xde, 0xa4, 0xc9, 0xd6,
	0x5b, 0x3d, 0xf5, 0x4b, 0x7f, 0x9f, 0x7c, 0x7e, 0x7f, 0x10, 0x95, 0x4a, 0xc9, 0xad, 0x08, 0x1a,
	0x0e, 0x81, 0x8d, 0x7d, 0x3a, 0x86, 0x81, 0xd0, 0x5a, 0x69, 0x08, 0xb4, 0x00, 0x75, 0xd0, 0xad,
	0xa8, 0x9b, 0xb6, 0x15, 0x00, 0x35, 0x17, 0xdd, 0x46, 0xf0, 0xda, 0x54, 0xf1, 0x5e, 0xab, 0x0f,
	0xe5, 0x7a, 0xf6, 0x1d, 0x6e, 0x38, 0xe0, 0x41, 0x81, 0x8f, 0x21, 0xb6, 0x0a, 0x1f, 0xd0, 0x7d,
	0x7e, 0xb2, 0x50, 0x23, 0x89, 0x8d, 0x83, 0xf5, 0x55, 0xd6, 0x1d, 0x76, 0x7e, 0x8e, 0xee, 0x46,
	0x01, 0xf7, 0x1a, 0xcd, 0x56, 0x69, 0x91, 0xb1, 0x28, 0x79, 0x4a, 0x58, 0x3c, 0xbf, 0x70, 0x67,
	0xe8, 0x6a, 0x95, 0x3e, 0xa7, 0x2f, 0xeb, 0x74, 0x3e, 0x71, 0x6f, 0xd1, 0xcd, 0x3a, 0x4f, 0x5e,
	0x59, 0x4d, 0xa3, 0x88, 0x15, 0x45, 0x1d, 0xb3, 0xb4, 0xa7, 0xa6, 0x8b, 0xdf, 0x09, 0xf2, 0x5b,
	0xb5, 0xc3, 0xff, 0xcf, 0xb6, 0xf0, 0x46, 0x1b, 0x67, 0xfd, 0x6e, 0xd9, 0xe4, 0x2d, 0x3e, 0x19,
	0xa4, 0xda, 0x36, 0x9d, 0xc4, 0x4a, 0xcb, 0x40, 0x8a, 0xce, 0x6c, 0x7e, 0x3e, 0xd8, 0x7e, 0x03,
	0x63, 0xf7, 0x7b, 0xb4, 0x9f, 0x4f, 0x67, 0xba, 0xa4, 0xf4, 0xcb, 0xf1, 0x96, 0x56, 0x46, 0x39,
	0x60, 0x1b, 0xfb, 0x54, 0x86, 0xd8, 0xb4, 0x84, 0xef, 0x33, 0x50, 0x51, 0x0e, 0xd5, 0x00, 0x54,
	0x65, 0x58, 0x59, 0xe0, 0xc7, 0xf1, 0xed, 0x5f, 0x42, 0x28, 0x07, 0x42, 0x06, 0x84, 0x90, 0x32,
	0x24, 0xc4, 0x42, 0xef, 0x97, 0x66, 0xba, 0x87, 0xbf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x45, 0x29,
	0x4d, 0xec, 0xdc, 0x01, 0x00, 0x00,
}
