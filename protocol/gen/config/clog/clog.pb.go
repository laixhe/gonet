// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.29.1
// source: config/clog/clog.proto

package clog

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// 开发模式
type RunType int32

const (
	RunType_console RunType = 0 // 终端
	RunType_file    RunType = 1 // 文件
)

// Enum value maps for RunType.
var (
	RunType_name = map[int32]string{
		0: "console",
		1: "file",
	}
	RunType_value = map[string]int32{
		"console": 0,
		"file":    1,
	}
)

func (x RunType) Enum() *RunType {
	p := new(RunType)
	*p = x
	return p
}

func (x RunType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (RunType) Descriptor() protoreflect.EnumDescriptor {
	return file_config_clog_clog_proto_enumTypes[0].Descriptor()
}

func (RunType) Type() protoreflect.EnumType {
	return &file_config_clog_clog_proto_enumTypes[0]
}

func (x RunType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use RunType.Descriptor instead.
func (RunType) EnumDescriptor() ([]byte, []int) {
	return file_config_clog_clog_proto_rawDescGZIP(), []int{0}
}

// 日志级别
type LevelType int32

const (
	LevelType_debug LevelType = 0
	LevelType_info  LevelType = 1
	LevelType_warn  LevelType = 2
	LevelType_error LevelType = 3
)

// Enum value maps for LevelType.
var (
	LevelType_name = map[int32]string{
		0: "debug",
		1: "info",
		2: "warn",
		3: "error",
	}
	LevelType_value = map[string]int32{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
	}
)

func (x LevelType) Enum() *LevelType {
	p := new(LevelType)
	*p = x
	return p
}

func (x LevelType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (LevelType) Descriptor() protoreflect.EnumDescriptor {
	return file_config_clog_clog_proto_enumTypes[1].Descriptor()
}

func (LevelType) Type() protoreflect.EnumType {
	return &file_config_clog_clog_proto_enumTypes[1]
}

func (x LevelType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use LevelType.Descriptor instead.
func (LevelType) EnumDescriptor() ([]byte, []int) {
	return file_config_clog_clog_proto_rawDescGZIP(), []int{1}
}

// 日志配置
type Log struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 日志模式 console file
	Run string `protobuf:"bytes,1,opt,name=run,proto3" json:"run,omitempty" mapstructure:"run"` // @gotags: mapstructure:"run"
	// 日志文件路径
	Path string `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty" mapstructure:"path"` // @gotags: mapstructure:"path"
	// 日志级别 debug  info  error
	Level string `protobuf:"bytes,3,opt,name=level,proto3" json:"level,omitempty" mapstructure:"level"` // @gotags: mapstructure:"level"
	// 每个日志文件保存大小 *M
	MaxSize int32 `protobuf:"varint,4,opt,name=max_size,json=maxSize,proto3" json:"max_size,omitempty" mapstructure:"max_size"` // @gotags: mapstructure:"max_size"
	// 保留 N 个备份
	MaxBackups int32 `protobuf:"varint,5,opt,name=max_backups,json=maxBackups,proto3" json:"max_backups,omitempty" mapstructure:"max_backups"` // @gotags: mapstructure:"max_backups"
	// 保留 N 天
	MaxAge int32 `protobuf:"varint,6,opt,name=max_age,json=maxAge,proto3" json:"max_age,omitempty" mapstructure:"max_age"` // @gotags: mapstructure:"max_age"
}

func (x *Log) Reset() {
	*x = Log{}
	mi := &file_config_clog_clog_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Log) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Log) ProtoMessage() {}

func (x *Log) ProtoReflect() protoreflect.Message {
	mi := &file_config_clog_clog_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Log.ProtoReflect.Descriptor instead.
func (*Log) Descriptor() ([]byte, []int) {
	return file_config_clog_clog_proto_rawDescGZIP(), []int{0}
}

func (x *Log) GetRun() string {
	if x != nil {
		return x.Run
	}
	return ""
}

func (x *Log) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *Log) GetLevel() string {
	if x != nil {
		return x.Level
	}
	return ""
}

func (x *Log) GetMaxSize() int32 {
	if x != nil {
		return x.MaxSize
	}
	return 0
}

func (x *Log) GetMaxBackups() int32 {
	if x != nil {
		return x.MaxBackups
	}
	return 0
}

func (x *Log) GetMaxAge() int32 {
	if x != nil {
		return x.MaxAge
	}
	return 0
}

var File_config_clog_clog_proto protoreflect.FileDescriptor

var file_config_clog_clog_proto_rawDesc = []byte{
	0x0a, 0x16, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x63, 0x6c, 0x6f, 0x67, 0x2f, 0x63, 0x6c,
	0x6f, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x63, 0x6c, 0x6f, 0x67, 0x22, 0x96,
	0x01, 0x0a, 0x03, 0x4c, 0x6f, 0x67, 0x12, 0x10, 0x0a, 0x03, 0x72, 0x75, 0x6e, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x72, 0x75, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x14, 0x0a, 0x05,
	0x6c, 0x65, 0x76, 0x65, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6c, 0x65, 0x76,
	0x65, 0x6c, 0x12, 0x19, 0x0a, 0x08, 0x6d, 0x61, 0x78, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x6d, 0x61, 0x78, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x1f, 0x0a,
	0x0b, 0x6d, 0x61, 0x78, 0x5f, 0x62, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x73, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x0a, 0x6d, 0x61, 0x78, 0x42, 0x61, 0x63, 0x6b, 0x75, 0x70, 0x73, 0x12, 0x17,
	0x0a, 0x07, 0x6d, 0x61, 0x78, 0x5f, 0x61, 0x67, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x06, 0x6d, 0x61, 0x78, 0x41, 0x67, 0x65, 0x2a, 0x20, 0x0a, 0x07, 0x52, 0x75, 0x6e, 0x54, 0x79,
	0x70, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x73, 0x6f, 0x6c, 0x65, 0x10, 0x00, 0x12,
	0x08, 0x0a, 0x04, 0x66, 0x69, 0x6c, 0x65, 0x10, 0x01, 0x2a, 0x35, 0x0a, 0x09, 0x4c, 0x65, 0x76,
	0x65, 0x6c, 0x54, 0x79, 0x70, 0x65, 0x12, 0x09, 0x0a, 0x05, 0x64, 0x65, 0x62, 0x75, 0x67, 0x10,
	0x00, 0x12, 0x08, 0x0a, 0x04, 0x69, 0x6e, 0x66, 0x6f, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x77,
	0x61, 0x72, 0x6e, 0x10, 0x02, 0x12, 0x09, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x10, 0x03,
	0x42, 0x37, 0x5a, 0x35, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c,
	0x61, 0x69, 0x78, 0x68, 0x65, 0x2f, 0x67, 0x6f, 0x6e, 0x65, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x6f, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f,
	0x63, 0x6c, 0x6f, 0x67, 0x3b, 0x63, 0x6c, 0x6f, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_config_clog_clog_proto_rawDescOnce sync.Once
	file_config_clog_clog_proto_rawDescData = file_config_clog_clog_proto_rawDesc
)

func file_config_clog_clog_proto_rawDescGZIP() []byte {
	file_config_clog_clog_proto_rawDescOnce.Do(func() {
		file_config_clog_clog_proto_rawDescData = protoimpl.X.CompressGZIP(file_config_clog_clog_proto_rawDescData)
	})
	return file_config_clog_clog_proto_rawDescData
}

var file_config_clog_clog_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_config_clog_clog_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_config_clog_clog_proto_goTypes = []any{
	(RunType)(0),   // 0: clog.RunType
	(LevelType)(0), // 1: clog.LevelType
	(*Log)(nil),    // 2: clog.Log
}
var file_config_clog_clog_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_config_clog_clog_proto_init() }
func file_config_clog_clog_proto_init() {
	if File_config_clog_clog_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_config_clog_clog_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_config_clog_clog_proto_goTypes,
		DependencyIndexes: file_config_clog_clog_proto_depIdxs,
		EnumInfos:         file_config_clog_clog_proto_enumTypes,
		MessageInfos:      file_config_clog_clog_proto_msgTypes,
	}.Build()
	File_config_clog_clog_proto = out.File
	file_config_clog_clog_proto_rawDesc = nil
	file_config_clog_clog_proto_goTypes = nil
	file_config_clog_clog_proto_depIdxs = nil
}
