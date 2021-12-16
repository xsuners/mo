package publisher

import (
	"sync"
	"time"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/misc/unats"
	"github.com/xsuners/mo/net/description"
)

type Options struct {
	Subject      string        `ini-name:"subject" long:"natsc-subject" description:"nats subject"`
	WaitResponse bool          `ini-name:"waitResponse" long:"natsc-waitResponse" description:"whether wait response"`
	Timeout      time.Duration `ini-name:"timeout" long:"natsc-timeout" description:"timeout secs"`
}

func (co *Options) Value() interface{} {
	return co
}

// AimSubject .
func AimSubject(ip string) description.CallOption {
	return description.NewFuncOption(func(o description.Options) {
		v, ok := o.Value().(*Options)
		if !ok {
			log.Fatalf("xnats: publisher call options type (%T) assertion error", o.Value())
		}
		v.Subject = "ip-" + v.Subject + "." + unats.IPSubject(ip)
	})
}

// AllSubject .
func AllSubject() description.CallOption {
	return description.NewFuncOption(func(o description.Options) {
		v, ok := o.Value().(*Options)
		if !ok {
			log.Fatalf("xnats: publisher call options type (%T) assertion error", o.Value())
		}
		v.Subject = "all-" + v.Subject
	})
}

// Subject .
func Subject(subject string) description.CallOption {
	return description.NewFuncOption(func(o description.Options) {
		v, ok := o.Value().(*Options)
		if !ok {
			log.Fatalf("xnats: publisher call options type (%T) assertion error", o.Value())
		}
		v.Subject = subject
	})
}

// Timeout .
func Timeout(duration time.Duration) description.CallOption {
	return description.NewFuncOption(func(o description.Options) {
		v, ok := o.Value().(*Options)
		if !ok {
			log.Fatalf("xnats: publisher call options type (%T) assertion error", o.Value())
		}
		v.Timeout = duration
	})
}

// WaitResponse .
func WaitResponse() description.CallOption {
	return description.NewFuncOption(func(o description.Options) {
		v, ok := o.Value().(*Options)
		if !ok {
			log.Fatalf("xnats: publisher call options type (%T) assertion error", o.Value())
		}
		v.WaitResponse = true
	})
}

var copool = sync.Pool{
	New: func() interface{} {
		return &Options{}
	},
}
