package xredis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// Config .
type Config struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
}

// Redis .
type Redis struct {
	*redis.Client
	opts *options
}

// New .
func New(c *Config, opt ...Option) *Redis {
	opts := defaultOptions
	opts.ropts.Addr = c.Addr
	opts.ropts.Password = c.Password
	for _, o := range opt {
		o(&opts)
	}
	client := redis.NewClient(opts.ropts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	return &Redis{
		Client: client,
	}
}

// Close .
func (r *Redis) Close() {
	r.Client.Close()
}
