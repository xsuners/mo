package mo

import (
	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/naming"
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

func PHTTP(port int) Option {
	return func(a *App) {
		a.httpport = port
	}
}

func PGRPC(port int) Option {
	return func(a *App) {
		a.grpcport = port
	}
}

func PTCP(port int) Option {
	return func(a *App) {
		a.tcpport = port
	}
}

func PWS(port int) Option {
	return func(a *App) {
		a.wsport = port
	}
}
