package startstopper

import (
	"sync"
)

// ChanCloser ...
//
//	done, doneCloseFn := ChanCloser(nil)
//
//	done, doneCloseFn := ChanCloser(make(chan struct{}, 1))
func ChanCloser(done chan struct{}) (chan struct{}, func()) {
	if done == nil {
		done = make(chan struct{})
	}

	return done, func() { close(done) }
}

// ChanCloserWaitGroup ...
func ChanCloserWaitGroup(done chan struct{}, delta int) (chan struct{}, func()) {
	if done == nil {
		done = make(chan struct{})
	}

	wg := sync.WaitGroup{}
	wg.Add(delta)

	go func() {
		wg.Wait()
		close(done)
	}()

	return done, wg.Done
}
