// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.28.3
// source: enum/eapp/platform.proto

package eapp

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

// 平台
type Platform int32

const (
	Platform_unknown Platform = 0
	// 安卓
	Platform_android Platform = 1
	// 苹果
	Platform_ios Platform = 2
	// 网页
	Platform_web Platform = 3
)

// Enum value maps for Platform.
var (
	Platform_name = map[int32]string{
		0: "unknown",
		1: "android",
		2: "ios",
		3: "web",
	}
	Platform_value = map[string]int32{
		"unknown": 0,
		"android": 1,
		"ios":     2,
		"web":     3,
	}
)

func (x Platform) Enum() *Platform {
	p := new(Platform)
	*p = x
	return p
}

func (x Platform) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Platform) Descriptor() protoreflect.EnumDescriptor {
	return file_enum_eapp_platform_proto_enumTypes[0].Descriptor()
}

func (Platform) Type() protoreflect.EnumType {
	return &file_enum_eapp_platform_proto_enumTypes[0]
}

func (x Platform) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Platform.Descriptor instead.
func (Platform) EnumDescriptor() ([]byte, []int) {
	return file_enum_eapp_platform_proto_rawDescGZIP(), []int{0}
}

var File_enum_eapp_platform_proto protoreflect.FileDescriptor

var file_enum_eapp_platform_proto_rawDesc = []byte{
	0x0a, 0x18, 0x65, 0x6e, 0x75, 0x6d, 0x2f, 0x65, 0x61, 0x70, 0x70, 0x2f, 0x70, 0x6c, 0x61, 0x74,
	0x66, 0x6f, 0x72, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x65, 0x61, 0x70, 0x70,
	0x2a, 0x36, 0x0a, 0x08, 0x70, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x12, 0x0b, 0x0a, 0x07,
	0x75, 0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x61, 0x6e, 0x64,
	0x72, 0x6f, 0x69, 0x64, 0x10, 0x01, 0x12, 0x07, 0x0a, 0x03, 0x69, 0x6f, 0x73, 0x10, 0x02, 0x12,
	0x07, 0x0a, 0x03, 0x77, 0x65, 0x62, 0x10, 0x03, 0x42, 0x32, 0x5a, 0x30, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x61, 0x69, 0x78, 0x68, 0x65, 0x2f, 0x67, 0x6f,
	0x6e, 0x65, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x65, 0x6e,
	0x75, 0x6d, 0x2f, 0x65, 0x61, 0x70, 0x70, 0x3b, 0x65, 0x61, 0x70, 0x70, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_enum_eapp_platform_proto_rawDescOnce sync.Once
	file_enum_eapp_platform_proto_rawDescData = file_enum_eapp_platform_proto_rawDesc
)

func file_enum_eapp_platform_proto_rawDescGZIP() []byte {
	file_enum_eapp_platform_proto_rawDescOnce.Do(func() {
		file_enum_eapp_platform_proto_rawDescData = protoimpl.X.CompressGZIP(file_enum_eapp_platform_proto_rawDescData)
	})
	return file_enum_eapp_platform_proto_rawDescData
}

var file_enum_eapp_platform_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_enum_eapp_platform_proto_goTypes = []any{
	(Platform)(0), // 0: eapp.platform
}
var file_enum_eapp_platform_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_enum_eapp_platform_proto_init() }
func file_enum_eapp_platform_proto_init() {
	if File_enum_eapp_platform_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_enum_eapp_platform_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_enum_eapp_platform_proto_goTypes,
		DependencyIndexes: file_enum_eapp_platform_proto_depIdxs,
		EnumInfos:         file_enum_eapp_platform_proto_enumTypes,
	}.Build()
	File_enum_eapp_platform_proto = out.File
	file_enum_eapp_platform_proto_rawDesc = nil
	file_enum_eapp_platform_proto_goTypes = nil
	file_enum_eapp_platform_proto_depIdxs = nil
}
