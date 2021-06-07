package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/xsuners/mo/cache/memcache"
	"github.com/xsuners/mo/cache/xredis"
	"github.com/xsuners/mo/database/xmongo"
	"github.com/xsuners/mo/database/xsql"
	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/naming"
	"github.com/xsuners/mo/net/xgrpc"
	"github.com/xsuners/mo/net/xhttp"
	"github.com/xsuners/mo/net/xnats"
	"github.com/xsuners/mo/net/xnsq"
	"github.com/xsuners/mo/net/xtcp"
	"github.com/xsuners/mo/net/xws"
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
	App          *App             `json:"app"`
	Log          *log.Config      `json:"log"`
	HTTP         *xhttp.Config    `json:"http"`
	GRPC         *xgrpc.Config    `json:"grpc"`
	TCP          *xtcp.Config     `json:"tcp"`
	WS           *xws.Config      `json:"ws"`
	NATSConsumer *xnats.Config    `json:"nats"`
	NSQ          *xnsq.Config     `json:"nsq"`
	Redis        *xredis.Config   `json:"redis"`
	Naming       *naming.Config   `json:"naming"`
	Memcache     *memcache.Config `json:"memcache"`
	SQL          *xsql.Config     `json:"sql"`
	Mongo        *xmongo.Config   `json:"mongo"`
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
