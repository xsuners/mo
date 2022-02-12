package client

import (
	"sync"
	"time"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/net/description"
)

type CallOptions struct {
	Subject      string
	WaitResponse bool
	Timeout      time.Duration
}

func (co *CallOptions) Value() interface{} {
	return co
}

// Timeout .
func Timeout(duration time.Duration) description.CallOption {
	return description.NewFuncOption(func(o description.Options) {
		v, ok := o.Value().(*CallOptions)
		if !ok {
			log.Fatalf("xnats: publisher call options type (%T) assertion error", o.Value())
		}
		v.Timeout = duration
	})
}

var copool = sync.Pool{
	New: func() interface{} {
		return &CallOptions{}
	},
}
