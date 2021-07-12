package json

import (
	"encoding/json"
	"fmt"

	"github.com/xsuners/mo/net/encoding"
	"google.golang.org/protobuf/proto"
)

// Name is the name registered for the proto compressor.
const Name = "json"

func init() {
	encoding.RegisterCodec(Codec{})
}

// Codec is a Codec implementation with protobuf. It is the default Codec for gRPC.
type Codec struct{}

func (Codec) Marshal(v interface{}) ([]byte, error) {
	vv, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("failed to marshal, message is %T, want proto.Message", v)
	}
	return json.Marshal(vv)
}

func (Codec) Unmarshal(data []byte, v interface{}) error {
	vv, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("failed to unmarshal, message is %T, want proto.Message", v)
	}
	return json.Unmarshal(data, vv)
}

func (Codec) Name() string {
	return Name
}
