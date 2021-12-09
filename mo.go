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
	"github.com/xsuners/mo/misc/ip"
	"github.com/xsuners/mo/naming"
	"github.com/xsuners/mo/net/description"
	"github.com/xsuners/mo/net/xgrpc"
	"github.com/xsuners/mo/net/xhttp"
	"github.com/xsuners/mo/net/xnats"
	"github.com/xsuners/mo/net/xtcp"
	"github.com/xsuners/mo/net/xws"
)

func sport(p int) string {
	return fmt.Sprintf(":%d", p)
}

// App
type App struct {
	wsport   int
	tcpport  int
	httpport int
	grpcport int

	service interface{}

	ws   *xws.Server
	tcp  *xtcp.Server
	grpc *xgrpc.Server
	http *xhttp.Server
	nats *xnats.Server

	log    *log.Log
	naming naming.Naming

	cs   []func()
	cf   func()
	logc func()
}

func New(service interface{}, cf func(), opt ...Option) *App {
	app := &App{
		wsport:   5000,
		tcpport:  6000,
		httpport: 8000,
		grpcport: 9000,
		service:  service,
		cf:       cf,
	}
	for _, o := range opt {
		o(app)
	}
	return app
}

func (app *App) Serve() (err error) {
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
			if app.cf != nil {
				app.cf()
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

// WS .
func (app *App) WS(opts []xws.Option, sds ...*description.ServiceDesc) {
	var c func()
	app.ws, c = xws.New(opts...)
	app.cs = append(app.cs, c)
	app.ws.Register(app.service, sds...)
	l, err := net.Listen("tcp", sport(app.wsport))
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
			Port:     app.wsport,
			Tag:      []string{"ws"},
		}
		if err := app.naming.Register(ins); err != nil {
			panic(err)
		}
	}
}

// TCP .
func (app *App) TCP(opts []xtcp.Option, sds ...*description.ServiceDesc) {
	var c func()
	app.tcp, c = xtcp.New(opts...)
	app.cs = append(app.cs, c)
	app.tcp.Register(app.service, sds...)
	l, err := net.Listen("tcp", sport(app.tcpport))
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
			Port:     app.tcpport,
			Tag:      []string{"tcp"},
		}
		if err := app.naming.Register(ins); err != nil {
			panic(err)
		}
	}
}

// NATS .
func (app *App) NATS(opts []xnats.Option, sds ...*description.ServiceDesc) {
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

// HTTP .
func (app *App) HTTP(opts []xhttp.Option, sds ...*description.ServiceDesc) {
	var c func()
	app.http, c = xhttp.New(opts...)
	app.cs = append(app.cs, c)
	go func() {
		if err := app.http.Serve(app.httpport); err != nil {
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
			Port:     app.httpport,
			Tag:      []string{"http"},
		}
		if err := app.naming.Register(ins); err != nil {
			panic(err)
		}
	}
}

// GRPC .
func (app *App) GRPC(opts []xgrpc.Option, sds ...*description.ServiceDesc) {
	var c func()
	app.grpc, c = xgrpc.New(opts...)
	app.cs = append(app.cs, c)
	app.grpc.Register(app.service, sds...)
	l, err := net.Listen("tcp", sport(app.grpcport))
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
			Port:     app.grpcport,
			Tag:      []string{"grpc"},
		}
		if err = app.naming.Register(ins); err != nil {
			panic(err)
		}
	}
}

func (app *App) Profile(port int) {
	runtime.GOMAXPROCS(1)
	runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)
	go func() {
		if err := http.ListenAndServe(sport(port), nil); err != nil {
			log.Fatal(err)
		}
	}()
}
