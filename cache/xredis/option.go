package xredis

import "github.com/go-redis/redis/v8"

type options struct {
	ropts *redis.Options
}

var defaultOptions = options{
	ropts: &redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "PaSsWoRd",
	},
}

// Option sets server options.
type Option func(*options)

// DB returns a Option that will set TLS credentials for server
// connections.
func DB(db int) Option {
	return func(o *options) {
		o.ropts.DB = db
	}
}

func Addr(addr string) Option {
	return func(o *options) {
		o.ropts.Addr = addr
	}
}

func Password(pwd string) Option {
	return func(o *options) {
		o.ropts.Password = pwd
	}
}
