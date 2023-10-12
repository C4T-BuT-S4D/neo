// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: logs/api.proto

package logs

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type LogLine struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Exploit string `protobuf:"bytes,1,opt,name=exploit,proto3" json:"exploit,omitempty"`
	Version int64  `protobuf:"varint,2,opt,name=version,proto3" json:"version,omitempty"`
	Message string `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
	Level   string `protobuf:"bytes,4,opt,name=level,proto3" json:"level,omitempty"`
	Team    string `protobuf:"bytes,5,opt,name=team,proto3" json:"team,omitempty"`
}

func (x *LogLine) Reset() {
	*x = LogLine{}
	if protoimpl.UnsafeEnabled {
		mi := &file_logs_api_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LogLine) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogLine) ProtoMessage() {}

func (x *LogLine) ProtoReflect() protoreflect.Message {
	mi := &file_logs_api_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogLine.ProtoReflect.Descriptor instead.
func (*LogLine) Descriptor() ([]byte, []int) {
	return file_logs_api_proto_rawDescGZIP(), []int{0}
}

func (x *LogLine) GetExploit() string {
	if x != nil {
		return x.Exploit
	}
	return ""
}

func (x *LogLine) GetVersion() int64 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *LogLine) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *LogLine) GetLevel() string {
	if x != nil {
		return x.Level
	}
	return ""
}

func (x *LogLine) GetTeam() string {
	if x != nil {
		return x.Team
	}
	return ""
}

type AddLogLinesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Lines []*LogLine `protobuf:"bytes,1,rep,name=lines,proto3" json:"lines,omitempty"`
}

func (x *AddLogLinesRequest) Reset() {
	*x = AddLogLinesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_logs_api_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddLogLinesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddLogLinesRequest) ProtoMessage() {}

func (x *AddLogLinesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_logs_api_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddLogLinesRequest.ProtoReflect.Descriptor instead.
func (*AddLogLinesRequest) Descriptor() ([]byte, []int) {
	return file_logs_api_proto_rawDescGZIP(), []int{1}
}

func (x *AddLogLinesRequest) GetLines() []*LogLine {
	if x != nil {
		return x.Lines
	}
	return nil
}

type SearchLogLinesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Exploit string `protobuf:"bytes,1,opt,name=exploit,proto3" json:"exploit,omitempty"`
	Version int64  `protobuf:"varint,2,opt,name=version,proto3" json:"version,omitempty"`
}

func (x *SearchLogLinesRequest) Reset() {
	*x = SearchLogLinesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_logs_api_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SearchLogLinesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchLogLinesRequest) ProtoMessage() {}

func (x *SearchLogLinesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_logs_api_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SearchLogLinesRequest.ProtoReflect.Descriptor instead.
func (*SearchLogLinesRequest) Descriptor() ([]byte, []int) {
	return file_logs_api_proto_rawDescGZIP(), []int{2}
}

func (x *SearchLogLinesRequest) GetExploit() string {
	if x != nil {
		return x.Exploit
	}
	return ""
}

func (x *SearchLogLinesRequest) GetVersion() int64 {
	if x != nil {
		return x.Version
	}
	return 0
}

type SearchLogLinesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Lines []*LogLine `protobuf:"bytes,1,rep,name=lines,proto3" json:"lines,omitempty"`
}

func (x *SearchLogLinesResponse) Reset() {
	*x = SearchLogLinesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_logs_api_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SearchLogLinesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchLogLinesResponse) ProtoMessage() {}

func (x *SearchLogLinesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_logs_api_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SearchLogLinesResponse.ProtoReflect.Descriptor instead.
func (*SearchLogLinesResponse) Descriptor() ([]byte, []int) {
	return file_logs_api_proto_rawDescGZIP(), []int{3}
}

func (x *SearchLogLinesResponse) GetLines() []*LogLine {
	if x != nil {
		return x.Lines
	}
	return nil
}

var File_logs_api_proto protoreflect.FileDescriptor

var file_logs_api_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x6c, 0x6f, 0x67, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x04, 0x6c, 0x6f, 0x67, 0x73, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x81, 0x01, 0x0a, 0x07, 0x4c, 0x6f, 0x67, 0x4c, 0x69, 0x6e, 0x65, 0x12,
	0x18, 0x0a, 0x07, 0x65, 0x78, 0x70, 0x6c, 0x6f, 0x69, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x65, 0x78, 0x70, 0x6c, 0x6f, 0x69, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x14, 0x0a,
	0x05, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6c, 0x65,
	0x76, 0x65, 0x6c, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x65, 0x61, 0x6d, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x74, 0x65, 0x61, 0x6d, 0x22, 0x39, 0x0a, 0x12, 0x41, 0x64, 0x64, 0x4c, 0x6f,
	0x67, 0x4c, 0x69, 0x6e, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x23, 0x0a,
	0x05, 0x6c, 0x69, 0x6e, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x6c,
	0x6f, 0x67, 0x73, 0x2e, 0x4c, 0x6f, 0x67, 0x4c, 0x69, 0x6e, 0x65, 0x52, 0x05, 0x6c, 0x69, 0x6e,
	0x65, 0x73, 0x22, 0x4b, 0x0a, 0x15, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x4c, 0x6f, 0x67, 0x4c,
	0x69, 0x6e, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x65,
	0x78, 0x70, 0x6c, 0x6f, 0x69, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x65, 0x78,
	0x70, 0x6c, 0x6f, 0x69, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22,
	0x3d, 0x0a, 0x16, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x4c, 0x6f, 0x67, 0x4c, 0x69, 0x6e, 0x65,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x23, 0x0a, 0x05, 0x6c, 0x69, 0x6e,
	0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x6c, 0x6f, 0x67, 0x73, 0x2e,
	0x4c, 0x6f, 0x67, 0x4c, 0x69, 0x6e, 0x65, 0x52, 0x05, 0x6c, 0x69, 0x6e, 0x65, 0x73, 0x32, 0x9d,
	0x01, 0x0a, 0x07, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x41, 0x0a, 0x0b, 0x41, 0x64,
	0x64, 0x4c, 0x6f, 0x67, 0x4c, 0x69, 0x6e, 0x65, 0x73, 0x12, 0x18, 0x2e, 0x6c, 0x6f, 0x67, 0x73,
	0x2e, 0x41, 0x64, 0x64, 0x4c, 0x6f, 0x67, 0x4c, 0x69, 0x6e, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x4f, 0x0a,
	0x0e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x4c, 0x6f, 0x67, 0x4c, 0x69, 0x6e, 0x65, 0x73, 0x12,
	0x1b, 0x2e, 0x6c, 0x6f, 0x67, 0x73, 0x2e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x4c, 0x6f, 0x67,
	0x4c, 0x69, 0x6e, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x6c,
	0x6f, 0x67, 0x73, 0x2e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x4c, 0x6f, 0x67, 0x4c, 0x69, 0x6e,
	0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x42, 0x71,
	0x0a, 0x08, 0x63, 0x6f, 0x6d, 0x2e, 0x6c, 0x6f, 0x67, 0x73, 0x42, 0x08, 0x41, 0x70, 0x69, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x2b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x63, 0x34, 0x74, 0x2d, 0x62, 0x75, 0x74, 0x2d, 0x73, 0x34, 0x64, 0x2f, 0x6e,
	0x65, 0x6f, 0x2f, 0x76, 0x32, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x2f, 0x6c,
	0x6f, 0x67, 0x73, 0xa2, 0x02, 0x03, 0x4c, 0x58, 0x58, 0xaa, 0x02, 0x04, 0x4c, 0x6f, 0x67, 0x73,
	0xca, 0x02, 0x04, 0x4c, 0x6f, 0x67, 0x73, 0xe2, 0x02, 0x10, 0x4c, 0x6f, 0x67, 0x73, 0x5c, 0x47,
	0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x04, 0x4c, 0x6f, 0x67,
	0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_logs_api_proto_rawDescOnce sync.Once
	file_logs_api_proto_rawDescData = file_logs_api_proto_rawDesc
)

func file_logs_api_proto_rawDescGZIP() []byte {
	file_logs_api_proto_rawDescOnce.Do(func() {
		file_logs_api_proto_rawDescData = protoimpl.X.CompressGZIP(file_logs_api_proto_rawDescData)
	})
	return file_logs_api_proto_rawDescData
}

var file_logs_api_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_logs_api_proto_goTypes = []interface{}{
	(*LogLine)(nil),                // 0: logs.LogLine
	(*AddLogLinesRequest)(nil),     // 1: logs.AddLogLinesRequest
	(*SearchLogLinesRequest)(nil),  // 2: logs.SearchLogLinesRequest
	(*SearchLogLinesResponse)(nil), // 3: logs.SearchLogLinesResponse
	(*emptypb.Empty)(nil),          // 4: google.protobuf.Empty
}
var file_logs_api_proto_depIdxs = []int32{
	0, // 0: logs.AddLogLinesRequest.lines:type_name -> logs.LogLine
	0, // 1: logs.SearchLogLinesResponse.lines:type_name -> logs.LogLine
	1, // 2: logs.Service.AddLogLines:input_type -> logs.AddLogLinesRequest
	2, // 3: logs.Service.SearchLogLines:input_type -> logs.SearchLogLinesRequest
	4, // 4: logs.Service.AddLogLines:output_type -> google.protobuf.Empty
	3, // 5: logs.Service.SearchLogLines:output_type -> logs.SearchLogLinesResponse
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_logs_api_proto_init() }
func file_logs_api_proto_init() {
	if File_logs_api_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_logs_api_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LogLine); i {
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
		file_logs_api_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddLogLinesRequest); i {
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
		file_logs_api_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SearchLogLinesRequest); i {
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
		file_logs_api_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SearchLogLinesResponse); i {
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
			RawDescriptor: file_logs_api_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_logs_api_proto_goTypes,
		DependencyIndexes: file_logs_api_proto_depIdxs,
		MessageInfos:      file_logs_api_proto_msgTypes,
	}.Build()
	File_logs_api_proto = out.File
	file_logs_api_proto_rawDesc = nil
	file_logs_api_proto_goTypes = nil
	file_logs_api_proto_depIdxs = nil
}
