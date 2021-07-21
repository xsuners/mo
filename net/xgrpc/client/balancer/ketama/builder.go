package ketama

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/xsuners/mo/log"
	"go.uber.org/zap"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

var _ base.PickerBuilder = &ketamaPickerBuilder{}
var _ balancer.Picker = &ketamaPicker{}

// Name is balancer name.
const Name = "ketama"

// Key .
type hashKey struct{}
type addrKey struct{}

func init() {
	// balancer.Register(Builder())
	rand.Seed(time.Now().UnixNano())
}

// To .
func To(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, hashKey{}, key)
}

// Addr .
func Addr(ctx context.Context, addr string) context.Context {
	return context.WithValue(ctx, addrKey{}, addr)
}

// Builder creates a new ConsistanceHash balancer builder.
func Builder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &ketamaPickerBuilder{}, base.Config{HealthCheck: true})
}

type ketamaPickerBuilder struct{}

// func (b *ketamaPickerBuilder) Build(readySCs map[resolver.Address]balancer.SubConn) balancer.Picker {
func (b *ketamaPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	// log.Info("ketama balancer: newPicker called with readySCs: %v", readySCs)
	readySCs := info.ReadySCs
	if len(readySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	picker := &ketamaPicker{
		subConns: make(map[string]balancer.SubConn),
		hash:     NewKetama(10, nil),
	}

	for sc, sci := range readySCs {
		picker.subConns[sci.Address.Addr] = sc
		weight := getWeight(sci.Address)
		for i := 0; i < weight; i++ {
			node := wrapAddr(sci.Address.Addr, i)
			picker.hash.Add(node)
			picker.subConns[node] = sc
		}
	}
	return picker
}

type ketamaPicker struct {
	mu       sync.Mutex
	hash     *Ketama
	subConns map[string]balancer.SubConn
}

// func (p *ketamaPicker) Pick(ctx context.Context, opts balancer.PickInfo) (balancer.SubConn, func(balancer.DoneInfo), error) {
func (p *ketamaPicker) Pick(info balancer.PickInfo) (result balancer.PickResult, err error) {
	var sc balancer.SubConn

	p.mu.Lock()
	defer p.mu.Unlock()

	// if use addr type
	addr, ok := info.Ctx.Value(addrKey{}).(string)
	if ok {
		log.Debugsc(info.Ctx, "Pick:Addr", zap.String("addr", addr))
		sc, ok = p.subConns[addr]
		if !ok {
			err = fmt.Errorf("addr(%s) no conn found", addr)
			return
		}
		result.SubConn = sc
		result.Done = func(di balancer.DoneInfo) {
			// TODO
			// log.Infow("TODO")
		}
		return
	}

	key, ok := info.Ctx.Value(hashKey{}).(string)
	if !ok {
		// key = strconv.Itoa(rand.Intn(99999999))
		key = strconv.Itoa(99999999)
		// log.Infof("ketama balancer: fallback to random strategy. key: %s", key)
	}

	targetAddr, ok := p.hash.Get(key)
	if ok {
		sc = p.subConns[targetAddr]
	} else {
		log.Errorf("ketama balancer: get targetAddr failed: %v", targetAddr)
		err = fmt.Errorf("ketama balancer: can not get sub conn with addr: %s", targetAddr)
		return
	}

	result.SubConn = sc
	result.Done = func(di balancer.DoneInfo) {
		// TODO
		// log.Infow("TODO")
	}

	return
}

func wrapAddr(addr string, idx int) string {
	return fmt.Sprintf("%s-%d", addr, idx)
}

const (
	// WeightKey .
	WeightKey = "weight"
)

func getWeight(addr resolver.Address) int {
	w, ok := addr.Attributes.Value(WeightKey).(int)
	if ok {
		return w
	}
	return 1
}
