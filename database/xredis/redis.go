package xredis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// Redis .
type Redis struct {
	*redis.Client
	opts *options
}

// New .
func New(opt ...Option) (*Redis, func(), error) {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	client := redis.NewClient(opts.ropts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	r := &Redis{
		opts:   &opts,
		Client: client,
	}
	return r, func() {
		r.Close()
	}, nil
}
