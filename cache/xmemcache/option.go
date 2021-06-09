package xmemcache

type options struct {
	addr     string
	password string
}

var defaultOptions = options{}

// Option sets server options.
type Option func(*options)

func Addr(addr string) Option {
	return func(o *options) {
		o.addr = addr
	}
}

func Password(pwd string) Option {
	return func(o *options) {
		o.password = pwd
	}
}
