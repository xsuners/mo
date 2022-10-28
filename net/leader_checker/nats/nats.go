package nats

import (
	"context"
	"sync"

	"github.com/nats-io/graft"
	"github.com/nats-io/nats.go"
	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/net/leader_checker"
	"go.uber.org/zap"
)

type natslc struct {
	node *graft.Node
	opts Options
}

func (lc *natslc) IsLeader() bool {
	return lc.node.State() == graft.LEADER
}

var (
	_lc     *natslc
	_lcCls  func()
	_lcOnce sync.Once
)

func New(cluster string, opt ...Option) (leader_checker.Checker, func(), error) {
	_lcOnce.Do(func() {
		_lc = &natslc{
			opts: defaultOptions,
		}
		for _, o := range opt {
			o(&_lc.opts)
		}

		opt := nats.GetDefaultOptions()
		opt.Url = _lc.opts.Urls

		rpc, err := graft.NewNatsRpc(&opt)
		if err != nil {
			panic(err)
		}

		ec := make(chan error)
		scc := make(chan graft.StateChange)
		handler := graft.NewChanHandler(scc, ec)

		ci := graft.ClusterInfo{Name: cluster, Size: _lc.opts.Members}

		_lc.node, err = graft.New(ci, handler, rpc, _lc.opts.LogPath)
		if err != nil {
			panic(err)
		}

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			for {
				select {
				case sc := <-scc:
					log.Infos("nats lc", zap.Any("sc", sc))
				case err := <-ec:
					log.Errors("nats lc", zap.Error(err))
				case <-ctx.Done():
					_lc.node.Close()
					return
				}
			}
		}()

		_lcCls = func() {
			cancel()
		}
	})

	return _lc, _lcCls, nil
}
