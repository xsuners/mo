package cache

import (
	"sync"
	"time"
)

type Cache interface {
	Get(key any) (any, bool)
	Set(key any, value any, duration time.Duration)
}

type Options struct {
	CleaningInterval time.Duration `ini-cleanInterval:"proxy" long:"gpt-clean-interval" description:"clean interval"`
}

type Option func(*Options)

func CleaningInterval(cleaningInterval time.Duration) Option {
	return func(o *Options) {
		o.CleaningInterval = cleaningInterval
	}
}

// impl stores arbitrary data with expiration time.
type impl struct {
	items sync.Map
	close chan struct{}
	opts  Options
}

// An item represents arbitrary data with expiration time.
type item struct {
	data    any
	expires int64
}

// New creates a new cache that asynchronously cleans
// expired entries after the given time passes.
func New(opts ...Option) Cache {
	cache := &impl{
		close: make(chan struct{}),
		opts: Options{
			CleaningInterval: time.Minute,
		},
	}
	for _, opt := range opts {
		opt(&cache.opts)
	}
	go func() {
		ticker := time.NewTicker(cache.opts.CleaningInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				now := time.Now().UnixNano()

				cache.items.Range(func(key, value any) bool {
					item := value.(item)

					if item.expires > 0 && now > item.expires {
						cache.items.Delete(key)
					}

					return true
				})

			case <-cache.close:
				return
			}
		}
	}()

	return cache
}

// Get gets the value for the given key.
func (cache *impl) Get(key any) (any, bool) {
	obj, exists := cache.items.Load(key)

	if !exists {
		return nil, false
	}

	item := obj.(item)

	if item.expires > 0 && time.Now().UnixNano() > item.expires {
		return nil, false
	}

	return item.data, true
}

// Set sets a value for the given key with an expiration duration.
// If the duration is 0 or less, it will be stored forever.
func (cache *impl) Set(key any, value any, duration time.Duration) {
	var expires int64

	if duration > 0 {
		expires = time.Now().Add(duration).UnixNano()
	}

	cache.items.Store(key, item{
		data:    value,
		expires: expires,
	})
}

// Range calls f sequentially for each key and value present in the cache.
// If f returns false, range stops the iteration.
func (cache *impl) Range(f func(key, value any) bool) {
	now := time.Now().UnixNano()

	fn := func(key, value any) bool {
		item := value.(item)

		if item.expires > 0 && now > item.expires {
			return true
		}

		return f(key, item.data)
	}

	cache.items.Range(fn)
}

// Delete deletes the key and its value from the cache.
func (cache *impl) Delete(key any) {
	cache.items.Delete(key)
}

// Close closes the cache and frees up resources.
func (cache *impl) Close() {
	cache.close <- struct{}{}
	cache.items = sync.Map{}
}
