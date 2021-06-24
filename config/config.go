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

type Config interface {
	Clone() Config
	Add(opt ...Option) Config
	WSOptions() []xws.Option
	TCPOptions() []xtcp.Option
	GRPCOptions() []xgrpc.Option
	NATSOptions() []xnats.Option
	HTTPOptions() []xhttp.Option
	LogOptions() []log.Option
	NamingOptions() []naming.Option
}

// Option sets server options.
type Option func(*configure)

// WSOptions .
func WSOptions(opts ...xws.Option) Option {
	return func(c *configure) {
		c.wsopts = append(c.wsopts, opts...)
	}
}

// TCPOptions .
func TCPOptions(opts ...xtcp.Option) Option {
	return func(c *configure) {
		c.tcpopts = append(c.tcpopts, opts...)
	}
}

// GRPCOptions .
func GRPCOptions(opts ...xgrpc.Option) Option {
	return func(c *configure) {
		c.grpcopts = append(c.grpcopts, opts...)
	}
}

// HTTPOptions .
func HTTPOptions(opts ...xhttp.Option) Option {
	return func(c *configure) {
		c.httpopts = append(c.httpopts, opts...)
	}
}

// NATSOptions .
func NATSOptions(opts ...xnats.Option) Option {
	return func(c *configure) {
		c.natsopts = append(c.natsopts, opts...)
	}
}

// LogOptions .
func LogOptions(opts ...log.Option) Option {
	return func(c *configure) {
		c.logopts = append(c.logopts, opts...)
	}
}

// NamingOptions .
func NamingOptions(opts ...naming.Option) Option {
	return func(c *configure) {
		c.namingopts = append(c.namingopts, opts...)
	}
}

// configure .
type configure struct {
	wsopts     []xws.Option
	tcpopts    []xtcp.Option
	grpcopts   []xgrpc.Option
	natsopts   []xnats.Option
	httpopts   []xhttp.Option
	logopts    []log.Option
	namingopts []naming.Option
}

var _ Config = (*configure)(nil)

// New .
func New(opt ...Option) *configure {
	c := new(configure)
	for _, o := range opt {
		o(c)
	}
	return c
}

func (c *configure) Clone() Config {
	n := new(configure)
	n.wsopts = append(n.wsopts, c.wsopts...)
	n.tcpopts = append(n.tcpopts, c.tcpopts...)
	n.grpcopts = append(n.grpcopts, c.grpcopts...)
	n.natsopts = append(n.natsopts, c.natsopts...)
	n.httpopts = append(n.httpopts, c.httpopts...)
	return n
}

func (c *configure) Add(opt ...Option) Config {
	for _, o := range opt {
		o(c)
	}
	return c
}

// WSOptions .
func (c *configure) WSOptions() []xws.Option {
	return c.wsopts
}

// TCPOptions .
func (c *configure) TCPOptions() []xtcp.Option {
	return c.tcpopts
}

// GRPCOptions .
func (c *configure) GRPCOptions() []xgrpc.Option {
	return c.grpcopts
}

// NATSOptions .
func (c *configure) NATSOptions() []xnats.Option {
	return c.natsopts
}

// HTTPOptions .
func (c *configure) HTTPOptions() []xhttp.Option {
	return c.httpopts
}

// LogOptions .
func (c *configure) LogOptions() []log.Option {
	return c.logopts
}

// NamingOptions .
func (c *configure) NamingOptions() []naming.Option {
	return c.namingopts
}
