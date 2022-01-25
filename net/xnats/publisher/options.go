package publisher

import (
	"sync"
	"time"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/misc/unats"
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

// AimSubject .
func AimSubject(ip string) description.CallOption {
	return description.NewFuncOption(func(o description.Options) {
		v, ok := o.Value().(*CallOptions)
		if !ok {
			log.Fatalf("xnats: publisher call options type (%T) assertion error", o.Value())
		}
		v.Subject = "ip-" + v.Subject + "." + unats.IPSubject(ip)
	})
}

// AllSubject .
func AllSubject() description.CallOption {
	return description.NewFuncOption(func(o description.Options) {
		v, ok := o.Value().(*CallOptions)
		if !ok {
			log.Fatalf("xnats: publisher call options type (%T) assertion error", o.Value())
		}
		v.Subject = "all-" + v.Subject
	})
}

// Subject .
func Subject(subject string) description.CallOption {
	return description.NewFuncOption(func(o description.Options) {
		v, ok := o.Value().(*CallOptions)
		if !ok {
			log.Fatalf("xnats: publisher call options type (%T) assertion error", o.Value())
		}
		v.Subject = subject
	})
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

// WaitResponse .
func WaitResponse() description.CallOption {
	return description.NewFuncOption(func(o description.Options) {
		v, ok := o.Value().(*CallOptions)
		if !ok {
			log.Fatalf("xnats: publisher call options type (%T) assertion error", o.Value())
		}
		v.WaitResponse = true
	})
}

var copool = sync.Pool{
	New: func() interface{} {
		return &CallOptions{}
	},
}
