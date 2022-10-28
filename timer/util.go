package timer

import (
	"context"
	"time"

	"github.com/xsuners/mo/log"
	"github.com/xsuners/mo/sync/workerpool"
)

var (
	wheel *Wheel
	wp    *workerpool.WorkerPool
)

func init() {
	wheel = NewWheel(context.Background())
	wp = workerpool.New(1024)
	go func() {
		for {
			e, ok := <-wheel.Expired()
			if !ok {
				log.Info("global timer closed")
				return
			}
			cb, ok := e.Job.(func(int64))
			if !ok {
				log.Errorf("global timer job (%T) error", e.Job)
				continue
			}
			wp.Submit(func() {
				cb(e.ID)
			})
		}
	}()
}

// Cancel .
func Cancel(tid int64) {
	wheel.Cancel(tid)
}

// Size .
func Size() int {
	return wheel.Size()
}

// Stop .
func Stop() {
	wheel.Stop()
	wp.StopWait()
}

// RunAt .
func RunAt(ts time.Time, cb func(tid int64)) int64 {
	return wheel.Add(ts, 0, cb)
}

// RunAfter .
func RunAfter(d time.Duration, cb func(tid int64)) int64 {
	delay := time.Now().Add(d)
	return RunAt(delay, cb)
}

// RunEvery .
func RunEvery(d time.Duration, cb func(tid int64)) int64 {
	delay := time.Now().Add(d)
	return wheel.Add(delay, d, cb)
}
