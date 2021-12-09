package xrand

import (
	"math/rand"
	"sync"
	"time"
)

var (
	r  = rand.New(rand.NewSource(time.Now().UnixNano()))
	mu sync.Mutex
)

// Int63n implements rand.Int63n on the xrand global source.
func Int63n(n int64) int64 {
	mu.Lock()
	res := r.Int63n(n)
	mu.Unlock()
	return res
}

// Intn implements rand.Intn on the xrand global source.
func Intn(n int) int {
	mu.Lock()
	res := r.Intn(n)
	mu.Unlock()
	return res
}

// Float64 implements rand.Float64 on the xrand global source.
func Float64() float64 {
	mu.Lock()
	res := r.Float64()
	mu.Unlock()
	return res
}
