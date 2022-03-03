package metadata

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/xsuners/mo/net/encoding"
	jsonc "github.com/xsuners/mo/net/encoding/json"
	protoc "github.com/xsuners/mo/net/encoding/proto"
	"github.com/xsuners/mo/net/message"
	md "google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

////////////////////////////

var (
	// MK name is used in grpc md.
	// MK  = string(proto.MessageName((*Metadata)(nil)))
	MK  = "metadata"
	JMK = MK + ".json"
	// HK name is used in http header
	HK = "x-mo-meta"
)

type mk struct{}

// NewContext .
func NewContext(ctx context.Context, md *Metadata) context.Context {
	return context.WithValue(ctx, mk{}, md)
}

// FromContext .
func FromContext(ctx context.Context) *Metadata {
	if md, ok := ctx.Value(mk{}).(*Metadata); ok {
		return md
	}
	return &Metadata{
		Sn: math.MaxInt64,
	}
}

// Clone .
func Clone(in *Metadata) (out *Metadata) {
	if in == nil {
		return &Metadata{
			Time: time.Now().Unix(),
		}
	}
	return proto.Clone(in).(*Metadata)
}

// NewOutgoingContext .
func NewOutgoingContext(ctx context.Context, smd *Metadata) context.Context {
	data, _ := proto.Marshal(smd)
	encodeString := base64.StdEncoding.EncodeToString(data)
	m, ok := md.FromOutgoingContext(ctx)
	if ok {
		m = m.Copy()
		m.Set(MK, encodeString)
	} else {
		m = md.New(map[string]string{
			MK: encodeString,
		})
	}
	ctx = md.NewOutgoingContext(ctx, m)
	return ctx
}

// FromIncomingContext .
func FromIncomingContext(ctx context.Context) (*Metadata, bool) {
	md, ok := md.FromIncomingContext(ctx)
	if !ok {
		return nil, false
	}
	mds := md.Get(MK)
	if len(mds) > 0 {
		data, _ := base64.StdEncoding.DecodeString(mds[0])
		smd := &Metadata{}
		err := proto.Unmarshal(data, smd)
		if err != nil {
			return nil, false
		}
		return smd, true
	}
	mds = md.Get(JMK)
	if len(mds) > 0 {
		data, _ := base64.StdEncoding.DecodeString(mds[0])
		smd := &Metadata{}
		err := json.Unmarshal(data, smd)
		if err != nil {
			return nil, false
		}
		return smd, true
	}
	return nil, false
}

// Base64 .
// TODO 对 codec 的支持更自由
func Base64(mt *Metadata, codec encoding.Codec) (*message.Meta, error) {
	data, err := codec.Marshal(mt)
	if err != nil {
		return nil, err
	}
	switch codec.Name() {
	case jsonc.Name:
		return &message.Meta{
			Name:  JMK,
			Value: base64.StdEncoding.EncodeToString(data),
		}, nil
	case protoc.Name:
		return &message.Meta{
			Name:  MK,
			Value: base64.StdEncoding.EncodeToString(data),
		}, nil
	}
	return nil, errors.New("unknown codec")
}
