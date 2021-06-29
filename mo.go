package mo

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/log/extractor"
	"github.com/xsuners/mo/naming"
	"github.com/xsuners/mo/net/description"
	"github.com/xsuners/mo/net/util/interceptor"
	"github.com/xsuners/mo/net/xgrpc"
	"github.com/xsuners/mo/net/xhttp"
	"github.com/xsuners/mo/net/xnats"
	"github.com/xsuners/mo/net/xtcp"
	"github.com/xsuners/mo/net/xws"
	"github.com/xsuners/mo/util/ip"
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
	naming *naming.Naming

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
		opts = append(opts, log.WithExtractor(extractor.MDExtractor{}))
		app.log, app.logc = log.New(opts...)
	}
}

// Naming .
func Naming(opts ...naming.Option) Option {
	return func(app *App) {
		n, c, err := naming.New(opts...)
		if err != nil {
			panic(err)
		}
		app.naming = n
		app.cs = append(app.cs, c)
	}
}

// WSServer .
func WSServer(port int, sds []*description.ServiceDesc, opts ...xws.Option) Option {
	return func(app *App) {
		s, c := xws.New(opts...)
		app.cs = append(app.cs, c)
		s.Register(app.service, sds...)
		l, err := net.Listen("tcp", sport(port))
		if err != nil {
			panic(err)
		}
		go func() {
			if err := s.Serve(l); err != nil {
				panic(err)
			}
		}()
		if app.naming == nil {
			panic("ws server need naming")
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
		app.ws = s
	}
}

// TCPServer .
func TCPServer(port int, sds []*description.ServiceDesc, opts ...xtcp.Option) Option {
	return func(app *App) {
		s, c := xtcp.New(opts...)
		app.cs = append(app.cs, c)
		s.Register(app.service, sds...)
		l, err := net.Listen("tcp", sport(port))
		if err != nil {
			panic(err)
		}
		go func() {
			if err := s.Serve(l); err != nil {
				panic(err)
			}
		}()
		if app.naming == nil {
			panic("tcp server need naming")
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
		app.tcp = s
	}
}

// NATSServer .
func NATSServer(sds []*description.ServiceDesc, opts ...xnats.Option) Option {
	return func(app *App) {
		opts = append(opts, xnats.UnaryInterceptor(interceptor.MetaServerInterceptor()))
		s, c, err := xnats.New(opts...)
		if err != nil {
			panic(err)
		}
		app.cs = append(app.cs, c)
		s.Register(app.service, sds...)
		go func() {
			if err := s.Serve(); err != nil {
				panic(err)
			}
		}()
		app.nats = s
	}
}

// HTTPServer .
func HTTPServer(port int, sds []*description.ServiceDesc, opts ...xhttp.Option) Option {
	return func(app *App) {
		s, c := xhttp.New(opts...)
		app.cs = append(app.cs, c)
		go func() {
			if err := s.Serve(port); err != nil {
				panic(err)
			}
		}()
		if app.naming == nil {
			panic("http server need naming")
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
		app.http = s
	}
}

// GRPCServer .
func GRPCServer(port int, sds []*description.ServiceDesc, opts ...xgrpc.Option) Option {
	return func(app *App) {
		opts = append(opts, xgrpc.UnaryInterceptor(interceptor.MetaServerInterceptor()))
		s, c := xgrpc.New(opts...)
		app.cs = append(app.cs, c)
		s.Register(app.service, sds...)
		l, err := net.Listen("tcp", sport(port))
		if err != nil {
			panic(err)
		}
		go func() {
			if err := s.Serve(l); err != nil {
				panic(err)
			}
		}()
		if app.naming == nil {
			panic("grpc server need naming")
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
		app.grpc = s
	}
}
