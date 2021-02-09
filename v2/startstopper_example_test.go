package startstopper_test

import (
	"fmt"

	"github.com/Darigaaz/startstopper/v2"
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

func (srv *Srv) Go() error {
	readyCh := make(chan error, 1)

	go srv.Start(readyCh)

	return <-readyCh
}

func (srv *Srv) Start(signalCh chan<- error) {
	err := srv.init()
	startstopper.Notify(signalCh, err, false)
	if err != nil {
		return
	}

	closingCh, err := srv.StartStopper.StartNotify(signalCh)
	if err != nil {
		return
	}

	for {
		select {
		// listen to closingCh in main loop
		case <-closingCh:
			// serving is done

			// dont forget to cleanup in Stop
			return

		//	other cases
		case v := <-srv.C:
			fmt.Println(v)
		}
	}
}

func (srv *Srv) Stop() error {
	err := srv.StartStopper.Stop()
	if err != nil {
		return err
	}

	return srv.cleanup()
}

func (srv *Srv) init() error {
	return nil
}

func (srv *Srv) cleanup() error {
	return nil
}

func ExampleStartStopper() {
	srv := NewSrv()

	fmt.Println(srv.Go())

	srv.DoStuff(1)
	srv.DoStuff(2)
	srv.DoStuff(3)

	fmt.Println(srv.Stop())

	// Output:
	// <nil>
	// 1
	// 2
	// 3
	// <nil>
}
