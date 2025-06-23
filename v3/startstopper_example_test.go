package startstopper_test

import (
	"context"
	"fmt"
	"time"

	"github.com/Darigaaz/startstopper/v3"
)

type Srv struct {
	// embed or include startstopper.StartStopper into your struct
	startstopper.StartStopper

	C chan int
}

func NewSrv(ctx context.Context) *Srv {
	// The zero value for StartStopper is ready to use.
	// But dont forget to init it once.
	service := &Srv{
		C: make(chan int),
	}

	// @NOTE Init startstopper
	_ = service.StartStopper.Init(ctx, nil)

	return service
}

func (srv *Srv) DoStuff(n int) {
	// precondition: do not produce if we are shutting down. synchronization is up to you
	// if !running { return }

	// your logic here, example
	srv.C <- n
}

func (srv *Srv) Go(ctx context.Context) (<-chan struct{}, error) {
	readyCh := make(chan error, 1)

	go srv.Start(ctx, readyCh)

	return srv.Done(), <-readyCh
}

func (srv *Srv) process(msg int) error {
	// process message
	fmt.Println(msg)
	return nil
}

func (srv *Srv) start() error {
	// setup before each start

	return nil
}

func (srv *Srv) Start(ctx context.Context, readyCh chan<- error) error {
	err := srv.StartStopper.InitNotify(ctx, readyCh, nil)
	if err != nil {
		return err
	}

	cleanupDone, doneFn := startstopper.ChanCloserWaitGroup(nil, 2)

	ctx, killCtx, done, err := srv.StartStopper.Start(ctx, cleanupDone, readyCh, srv.start)
	if err != nil {
		return err
	}

	go srv.loop(ctx, killCtx, doneFn)
	go srv.monitor(ctx, killCtx, doneFn)

	// wait for everything to stop
	<-done // <-srv.StartStopper.Done()

	return nil
}

func (srv *Srv) monitor(
	ctx context.Context,
	killCtx context.Context,
	doneFn func(),
) {
	defer doneFn() // signal startstopper this loop is done

	ticker := time.NewTicker(time.Second)
	defer func() {
		// clean up
		ticker.Stop()
	}()

	ctxDone := ctx.Done()
	killChan := killCtx.Done()

	for {
		select {
		case <-killChan:
			// terminate now
			return

		case <-ctxDone:
			// begin graceful shutdown
			// but its easy case - nothing to do, just return
			return

		case <-ticker.C:
			// do work
			fmt.Println("tick")
			ticker.Reset(time.Second)
		}
	}
}

func (srv *Srv) loop(
	ctx context.Context,
	killCtx context.Context,
	doneFn func(),
) {
	defer doneFn() // signal startstopper this loop is done

	ctxDone := ctx.Done()
	killChan := killCtx.Done()
	counter := 1

	var (
		dieNow     bool
		inShutdown bool
		inChan     chan int
	)

	defer srv.cleanupLoop(ctx)

	for {
		if dieNow {
			// single return
			return
		}

		inChan = srv.C
		if inShutdown && counter == 0 {
			// e.g., stop reading from chan
			inChan = nil
		}

		// prioritize kill (optional implementation)
		select {
		case <-killChan:
			killChan = nil
			dieNow = true
			continue
		default:
		}

		select {
		case <-killChan:
			killChan = nil // we do not want more signals
			dieNow = true
			continue

		case <-ctxDone:
			ctxDone = nil // we do not want more signals for shutdown
			inShutdown = true

		//	business cases
		case v := <-inChan:
			_ = srv.process(v)

			if inShutdown {
				counter--
				if counter == 0 {
					dieNow = true
				}
			}
		}
	}
}

func (srv *Srv) Close() error {
	srv.StartStopper.Close()
	return nil
}

func (srv *Srv) init(_ context.Context) error {
	// your logic here
	return nil
}

func (srv *Srv) cleanupLoop(_ context.Context) {
	// your logic here
	close(srv.C)
}

func ExampleStartStopper() {
	ctx := context.Background()

	srv := NewSrv(ctx)

	done, err := srv.Go(ctx)
	if err != nil {
		fmt.Println("Go returned error:", err)
		return
	}

	srv.DoStuff(1)
	srv.DoStuff(2)
	srv.DoStuff(3)

	srv.CloseAsync()

	srv.DoStuff(4)

	fmt.Println(srv.Close())

	<-done
	fmt.Println("done")

	// Output:
	// 1
	// 2
	// 3
	// 4
	// <nil>
	// done
}
