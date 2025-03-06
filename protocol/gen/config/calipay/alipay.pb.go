// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.30.2
// source: config/calipay/alipay.proto

package calipay

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

// 支付宝配置
type Alipay struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// 应用唯一标识
	AppId string `protobuf:"bytes,1,opt,name=app_id,json=appId,proto3" json:"app_id,omitempty" mapstructure:"app_id"` // @gotags: mapstructure:"app_id"
	// 应用私钥，支持PKCS1和PKCS8
	PrivateKey string `protobuf:"bytes,2,opt,name=private_key,json=privateKey,proto3" json:"private_key,omitempty" mapstructure:"private_key"` // @gotags: mapstructure:"private_key"
	// 是否是正式环境
	IsProduction bool `protobuf:"varint,3,opt,name=is_production,json=isProduction,proto3" json:"is_production,omitempty" mapstructure:"is_production"` // @gotags: mapstructure:"is_production"
	// 应用公钥证书路径
	AppCertPublicKeyFile string `protobuf:"bytes,4,opt,name=app_cert_public_key_file,json=appCertPublicKeyFile,proto3" json:"app_cert_public_key_file,omitempty" mapstructure:"app_cert_public_key_file"` // @gotags: mapstructure:"app_cert_public_key_file"
	// 支付宝根证书路径
	AlipayRootCertFile string `protobuf:"bytes,5,opt,name=alipay_root_cert_file,json=alipayRootCertFile,proto3" json:"alipay_root_cert_file,omitempty" mapstructure:"alipay_root_cert_file"` // @gotags: mapstructure:"alipay_root_cert_file"
	// 支付宝公钥证书路径
	AlipayCertPublicKeyFile string `protobuf:"bytes,6,opt,name=alipay_cert_public_key_file,json=alipayCertPublicKeyFile,proto3" json:"alipay_cert_public_key_file,omitempty" mapstructure:"alipay_cert_public_key_file"` // @gotags: mapstructure:"alipay_cert_public_key_file"
	// 同步回调
	ReturnUrl string `protobuf:"bytes,7,opt,name=return_url,json=returnUrl,proto3" json:"return_url,omitempty" mapstructure:"return_url"` // @gotags: mapstructure:"return_url"
	// 异步通知
	NotifyUrl     string `protobuf:"bytes,8,opt,name=notify_url,json=notifyUrl,proto3" json:"notify_url,omitempty" mapstructure:"notify_url"` // @gotags: mapstructure:"notify_url"
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Alipay) Reset() {
	*x = Alipay{}
	mi := &file_config_calipay_alipay_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Alipay) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Alipay) ProtoMessage() {}

func (x *Alipay) ProtoReflect() protoreflect.Message {
	mi := &file_config_calipay_alipay_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Alipay.ProtoReflect.Descriptor instead.
func (*Alipay) Descriptor() ([]byte, []int) {
	return file_config_calipay_alipay_proto_rawDescGZIP(), []int{0}
}

func (x *Alipay) GetAppId() string {
	if x != nil {
		return x.AppId
	}
	return ""
}

func (x *Alipay) GetPrivateKey() string {
	if x != nil {
		return x.PrivateKey
	}
	return ""
}

func (x *Alipay) GetIsProduction() bool {
	if x != nil {
		return x.IsProduction
	}
	return false
}

func (x *Alipay) GetAppCertPublicKeyFile() string {
	if x != nil {
		return x.AppCertPublicKeyFile
	}
	return ""
}

func (x *Alipay) GetAlipayRootCertFile() string {
	if x != nil {
		return x.AlipayRootCertFile
	}
	return ""
}

func (x *Alipay) GetAlipayCertPublicKeyFile() string {
	if x != nil {
		return x.AlipayCertPublicKeyFile
	}
	return ""
}

func (x *Alipay) GetReturnUrl() string {
	if x != nil {
		return x.ReturnUrl
	}
	return ""
}

func (x *Alipay) GetNotifyUrl() string {
	if x != nil {
		return x.NotifyUrl
	}
	return ""
}

var File_config_calipay_alipay_proto protoreflect.FileDescriptor

const file_config_calipay_alipay_proto_rawDesc = "" +
	"\n" +
	"\x1bconfig/calipay/alipay.proto\x12\acalipay\"\xcc\x02\n" +
	"\x06Alipay\x12\x15\n" +
	"\x06app_id\x18\x01 \x01(\tR\x05appId\x12\x1f\n" +
	"\vprivate_key\x18\x02 \x01(\tR\n" +
	"privateKey\x12#\n" +
	"\ris_production\x18\x03 \x01(\bR\fisProduction\x126\n" +
	"\x18app_cert_public_key_file\x18\x04 \x01(\tR\x14appCertPublicKeyFile\x121\n" +
	"\x15alipay_root_cert_file\x18\x05 \x01(\tR\x12alipayRootCertFile\x12<\n" +
	"\x1balipay_cert_public_key_file\x18\x06 \x01(\tR\x17alipayCertPublicKeyFile\x12\x1d\n" +
	"\n" +
	"return_url\x18\a \x01(\tR\treturnUrl\x12\x1d\n" +
	"\n" +
	"notify_url\x18\b \x01(\tR\tnotifyUrlB=Z;github.com/laixhe/gonet/protocol/gen/config/calipay;calipayb\x06proto3"

var (
	file_config_calipay_alipay_proto_rawDescOnce sync.Once
	file_config_calipay_alipay_proto_rawDescData []byte
)

func file_config_calipay_alipay_proto_rawDescGZIP() []byte {
	file_config_calipay_alipay_proto_rawDescOnce.Do(func() {
		file_config_calipay_alipay_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_config_calipay_alipay_proto_rawDesc), len(file_config_calipay_alipay_proto_rawDesc)))
	})
	return file_config_calipay_alipay_proto_rawDescData
}

var file_config_calipay_alipay_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_config_calipay_alipay_proto_goTypes = []any{
	(*Alipay)(nil), // 0: calipay.Alipay
}
var file_config_calipay_alipay_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_config_calipay_alipay_proto_init() }
func file_config_calipay_alipay_proto_init() {
	if File_config_calipay_alipay_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_config_calipay_alipay_proto_rawDesc), len(file_config_calipay_alipay_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_config_calipay_alipay_proto_goTypes,
		DependencyIndexes: file_config_calipay_alipay_proto_depIdxs,
		MessageInfos:      file_config_calipay_alipay_proto_msgTypes,
	}.Build()
	File_config_calipay_alipay_proto = out.File
	file_config_calipay_alipay_proto_goTypes = nil
	file_config_calipay_alipay_proto_depIdxs = nil
}
