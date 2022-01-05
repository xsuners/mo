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

func WSSDS(s interface{}, sds ...*description.ServiceDesc) Option {
	return func(a *App) {
		a.wssds = append(a.wssds, &ssd{
			server: s,
			sds:    sds,
		})
	}
}

func TCPSDS(s interface{}, sds ...*description.ServiceDesc) Option {
	return func(a *App) {
		a.tcpsds = append(a.tcpsds, &ssd{
			server: s,
			sds:    sds,
		})
	}
}

func GRPCSDS(s interface{}, sds ...*description.ServiceDesc) Option {
	return func(a *App) {
		a.grpcsds = append(a.grpcsds, &ssd{
			server: s,
			sds:    sds,
		})
	}
}

func HTTPSDS(s interface{}, sds ...*description.ServiceDesc) Option {
	return func(a *App) {
		a.httpsds = append(a.httpsds, &ssd{
			server: s,
			sds:    sds,
		})
	}
}

func NATSSDS(s interface{}, sds ...*description.ServiceDesc) Option {
	return func(a *App) {
		a.natssds = append(a.natssds, &ssd{
			server: s,
			sds:    sds,
		})
	}
}

func WS(port int, opts ...xws.Option) Option {
	return func(app *App) {
		var c func()
		if port < 1 {
			app.wsport = 5000
		} else {
			app.wsport = port
		}
		app.ws, c = xws.New(opts...)
		app.cs = append(app.cs, c)
	}
}

func TCP(port int, opts ...xtcp.Option) Option {
	return func(app *App) {
		var c func()
		if port < 1 {
			app.tcpport = 6000
		} else {
			app.tcpport = port
		}
		app.tcp, c = xtcp.New(opts...)
		app.cs = append(app.cs, c)
	}
}

func NATS(opts ...xnats.Option) Option {
	return func(app *App) {
		var err error
		var c func()
		app.nats, c, err = xnats.New(opts...)
		if err != nil {
			panic(err)
		}
		app.cs = append(app.cs, c)
	}
}

func HTTP(port int, opts ...xhttp.Option) Option {
	return func(app *App) {
		var c func()
		if port < 1 {
			app.httpport = 8000
		} else {
			app.httpport = port
		}
		app.http, c = xhttp.New(opts...)
		app.cs = append(app.cs, c)
	}
}

func GRPC(port int, opts ...xgrpc.Option) Option {
	return func(app *App) {
		var c func()
		if port < 1 {
			app.grpcport = 9000
		} else {
			app.grpcport = port
		}
		app.grpc, c = xgrpc.New(opts...)
		app.cs = append(app.cs, c)
	}
}
