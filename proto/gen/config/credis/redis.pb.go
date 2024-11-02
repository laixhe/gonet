// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v5.28.3
// source: config/credis/redis.proto

package credis

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

// Redis配置
type Redis struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 连接地址
	Addr string `protobuf:"bytes,1,opt,name=addr,proto3" json:"addr,omitempty" mapstructure:"addr"` // @gotags: mapstructure:"addr"
	// 选择N号数据库
	DbNum int32 `protobuf:"varint,2,opt,name=db_num,json=dbNum,proto3" json:"db_num,omitempty" mapstructure:"db_num"` // @gotags: mapstructure:"db_num"
	// 设置打开数据库连接的最大数量
	Password string `protobuf:"bytes,3,opt,name=password,proto3" json:"password,omitempty" mapstructure:"password"` // @gotags: mapstructure:"password"
	// 最大链接数
	PoolSize int32 `protobuf:"varint,4,opt,name=pool_size,json=poolSize,proto3" json:"pool_size,omitempty" mapstructure:"pool_size"` // @gotags: mapstructure:"pool_size"
	// 空闲链接数
	MinIdleConn int32 `protobuf:"varint,5,opt,name=min_idle_conn,json=minIdleConn,proto3" json:"min_idle_conn,omitempty" mapstructure:"min_idle_conn"` // @gotags: mapstructure:"min_idle_conn"
}

func (x *Redis) Reset() {
	*x = Redis{}
	mi := &file_config_credis_redis_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Redis) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Redis) ProtoMessage() {}

func (x *Redis) ProtoReflect() protoreflect.Message {
	mi := &file_config_credis_redis_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Redis.ProtoReflect.Descriptor instead.
func (*Redis) Descriptor() ([]byte, []int) {
	return file_config_credis_redis_proto_rawDescGZIP(), []int{0}
}

func (x *Redis) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

func (x *Redis) GetDbNum() int32 {
	if x != nil {
		return x.DbNum
	}
	return 0
}

func (x *Redis) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

func (x *Redis) GetPoolSize() int32 {
	if x != nil {
		return x.PoolSize
	}
	return 0
}

func (x *Redis) GetMinIdleConn() int32 {
	if x != nil {
		return x.MinIdleConn
	}
	return 0
}

var File_config_credis_redis_proto protoreflect.FileDescriptor

var file_config_credis_redis_proto_rawDesc = []byte{
	0x0a, 0x19, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x63, 0x72, 0x65, 0x64, 0x69, 0x73, 0x2f,
	0x72, 0x65, 0x64, 0x69, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x63, 0x72, 0x65,
	0x64, 0x69, 0x73, 0x22, 0x8f, 0x01, 0x0a, 0x05, 0x52, 0x65, 0x64, 0x69, 0x73, 0x12, 0x12, 0x0a,
	0x04, 0x61, 0x64, 0x64, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x61, 0x64, 0x64,
	0x72, 0x12, 0x15, 0x0a, 0x06, 0x64, 0x62, 0x5f, 0x6e, 0x75, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x05, 0x64, 0x62, 0x4e, 0x75, 0x6d, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x61, 0x73, 0x73,
	0x77, 0x6f, 0x72, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x61, 0x73, 0x73,
	0x77, 0x6f, 0x72, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x6f, 0x6f, 0x6c, 0x5f, 0x73, 0x69, 0x7a,
	0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x6f, 0x6f, 0x6c, 0x53, 0x69, 0x7a,
	0x65, 0x12, 0x22, 0x0a, 0x0d, 0x6d, 0x69, 0x6e, 0x5f, 0x69, 0x64, 0x6c, 0x65, 0x5f, 0x63, 0x6f,
	0x6e, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x6d, 0x69, 0x6e, 0x49, 0x64, 0x6c,
	0x65, 0x43, 0x6f, 0x6e, 0x6e, 0x42, 0x38, 0x5a, 0x36, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x61, 0x69, 0x78, 0x68, 0x65, 0x2f, 0x67, 0x6f, 0x6e, 0x65, 0x74,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x2f, 0x63, 0x72, 0x65, 0x64, 0x69, 0x73, 0x3b, 0x63, 0x72, 0x65, 0x64, 0x69, 0x73, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_config_credis_redis_proto_rawDescOnce sync.Once
	file_config_credis_redis_proto_rawDescData = file_config_credis_redis_proto_rawDesc
)

func file_config_credis_redis_proto_rawDescGZIP() []byte {
	file_config_credis_redis_proto_rawDescOnce.Do(func() {
		file_config_credis_redis_proto_rawDescData = protoimpl.X.CompressGZIP(file_config_credis_redis_proto_rawDescData)
	})
	return file_config_credis_redis_proto_rawDescData
}

var file_config_credis_redis_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_config_credis_redis_proto_goTypes = []any{
	(*Redis)(nil), // 0: credis.Redis
}
var file_config_credis_redis_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_config_credis_redis_proto_init() }
func file_config_credis_redis_proto_init() {
	if File_config_credis_redis_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_config_credis_redis_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_config_credis_redis_proto_goTypes,
		DependencyIndexes: file_config_credis_redis_proto_depIdxs,
		MessageInfos:      file_config_credis_redis_proto_msgTypes,
	}.Build()
	File_config_credis_redis_proto = out.File
	file_config_credis_redis_proto_rawDesc = nil
	file_config_credis_redis_proto_goTypes = nil
	file_config_credis_redis_proto_depIdxs = nil
}
