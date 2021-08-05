package xmemcache

type Memcache struct{}

func New(opt ...Option) (*Memcache, func(), error) {
	// TODO
	return &Memcache{}, func() {}, nil
}
