package mo

import (
	"fmt"
	"net"
	"os"
	"os/signal"
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

type ssd struct {
	server interface{}
	sds    []*description.ServiceDesc
}

// App
type App struct {
	wsport   int
	tcpport  int
	httpport int
	grpcport int

	ws   *xws.Server
	tcp  *xtcp.Server
	grpc *xgrpc.Server
	http *xhttp.Server
	nats *xnats.Server

	wssds   []*ssd
	tcpsds  []*ssd
	grpcsds []*ssd
	httpsds []*ssd
	natssds []*ssd

	log    *log.Log
	naming naming.Naming

	cs   []func()
	cf   func()
	logc func()
}

func New(cf func(), opt ...Option) *App {
	app := &App{
		wsport:   5000,
		tcpport:  6000,
		httpport: 8000,
		grpcport: 9000,
		cf:       cf,
	}
	for _, o := range opt {
		o(app)
	}
	return app
}

func (app *App) Serve() (err error) {
	app.run()
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

func (app *App) run() {
	if app.ws != nil {
		for _, sd := range app.wssds {
			app.ws.Register(sd.server, sd.sds...)
		}
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
		for _, sd := range app.wssds {
			for _, service := range sd.sds {
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
	}

	if app.tcp != nil {
		for _, sd := range app.tcpsds {
			app.tcp.Register(sd.server, sd.sds...)
		}
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
		for _, sd := range app.tcpsds {
			for _, service := range sd.sds {
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
	}

	if app.nats != nil {
		for _, sd := range app.natssds {
			app.nats.Register(sd.server, sd.sds...)
		}
		go func() {
			if err := app.nats.Serve(); err != nil {
				panic(err)
			}
		}()
	}

	if app.http != nil {
		go func() {
			if err := app.http.Serve(app.httpport); err != nil {
				panic(err)
			}
		}()
		if app.naming == nil {
			log.Infos("naming is nil")
			return
		}
		for _, sd := range app.httpsds {
			for _, service := range sd.sds {
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
	}

	if app.grpc != nil {
		for _, sd := range app.grpcsds {
			app.grpc.Register(sd.server, sd.sds...)
		}
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
		for _, sd := range app.grpcsds {
			for _, service := range sd.sds {
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
	}
}
