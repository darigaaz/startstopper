package startstopper

import (
	"errors"
	"sync"
)

var (
	ErrAlreadyStarted = errors.New("already started")
	ErrAlreadyStopped = errors.New("already stopped")
)

type StartStopper struct {
	mu      sync.Mutex
	closing chan struct{}
	started bool
}

// Mutex must be held already
func (startStopper *StartStopper) init() {
	if startStopper.closing == nil {
		startStopper.closing = make(chan struct{})
	}
}

// Start
// Returns closingCh and error if service can not be started (already started)
func (startStopper *StartStopper) Start() (<-chan struct{}, error) {
	startStopper.mu.Lock()
	startStopper.init()

	closingCh := startStopper.closing
	wasStarted := startStopper.started

	if !wasStarted {
		startStopper.started = true
	}

	startStopper.mu.Unlock()

	if wasStarted {
		return nil, ErrAlreadyStarted
	}

	return closingCh, nil
}

func (startStopper *StartStopper) StartNotify(signalCh chan<- error) (<-chan struct{}, error) {
	closingCh, err := startStopper.Start()
	Notify(signalCh, err, true)
	return closingCh, err
}

func (startStopper *StartStopper) Stop() error {
	startStopper.mu.Lock()
	startStopper.init()

	closingCh := startStopper.closing
	wasStarted := startStopper.started

	startStopper.started = false

	startStopper.mu.Unlock()

	if !wasStarted {
		return ErrAlreadyStopped
	}

	// sync with main service's for { select } loop
	closingCh <- struct{}{}

	return nil
}
