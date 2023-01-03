package consul

import (
	"fmt"
	"sync"

	"github.com/hashicorp/consul/api"
)

var (
	conns = make(map[string]*api.Client)
	mu    sync.RWMutex
)

func conn(addr string) (*api.Client, error) {
	mu.RLock()
	cli, ok := conns[addr]
	if ok {
		mu.RUnlock()
		return cli, nil
	}
	mu.RUnlock()
	if !ok {
		mu.Lock()
		defer mu.Unlock()
		config := api.DefaultConfig()
		config.Address = addr
		client, err := api.NewClient(config)
		if err != nil {
			fmt.Printf("error create consul client: %v", err)
			return nil, err
		}
		conns[addr] = client
	}
	return conns[addr], nil
}
