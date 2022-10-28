package proxy

import (
	"bytes"
	"io"

	// "github.com/golang/protobuf/proto"

	"github.com/xsuners/mo/net/message"
	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/proto"
)

// Codec returns a proxying encoding.Codec with the default protobuf codec as parent.
//
// See CodecWithParent.
func Codec() encoding.Codec {
	return codec(&protoCodec{})
}

// codec .
func codec(fallback encoding.Codec) encoding.Codec {
	return &rawCodec{fallback}
}

type rawCodec struct {
	parentCodec encoding.Codec
}

func (c *rawCodec) Marshal(v interface{}) ([]byte, error) {
	switch out := v.(type) {
	case *message.Frame:
		return out.Data, nil
	case io.Reader:
		buf := bytes.NewBuffer([]byte{})
		_, err := io.Copy(buf, out)
		// fmt.Println("<<<<<<<<<<<<<<<<<", buf.Bytes(), "|||", buf.String())
		return buf.Bytes(), err
	}
	return c.parentCodec.Marshal(v)
}

func (c *rawCodec) Unmarshal(data []byte, v interface{}) error {
	switch dst := v.(type) {
	case *message.Frame:
		dst.Data = data
		return nil
	case io.Writer:
		// fmt.Println(">>>>>>>>>>>>>>>>", data)
		_, err := dst.Write(data)
		return err
	}
	return c.parentCodec.Unmarshal(data, v)
}

// func (c *rawCodec) String() string {
// 	return "proxy-codec"
// }

func (c *rawCodec) Name() string {
	return "proxy-codec"
}

// protoCodec is a Codec implementation with protobuf. It is the default rawCodec for gRPC.
type protoCodec struct{}

func (protoCodec) Marshal(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (protoCodec) Unmarshal(data []byte, v interface{}) error {
	return proto.Unmarshal(data, v.(proto.Message))
}

// func (protoCodec) String() string {
// 	return "proto"
// }

func (protoCodec) Name() string {
	return "proto"
}
