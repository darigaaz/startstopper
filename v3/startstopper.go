package startstopper

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	// KillTimeoutDefault ...
	KillTimeoutDefault = 100 * time.Millisecond
)

var (
	ErrStart = errors.New("can not be done in stopping state")
	errStart = NewErrorCode(ErrStart, "STARTSTOPPER_ERR_START")
)

func makeClosedChan[T any]() chan T {
	ch := make(chan T)
	close(ch)
	return ch
}

var alwaysClosedChan = makeClosedChan[struct{}]()

// StartStopper ...
// The zero value for StartStopper is ready to use (remember to call Init).
type StartStopper struct {
	initOnce sync.Once
	initErr  error

	mu   sync.Mutex
	done chan struct{} // closes when shutdown completes

	gracefulCtx           context.Context    // listen to begin graceful shutdown
	gracefulCtxCancelFunc context.CancelFunc // cancels gracefulCtx

	killTimeoutProvider func(ctx context.Context) time.Duration
	killCtx             context.Context    // listen to begin termination
	killCtxCancelFunc   context.CancelFunc // cancels killCtx
}

// New ...
func New(
	ctx context.Context,
	killTimeoutProvider func(ctx context.Context) time.Duration,
) *StartStopper {
	var inst StartStopper
	_ = inst.Init(ctx, killTimeoutProvider)
	return &inst
}

// must be called once via Init
func (startStopper *StartStopper) init(
	killTimeoutProvider func(ctx context.Context) time.Duration,
) {
	startStopper.done = alwaysClosedChan

	if killTimeoutProvider == nil {
		killTimeoutProvider = func(_ context.Context) time.Duration { return KillTimeoutDefault }
	}
	startStopper.killTimeoutProvider = killTimeoutProvider

	// startStopper.initErr = nil
}

// Init to be called from ctor.
func (startStopper *StartStopper) Init(
	ctx context.Context,
	killTimeoutProvider func(ctx context.Context) time.Duration,
) error {
	startStopper.initOnce.Do(func() {
		startStopper.init(killTimeoutProvider)
	})
	return startStopper.initErr
}

// InitNotify to be called from methods on zero value initialization.
func (startStopper *StartStopper) InitNotify(
	ctx context.Context,
	signalCh chan<- error,
	killTimeoutProvider func(ctx context.Context) time.Duration,
) error {
	err := startStopper.Init(ctx, killTimeoutProvider)
	Notify(signalCh, startStopper.initErr, NotifyCloseModeOnError)
	return err
}

// Start the loop.
// Safe to call after Init.
// Initialize your state in startFn (fallible).
// Returns gracefulContext, killContext, doneChan, error.
func (startStopper *StartStopper) Start(
	ctx context.Context,
	cleanupDoneChan <-chan struct{},
	readyCh chan<- error, // optional
	startFn func() error, // optional
) (
	context.Context, // graceful context
	context.Context, // killCtx context
	<-chan struct{}, // done chan
	error,
) {
	var (
		err                   error
		done                  chan struct{}
		gracefulCtx           context.Context
		gracefulCtxCancelFunc context.CancelFunc
		killCtx               context.Context
		killCtxCancelFunc     context.CancelFunc
	)

	WithMutex(&startStopper.mu, func() {
		if startStopper.done != alwaysClosedChan {
			err = errStart
			return
		}

		if startFn != nil {
			err = startFn()
			if err != nil {
				return
			}
		}

		done = make(chan struct{})
		gracefulCtx, gracefulCtxCancelFunc = context.WithCancel(ctx)
		killCtx, killCtxCancelFunc = context.WithCancel(context.WithoutCancel(gracefulCtx))

		// cancel killCtx with timeout
		context.AfterFunc(gracefulCtx, func() {
			time.AfterFunc(startStopper.killTimeoutProvider(killCtx), killCtxCancelFunc)
		})

		startStopper.gracefulCtx = gracefulCtx
		startStopper.gracefulCtxCancelFunc = gracefulCtxCancelFunc
		startStopper.killCtx = killCtx
		startStopper.killCtxCancelFunc = killCtxCancelFunc

		startStopper.done = done
	})

	if err == nil {
		// Setup cleanup.
		go func() {
			<-cleanupDoneChan

			// make sure contexts dont leak
			gracefulCtxCancelFunc()

			WithMutex(&startStopper.mu, func() {
				close(done)
				startStopper.done = alwaysClosedChan
			})
		}()
	}

	Notify(readyCh, err, NotifyCloseModeAlways)
	return gracefulCtx, killCtx, done, err
}

// Done returns a channel that's closed when work done and loop is stopped.
// Threadsafe.
func (startStopper *StartStopper) Done() <-chan struct{} {
	var ch <-chan struct{}

	WithMutex(&startStopper.mu, func() {
		ch = startStopper.done
	})

	return ch
}

// Context returns context for graceful shutdown.
// @TODO SAFETY ???
func (startStopper *StartStopper) Context() context.Context {
	return startStopper.gracefulCtx
}

// KillContext returns context for kill.
// @TODO SAFETY ???
func (startStopper *StartStopper) KillContext() context.Context {
	return startStopper.killCtx
}

// Close tries to stop the loop gracefully. Kill after timeout.
// Wait for completion.
func (startStopper *StartStopper) Close() {
	done := startStopper.Done()

	startStopper.CloseAsync()

	<-done
}

// Kill the loop as soon as possible.
// Wait for completion.
func (startStopper *StartStopper) Kill() {
	done := startStopper.Done()

	startStopper.KillAsync()

	<-done
}

// CloseAsync like Close but dont wait.
// @TODO SAFETY ???
func (startStopper *StartStopper) CloseAsync() {
	startStopper.gracefulCtxCancelFunc()
}

// KillAsync like Kill but dont wait.
// @TODO SAFETY ???
func (startStopper *StartStopper) KillAsync() {
	startStopper.gracefulCtxCancelFunc()
	startStopper.killCtxCancelFunc()
}
