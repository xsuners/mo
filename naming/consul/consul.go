package consul

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/naming"
	"go.uber.org/zap"
)

type Options struct {
	Host string `json:"host" ini-name:"host" long:"consul-host" description:"consul host"`
	Port int    `json:"port" ini-name:"port" long:"consul-port" description:"consul port"`
}

var defaultOptions = Options{
	Host: "127.0.0.1",
	Port: 8500,
}

// Option .
type Option func(*Options)

// IP .
func IP(ip string) Option {
	return func(o *Options) {
		o.Host = ip
	}
}

// Port .
func Port(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

type Naming struct {
	opts     Options
	client   *api.Client
	services []*naming.Service
}

var _ naming.Naming = (*Naming)(nil)

// TODO
func New(opt ...Option) (nm naming.Naming) {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	cfg := api.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	client, err := api.NewClient(cfg)
	if err != nil {
		log.Fatals("new consul client error", zap.Error(err))
	}
	nm = &Naming{
		opts:   opts,
		client: client,
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

func (n *Naming) Register(svc *naming.Service) (err error) {
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
	case naming.WS, naming.WSS, naming.TCP, naming.TCPS:
		reg.Check.TCP = fmt.Sprintf("%v:%v", svc.IP, svc.Port)
	case naming.HTTP, naming.HTTPS:
		reg.Check.HTTP = fmt.Sprintf("%s://%v:%v", svc.Protocol, svc.IP, svc.Port)
	case naming.GRPC, naming.GRPCS:
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

func serviceID(svc *naming.Service) string {
	return fmt.Sprintf("%v-%v-%v", svc.Name, svc.IP, svc.Port)
}
