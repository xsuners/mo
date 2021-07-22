package mo

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/naming"
	"github.com/xsuners/mo/net/description"
	"github.com/xsuners/mo/net/util/ip"
	"github.com/xsuners/mo/net/xgrpc"
	"github.com/xsuners/mo/net/xhttp"
	"github.com/xsuners/mo/net/xnats"
	"github.com/xsuners/mo/net/xtcp"
	"github.com/xsuners/mo/net/xws"
)

// App
type App struct {
	service interface{}

	ws   *xws.Server
	tcp  *xtcp.Server
	grpc *xgrpc.Server
	http *xhttp.Server
	nats *xnats.Server

	log    *log.Log
	naming naming.Naming

	cs   []func()
	logc func()
}

func sport(p int) string {
	return fmt.Sprintf(":%d", p)
}

func Serve(service interface{}, cf func(), opt ...Option) (err error) {
	app := &App{
		service: service,
	}
	for _, o := range opt {
		o(app)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			for _, c := range app.cs {
				c()
			}
			if cf != nil {
				cf()
			}
			if app.logc != nil {
				app.logc()
			}
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

type Option func(*App)

// Log .
func Log(opts ...log.Option) Option {
	return func(app *App) {
		app.log, app.logc = log.New(opts...)
	}
}

// Naming .
func Naming(n naming.Naming) Option {
	return func(app *App) {
		// n, c, err := naming.New(opts...)
		// if err != nil {
		// 	panic(err)
		// }
		app.naming = n
		// app.cs = append(app.cs, c)
	}
}

// WSServer .
func WSServer(port int, sds []*description.ServiceDesc, opts ...xws.Option) Option {
	return func(app *App) {
		var c func()
		app.ws, c = xws.New(opts...)
		app.cs = append(app.cs, c)
		app.ws.Register(app.service, sds...)
		l, err := net.Listen("tcp", sport(port))
		if err != nil {
			panic(err)
		}
		go func() {
			if err := app.ws.Serve(l); err != nil {
				panic(err)
			}
		}()
		if app.naming == nil {
			log.Infos("naming is nil")
			return
		}
		for _, service := range sds {
			ins := &naming.Service{
				Name:     service.ServiceName,
				Protocol: naming.WS,
				IP:       ip.Internal(),
				Port:     port,
				Tag:      []string{"ws"},
			}
			if err := app.naming.Register(ins); err != nil {
				panic(err)
			}
		}
	}
}

// TCPServer .
func TCPServer(port int, sds []*description.ServiceDesc, opts ...xtcp.Option) Option {
	return func(app *App) {
		var c func()
		app.tcp, c = xtcp.New(opts...)
		app.cs = append(app.cs, c)
		app.tcp.Register(app.service, sds...)
		l, err := net.Listen("tcp", sport(port))
		if err != nil {
			panic(err)
		}
		go func() {
			if err := app.tcp.Serve(l); err != nil {
				panic(err)
			}
		}()
		if app.naming == nil {
			log.Infos("naming is nil")
			return
		}
		for _, service := range sds {
			ins := &naming.Service{
				Name:     service.ServiceName,
				Protocol: naming.TCP,
				IP:       ip.Internal(),
				Port:     port,
				Tag:      []string{"tcp"},
			}
			if err := app.naming.Register(ins); err != nil {
				panic(err)
			}
		}
	}
}

// NATSServer .
func NATSServer(sds []*description.ServiceDesc, opts ...xnats.Option) Option {
	return func(app *App) {
		var err error
		var c func()
		app.nats, c, err = xnats.New(opts...)
		if err != nil {
			panic(err)
		}
		app.cs = append(app.cs, c)
		app.nats.Register(app.service, sds...)
		go func() {
			if err := app.nats.Serve(); err != nil {
				panic(err)
			}
		}()
	}
}

// HTTPServer .
func HTTPServer(port int, sds []*description.ServiceDesc, opts ...xhttp.Option) Option {
	return func(app *App) {
		var c func()
		app.http, c = xhttp.New(opts...)
		app.cs = append(app.cs, c)
		go func() {
			if err := app.http.Serve(port); err != nil {
				panic(err)
			}
		}()
		if app.naming == nil {
			log.Infos("naming is nil")
			return
		}
		for _, service := range sds {
			ins := &naming.Service{
				Name:     service.ServiceName,
				Protocol: naming.HTTP,
				IP:       ip.Internal(),
				Port:     port,
				Tag:      []string{"http"},
			}
			if err := app.naming.Register(ins); err != nil {
				panic(err)
			}
		}
	}
}

// GRPCServer .
func GRPCServer(port int, sds []*description.ServiceDesc, opts ...xgrpc.Option) Option {
	return func(app *App) {
		var c func()
		app.grpc, c = xgrpc.New(opts...)
		app.cs = append(app.cs, c)
		app.grpc.Register(app.service, sds...)
		l, err := net.Listen("tcp", sport(port))
		if err != nil {
			panic(err)
		}
		go func() {
			if err := app.grpc.Serve(l); err != nil {
				panic(err)
			}
		}()
		if app.naming == nil {
			log.Infos("naming is nil")
			return
		}
		for _, service := range sds {
			ins := &naming.Service{
				Name:     service.ServiceName,
				Protocol: naming.GRPC,
				IP:       ip.Internal(),
				Port:     port,
				Tag:      []string{"grpc"},
			}
			if err = app.naming.Register(ins); err != nil {
				panic(err)
			}
		}
	}
}

func Profile(port int) Option {
	return func(a *App) {
		runtime.GOMAXPROCS(1)
		runtime.SetMutexProfileFraction(1)
		runtime.SetBlockProfileRate(1)
		go func() {
			if err := http.ListenAndServe(sport(port), nil); err != nil {
				log.Fatal(err)
			}
		}()
	}
}
