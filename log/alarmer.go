package log

import (
	"context"
)

type Alarmer interface {
	Alarm(ctx context.Context, domain, title string, kvs ...string) error
}

type empty struct{}

var _ Alarmer = (*empty)(nil)

func (*empty) Alarm(ctx context.Context, domain, title string, kvs ...string) error {
	return nil
}
