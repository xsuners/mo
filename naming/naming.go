package naming

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/xsuners/mo/log"
	"go.uber.org/zap"
)

// Config .
type Config struct {
	IP   string
	Port int
}

// Service .
type Service struct {
	Name string
	IP   string
	Port int
	Tag  []string
}

type naming struct {
	client   *api.Client
	services []*Service
}

var nm *naming

// TODO
func init() {
	c := &Config{
		IP:   "127.0.0.1",
		Port: 8500,
	}
	cfg := api.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d", c.IP, c.Port)
	client, err := api.NewClient(cfg)
	if err != nil {
		log.Fatals("new consul client error", zap.Error(err))
		return
	}
	nm = &naming{client: client}
}

// Close .
func Close() {
	if nm == nil {
		return
	}
	nm.deregister()
}

func (n *naming) deregister() {
	for _, svc := range nm.services {
		err := n.client.Agent().ServiceDeregister(serviceID(svc))
		if err != nil {
			log.Errorw("service deregister error", "err", err)
		} else {
			log.Infow("service deregisted", "service", svc.Name)
		}
	}
}

func (n *naming) register(svc *Service) (err error) {
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
func Regitser(svc *Service) (err error) {
	return nm.register(svc)
}
