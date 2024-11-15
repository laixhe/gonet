// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.28.3
// source: config/cserver/server.proto

package cserver

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

// 服务器配置
type Server struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 运行IP
	Ip string `protobuf:"bytes,1,opt,name=ip,proto3" json:"ip,omitempty" mapstructure:"ip"` // @gotags: mapstructure:"ip"
	// 运行端口
	Port int32 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty" mapstructure:"port"` // @gotags: mapstructure:"port"
	// 超时时间(单位秒)
	Timeout int64 `protobuf:"varint,3,opt,name=timeout,proto3" json:"timeout,omitempty" mapstructure:"timeout"` // @gotags: mapstructure:"timeout"
}

func (x *Server) Reset() {
	*x = Server{}
	mi := &file_config_cserver_server_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Server) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Server) ProtoMessage() {}

func (x *Server) ProtoReflect() protoreflect.Message {
	mi := &file_config_cserver_server_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Server.ProtoReflect.Descriptor instead.
func (*Server) Descriptor() ([]byte, []int) {
	return file_config_cserver_server_proto_rawDescGZIP(), []int{0}
}

func (x *Server) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

func (x *Server) GetPort() int32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *Server) GetTimeout() int64 {
	if x != nil {
		return x.Timeout
	}
	return 0
}

// 服务器组
type Servers struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Http *Server `protobuf:"bytes,1,opt,name=http,proto3" json:"http,omitempty" mapstructure:"http"` // @gotags: mapstructure:"http"
}

func (x *Servers) Reset() {
	*x = Servers{}
	mi := &file_config_cserver_server_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Servers) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Servers) ProtoMessage() {}

func (x *Servers) ProtoReflect() protoreflect.Message {
	mi := &file_config_cserver_server_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Servers.ProtoReflect.Descriptor instead.
func (*Servers) Descriptor() ([]byte, []int) {
	return file_config_cserver_server_proto_rawDescGZIP(), []int{1}
}

func (x *Servers) GetHttp() *Server {
	if x != nil {
		return x.Http
	}
	return nil
}

var File_config_cserver_server_proto protoreflect.FileDescriptor

var file_config_cserver_server_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x63, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x63,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x22, 0x46, 0x0a, 0x06, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x70,
	0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04,
	0x70, 0x6f, 0x72, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x22, 0x2e,
	0x0a, 0x07, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x12, 0x23, 0x0a, 0x04, 0x68, 0x74, 0x74,
	0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x63, 0x73, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x52, 0x04, 0x68, 0x74, 0x74, 0x70, 0x42, 0x3a,
	0x5a, 0x38, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x61, 0x69,
	0x78, 0x68, 0x65, 0x2f, 0x67, 0x6f, 0x6e, 0x65, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f,
	0x67, 0x65, 0x6e, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x63, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x3b, 0x63, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_config_cserver_server_proto_rawDescOnce sync.Once
	file_config_cserver_server_proto_rawDescData = file_config_cserver_server_proto_rawDesc
)

func file_config_cserver_server_proto_rawDescGZIP() []byte {
	file_config_cserver_server_proto_rawDescOnce.Do(func() {
		file_config_cserver_server_proto_rawDescData = protoimpl.X.CompressGZIP(file_config_cserver_server_proto_rawDescData)
	})
	return file_config_cserver_server_proto_rawDescData
}

var file_config_cserver_server_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_config_cserver_server_proto_goTypes = []any{
	(*Server)(nil),  // 0: cserver.Server
	(*Servers)(nil), // 1: cserver.Servers
}
var file_config_cserver_server_proto_depIdxs = []int32{
	0, // 0: cserver.Servers.http:type_name -> cserver.Server
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_config_cserver_server_proto_init() }
func file_config_cserver_server_proto_init() {
	if File_config_cserver_server_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_config_cserver_server_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_config_cserver_server_proto_goTypes,
		DependencyIndexes: file_config_cserver_server_proto_depIdxs,
		MessageInfos:      file_config_cserver_server_proto_msgTypes,
	}.Build()
	File_config_cserver_server_proto = out.File
	file_config_cserver_server_proto_rawDesc = nil
	file_config_cserver_server_proto_goTypes = nil
	file_config_cserver_server_proto_depIdxs = nil
}
