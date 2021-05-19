package extracter

import (
	"context"
	"fmt"

	"github.com/xsuners/mo/net/util/meta"
	"go.uber.org/zap"
)

// Extracter .
type Extracter interface {
	WithMDKVs(ctx context.Context, keysAndValues []interface{}) []interface{}
	WithMDArgs(ctx context.Context, args []interface{}) []interface{}
	WithMDFormat(ctx context.Context, format string) string
	WithMDFields(ctx context.Context, fields []zap.Field) []zap.Field
}

var _ Extracter = Extractor{}

// Extractor .
type Extractor struct{}

// WithMDKVs .
func (Extractor) WithMDKVs(ctx context.Context, keysAndValues []interface{}) []interface{} {
	if ctx == nil {
		return keysAndValues
	}
	m := meta.FromContext(ctx)
	mdKVs := []interface{}{
		"srcappid", m.Srcappid,
		"aimappid", m.Aimappid,
		"srcid", m.Srcid,
		"aimid", m.Aimid,
		"hash", m.Hash,
		"time", m.Time,
		"sn", m.Sn,
	}
	return append(mdKVs, keysAndValues...)
}

// WithMDArgs .
func (Extractor) WithMDArgs(ctx context.Context, args []interface{}) []interface{} {
	if ctx == nil {
		return args
	}
	m := meta.FromContext(ctx)
	mdArgs := []interface{}{
		"srcappid" + ":", m.Srcappid,
		" " + "aimappid" + ":", m.Aimappid,
		" " + "srcid" + ":", m.Srcid,
		" " + "aimid" + ":", m.Aimid,
		" " + "time" + ":", m.Time,
		" " + "hash" + ":", m.Hash,
		" " + "sn" + ":", m.Sn,
		" ",
	}
	return append(mdArgs, args...)
}

// WithMDFormat .
func (Extractor) WithMDFormat(ctx context.Context, format string) string {
	if ctx == nil {
		return format
	}
	m := meta.FromContext(ctx)

	return format + fmt.Sprintf(" srcappid: %d aimappid: %d srcid: %d aimid: %d time: %d hash: %d sn: %d", m.Srcappid, m.Aimappid, m.Srcid, m.Aimid, m.Time, m.Hash, m.Sn)
}

// WithMDFields .
func (Extractor) WithMDFields(ctx context.Context, fields []zap.Field) []zap.Field {
	if ctx == nil {
		return fields
	}
	m := meta.FromContext(ctx)
	mdFields := []zap.Field{
		zap.Int64("srcappid", m.Srcappid),
		zap.Int64("aimappid", m.Aimappid),
		zap.Int64("srcid", m.Srcid),
		zap.Int64("aimid", m.Aimid),
		zap.Int64("time", m.Time),
		zap.Int64("hash", m.Hash),
		zap.Int64("sn", m.Sn),
	}
	return append(mdFields, fields...)
}
