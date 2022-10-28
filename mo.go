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

type App interface {
	Serve() error
}

// app
type app struct {
	servers []*server

	log    *log.Log
	naming naming.Naming

	cf   func()
	logc func()
}

func New(cf func(), opt ...Option) App {
	a := &app{
		cf: cf,
	}
	for _, o := range opt {
		o(a)
	}
	return a
}

// Serve .
func (a *app) Serve() (err error) {
	for _, s := range a.servers {
		for _, sd := range s.ssds {
			s.server.Register(sd.svc, sd.sds...)
		}
		go func(s description.Server) {
			if err := s.Serve(); err != nil {
				panic(err)
			}
		}(s.server)
		if a.naming == nil {
			log.Infos("naming is nil")
			return
		}
		time.Sleep(time.Second) // TODO 更优雅点
		if err := s.server.Naming(a.naming); err != nil {
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
			if a.naming != nil {
				a.naming.Deregister()
			}
			for _, s := range a.servers {
				s.cf()
			}
			if a.cf != nil {
				a.cf()
			}
			if a.logc != nil {
				a.logc()
			}
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
