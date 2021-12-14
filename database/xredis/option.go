package xredis

import "github.com/go-redis/redis/v8"

type Options struct {
	Addr     string `ini-name:"addr" long:"redis.addr" description:"redis addr"`
	Password string `ini-name:"password" long:"redis.password" description:"redis password"`

	ropts *redis.Options
}

var defaultOptions = Options{
	ropts: &redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "PaSsWoRd",
	},
}

// Option sets server options.
type Option func(*Options)

// DB returns a Option that will set TLS credentials for server
// connections.
func DB(db int) Option {
	return func(o *Options) {
		o.ropts.DB = db
	}
}

func Addr(addr string) Option {
	return func(o *Options) {
		o.ropts.Addr = addr
	}
}

func Password(pwd string) Option {
	return func(o *Options) {
		o.ropts.Password = pwd
	}
}
