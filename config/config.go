package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/xsuners/mo/database/xmongo"
)

// App .
type App struct {
	Service  string `json:"service"`
	Version  string `json:"version"`
	DC       string `json:"dc"`
	Env      string `json:"env"`
	Hostname string `json:"hostname"`
}

// Config .
type Config struct {
	App   *App           `json:"app"`
	Mongo *xmongo.Config `json:"mongo"`
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
