// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v5.28.3
// source: config/cmongodb/mongodb.proto

package cmongodb

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

// MongoDB数据库配置
type MongoDB struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 连接地址
	Uri string `protobuf:"bytes,1,opt,name=uri,proto3" json:"uri,omitempty" mapstructure:"uri"` // @gotags: mapstructure:"uri"
	// 指定数据库
	Database string `protobuf:"bytes,2,opt,name=database,proto3" json:"database,omitempty" mapstructure:"database"` // @gotags: mapstructure:"database"
	// 最大连接的数量
	MaxPoolSize uint64 `protobuf:"varint,3,opt,name=max_pool_size,json=maxPoolSize,proto3" json:"max_pool_size,omitempty" mapstructure:"max_pool_size"` // @gotags: mapstructure:"max_pool_size"
	// 最小连接的数量
	MinPoolSize uint64 `protobuf:"varint,4,opt,name=min_pool_size,json=minPoolSize,proto3" json:"min_pool_size,omitempty" mapstructure:"min_pool_size"` // @gotags: mapstructure:"min_pool_size"
	// 最大连接的空闲时间(设置了连接可复用的最大时间)(单位秒)
	MaxConnIdleTime int64 `protobuf:"varint,5,opt,name=max_conn_idle_time,json=maxConnIdleTime,proto3" json:"max_conn_idle_time,omitempty" mapstructure:"max_conn_idle_time"` // @gotags: mapstructure:"max_conn_idle_time"
}

func (x *MongoDB) Reset() {
	*x = MongoDB{}
	mi := &file_config_cmongodb_mongodb_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MongoDB) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MongoDB) ProtoMessage() {}

func (x *MongoDB) ProtoReflect() protoreflect.Message {
	mi := &file_config_cmongodb_mongodb_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MongoDB.ProtoReflect.Descriptor instead.
func (*MongoDB) Descriptor() ([]byte, []int) {
	return file_config_cmongodb_mongodb_proto_rawDescGZIP(), []int{0}
}

func (x *MongoDB) GetUri() string {
	if x != nil {
		return x.Uri
	}
	return ""
}

func (x *MongoDB) GetDatabase() string {
	if x != nil {
		return x.Database
	}
	return ""
}

func (x *MongoDB) GetMaxPoolSize() uint64 {
	if x != nil {
		return x.MaxPoolSize
	}
	return 0
}

func (x *MongoDB) GetMinPoolSize() uint64 {
	if x != nil {
		return x.MinPoolSize
	}
	return 0
}

func (x *MongoDB) GetMaxConnIdleTime() int64 {
	if x != nil {
		return x.MaxConnIdleTime
	}
	return 0
}

var File_config_cmongodb_mongodb_proto protoreflect.FileDescriptor

var file_config_cmongodb_mongodb_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x63, 0x6d, 0x6f, 0x6e, 0x67, 0x6f, 0x64,
	0x62, 0x2f, 0x6d, 0x6f, 0x6e, 0x67, 0x6f, 0x64, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x08, 0x63, 0x6d, 0x6f, 0x6e, 0x67, 0x6f, 0x64, 0x62, 0x22, 0xac, 0x01, 0x0a, 0x07, 0x4d, 0x6f,
	0x6e, 0x67, 0x6f, 0x44, 0x42, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x69, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x69, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x61, 0x74, 0x61, 0x62,
	0x61, 0x73, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x64, 0x61, 0x74, 0x61, 0x62,
	0x61, 0x73, 0x65, 0x12, 0x22, 0x0a, 0x0d, 0x6d, 0x61, 0x78, 0x5f, 0x70, 0x6f, 0x6f, 0x6c, 0x5f,
	0x73, 0x69, 0x7a, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b, 0x6d, 0x61, 0x78, 0x50,
	0x6f, 0x6f, 0x6c, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x22, 0x0a, 0x0d, 0x6d, 0x69, 0x6e, 0x5f, 0x70,
	0x6f, 0x6f, 0x6c, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b,
	0x6d, 0x69, 0x6e, 0x50, 0x6f, 0x6f, 0x6c, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x2b, 0x0a, 0x12, 0x6d,
	0x61, 0x78, 0x5f, 0x63, 0x6f, 0x6e, 0x6e, 0x5f, 0x69, 0x64, 0x6c, 0x65, 0x5f, 0x74, 0x69, 0x6d,
	0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0f, 0x6d, 0x61, 0x78, 0x43, 0x6f, 0x6e, 0x6e,
	0x49, 0x64, 0x6c, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x42, 0x3c, 0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x61, 0x69, 0x78, 0x68, 0x65, 0x2f, 0x67, 0x6f,
	0x6e, 0x65, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2f, 0x63, 0x6d, 0x6f, 0x6e, 0x67, 0x6f, 0x64, 0x62, 0x3b, 0x63, 0x6d,
	0x6f, 0x6e, 0x67, 0x6f, 0x64, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_config_cmongodb_mongodb_proto_rawDescOnce sync.Once
	file_config_cmongodb_mongodb_proto_rawDescData = file_config_cmongodb_mongodb_proto_rawDesc
)

func file_config_cmongodb_mongodb_proto_rawDescGZIP() []byte {
	file_config_cmongodb_mongodb_proto_rawDescOnce.Do(func() {
		file_config_cmongodb_mongodb_proto_rawDescData = protoimpl.X.CompressGZIP(file_config_cmongodb_mongodb_proto_rawDescData)
	})
	return file_config_cmongodb_mongodb_proto_rawDescData
}

var file_config_cmongodb_mongodb_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_config_cmongodb_mongodb_proto_goTypes = []any{
	(*MongoDB)(nil), // 0: cmongodb.MongoDB
}
var file_config_cmongodb_mongodb_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_config_cmongodb_mongodb_proto_init() }
func file_config_cmongodb_mongodb_proto_init() {
	if File_config_cmongodb_mongodb_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_config_cmongodb_mongodb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_config_cmongodb_mongodb_proto_goTypes,
		DependencyIndexes: file_config_cmongodb_mongodb_proto_depIdxs,
		MessageInfos:      file_config_cmongodb_mongodb_proto_msgTypes,
	}.Build()
	File_config_cmongodb_mongodb_proto = out.File
	file_config_cmongodb_mongodb_proto_rawDesc = nil
	file_config_cmongodb_mongodb_proto_goTypes = nil
	file_config_cmongodb_mongodb_proto_depIdxs = nil
}
