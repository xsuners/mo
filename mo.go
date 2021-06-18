package mo

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/naming"
	"github.com/xsuners/mo/net/description"
	"github.com/xsuners/mo/net/xgrpc"
	"github.com/xsuners/mo/net/xhttp"
	"github.com/xsuners/mo/net/xnats"
	"github.com/xsuners/mo/net/xtcp"
	"github.com/xsuners/mo/net/xws"
	"github.com/xsuners/mo/util/ip"
)

// App
type App struct {
	// wsport   int
	// tcpport  int
	// grpcport int
	// httpport int

	// wsoptions   []xws.Option
	// tcpoptions  []xtcp.Option
	// grpcoptions []xgrpc.Option
	// httpoptions []xhttp.Option
	// natsoptions []xnats.Option

	// wsservices   []*description.ServiceDesc // TODO 重命名
	// tcpservices  []*description.ServiceDesc
	// grpcservices []*description.ServiceDesc
	// httpservices []*description.ServiceDesc
	// natsservices []*description.ServiceDesc

	ws   *xws.Server
	tcp  *xtcp.Server
	grpc *xgrpc.Server
	http *xhttp.Server
	nats *xnats.Server

	log    *log.Log
	naming *naming.Naming

	cs   []func()
	logc func()

	service interface{}
}

func sport(p int) string {
	return fmt.Sprintf(":%d", p)
}

func New(service interface{}, opt ...Option) (app *App, cf func()) {
	app = &App{
		service: service,
	}
	for _, o := range opt {
		o(app)
	}
	cf = func() {
		if app.logc != nil {
			app.logc()
		}
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
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

type Option func(*App)

func WithLog(opts ...log.Option) Option {
	return func(app *App) {
		app.log, app.logc = log.New(opts...)
	}
}

func WithNaming(opts ...naming.Option) Option {
	return func(app *App) {
		n, c, err := naming.New(opts...)
		if err != nil {
			panic(err)
		}
		app.naming = n
		app.cs = append(app.cs, c)
	}
}

func WSServer(port int, sds []*description.ServiceDesc, opts ...xws.Option) Option {
	return func(app *App) {
		// app.wsport = port
		// app.wsservices = append(app.wsservices, sds...)
		// app.wsoptions = append(app.wsoptions, opts...)
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
		app.ws = s
	}
}

func TCPServer(port int, sds []*description.ServiceDesc, opts ...xtcp.Option) Option {
	return func(app *App) {
		// app.tcpport = port
		// app.tcpservices = append(app.tcpservices, sds...)
		// app.tcpoptions = append(app.tcpoptions, opts...)
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
		app.tcp = s
	}
}

func NATSServer(sds []*description.ServiceDesc, opts ...xnats.Option) Option {
	return func(app *App) {
		// app.natsservices = append(app.natsservices, sds...)
		// app.natsoptions = append(app.natsoptions, opts...)
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

func GRPCServer(port int, sds []*description.ServiceDesc, opts ...xgrpc.Option) Option {
	return func(app *App) {
		// app.grpcport = port
		// app.grpcservices = append(app.grpcservices, sds...)
		// app.grpcoptions = append(app.grpcoptions, opts...)
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
				Name: service.ServiceName,
				IP:   ip.Internal(),
				Port: port,
				Tag:  []string{"grpc"},
			}
			if err = app.naming.Regitser(ins); err != nil {
				log.Errorw("xtcp: register service error", "err", err)
				return
			}
		}
		app.grpc = s
	}
}

func HTTPServer(port int, sds []*description.ServiceDesc, opts ...xhttp.Option) Option {
	return func(app *App) {
		// app.httpport = port
		// app.httpservices = append(app.httpservices, sds...)
		// app.httpoptions = append(app.httpoptions, opts...)
		s, c := xhttp.New(opts...)
		app.cs = append(app.cs, c)
		go func() {
			if err := s.Serve(port); err != nil {
				panic(err)
			}
		}()
		app.http = s
	}
}
