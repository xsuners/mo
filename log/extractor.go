package log

import (
	"context"

	"go.uber.org/zap"
)

// Extractor .
type Extractor interface {
	WithMDKVs(ctx context.Context, keysAndValues []interface{}) []interface{}
	WithMDArgs(ctx context.Context, args []interface{}) []interface{}
	WithMDFormat(ctx context.Context, format string) string
	WithMDFields(ctx context.Context, fields []zap.Field) []zap.Field
}
