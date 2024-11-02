// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v5.28.3
// source: config/cdb/db.proto

package cdb

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

// 数据库配置
type DB struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 连接地址
	Dsn string `protobuf:"bytes,1,opt,name=dsn,proto3" json:"dsn,omitempty" mapstructure:"dsn"` // @gotags: mapstructure:"dsn"
	// 设置空闲连接池中连接的最大数量
	MaxIdleCount int32 `protobuf:"varint,2,opt,name=max_idle_count,json=maxIdleCount,proto3" json:"max_idle_count,omitempty" mapstructure:"max_idle_count"` // @gotags: mapstructure:"max_idle_count"
	// 设置打开数据库连接的最大数量
	MaxOpenCount int32 `protobuf:"varint,3,opt,name=max_open_count,json=maxOpenCount,proto3" json:"max_open_count,omitempty" mapstructure:"max_open_count"` // @gotags: mapstructure:"max_open_count"
	// 设置了连接可复用的最大时间(要比数据库设置连接超时时间少)(单位秒)
	MaxLifeTime int64 `protobuf:"varint,4,opt,name=max_life_time,json=maxLifeTime,proto3" json:"max_life_time,omitempty" mapstructure:"max_life_time"` // @gotags: mapstructure:"max_life_time"
}

func (x *DB) Reset() {
	*x = DB{}
	mi := &file_config_cdb_db_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DB) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DB) ProtoMessage() {}

func (x *DB) ProtoReflect() protoreflect.Message {
	mi := &file_config_cdb_db_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DB.ProtoReflect.Descriptor instead.
func (*DB) Descriptor() ([]byte, []int) {
	return file_config_cdb_db_proto_rawDescGZIP(), []int{0}
}

func (x *DB) GetDsn() string {
	if x != nil {
		return x.Dsn
	}
	return ""
}

func (x *DB) GetMaxIdleCount() int32 {
	if x != nil {
		return x.MaxIdleCount
	}
	return 0
}

func (x *DB) GetMaxOpenCount() int32 {
	if x != nil {
		return x.MaxOpenCount
	}
	return 0
}

func (x *DB) GetMaxLifeTime() int64 {
	if x != nil {
		return x.MaxLifeTime
	}
	return 0
}

var File_config_cdb_db_proto protoreflect.FileDescriptor

var file_config_cdb_db_proto_rawDesc = []byte{
	0x0a, 0x13, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x63, 0x64, 0x62, 0x2f, 0x64, 0x62, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x63, 0x64, 0x62, 0x22, 0x86, 0x01, 0x0a, 0x02, 0x44,
	0x42, 0x12, 0x10, 0x0a, 0x03, 0x64, 0x73, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x64, 0x73, 0x6e, 0x12, 0x24, 0x0a, 0x0e, 0x6d, 0x61, 0x78, 0x5f, 0x69, 0x64, 0x6c, 0x65, 0x5f,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x6d, 0x61, 0x78,
	0x49, 0x64, 0x6c, 0x65, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x24, 0x0a, 0x0e, 0x6d, 0x61, 0x78,
	0x5f, 0x6f, 0x70, 0x65, 0x6e, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x0c, 0x6d, 0x61, 0x78, 0x4f, 0x70, 0x65, 0x6e, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12,
	0x22, 0x0a, 0x0d, 0x6d, 0x61, 0x78, 0x5f, 0x6c, 0x69, 0x66, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x6d, 0x61, 0x78, 0x4c, 0x69, 0x66, 0x65, 0x54,
	0x69, 0x6d, 0x65, 0x42, 0x32, 0x5a, 0x30, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x6c, 0x61, 0x69, 0x78, 0x68, 0x65, 0x2f, 0x67, 0x6f, 0x6e, 0x65, 0x74, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f,
	0x63, 0x64, 0x62, 0x3b, 0x63, 0x64, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_config_cdb_db_proto_rawDescOnce sync.Once
	file_config_cdb_db_proto_rawDescData = file_config_cdb_db_proto_rawDesc
)

func file_config_cdb_db_proto_rawDescGZIP() []byte {
	file_config_cdb_db_proto_rawDescOnce.Do(func() {
		file_config_cdb_db_proto_rawDescData = protoimpl.X.CompressGZIP(file_config_cdb_db_proto_rawDescData)
	})
	return file_config_cdb_db_proto_rawDescData
}

var file_config_cdb_db_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_config_cdb_db_proto_goTypes = []any{
	(*DB)(nil), // 0: cdb.DB
}
var file_config_cdb_db_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_config_cdb_db_proto_init() }
func file_config_cdb_db_proto_init() {
	if File_config_cdb_db_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_config_cdb_db_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_config_cdb_db_proto_goTypes,
		DependencyIndexes: file_config_cdb_db_proto_depIdxs,
		MessageInfos:      file_config_cdb_db_proto_msgTypes,
	}.Build()
	File_config_cdb_db_proto = out.File
	file_config_cdb_db_proto_rawDesc = nil
	file_config_cdb_db_proto_goTypes = nil
	file_config_cdb_db_proto_depIdxs = nil
}
