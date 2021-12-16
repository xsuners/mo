package xmemcache

type Options struct {
	Addr     string `ini-name:"addr" long:"memcache-addr" description:"memcache addr"`
	Password string `ini-name:"password" long:"memcache-password" description:"memcache password"`
}

// var defaultOptions = Options{}

// Option sets server options.
type Option func(*Options)

func Addr(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}

func Password(pwd string) Option {
	return func(o *Options) {
		o.Password = pwd
	}
}
