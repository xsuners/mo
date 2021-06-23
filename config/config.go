package config

import (
	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/naming"
	"github.com/xsuners/mo/net/xgrpc"
	"github.com/xsuners/mo/net/xhttp"
	"github.com/xsuners/mo/net/xnats"
	"github.com/xsuners/mo/net/xtcp"
	"github.com/xsuners/mo/net/xws"
)

type options struct {
	localPath string
}

var defaultOptions = options{
	localPath: "../conf/conf.json",
}

// Option sets server options.
type Option func(*options)

// LocalPath returns a Option that will set TLS credentials for server
// connections.
func LocalPath(path string) Option {
	return func(o *options) {
		o.localPath = path
	}
}

type Config interface {
	WSOptions() []xws.Option
	TCPOptions() []xtcp.Option
	GRPCOptions() []xgrpc.Option
	NATSOptions() []xnats.Option
	HTTPOptions() []xhttp.Option
	LogOptions() []log.Option
	NamingOptions() []naming.Option
}

// Configure .
type Configure struct {
	opts options
}

// New .
func New(opt ...Option) *Configure {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	return &Configure{
		opts: opts,
	}
}
