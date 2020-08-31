// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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
// source: google/ads/googleads/v3/resources/dynamic_search_ads_search_term_view.proto

package resources

import (
	reflect "reflect"
	sync "sync"

	proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
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

// A dynamic search ads search term view.
type DynamicSearchAdsSearchTermView struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Output only. The resource name of the dynamic search ads search term view.
	// Dynamic search ads search term view resource names have the form:
	//
	// `customers/{customer_id}/dynamicSearchAdsSearchTermViews/{ad_group_id}~{search_term_fp}~{headline_fp}~{landing_page_fp}~{page_url_fp}`
	ResourceName string `protobuf:"bytes,1,opt,name=resource_name,json=resourceName,proto3" json:"resource_name,omitempty"`
	// Output only. Search term
	//
	// This field is read-only.
	SearchTerm *wrapperspb.StringValue `protobuf:"bytes,2,opt,name=search_term,json=searchTerm,proto3" json:"search_term,omitempty"`
	// Output only. The dynamically generated headline of the Dynamic Search Ad.
	//
	// This field is read-only.
	Headline *wrapperspb.StringValue `protobuf:"bytes,3,opt,name=headline,proto3" json:"headline,omitempty"`
	// Output only. The dynamically selected landing page URL of the impression.
	//
	// This field is read-only.
	LandingPage *wrapperspb.StringValue `protobuf:"bytes,4,opt,name=landing_page,json=landingPage,proto3" json:"landing_page,omitempty"`
	// Output only. The URL of page feed item served for the impression.
	//
	// This field is read-only.
	PageUrl *wrapperspb.StringValue `protobuf:"bytes,5,opt,name=page_url,json=pageUrl,proto3" json:"page_url,omitempty"`
	// Output only. True if query matches a negative keyword.
	//
	// This field is read-only.
	HasNegativeKeyword *wrapperspb.BoolValue `protobuf:"bytes,6,opt,name=has_negative_keyword,json=hasNegativeKeyword,proto3" json:"has_negative_keyword,omitempty"`
	// Output only. True if query is added to targeted keywords.
	//
	// This field is read-only.
	HasMatchingKeyword *wrapperspb.BoolValue `protobuf:"bytes,7,opt,name=has_matching_keyword,json=hasMatchingKeyword,proto3" json:"has_matching_keyword,omitempty"`
	// Output only. True if query matches a negative url.
	//
	// This field is read-only.
	HasNegativeUrl *wrapperspb.BoolValue `protobuf:"bytes,8,opt,name=has_negative_url,json=hasNegativeUrl,proto3" json:"has_negative_url,omitempty"`
}

func (x *DynamicSearchAdsSearchTermView) Reset() {
	*x = DynamicSearchAdsSearchTermView{}
	if protoimpl.UnsafeEnabled {
		mi := &file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DynamicSearchAdsSearchTermView) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DynamicSearchAdsSearchTermView) ProtoMessage() {}

