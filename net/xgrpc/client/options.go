package client

import (
	"sync"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/net/description"
	"google.golang.org/grpc"
)

type Options struct {
	Resolver string `ini-name:"resolver" long:"grpcc.resolver" description:"grpcc resolver"`

	copts []grpc.CallOption
}

func (co *Options) Value() interface{} {
	return co
}

// CallOption .
func CallOption(copt grpc.CallOption) description.CallOption {
	return description.NewFuncOption(func(o description.Options) {
		v, ok := o.Value().(*Options)
		if !ok {
			log.Fatalf("xnats: publisher call options type (%T) assertion error", o.Value())
		}
		v.copts = append(v.copts, copt)
	})
}

var copool = sync.Pool{
	New: func() interface{} {
		return &Options{}
	},
}
