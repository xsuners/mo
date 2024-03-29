// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.23.4
// source: option/option.proto

package option

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Cron struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Spec string `protobuf:"bytes,1,opt,name=spec,proto3" json:"spec"`
	Cl   bool   `protobuf:"varint,2,opt,name=cl,proto3" json:"cl"`
}

func (x *Cron) Reset() {
	*x = Cron{}
	if protoimpl.UnsafeEnabled {
		mi := &file_option_option_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Cron) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Cron) ProtoMessage() {}

func (x *Cron) ProtoReflect() protoreflect.Message {
	mi := &file_option_option_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Cron.ProtoReflect.Descriptor instead.
func (*Cron) Descriptor() ([]byte, []int) {
	return file_option_option_proto_rawDescGZIP(), []int{0}
}

func (x *Cron) GetSpec() string {
	if x != nil {
		return x.Spec
	}
	return ""
}

func (x *Cron) GetCl() bool {
	if x != nil {
		return x.Cl
	}
	return false
}

type Event struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Broadcast bool `protobuf:"varint,1,opt,name=broadcast,proto3" json:"broadcast"`
	Ip        bool `protobuf:"varint,2,opt,name=ip,proto3" json:"ip"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_option_option_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_option_option_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event.ProtoReflect.Descriptor instead.
func (*Event) Descriptor() ([]byte, []int) {
	return file_option_option_proto_rawDescGZIP(), []int{1}
}

func (x *Event) GetBroadcast() bool {
	if x != nil {
		return x.Broadcast
	}
	return false
}

func (x *Event) GetIp() bool {
	if x != nil {
		return x.Ip
	}
	return false
}

var file_option_option_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.ServiceOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         10001,
		Name:          "mo.option.type",
		Tag:           "bytes,10001,opt,name=type",
		Filename:      "option/option.proto",
	},
	{
		ExtendedType:  (*descriptorpb.MethodOptions)(nil),
		ExtensionType: (*Event)(nil),
		Field:         10001,
		Name:          "mo.option.event",
		Tag:           "bytes,10001,opt,name=event",
		Filename:      "option/option.proto",
	},
	{
		ExtendedType:  (*descriptorpb.MethodOptions)(nil),
		ExtensionType: (*Cron)(nil),
		Field:         10002,
		Name:          "mo.option.cron",
		Tag:           "bytes,10002,opt,name=cron",
		Filename:      "option/option.proto",
	},
}

// Extension fields to descriptorpb.ServiceOptions.
var (
	// optional string type = 10001;
	E_Type = &file_option_option_proto_extTypes[0]
)

// Extension fields to descriptorpb.MethodOptions.
var (
	//	repeated mo.status.Status status = 10001;
	//	Job job = 10001;
	//
	// optional mo.option.Event event = 10001;
	E_Event = &file_option_option_proto_extTypes[1]
	// optional mo.option.Cron cron = 10002;
	E_Cron = &file_option_option_proto_extTypes[2]
)

var File_option_option_proto protoreflect.FileDescriptor

var file_option_option_proto_rawDesc = []byte{
	0x0a, 0x13, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x6d, 0x6f, 0x2e, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0x2a, 0x0a, 0x04, 0x43, 0x72, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x70,
	0x65, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x73, 0x70, 0x65, 0x63, 0x12, 0x0e,
	0x0a, 0x02, 0x63, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x02, 0x63, 0x6c, 0x22, 0x35,
	0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x62, 0x72, 0x6f, 0x61, 0x64,
	0x63, 0x61, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x62, 0x72, 0x6f, 0x61,
	0x64, 0x63, 0x61, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x02, 0x69, 0x70, 0x3a, 0x34, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x1f, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x91,
	0x4e, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x3a, 0x47, 0x0a, 0x05, 0x65,
	0x76, 0x65, 0x6e, 0x74, 0x12, 0x1e, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x4f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x18, 0x91, 0x4e, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x6d, 0x6f,
	0x2e, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x05, 0x65,
	0x76, 0x65, 0x6e, 0x74, 0x3a, 0x44, 0x0a, 0x04, 0x63, 0x72, 0x6f, 0x6e, 0x12, 0x1e, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d,
	0x65, 0x74, 0x68, 0x6f, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x92, 0x4e, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x6d, 0x6f, 0x2e, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x2e,
	0x43, 0x72, 0x6f, 0x6e, 0x52, 0x04, 0x63, 0x72, 0x6f, 0x6e, 0x42, 0x2b, 0x5a, 0x29, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x78, 0x73, 0x75, 0x6e, 0x65, 0x72, 0x73,
	0x2f, 0x6d, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x2f, 0x67, 0x6f,
	0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_option_option_proto_rawDescOnce sync.Once
	file_option_option_proto_rawDescData = file_option_option_proto_rawDesc
)

func file_option_option_proto_rawDescGZIP() []byte {
	file_option_option_proto_rawDescOnce.Do(func() {
		file_option_option_proto_rawDescData = protoimpl.X.CompressGZIP(file_option_option_proto_rawDescData)
	})
	return file_option_option_proto_rawDescData
}

var file_option_option_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_option_option_proto_goTypes = []interface{}{
	(*Cron)(nil),                        // 0: mo.option.Cron
	(*Event)(nil),                       // 1: mo.option.Event
	(*descriptorpb.ServiceOptions)(nil), // 2: google.protobuf.ServiceOptions
	(*descriptorpb.MethodOptions)(nil),  // 3: google.protobuf.MethodOptions
}
var file_option_option_proto_depIdxs = []int32{
	2, // 0: mo.option.type:extendee -> google.protobuf.ServiceOptions
	3, // 1: mo.option.event:extendee -> google.protobuf.MethodOptions
	3, // 2: mo.option.cron:extendee -> google.protobuf.MethodOptions
	1, // 3: mo.option.event:type_name -> mo.option.Event
	0, // 4: mo.option.cron:type_name -> mo.option.Cron
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	3, // [3:5] is the sub-list for extension type_name
	0, // [0:3] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_option_option_proto_init() }
func file_option_option_proto_init() {
	if File_option_option_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_option_option_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Cron); i {
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
		file_option_option_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Event); i {
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
			RawDescriptor: file_option_option_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 3,
			NumServices:   0,
		},
		GoTypes:           file_option_option_proto_goTypes,
		DependencyIndexes: file_option_option_proto_depIdxs,
		MessageInfos:      file_option_option_proto_msgTypes,
		ExtensionInfos:    file_option_option_proto_extTypes,
	}.Build()
	File_option_option_proto = out.File
	file_option_option_proto_rawDesc = nil
	file_option_option_proto_goTypes = nil
	file_option_option_proto_depIdxs = nil
}
