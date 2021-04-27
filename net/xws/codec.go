package xws

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

// Codec is the interface for message coder and decoder.
// Application programmer can define a custom codec themselves.
type Codec interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

type baseCodec struct {
	marshaler   proto.MarshalOptions
	unmarshaler proto.UnmarshalOptions
}

var _ Codec = baseCodec{}

// NewBaseCodec .
func NewBaseCodec() Codec {
	// TODO config codec
	return baseCodec{
		marshaler:   proto.MarshalOptions{},
		unmarshaler: proto.UnmarshalOptions{},
	}
}

// Decode decodes the bytes data into Message
func (codec baseCodec) Unmarshal(data []byte, v interface{}) error {
	message, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("type v (%T) is not *proto.Message", v)
	}
	return codec.unmarshaler.Unmarshal(data, message)
}

// Encode encodes the message into bytes data.
func (codec baseCodec) Marshal(v interface{}) ([]byte, error) {
	// TODO
	return nil, nil
}
