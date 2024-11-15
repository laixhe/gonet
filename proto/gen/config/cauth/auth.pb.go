// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.28.3
// source: config/cauth/auth.proto

package cauth

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

// 鉴权配置
type Jwt struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// jwt secret key
	SecretKey string `protobuf:"bytes,1,opt,name=secret_key,json=secretKey,proto3" json:"secret_key,omitempty" mapstructure:"secret_key"` // @gotags: mapstructure:"secret_key"
	// 过期时长(单位秒)
	ExpireTime int64 `protobuf:"varint,2,opt,name=expire_time,json=expireTime,proto3" json:"expire_time,omitempty" mapstructure:"expire_time"` // @gotags: mapstructure:"expire_time"
}

func (x *Jwt) Reset() {
	*x = Jwt{}
	mi := &file_config_cauth_auth_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Jwt) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Jwt) ProtoMessage() {}

func (x *Jwt) ProtoReflect() protoreflect.Message {
	mi := &file_config_cauth_auth_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Jwt.ProtoReflect.Descriptor instead.
func (*Jwt) Descriptor() ([]byte, []int) {
	return file_config_cauth_auth_proto_rawDescGZIP(), []int{0}
}

func (x *Jwt) GetSecretKey() string {
	if x != nil {
		return x.SecretKey
	}
	return ""
}

func (x *Jwt) GetExpireTime() int64 {
	if x != nil {
		return x.ExpireTime
	}
	return 0
}

var File_config_cauth_auth_proto protoreflect.FileDescriptor

var file_config_cauth_auth_proto_rawDesc = []byte{
	0x0a, 0x17, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x63, 0x61, 0x75, 0x74, 0x68, 0x2f, 0x61,
	0x75, 0x74, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x63, 0x61, 0x75, 0x74, 0x68,
	0x22, 0x45, 0x0a, 0x03, 0x4a, 0x77, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x63, 0x72, 0x65,
	0x74, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x65, 0x63,
	0x72, 0x65, 0x74, 0x4b, 0x65, 0x79, 0x12, 0x1f, 0x0a, 0x0b, 0x65, 0x78, 0x70, 0x69, 0x72, 0x65,
	0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x65, 0x78, 0x70,
	0x69, 0x72, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x42, 0x36, 0x5a, 0x34, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x61, 0x69, 0x78, 0x68, 0x65, 0x2f, 0x67, 0x6f, 0x6e,
	0x65, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x63, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x2f, 0x63, 0x61, 0x75, 0x74, 0x68, 0x3b, 0x63, 0x61, 0x75, 0x74, 0x68, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_config_cauth_auth_proto_rawDescOnce sync.Once
	file_config_cauth_auth_proto_rawDescData = file_config_cauth_auth_proto_rawDesc
)

func file_config_cauth_auth_proto_rawDescGZIP() []byte {
	file_config_cauth_auth_proto_rawDescOnce.Do(func() {
		file_config_cauth_auth_proto_rawDescData = protoimpl.X.CompressGZIP(file_config_cauth_auth_proto_rawDescData)
	})
	return file_config_cauth_auth_proto_rawDescData
}

var file_config_cauth_auth_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_config_cauth_auth_proto_goTypes = []any{
	(*Jwt)(nil), // 0: cauth.Jwt
}
var file_config_cauth_auth_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_config_cauth_auth_proto_init() }
func file_config_cauth_auth_proto_init() {
	if File_config_cauth_auth_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_config_cauth_auth_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_config_cauth_auth_proto_goTypes,
		DependencyIndexes: file_config_cauth_auth_proto_depIdxs,
		MessageInfos:      file_config_cauth_auth_proto_msgTypes,
	}.Build()
	File_config_cauth_auth_proto = out.File
	file_config_cauth_auth_proto_rawDesc = nil
	file_config_cauth_auth_proto_goTypes = nil
	file_config_cauth_auth_proto_depIdxs = nil
}
