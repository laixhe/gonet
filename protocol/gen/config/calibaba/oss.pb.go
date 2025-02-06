// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.4
// 	protoc        v5.29.3
// source: config/calibaba/oss.proto

package calibaba

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// 阿里云对象存储配置
type Oss struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// 标识用户ID
	AccessKeyId string `protobuf:"bytes,1,opt,name=access_key_id,json=accessKeyId,proto3" json:"access_key_id,omitempty" mapstructure:"access_key_id" toml:"access_key_id" yaml:"access_key_id"` // @gotags: mapstructure:"access_key_id" toml:"access_key_id" yaml:"access_key_id"
	// 密钥
	AccessKeySecret string `protobuf:"bytes,2,opt,name=access_key_secret,json=accessKeySecret,proto3" json:"access_key_secret,omitempty" mapstructure:"access_key_secret" toml:"access_key_secret" yaml:"access_key_secret"` // @gotags: mapstructure:"access_key_secret" toml:"access_key_secret" yaml:"access_key_secret"
	// 地域(如: cn-shenzhen)
	Region string `protobuf:"bytes,3,opt,name=region,proto3" json:"region,omitempty" mapstructure:"region" toml:"region" yaml:"region"` // @gotags: mapstructure:"region" toml:"region" yaml:"region"
	// 访问域名(如: https://oss-cn-shenzhen.aliyuncs.com)
	Endpoint string `protobuf:"bytes,4,opt,name=endpoint,proto3" json:"endpoint,omitempty" mapstructure:"endpoint" toml:"endpoint" yaml:"endpoint"` // @gotags: mapstructure:"endpoint" toml:"endpoint" yaml:"endpoint"
	// 桶名(存储空间如: test)
	Bucket        string `protobuf:"bytes,5,opt,name=bucket,proto3" json:"bucket,omitempty" mapstructure:"bucket" toml:"bucket" yaml:"bucket"` // @gotags: mapstructure:"bucket" toml:"bucket" yaml:"bucket"
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Oss) Reset() {
	*x = Oss{}
	mi := &file_config_calibaba_oss_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Oss) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Oss) ProtoMessage() {}

func (x *Oss) ProtoReflect() protoreflect.Message {
	mi := &file_config_calibaba_oss_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Oss.ProtoReflect.Descriptor instead.
func (*Oss) Descriptor() ([]byte, []int) {
	return file_config_calibaba_oss_proto_rawDescGZIP(), []int{0}
}

func (x *Oss) GetAccessKeyId() string {
	if x != nil {
		return x.AccessKeyId
	}
	return ""
}

func (x *Oss) GetAccessKeySecret() string {
	if x != nil {
		return x.AccessKeySecret
	}
	return ""
}

func (x *Oss) GetRegion() string {
	if x != nil {
		return x.Region
	}
	return ""
}

func (x *Oss) GetEndpoint() string {
	if x != nil {
		return x.Endpoint
	}
	return ""
}

func (x *Oss) GetBucket() string {
	if x != nil {
		return x.Bucket
	}
	return ""
}

var File_config_calibaba_oss_proto protoreflect.FileDescriptor

var file_config_calibaba_oss_proto_rawDesc = string([]byte{
	0x0a, 0x19, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x63, 0x61, 0x6c, 0x69, 0x62, 0x61, 0x62,
	0x61, 0x2f, 0x6f, 0x73, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x63, 0x61, 0x6c,
	0x69, 0x62, 0x61, 0x62, 0x61, 0x22, 0xa1, 0x01, 0x0a, 0x03, 0x4f, 0x73, 0x73, 0x12, 0x22, 0x0a,
	0x0d, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x6b, 0x65, 0x79, 0x5f, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x4b, 0x65, 0x79, 0x49,
	0x64, 0x12, 0x2a, 0x0a, 0x11, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x6b, 0x65, 0x79, 0x5f,
	0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x61, 0x63,
	0x63, 0x65, 0x73, 0x73, 0x4b, 0x65, 0x79, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x12, 0x16, 0x0a,
	0x06, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72,
	0x65, 0x67, 0x69, 0x6f, 0x6e, 0x12, 0x1a, 0x0a, 0x08, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x12, 0x16, 0x0a, 0x06, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x42, 0x3f, 0x5a, 0x3d, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x61, 0x69, 0x78, 0x68, 0x65, 0x2f, 0x67,
	0x6f, 0x6e, 0x65, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2f, 0x67, 0x65,
	0x6e, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x63, 0x61, 0x6c, 0x69, 0x62, 0x61, 0x62,
	0x61, 0x3b, 0x63, 0x61, 0x6c, 0x69, 0x62, 0x61, 0x62, 0x61, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
})

var (
	file_config_calibaba_oss_proto_rawDescOnce sync.Once
	file_config_calibaba_oss_proto_rawDescData []byte
)

func file_config_calibaba_oss_proto_rawDescGZIP() []byte {
	file_config_calibaba_oss_proto_rawDescOnce.Do(func() {
		file_config_calibaba_oss_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_config_calibaba_oss_proto_rawDesc), len(file_config_calibaba_oss_proto_rawDesc)))
	})
	return file_config_calibaba_oss_proto_rawDescData
}

var file_config_calibaba_oss_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_config_calibaba_oss_proto_goTypes = []any{
	(*Oss)(nil), // 0: calibaba.Oss
}
var file_config_calibaba_oss_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_config_calibaba_oss_proto_init() }
func file_config_calibaba_oss_proto_init() {
	if File_config_calibaba_oss_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_config_calibaba_oss_proto_rawDesc), len(file_config_calibaba_oss_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_config_calibaba_oss_proto_goTypes,
		DependencyIndexes: file_config_calibaba_oss_proto_depIdxs,
		MessageInfos:      file_config_calibaba_oss_proto_msgTypes,
	}.Build()
	File_config_calibaba_oss_proto = out.File
	file_config_calibaba_oss_proto_goTypes = nil
	file_config_calibaba_oss_proto_depIdxs = nil
}
