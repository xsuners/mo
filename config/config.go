package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/xsuners/mo/database/xmongo"
	"github.com/xsuners/mo/net/xnsq"
)

// App .
type App struct {
	Service  string `json:"service"`
	Version  string `json:"version"`
	DC       string `json:"dc"`
	Env      string `json:"env"`
	Hostname string `json:"hostname"`
}

// // Configc .
// type Configc struct {
// 	NATS *publisher.Config `json:"nats"`
// 	GRPC *client.Config    `json:"grpc"`
// }

// Config .
type Config struct {
	// Log          *log.Config      `json:"log"`
	// GRPC         *xgrpc.Config    `json:"grpc"`
	// NATSConsumer *xnats.Config    `json:"nats"`
	// Naming *naming.Config `json:"naming"`
	// HTTP  *xhttp.Config  `json:"http"`
	// TCP   *xtcp.Config   `json:"tcp"`
	// WS    *xws.Config    `json:"ws"`
	App   *App           `json:"app"`
	NSQ   *xnsq.Config   `json:"nsq"`
	Mongo *xmongo.Config `json:"mongo"`
	// SQL   *xsql.Config   `json:"sql"`
	// Redis    *xredis.Config   `json:"redis"`
	// Memcache *memcache.Config `json:"memcache"`
	// Services     map[string]*Configc `json:"services"`
}

// Configure .
type Configure struct {
	opts options
}

var (
	config     *Configure
	configOnce sync.Once
)

// Instance .
func Instance(opt ...Option) *Configure {
	configOnce.Do(func() {
		opts := defaultOptions
		for _, o := range opt {
			o(&opts)
		}
		config = &Configure{
			opts: opts,
		}
	})
	return config
}

// Unmarshal .
func (configure *Configure) Unmarshal(conf interface{}) error {
	path := config.opts.localPath
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open config file %s err: %w", path, err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("read config file %s err: %w", path, err)
	}
	err = json.Unmarshal(buf, conf)
	if err != nil {
		return fmt.Errorf("unmarshal config file %s err: %w", path, err)
	}
	return nil
}

// Unmarshal .
func Unmarshal(conf interface{}, opt ...Option) (err error) {
	return Instance(opt...).Unmarshal(conf)
}

// Watch .
func Watch(conf interface{}) {}

// Close .
func Close() {}
