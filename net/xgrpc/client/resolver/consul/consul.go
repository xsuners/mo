package consul

import (
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/resolver"
)

const (
	defaultPort = "8500"
)

var (
	errMissingAddr  = errors.New("consul resolver: missing address")
	errAddrMisMatch = errors.New("consul resolver: invalied uri")
	regexConsul, _  = regexp.Compile("^([A-z0-9.]+)(:[0-9]{1,5})?/([A-z0-9._]+)$")
)

// Init .
func Init() {
	// log.Infof("calling consul init")
	resolver.Register(NewBuilder())
}

type consulBuilder struct{}

// NewBuilder .
func NewBuilder() resolver.Builder {
	return &consulBuilder{}
}

func (cb *consulBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	host, port, name, err := parseTarget(fmt.Sprintf("%s/%s", target.Authority, target.Endpoint))
	if err != nil {
		return nil, err
	}
	cr := &consulResolver{
		address:              fmt.Sprintf("%s%s", host, port),
		name:                 name,
		cc:                   cc,
		disableServiceConfig: opts.DisableServiceConfig,
		lastIndex:            0,
	}
	cr.wg.Add(1)
	go cr.watcher()
	return cr, nil

}

func (cb *consulBuilder) Scheme() string {
	return "consul"
}

type consulResolver struct {
	address              string
	wg                   sync.WaitGroup
	cc                   resolver.ClientConn
	name                 string
	disableServiceConfig bool
	lastIndex            uint64
}

func (cr *consulResolver) watcher() {
	config := api.DefaultConfig()
	config.Address = cr.address
	client, err := api.NewClient(config)
	if err != nil {
		fmt.Printf("error create consul client: %v", err)
		return
	}
	for {
		services, metainfo, err := client.Health().Service(cr.name, "", true, &api.QueryOptions{WaitIndex: cr.lastIndex})
		if err != nil {
			fmt.Printf("error retrieving instances from Consul: %v", err)
			// TODO 更好的重试机制
			time.Sleep(time.Second)
			continue
		}

		cr.lastIndex = metainfo.LastIndex
		var newAddrs []resolver.Address
		for _, service := range services {
			addr := fmt.Sprintf("%v:%v", service.Service.Address, service.Service.Port)
			newAddrs = append(newAddrs, resolver.Address{Addr: addr})
		}
		// log.Info("adding service addrs")
		// log.Infow("service addrs", "addrs", newAddrs)
		// cr.cc.NewAddress(newAddrs)
		// cr.cc.NewServiceConfig(cr.name)
		cr.cc.UpdateState(resolver.State{Addresses: newAddrs})
	}
}

func (cr *consulResolver) ResolveNow(opt resolver.ResolveNowOptions) {}

func (cr *consulResolver) Close() {}

func parseTarget(target string) (host, port, name string, err error) {

	// log.Infof("target uri: %v", target)
	if target == "" {
		return "", "", "", errMissingAddr
	}

	if !regexConsul.MatchString(target) {
		return "", "", "", errAddrMisMatch
	}

	groups := regexConsul.FindStringSubmatch(target)
	host = groups[1]
	port = groups[2]
	name = groups[3]
	if port == "" {
		port = defaultPort
	}
	return host, port, name, nil
}
