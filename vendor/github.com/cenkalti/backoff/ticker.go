package backoff

import (
	"runtime"
	"sync"
	"time"
)

// Ticker holds a channel that delivers `ticks' of a clock at times reported by a BackOff.
//
// Ticks will continue to arrive when the previous operation is still running,
// so operations that take a while to fail could run in quick succession.
//
// Usage:
// 	operation := func() error {
// 		// An operation that may fail
// 	}
//
//	b := backoff.NewExponentialBackOff()
//	ticker := backoff.NewTicker(b)
//
// 	var err error
//	for _ = range ticker.C {
//		if err = operation(); err != nil {
//			log.Println(err, "will retry...")
//			continue
//		}
//
//		ticker.Stop()
//		break
//	}
//
// 	if err != nil {
// 		// Operation has failed.
// 	}
//
// 	// Operation is successfull.
//
type Ticker struct {
	C        <-chan time.Time
	c        chan time.Time
	b        BackOff
	stop     chan struct{}
	stopOnce sync.Once
}

// NewTicker returns a new Ticker containing a channel that will send the time at times
// specified by the BackOff argument. Ticker is guaranteed to tick at least once.
// The channel is closed when Stop method is called or BackOff stops.
func NewTicker(b BackOff) *Ticker {
	c := make(chan time.Time)
	t := &Ticker{
		C:    c,
		c:    c,
		b:    b,
		stop: make(chan struct{}),
	}
	go t.run()
	runtime.SetFinalizer(t, (*Ticker).Stop)
	return t
}

// Stop turns off a ticker. After Stop, no more ticks will be sent.
func (t *Ticker) Stop() {
	t.stopOnce.Do(func() { close(t.stop) })
}

func (t *Ticker) run() {
	c := t.c
	defer close(c)
	t.b.Reset()

	// Ticker is guaranteed to tick at least once.
	afterC := t.send(time.Now())

	for {
		if afterC == nil {
			return
		}

		select {
		case tick := <-afterC:
			afterC = t.send(tick)
		case <-t.stop:
			t.c = nil // Prevent future ticks from being sent to the channel.
			return
		}
	}
}

func (t *Ticker) send(tick time.Time) <-chan time.Time {
	select {
	case t.c <- tick:
	case <-t.stop:
		return nil
	}

	next := t.b.NextBackOff()
	if next == Stop {
		t.Stop()
		return nil
	}

	return time.After(next)
}
