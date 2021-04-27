package timer

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xsuners/mo/sync/atom"
)

const (
	tickPeriod time.Duration = 500 * time.Millisecond
	bufferSize               = 1024
)

var timerIds *atom.Int64

func init() {
	timerIds = atom.NewInt64(0)
}

// timers is a heap-based priority queue
type timers []*timer

func (heap timers) index(id int64) int {
	for _, t := range heap {
		if t.id == id {
			return t.index
		}
	}
	return -1
}

func (heap timers) Len() int {
	return len(heap)
}

func (heap timers) Less(i, j int) bool {
	return heap[i].expiration.UnixNano() < heap[j].expiration.UnixNano()
}

func (heap timers) Swap(i, j int) {
	heap[i], heap[j] = heap[j], heap[i]
	heap[i].index = i
	heap[j].index = j
}

func (heap *timers) Push(x interface{}) {
	n := len(*heap)
	timer := x.(*timer)
	timer.index = n
	*heap = append(*heap, timer)
}

func (heap *timers) Pop() interface{} {
	old := *heap
	n := len(old)
	timer := old[n-1]
	timer.index = -1
	*heap = old[0 : n-1]
	return timer
}

/* 'expiration' is the time when timer time out, if 'interval' > 0
the timer will time out periodically, 'timeout' contains the callback
to be called when times out */
type timer struct {
	id         int64
	expiration time.Time
	interval   time.Duration
	timeout    interface{}
	index      int // for container/heap
}

func newTimer(when time.Time, interv time.Duration, to interface{}) *timer {
	return &timer{
		id:         timerIds.GetAndIncrement(),
		expiration: when,
		interval:   interv,
		timeout:    to,
	}
}

// type resetConfig struct {
// 	id         int64
// 	expiration time.Time
// }

// Outer .
type Outer struct {
	ID  int64
	Job interface{}
}

// Wheel manages all the timed task.
type Wheel struct {
	expiredChan chan *Outer
	timers      timers
	ticker      *time.Ticker
	wg          *sync.WaitGroup
	addChan     chan *timer // add timer in loop
	cancelChan  chan int64  // cancel timer in loop
	sizeChan    chan int    // get size in loop
	ctx         context.Context
	cancel      context.CancelFunc
	// resetChan   chan *resetConfig
}

// NewWheel returns a *Wheel ready for use.
func NewWheel(ctx context.Context) *Wheel {
	wheel := &Wheel{
		timers:      make(timers, 0),
		ticker:      time.NewTicker(tickPeriod),
		wg:          &sync.WaitGroup{},
		expiredChan: make(chan *Outer, bufferSize),
		addChan:     make(chan *timer, bufferSize),
		cancelChan:  make(chan int64, bufferSize),
		// resetChan:   make(chan *resetConfig, bufferSize),
		sizeChan: make(chan int),
	}
	wheel.ctx, wheel.cancel = context.WithCancel(ctx)

	heap.Init(&wheel.timers)

	wheel.wg.Add(1)
	go func() {
		wheel.start()
		wheel.wg.Done()
	}()

	return wheel
}

// Expired returns the timeout channel.
func (tw *Wheel) Expired() chan *Outer {
	return tw.expiredChan
}

// Add adds new timed task.
func (tw *Wheel) Add(when time.Time, interv time.Duration, to interface{}) int64 {
	if to == nil {
		return int64(-1)
	}
	timer := newTimer(when, interv, to)
	tw.addChan <- timer
	return timer.id
}

// // Reset reset the timer expiration
// func (tw *Wheel) Reset(id int64, t time.Time) {
// 	tw.resetChan <- &resetConfig{
// 		id:         id,
// 		expiration: t,
// 	}
// }

// Size returns the number of timed tasks.
func (tw *Wheel) Size() int {
	return <-tw.sizeChan
}

// Cancel cancels a timed task with specified timer ID.
func (tw *Wheel) Cancel(timerID int64) {
	tw.cancelChan <- timerID
}

// Stop stops the Wheel.
func (tw *Wheel) Stop() {
	tw.cancel()
	tw.wg.Wait()
}

func (tw *Wheel) expired() []*timer {
	expired := make([]*timer, 0)
	for tw.timers.Len() > 0 {
		timer := heap.Pop(&tw.timers).(*timer)
		elapsed := time.Since(timer.expiration).Seconds()
		if elapsed > 1.0 {
			// log.Warn("elapsed %f", elapsed)
			fmt.Printf("elapsed %f", elapsed)
		}
		if elapsed > 0.0 {
			expired = append(expired, timer)
			continue
		} else {
			heap.Push(&tw.timers, timer)
			break
		}
	}
	return expired
}

func (tw *Wheel) update(timers []*timer) {
	if timers != nil {
		for _, t := range timers {
			if t.interval > 0 { // repeatable timer task
				t.expiration = t.expiration.Add(t.interval)
				// if task time out for at least 10 seconds, the expiration time needs
				// to be updated in case this task executes every time timer wakes up.
				if time.Since(t.expiration).Seconds() >= 10.0 {
					t.expiration = time.Now()
				}
				heap.Push(&tw.timers, t)
			}
		}
	}
}

func (tw *Wheel) start() {
	for {
		select {
		case timerID := <-tw.cancelChan:
			index := tw.timers.index(timerID)
			if index >= 0 {
				heap.Remove(&tw.timers, index)
			}

		// case cfg := <-tw.resetChan:
		// 	index := tw.timers.index(cfg.id)
		// 	if index >= 0 {
		// 		tm := heap.Remove(&tw.timers, index).(*timer)
		// 		tm.expiration = cfg.expiration
		// 		heap.Push(&tw.timers, tm)
		// 	}

		case tw.sizeChan <- tw.timers.Len():

		case <-tw.ctx.Done():
			tw.ticker.Stop()
			return

		case timer := <-tw.addChan:
			heap.Push(&tw.timers, timer)

		case <-tw.ticker.C:
			timers := tw.expired()
			for _, t := range timers {
				tw.expiredChan <- &Outer{
					ID:  t.id,
					Job: t.timeout,
				}
			}
			tw.update(timers)
		}
	}
}
