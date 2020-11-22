package startstopper_test

import (
	"fmt"

	"github.com/Darigaaz/startstopper"
)

type Srv struct {
	startstopper.StartStopper

	C chan int
}

func NewSrv() *Srv {
	service := &Srv{
		C: make(chan int),
	}

	return service
}

func (srv *Srv) DoStuff(n int) {
	srv.C <- n
}

func (srv *Srv) Start(signal chan error) {
	var lastError error

	closingCh, err := srv.StartStopper.Start(signal)
	if err != nil {
		return
	}

	for {
		select {
		// listen to closingCh in main loop
		case done := <-closingCh:
			lastError = srv.cleanup()

			// response with error or nil when you are done cleaning up
			done(lastError)

			// dont forget to return
			return

		//	other cases
		case v := <-srv.C:
			fmt.Println(v)
		}
	}
}

func (srv *Srv) Stop(errCh chan error) {
	srv.StartStopper.Stop(errCh)
}

func (srv *Srv) cleanup() error {
	return nil
}

func ExampleStartStopper() {
	srv := NewSrv()

	readyCh := make(chan error, 1)
	// pass readyCh or nil if you dont want to be notified on 'started'
	go srv.Start(readyCh)
	<-readyCh

	srv.DoStuff(1)
	srv.DoStuff(2)
	srv.DoStuff(3)

	stoppedCh := make(chan error, 1)
	// pass stoppedCh or nil if you dont want to be notified on 'stopped'
	srv.Stop(stoppedCh)
	fmt.Println(<-stoppedCh)

	// Output:
	// 1
	// 2
	// 3
	// <nil>
}
