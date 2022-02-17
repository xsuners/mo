package mo

import (
	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/naming"
	"github.com/xsuners/mo/net/description"
	"github.com/xsuners/mo/net/xgrpc"
	"github.com/xsuners/mo/net/xhttp"
	"github.com/xsuners/mo/net/xnats"
	"github.com/xsuners/mo/net/xtcp"
	"github.com/xsuners/mo/net/xws"
)

type Option func(*App)

// Log .
func Log(opts []log.Option) Option {
	return func(app *App) {
		app.log, app.logc = log.New(opts...)
	}
}

// Naming .
func Naming(n naming.Naming) Option {
	return func(app *App) {
		app.naming = n
	}
}

func WSSDS(svc interface{}, sds ...*description.ServiceDesc) Option {
	return func(app *App) {
		for _, s := range app.servers {
			if _, ok := s.server.(*xws.Server); ok {
				s.ssds = append(s.ssds, &ssd{
					svc: svc,
					sds: sds,
				})
				return
			}
		}
		panic("xws not exist")
	}
}

func TCPSDS(svc interface{}, sds ...*description.ServiceDesc) Option {
	return func(app *App) {
		for _, s := range app.servers {
			if _, ok := s.server.(*xtcp.Server); ok {
				s.ssds = append(s.ssds, &ssd{
					svc: svc,
					sds: sds,
				})
				return
			}
		}
		panic("xtcp not exist")
	}
}

func GRPCSDS(svc interface{}, sds ...*description.ServiceDesc) Option {
	return func(app *App) {
		for _, s := range app.servers {
			if _, ok := s.server.(*xgrpc.Server); ok {
				s.ssds = append(s.ssds, &ssd{
					svc: svc,
					sds: sds,
				})
				return
			}
		}
		panic("xgrpc not exist")
	}
}

func HTTPSDS(svc interface{}, sds ...*description.ServiceDesc) Option {
	return func(app *App) {
		for _, s := range app.servers {
			if _, ok := s.server.(*xhttp.Server); ok {
				s.ssds = append(s.ssds, &ssd{
					svc: svc,
					sds: sds,
				})
				return
			}
		}
		panic("xhttp not exist")
	}
}

func NATSSDS(svc interface{}, sds ...*description.ServiceDesc) Option {
	return func(app *App) {
		for _, s := range app.servers {
			if _, ok := s.server.(*xnats.Server); ok {
				s.ssds = append(s.ssds, &ssd{
					svc: svc,
					sds: sds,
				})
				return
			}
		}
		panic("xnats not exist")
	}
}

func WS(opts ...xws.Option) Option {
	return func(app *App) {
		for _, s := range app.servers {
			if _, ok := s.server.(*xws.Server); ok {
				panic("server exists")
			}
		}
		s := new(server)
		s.server, s.cf = xws.New(opts...)
		app.servers = append(app.servers, s)
	}
}

func TCP(opts ...xtcp.Option) Option {
	return func(app *App) {
		for _, s := range app.servers {
			if _, ok := s.server.(*xtcp.Server); ok {
				panic("server exists")
			}
		}
		s := new(server)
		s.server, s.cf = xtcp.New(opts...)
		app.servers = append(app.servers, s)
	}
}

func NATS(opts ...xnats.Option) Option {
	return func(app *App) {
		for _, s := range app.servers {
			if _, ok := s.server.(*xnats.Server); ok {
				panic("server exists")
			}
		}
		var err error
		s := new(server)
		s.server, s.cf, err = xnats.New(opts...)
		if err != nil {
			panic(err)
		}
		app.servers = append(app.servers, s)
	}
}

func HTTP(opts ...xhttp.Option) Option {
	return func(app *App) {
		for _, s := range app.servers {
			if _, ok := s.server.(*xhttp.Server); ok {
				panic("server exists")
			}
		}
		s := new(server)
		s.server, s.cf = xhttp.New(opts...)
		app.servers = append(app.servers, s)
	}
}

func GRPC(opts ...xgrpc.Option) Option {
	return func(app *App) {
		for _, s := range app.servers {
			if _, ok := s.server.(*xgrpc.Server); ok {
				panic("server exists")
			}
		}
		s := new(server)
		s.server, s.cf = xgrpc.New(opts...)
		app.servers = append(app.servers, s)
	}
}
