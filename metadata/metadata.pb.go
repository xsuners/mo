// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0-devel
// 	protoc        v3.19.1
// source: mo/metadata/metadata.proto

package metadata

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

type Metadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hash   int64             `protobuf:"varint,1,opt,name=hash,proto3" json:"hash,omitempty"`
	Time   int64             `protobuf:"varint,2,opt,name=time,proto3" json:"time,omitempty"`
	Sn     int64             `protobuf:"varint,3,opt,name=sn,proto3" json:"sn,omitempty"`
	Addr   string            `protobuf:"bytes,4,opt,name=addr,proto3" json:"addr,omitempty"`
	Appid  int64             `protobuf:"varint,10,opt,name=appid,proto3" json:"appid,omitempty"`
	Id     int64             `protobuf:"varint,11,opt,name=id,proto3" json:"id,omitempty"`
	Name   string            `protobuf:"bytes,12,opt,name=name,proto3" json:"name,omitempty"`
	Device int32             `protobuf:"varint,13,opt,name=device,proto3" json:"device,omitempty"`
	Ints   map[string]int64  `protobuf:"bytes,21,rep,name=ints,proto3" json:"ints,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	Strs   map[string]string `protobuf:"bytes,22,rep,name=strs,proto3" json:"strs,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Objs   map[string][]byte `protobuf:"bytes,23,rep,name=objs,proto3" json:"objs,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Metadata) Reset() {
	*x = Metadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mo_metadata_metadata_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Metadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Metadata) ProtoMessage() {}

func (x *Metadata) ProtoReflect() protoreflect.Message {
	mi := &file_mo_metadata_metadata_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Metadata.ProtoReflect.Descriptor instead.
func (*Metadata) Descriptor() ([]byte, []int) {
	return file_mo_metadata_metadata_proto_rawDescGZIP(), []int{0}
}

func (x *Metadata) GetHash() int64 {
	if x != nil {
		return x.Hash
	}
	return 0
}

func (x *Metadata) GetTime() int64 {
	if x != nil {
		return x.Time
	}
	return 0
}

func (x *Metadata) GetSn() int64 {
	if x != nil {
		return x.Sn
	}
	return 0
}

func (x *Metadata) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

func (x *Metadata) GetAppid() int64 {
	if x != nil {
		return x.Appid
	}
	return 0
}

func (x *Metadata) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Metadata) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Metadata) GetDevice() int32 {
	if x != nil {
		return x.Device
	}
	return 0
}

func (x *Metadata) GetInts() map[string]int64 {
	if x != nil {
		return x.Ints
	}
	return nil
}

func (x *Metadata) GetStrs() map[string]string {
	if x != nil {
		return x.Strs
	}
	return nil
}

func (x *Metadata) GetObjs() map[string][]byte {
	if x != nil {
		return x.Objs
	}
	return nil
}

var File_mo_metadata_metadata_proto protoreflect.FileDescriptor

var file_mo_metadata_metadata_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x6d, 0x6f, 0x2f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2f, 0x6d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x6d, 0x6f,
	0x2e, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x22, 0xf2, 0x03, 0x0a, 0x08, 0x4d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x69,
	0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x0e,
	0x0a, 0x02, 0x73, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x73, 0x6e, 0x12, 0x12,
	0x0a, 0x04, 0x61, 0x64, 0x64, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x61, 0x64,
	0x64, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x61, 0x70, 0x70, 0x69, 0x64, 0x18, 0x0a, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x05, 0x61, 0x70, 0x70, 0x69, 0x64, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x0b,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x0c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06,
	0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x64, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x33, 0x0a, 0x04, 0x69, 0x6e, 0x74, 0x73, 0x18, 0x15, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x6d, 0x6f, 0x2e, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x49, 0x6e, 0x74, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x04, 0x69, 0x6e, 0x74, 0x73, 0x12, 0x33, 0x0a, 0x04, 0x73, 0x74, 0x72,
	0x73, 0x18, 0x16, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x6d, 0x6f, 0x2e, 0x6d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x53,
	0x74, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x73, 0x74, 0x72, 0x73, 0x12, 0x33,
	0x0a, 0x04, 0x6f, 0x62, 0x6a, 0x73, 0x18, 0x17, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x6d,
	0x6f, 0x2e, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x2e, 0x4f, 0x62, 0x6a, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x6f,
	0x62, 0x6a, 0x73, 0x1a, 0x37, 0x0a, 0x09, 0x49, 0x6e, 0x74, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x37, 0x0a, 0x09,
	0x53, 0x74, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x37, 0x0a, 0x09, 0x4f, 0x62, 0x6a, 0x73, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x28,
	0x5a, 0x26, 0x73, 0x68, 0x65, 0x70, 0x69, 0x6e, 0x2e, 0x6c, 0x69, 0x76, 0x65, 0x2f, 0x62, 0x61,
	0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x61, 0x70, 0x69, 0x5f, 0x67, 0x6f, 0x2f, 0x6d, 0x6f, 0x2f,
	0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mo_metadata_metadata_proto_rawDescOnce sync.Once
	file_mo_metadata_metadata_proto_rawDescData = file_mo_metadata_metadata_proto_rawDesc
)

func file_mo_metadata_metadata_proto_rawDescGZIP() []byte {
	file_mo_metadata_metadata_proto_rawDescOnce.Do(func() {
		file_mo_metadata_metadata_proto_rawDescData = protoimpl.X.CompressGZIP(file_mo_metadata_metadata_proto_rawDescData)
	})
	return file_mo_metadata_metadata_proto_rawDescData
}

var file_mo_metadata_metadata_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_mo_metadata_metadata_proto_goTypes = []interface{}{
	(*Metadata)(nil), // 0: mo.metadata.Metadata
	nil,              // 1: mo.metadata.Metadata.IntsEntry
	nil,              // 2: mo.metadata.Metadata.StrsEntry
	nil,              // 3: mo.metadata.Metadata.ObjsEntry
}
var file_mo_metadata_metadata_proto_depIdxs = []int32{
	1, // 0: mo.metadata.Metadata.ints:type_name -> mo.metadata.Metadata.IntsEntry
	2, // 1: mo.metadata.Metadata.strs:type_name -> mo.metadata.Metadata.StrsEntry
	3, // 2: mo.metadata.Metadata.objs:type_name -> mo.metadata.Metadata.ObjsEntry
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_mo_metadata_metadata_proto_init() }
func file_mo_metadata_metadata_proto_init() {
	if File_mo_metadata_metadata_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mo_metadata_metadata_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Metadata); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_mo_metadata_metadata_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_mo_metadata_metadata_proto_goTypes,
		DependencyIndexes: file_mo_metadata_metadata_proto_depIdxs,
		MessageInfos:      file_mo_metadata_metadata_proto_msgTypes,
	}.Build()
	File_mo_metadata_metadata_proto = out.File
	file_mo_metadata_metadata_proto_rawDesc = nil
	file_mo_metadata_metadata_proto_goTypes = nil
	file_mo_metadata_metadata_proto_depIdxs = nil
}
