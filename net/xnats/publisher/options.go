package publisher

import (
	"sync"
	"time"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/net/description"
)

type callOptions struct {
	subject      string
	timeout      time.Duration
	waitResponse bool
}

func (co *callOptions) Value() interface{} {
	return co
}

// UseSubject .
func UseSubject(subject string) description.CallOption {
	return description.NewFuncOption(func(o description.Options) {
		v, ok := o.Value().(*callOptions)
		if !ok {
			log.Fatalf("xnats: publisher call options type (%T) assertion error", o.Value())
		}
		v.subject = subject
	})
}

// Timeout .
func Timeout(duration time.Duration) description.CallOption {
	return description.NewFuncOption(func(o description.Options) {
		v, ok := o.Value().(*callOptions)
		if !ok {
			log.Fatalf("xnats: publisher call options type (%T) assertion error", o.Value())
		}
		v.timeout = duration
	})
}

// WaitResponse .
func WaitResponse() description.CallOption {
	return description.NewFuncOption(func(o description.Options) {
		v, ok := o.Value().(*callOptions)
		if !ok {
			log.Fatalf("xnats: publisher call options type (%T) assertion error", o.Value())
		}
		v.waitResponse = true
	})
}

var copool = sync.Pool{
	New: func() interface{} {
		return &callOptions{}
	},
}
