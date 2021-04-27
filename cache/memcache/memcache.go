package memcache

type Config struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

type Memcache struct{}

func New(c *Config) *Memcache {
	return &Memcache{}
}

func (mc *Memcache) Close() {}
