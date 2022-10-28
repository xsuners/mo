package mf

import (
	"google.golang.org/protobuf/proto"
)

var factory = make(map[string]proto.Message)

func Add(msg proto.Message) {
	factory[string(msg.ProtoReflect().Descriptor().FullName())] = msg
}

func Get(name string) (proto.Message, bool) {
	msg, ok := factory[name]
	return proto.Clone(msg), ok
}
