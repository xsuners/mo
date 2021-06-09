package naming

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/xsuners/mo/log"
	"go.uber.org/zap"
)

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

// Service .
type Service struct {
	Name string
	IP   string
	Port int
	Tag  []string
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
		nm.deregister()
		log.Infos("naming is closed.")
	}
	return
}

// // Close .
// func Close() {
// 	if nm == nil {
// 		return
// 	}
// 	nm.deregister()
// }

func (n *Naming) deregister() {
	for _, svc := range n.services {
		err := n.client.Agent().ServiceDeregister(serviceID(svc))
		if err != nil {
			log.Errorw("service deregister error", "err", err)
		} else {
			log.Infow("service deregisted", "service", svc.Name)
		}
	}
}

func (n *Naming) register(svc *Service) (err error) {
	reg := &api.AgentServiceRegistration{
		ID:      serviceID(svc),
		Name:    svc.Name,
		Tags:    svc.Tag,
		Address: svc.IP,
		Port:    svc.Port,
	}
	interval := time.Duration(10) * time.Second
	deregister := time.Duration(1) * time.Minute
	reg.Check = &api.AgentServiceCheck{
		GRPC:                           fmt.Sprintf("%v:%v/%v", svc.IP, svc.Port, svc.Name),
		Interval:                       interval.String(),
		DeregisterCriticalServiceAfter: deregister.String(),
	}
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

// Regitser .
func (n *Naming) Regitser(svc *Service) (err error) {
	return n.register(svc)
}
