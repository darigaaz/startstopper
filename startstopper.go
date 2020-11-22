package startstopper

import (
	"errors"
	"sync"
)

var ErrAlreadyStarted = errors.New("already started")
var ErrAlreadyClosed = errors.New("already stopped")

type StartStopper struct {
	mu      sync.Mutex
	closing chan func(error)
	started bool
}

// Mutex must be held already
func (startStopper *StartStopper) init() {
	if startStopper.closing == nil {
		startStopper.closing = make(chan func(error))
	}
}

// Returns closingCh and error if service can not be started (already started)
// readyCh will be closed
func (startStopper *StartStopper) Start(readyCh chan error) (<-chan func(error), error) {
	startStopper.mu.Lock()
	startStopper.init()

	closingCh := startStopper.closing
	wasStarted := startStopper.started

	if !wasStarted {
		startStopper.started = true
	}

	startStopper.mu.Unlock()

	if wasStarted {
		if readyCh != nil {
			readyCh <- ErrAlreadyStarted
			close(readyCh)
		}
		return nil, ErrAlreadyStarted
	}

	if readyCh != nil {
		close(readyCh)
	}

	return closingCh, nil
}

// errCh will be closed on notifyFunc
func (startStopper *StartStopper) Stop(errCh chan error) {
	startStopper.mu.Lock()
	startStopper.init()

	closingCh := startStopper.closing
	wasStarted := startStopper.started

	startStopper.started = false

	startStopper.mu.Unlock()

	notifyFunc := func(err error) {
		if errCh != nil {
			if err != nil {
				errCh <- err
			}
			close(errCh)
		}
	}

	if !wasStarted {
		notifyFunc(ErrAlreadyClosed)
		return
	}

	closingCh <- notifyFunc
}
