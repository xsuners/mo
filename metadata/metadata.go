package metadata

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
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
		if md.Sn < 1 {
			md.Sn = time.Now().UnixNano()
		}
		return md
	}
	return &Metadata{
		Sn: time.Now().UnixNano(),
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
		smd := &Metadata{
			Sn: time.Now().UnixNano(),
		}
		err := proto.Unmarshal(data, smd)
		if err != nil {
			return nil, false
		}
		return smd, true
	}
	mds = md.Get(JMK)
	if len(mds) > 0 {
		// data, _ := base64.StdEncoding.DecodeString(mds[0])
		smd := &Metadata{
			Sn: time.Now().UnixNano(),
		}
		err := json.Unmarshal([]byte(mds[0]), smd)
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

func (x *Metadata) Int64(key string) int64 {
	return x.Ints[key]
}

func (x *Metadata) Str(key string) string {
	return x.Strs[key]
}

func (x *Metadata) Fetch(key string, val interface{}) (ok bool, err error) {
	bo, ok := x.Objs[key]
	if !ok {
		return
	}
	if val == nil {
		return
	}
	err = json.Unmarshal(bo, val)
	if err != nil {
		return
	}
	return
}
