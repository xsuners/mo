package consul

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/xsuners/mo/log"
	"go.uber.org/zap"
)

type Protocol string

const (
	WS    Protocol = "ws"
	WSS   Protocol = "wss"
	TCP   Protocol = "tcp"
	TCPS  Protocol = "tcps"
	HTTP  Protocol = "http"
	HTTPS Protocol = "https"
	GRPC  Protocol = "grpc"
	GRPCS Protocol = "grpcs"
)

// Service .
type Service struct {
	// Check    bool
	Name     string
	Protocol Protocol
	IP       string
	Port     int
	Tag      []string
}

type options struct {
	ip   string
	port int
}

var defaultOptions = options{
	ip:   "127.0.0.1",
	port: 8500,
}

// Option .
type Option func(*options)

// IP .
func IP(ip string) Option {
	return func(o *options) {
		o.ip = ip
	}
}

// Port .
func Port(port int) Option {
	return func(o *options) {
		o.port = port
	}
}

type Naming struct {
	opts     options
	client   *api.Client
	services []*Service
}

// var nm *naming

// TODO
func New(opt ...Option) (nm *Naming, cf func(), err error) {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	cfg := api.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d", opts.ip, opts.port)
	client, err := api.NewClient(cfg)
	if err != nil {
		log.Fatals("new consul client error", zap.Error(err))
		return
	}
	nm = &Naming{
		opts:   opts,
		client: client,
	}
	cf = func() {
		log.Infos("naming is closing...")
		log.Infos("naming is closed.")
	}
	return
}

func (n *Naming) Deregister() {
	for _, svc := range n.services {
		err := n.client.Agent().ServiceDeregister(serviceID(svc))
		if err != nil {
			log.Errorw("service deregister error", "err", err)
		} else {
			log.Infow("service deregisted", "service", svc.Name)
		}
	}
}

func (n *Naming) Register(svc *Service) (err error) {
	reg := &api.AgentServiceRegistration{
		ID:      serviceID(svc),
		Name:    svc.Name,
		Tags:    svc.Tag,
		Address: svc.IP,
		Port:    svc.Port,
	}
	interval := time.Duration(10) * time.Second
	deregister := time.Duration(1) * time.Minute
	// if svc.Check {
	// TODO handle tls
	reg.Check = &api.AgentServiceCheck{
		Interval:                       interval.String(),
		DeregisterCriticalServiceAfter: deregister.String(),
	}
	switch svc.Protocol {
	case WS, WSS, TCP, TCPS:
		reg.Check.TCP = fmt.Sprintf("%v:%v", svc.IP, svc.Port)
	case HTTP, HTTPS:
		reg.Check.HTTP = fmt.Sprintf("%s://%v:%v", svc.Protocol, svc.IP, svc.Port)
	case GRPC, GRPCS:
		reg.Check.GRPC = fmt.Sprintf("%v:%v/%v", svc.IP, svc.Port, svc.Name)
	default:
		err = fmt.Errorf("unknown protocol %s", svc.Protocol)
		log.Errors("naming:register", zap.Error(err))
		return
	}
	// }
	if err = n.client.Agent().ServiceRegister(reg); err != nil {
		log.Errorw("register service error", "err", err)
		return
	}
	log.Infow("service registed", "service", svc.Name)
	n.services = append(n.services, svc)
	return
}

func serviceID(svc *Service) string {
	return fmt.Sprintf("%v-%v-%v", svc.Name, svc.IP, svc.Port)
}