func (x *DynamicSearchAdsSearchTermView) ProtoReflect() protoreflect.Message {
	mi := &file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DynamicSearchAdsSearchTermView.ProtoReflect.Descriptor instead.
func (*DynamicSearchAdsSearchTermView) Descriptor() ([]byte, []int) {
	return file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_rawDescGZIP(), []int{0}
}

func (x *DynamicSearchAdsSearchTermView) GetResourceName() string {
	if x != nil {
		return x.ResourceName
	}
	return ""
}

func (x *DynamicSearchAdsSearchTermView) GetSearchTerm() *wrapperspb.StringValue {
	if x != nil {
		return x.SearchTerm
	}
	return nil
}

func (x *DynamicSearchAdsSearchTermView) GetHeadline() *wrapperspb.StringValue {
	if x != nil {
		return x.Headline
	}
	return nil
}

func (x *DynamicSearchAdsSearchTermView) GetLandingPage() *wrapperspb.StringValue {
	if x != nil {
		return x.LandingPage
	}
	return nil
}

func (x *DynamicSearchAdsSearchTermView) GetPageUrl() *wrapperspb.StringValue {
	if x != nil {
		return x.PageUrl
	}
	return nil
}

func (x *DynamicSearchAdsSearchTermView) GetHasNegativeKeyword() *wrapperspb.BoolValue {
	if x != nil {
		return x.HasNegativeKeyword
	}
	return nil
}

func (x *DynamicSearchAdsSearchTermView) GetHasMatchingKeyword() *wrapperspb.BoolValue {
	if x != nil {
		return x.HasMatchingKeyword
	}
	return nil
}

func (x *DynamicSearchAdsSearchTermView) GetHasNegativeUrl() *wrapperspb.BoolValue {
	if x != nil {
		return x.HasNegativeUrl
	}
	return nil
}

var File_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto protoreflect.FileDescriptor

var file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_rawDesc = []byte{
	0x0a, 0x4b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x64, 0x73, 0x2f, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x61, 0x64, 0x73, 0x2f, 0x76, 0x33, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x73, 0x2f, 0x64, 0x79, 0x6e, 0x61, 0x6d, 0x69, 0x63, 0x5f, 0x73, 0x65, 0x61, 0x72,
	0x63, 0x68, 0x5f, 0x61, 0x64, 0x73, 0x5f, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x5f, 0x74, 0x65,
	0x72, 0x6d, 0x5f, 0x76, 0x69, 0x65, 0x77, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x21, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x61, 0x64, 0x73, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x61, 0x64, 0x73, 0x2e, 0x76, 0x33, 0x2e, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73,
	0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x66, 0x69, 0x65,
	0x6c, 0x64, 0x5f, 0x62, 0x65, 0x68, 0x61, 0x76, 0x69, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x72, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x77, 0x72,
	0x61, 0x70, 0x70, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x9a, 0x06, 0x0a, 0x1e, 0x44,
	0x79, 0x6e, 0x61, 0x6d, 0x69, 0x63, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x41, 0x64, 0x73, 0x53,
	0x65, 0x61, 0x72, 0x63, 0x68, 0x54, 0x65, 0x72, 0x6d, 0x56, 0x69, 0x65, 0x77, 0x12, 0x64, 0x0a,
	0x0d, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x3f, 0xe0, 0x41, 0x03, 0xfa, 0x41, 0x39, 0x0a, 0x37, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x61, 0x64, 0x73, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x61, 0x70,
	0x69, 0x73, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x44, 0x79, 0x6e, 0x61, 0x6d, 0x69, 0x63, 0x53, 0x65,
	0x61, 0x72, 0x63, 0x68, 0x41, 0x64, 0x73, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x54, 0x65, 0x72,
	0x6d, 0x56, 0x69, 0x65, 0x77, 0x52, 0x0c, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4e,
	0x61, 0x6d, 0x65, 0x12, 0x42, 0x0a, 0x0b, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x5f, 0x74, 0x65,
	0x72, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e,
	0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x0a, 0x73, 0x65, 0x61,
	0x72, 0x63, 0x68, 0x54, 0x65, 0x72, 0x6d, 0x12, 0x3d, 0x0a, 0x08, 0x68, 0x65, 0x61, 0x64, 0x6c,
	0x69, 0x6e, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x69,
	0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x08, 0x68, 0x65,
	0x61, 0x64, 0x6c, 0x69, 0x6e, 0x65, 0x12, 0x44, 0x0a, 0x0c, 0x6c, 0x61, 0x6e, 0x64, 0x69, 0x6e,
	0x67, 0x5f, 0x70, 0x61, 0x67, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52,
	0x0b, 0x6c, 0x61, 0x6e, 0x64, 0x69, 0x6e, 0x67, 0x50, 0x61, 0x67, 0x65, 0x12, 0x3c, 0x0a, 0x08,
	0x70, 0x61, 0x67, 0x65, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x42, 0x03, 0xe0, 0x41,
	0x03, 0x52, 0x07, 0x70, 0x61, 0x67, 0x65, 0x55, 0x72, 0x6c, 0x12, 0x51, 0x0a, 0x14, 0x68, 0x61,
	0x73, 0x5f, 0x6e, 0x65, 0x67, 0x61, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x77, 0x6f,
	0x72, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x42, 0x6f, 0x6f, 0x6c, 0x56,
	0x61, 0x6c, 0x75, 0x65, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x12, 0x68, 0x61, 0x73, 0x4e, 0x65,
	0x67, 0x61, 0x74, 0x69, 0x76, 0x65, 0x4b, 0x65, 0x79, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x51, 0x0a,
	0x14, 0x68, 0x61, 0x73, 0x5f, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x69, 0x6e, 0x67, 0x5f, 0x6b, 0x65,
	0x79, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x42, 0x6f,
	0x6f, 0x6c, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x12, 0x68, 0x61,
	0x73, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x69, 0x6e, 0x67, 0x4b, 0x65, 0x79, 0x77, 0x6f, 0x72, 0x64,
	0x12, 0x49, 0x0a, 0x10, 0x68, 0x61, 0x73, 0x5f, 0x6e, 0x65, 0x67, 0x61, 0x74, 0x69, 0x76, 0x65,
	0x5f, 0x75, 0x72, 0x6c, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x42, 0x6f, 0x6f,
	0x6c, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x42, 0x03, 0xe0, 0x41, 0x03, 0x52, 0x0e, 0x68, 0x61, 0x73,
	0x4e, 0x65, 0x67, 0x61, 0x74, 0x69, 0x76, 0x65, 0x55, 0x72, 0x6c, 0x3a, 0x99, 0x01, 0xea, 0x41,
	0x95, 0x01, 0x0a, 0x37, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x61, 0x64, 0x73, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x61, 0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x44, 0x79, 0x6e,
	0x61, 0x6d, 0x69, 0x63, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x41, 0x64, 0x73, 0x53, 0x65, 0x61,
	0x72, 0x63, 0x68, 0x54, 0x65, 0x72, 0x6d, 0x56, 0x69, 0x65, 0x77, 0x12, 0x5a, 0x63, 0x75, 0x73,
	0x74, 0x6f, 0x6d, 0x65, 0x72, 0x73, 0x2f, 0x7b, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x65, 0x72,
	0x7d, 0x2f, 0x64, 0x79, 0x6e, 0x61, 0x6d, 0x69, 0x63, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x41,
	0x64, 0x73, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x54, 0x65, 0x72, 0x6d, 0x56, 0x69, 0x65, 0x77,
	0x73, 0x2f, 0x7b, 0x64, 0x79, 0x6e, 0x61, 0x6d, 0x69, 0x63, 0x5f, 0x73, 0x65, 0x61, 0x72, 0x63,
	0x68, 0x5f, 0x61, 0x64, 0x73, 0x5f, 0x73, 0x65, 0x61, 0x72, 0x63, 0x68, 0x5f, 0x74, 0x65, 0x72,
	0x6d, 0x5f, 0x76, 0x69, 0x65, 0x77, 0x7d, 0x42, 0x90, 0x02, 0x0a, 0x25, 0x63, 0x6f, 0x6d, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x61, 0x64, 0x73, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x61, 0x64, 0x73, 0x2e, 0x76, 0x33, 0x2e, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x73, 0x42, 0x23, 0x44, 0x79, 0x6e, 0x61, 0x6d, 0x69, 0x63, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68,
	0x41, 0x64, 0x73, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x54, 0x65, 0x72, 0x6d, 0x56, 0x69, 0x65,
	0x77, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x4a, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x67, 0x6f, 0x6c, 0x61, 0x6e, 0x67, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x67, 0x65, 0x6e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x61, 0x70, 0x69, 0x73, 0x2f,
	0x61, 0x64, 0x73, 0x2f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x61, 0x64, 0x73, 0x2f, 0x76, 0x33,
	0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x3b, 0x72, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x73, 0xa2, 0x02, 0x03, 0x47, 0x41, 0x41, 0xaa, 0x02, 0x21, 0x47, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x41, 0x64, 0x73, 0x2e, 0x47, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x41, 0x64,
	0x73, 0x2e, 0x56, 0x33, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0xca, 0x02,
	0x21, 0x47, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x5c, 0x41, 0x64, 0x73, 0x5c, 0x47, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x41, 0x64, 0x73, 0x5c, 0x56, 0x33, 0x5c, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x73, 0xea, 0x02, 0x25, 0x47, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x3a, 0x3a, 0x41, 0x64, 0x73,
	0x3a, 0x3a, 0x47, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x41, 0x64, 0x73, 0x3a, 0x3a, 0x56, 0x33, 0x3a,
	0x3a, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_rawDescOnce sync.Once
	file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_rawDescData = file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_rawDesc
)

func file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_rawDescGZIP() []byte {
	file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_rawDescOnce.Do(func() {
		file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_rawDescData = protoimpl.X.CompressGZIP(file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_rawDescData)
	})
	return file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_rawDescData
}

var file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_goTypes = []interface{}{
	(*DynamicSearchAdsSearchTermView)(nil), // 0: google.ads.googleads.v3.resources.DynamicSearchAdsSearchTermView
	(*wrapperspb.StringValue)(nil),         // 1: google.protobuf.StringValue
	(*wrapperspb.BoolValue)(nil),           // 2: google.protobuf.BoolValue
}
var file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_depIdxs = []int32{
	1, // 0: google.ads.googleads.v3.resources.DynamicSearchAdsSearchTermView.search_term:type_name -> google.protobuf.StringValue
	1, // 1: google.ads.googleads.v3.resources.DynamicSearchAdsSearchTermView.headline:type_name -> google.protobuf.StringValue
	1, // 2: google.ads.googleads.v3.resources.DynamicSearchAdsSearchTermView.landing_page:type_name -> google.protobuf.StringValue
	1, // 3: google.ads.googleads.v3.resources.DynamicSearchAdsSearchTermView.page_url:type_name -> google.protobuf.StringValue
	2, // 4: google.ads.googleads.v3.resources.DynamicSearchAdsSearchTermView.has_negative_keyword:type_name -> google.protobuf.BoolValue
	2, // 5: google.ads.googleads.v3.resources.DynamicSearchAdsSearchTermView.has_matching_keyword:type_name -> google.protobuf.BoolValue
	2, // 6: google.ads.googleads.v3.resources.DynamicSearchAdsSearchTermView.has_negative_url:type_name -> google.protobuf.BoolValue
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_init() }
func file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_init() {
	if File_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DynamicSearchAdsSearchTermView); i {
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
			RawDescriptor: file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_goTypes,
		DependencyIndexes: file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_depIdxs,
		MessageInfos:      file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_msgTypes,
	}.Build()
	File_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto = out.File
	file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_rawDesc = nil
	file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_goTypes = nil
	file_google_ads_googleads_v3_resources_dynamic_search_ads_search_term_view_proto_depIdxs = nil
}
