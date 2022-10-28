package consul

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/xsuners/mo/log"
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
	// host, port, name, err := parseTarget(fmt.Sprintf("%s/%s", target.Authority, target.Endpoint))
	host, port, name, err := parseTarget(fmt.Sprintf("%s%s", target.URL.Host, target.URL.Path))
	// fmt.Printf("%#v\n", target.URL)
	// fmt.Printf("%#v\n", target)
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
	cr.ctx, cr.cancel = context.WithCancel(context.Background())
	go cr.watcher()
	return cr, nil
}

func (cb *consulBuilder) Scheme() string {
	return "consul"
}

type consulResolver struct {
	ctx                  context.Context
	cancel               func()
	address              string
	cc                   resolver.ClientConn
	name                 string
	disableServiceConfig bool
	lastIndex            uint64
}

func (cr *consulResolver) watcher() {
	// config := api.DefaultConfig()
	// config.Address = cr.address
	// client, err := api.NewClient(config)
	// if err != nil {
	// 	fmt.Printf("error create consul client: %v", err)
	// 	return
	// }
	for {
		select {
		case <-cr.ctx.Done():
			fmt.Println("=========================================走走")
			return
		default:
			client, err := conn(cr.address)
			if err != nil {
				fmt.Printf("error retrieving instances from Consul: %v\n", err)
				continue
			}

			services, metainfo, err := client.Health().Service(cr.name, "", true, &api.QueryOptions{WaitIndex: cr.lastIndex})
			if err != nil {
				fmt.Printf("error retrieving instances from Consul: %v\n", err)
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
			err = cr.cc.UpdateState(resolver.State{Addresses: newAddrs})
			if err != nil {
				// fmt.Printf("update connection state error: %v\n", err)
				continue
			}
		}
	}
}

func (cr *consulResolver) ResolveNow(opt resolver.ResolveNowOptions) {}

func (cr *consulResolver) Close() {
	cr.cancel()
}

func parseTarget(target string) (host, port, name string, err error) {

	log.Infof("target uri: %v", target)
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
