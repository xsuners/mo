package publisher

import (
	"sync"
	"time"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/misc/unats"
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

// AimSubject .
func AimSubject(ip string) description.CallOption {
	return description.NewFuncOption(func(o description.Options) {
		v, ok := o.Value().(*callOptions)
		if !ok {
			log.Fatalf("xnats: publisher call options type (%T) assertion error", o.Value())
		}
		v.subject = "ip-" + v.subject + "." + unats.IPSubject(ip)
	})
}

// AllSubject .
func AllSubject() description.CallOption {
	return description.NewFuncOption(func(o description.Options) {
		v, ok := o.Value().(*callOptions)
		if !ok {
			log.Fatalf("xnats: publisher call options type (%T) assertion error", o.Value())
		}
		v.subject = "all-" + v.subject
	})
}

// Subject .
func Subject(subject string) description.CallOption {
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
