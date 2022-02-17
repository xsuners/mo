package mo

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/naming"
	"github.com/xsuners/mo/net/description"
)

type ssd struct {
	svc interface{}
	sds []*description.ServiceDesc
}

type server struct {
	server description.Server
	cf     func()
	ssds   []*ssd
}

// App
type App struct {
	servers []*server

	log    *log.Log
	naming naming.Naming

	cf   func()
	logc func()
}

func New(cf func(), opt ...Option) *App {
	app := &App{
		cf: cf,
	}
	for _, o := range opt {
		o(app)
	}
	return app
}

// Serve .
func (app *App) Serve() (err error) {
	for _, s := range app.servers {
		for _, sd := range s.ssds {
			s.server.Register(sd.svc, sd.sds...)
		}
		go func(s description.Server) {
			if err := s.Serve(); err != nil {
				panic(err)
			}
		}(s.server)
		if app.naming == nil {
			log.Infos("naming is nil")
			return
		}
		if err := s.server.Naming(app.naming); err != nil {
			panic(err)
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	for {
		s := <-c
		log.Infof("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			for _, s := range app.servers {
				s.cf()
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
