package client

import (
	"sync"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/net/description"
	"google.golang.org/grpc"
)

type callOptions struct {
	copts []grpc.CallOption
}

func (co *callOptions) Value() interface{} {
	return co
}

// CallOption .
func CallOption(copt grpc.CallOption) description.CallOption {
	return description.NewFuncOption(func(o description.Options) {
		v, ok := o.Value().(*callOptions)
		if !ok {
			log.Fatalf("xnats: publisher call options type (%T) assertion error", o.Value())
		}
		v.copts = append(v.copts, copt)
	})
}

var copool = sync.Pool{
	New: func() interface{} {
		return &callOptions{}
	},
}
