// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v5.28.3
// source: ecode/ecode.proto

package ecode

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

// 错误状态码
type ECode int32

const (
	ECode_Success     ECode = 0 // 成功
	ECode_Service     ECode = 1 // 服务错误
	ECode_Param       ECode = 2 // 参数错误
	ECode_TipMessage  ECode = 3 // 提示错误消息
	ECode_AuthInvalid ECode = 4 // 授权无效
	ECode_AuthExpire  ECode = 5 // 授权过期
)

// Enum value maps for ECode.
var (
	ECode_name = map[int32]string{
		0: "Success",
		1: "Service",
		2: "Param",
		3: "TipMessage",
		4: "AuthInvalid",
		5: "AuthExpire",
	}
	ECode_value = map[string]int32{
		"Success":     0,
		"Service":     1,
		"Param":       2,
		"TipMessage":  3,
		"AuthInvalid": 4,
		"AuthExpire":  5,
	}
)

func (x ECode) Enum() *ECode {
	p := new(ECode)
	*p = x
	return p
}

func (x ECode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ECode) Descriptor() protoreflect.EnumDescriptor {
	return file_ecode_ecode_proto_enumTypes[0].Descriptor()
}

func (ECode) Type() protoreflect.EnumType {
	return &file_ecode_ecode_proto_enumTypes[0]
}

func (x ECode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ECode.Descriptor instead.
func (ECode) EnumDescriptor() ([]byte, []int) {
	return file_ecode_ecode_proto_rawDescGZIP(), []int{0}
}

var File_ecode_ecode_proto protoreflect.FileDescriptor

var file_ecode_ecode_proto_rawDesc = []byte{
	0x0a, 0x11, 0x65, 0x63, 0x6f, 0x64, 0x65, 0x2f, 0x65, 0x63, 0x6f, 0x64, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x05, 0x65, 0x63, 0x6f, 0x64, 0x65, 0x2a, 0x5d, 0x0a, 0x05, 0x45, 0x43,
	0x6f, 0x64, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x10, 0x00,
	0x12, 0x0b, 0x0a, 0x07, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x10, 0x01, 0x12, 0x09, 0x0a,
	0x05, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x10, 0x02, 0x12, 0x0e, 0x0a, 0x0a, 0x54, 0x69, 0x70, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x10, 0x03, 0x12, 0x0f, 0x0a, 0x0b, 0x41, 0x75, 0x74, 0x68,
	0x49, 0x6e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x10, 0x04, 0x12, 0x0e, 0x0a, 0x0a, 0x41, 0x75, 0x74,
	0x68, 0x45, 0x78, 0x70, 0x69, 0x72, 0x65, 0x10, 0x05, 0x42, 0x2f, 0x5a, 0x2d, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x61, 0x69, 0x78, 0x68, 0x65, 0x2f, 0x67,
	0x6f, 0x6e, 0x65, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x65,
	0x63, 0x6f, 0x64, 0x65, 0x3b, 0x65, 0x63, 0x6f, 0x64, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_ecode_ecode_proto_rawDescOnce sync.Once
	file_ecode_ecode_proto_rawDescData = file_ecode_ecode_proto_rawDesc
)

func file_ecode_ecode_proto_rawDescGZIP() []byte {
	file_ecode_ecode_proto_rawDescOnce.Do(func() {
		file_ecode_ecode_proto_rawDescData = protoimpl.X.CompressGZIP(file_ecode_ecode_proto_rawDescData)
	})
	return file_ecode_ecode_proto_rawDescData
}

var file_ecode_ecode_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_ecode_ecode_proto_goTypes = []any{
	(ECode)(0), // 0: ecode.ECode
}
var file_ecode_ecode_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_ecode_ecode_proto_init() }
func file_ecode_ecode_proto_init() {
	if File_ecode_ecode_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_ecode_ecode_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ecode_ecode_proto_goTypes,
		DependencyIndexes: file_ecode_ecode_proto_depIdxs,
		EnumInfos:         file_ecode_ecode_proto_enumTypes,
	}.Build()
	File_ecode_ecode_proto = out.File
	file_ecode_ecode_proto_rawDesc = nil
	file_ecode_ecode_proto_goTypes = nil
	file_ecode_ecode_proto_depIdxs = nil
}
