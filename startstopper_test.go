package startstopper_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Darigaaz/startstopper"
)

var someError = errors.New("some error")

func TestStartStopper_Start(t *testing.T) {
	t.Run("Start", func(t *testing.T) {
		startStopper := startstopper.StartStopper{}
		closingCh, err := startStopper.Start(nil)
		if closingCh == nil || err != nil {
			t.Errorf("want channel and nil error, got: %v, %v", closingCh, err)
		}
	})

	t.Run("Start with ready", func(t *testing.T) {
		startStopper := startstopper.StartStopper{}
		ready := make(chan error, 1)
		closingCh, err := startStopper.Start(ready)

		if closingCh == nil || err != nil {
			t.Errorf("want channel and nil error, got: %v, %v", closingCh, err)
		}

		select {
		case err := <-ready:
			if err != nil {
				t.Errorf("want start without error, got: %v", err)
			}
			return
		case <-time.After(time.Second):
			t.Errorf("failed to start in 1 sec")
		}
	})

	t.Run("Start started", func(t *testing.T) {
		var err error

		startStopper := startstopper.StartStopper{}

		ready1 := make(chan error, 1)
		closingCh1, err1 := startStopper.Start(ready1)
		errReady1 := <-ready1
		if closingCh1 == nil || err1 != nil {
			t.Errorf("want channel and nil error, got: %v, %v", closingCh1, err1)
		}
		if errReady1 != nil {
			t.Errorf("want ready notification without error, got: %v", err)
		}

		ready2 := make(chan error, 1)
		closingCh2, err2 := startStopper.Start(ready2)
		errReady2 := <-ready2
		if closingCh2 != nil || err2 != startstopper.ErrAlreadyStarted {
			t.Errorf("want nil channel and ErrAlreadyStarted error, got: %v, %v", closingCh2, err2)
		}
		if errReady2 != startstopper.ErrAlreadyStarted {
			t.Errorf("want ready notification with startstopper.ErrAlreadyStarted error, got: %v", err)
		}
	})
}

func TestStartStopper_Stop(t *testing.T) {
	t.Run("Stop without notify", func(t *testing.T) {
		startStopper := startstopper.StartStopper{}
		ready := make(chan error, 1)
		closingCh, err := startStopper.Start(ready)
		<-ready

		if closingCh == nil || err != nil {
			t.Errorf("want channel and nil error, got: %v, %v", closingCh, err)
		}

		go func() {
			done := <-closingCh
			done(nil)
		}()

		startStopper.Stop(nil)
	})

	t.Run("Stop without notify with error", func(t *testing.T) {
		startStopper := startstopper.StartStopper{}
		ready := make(chan error, 1)
		closingCh, err := startStopper.Start(ready)
		<-ready

		if closingCh == nil || err != nil {
			t.Errorf("want channel and nil error, got: %v, %v", closingCh, err)
		}

		go func() {
			done := <-closingCh
			done(someError)
		}()

		// error is ignored
		startStopper.Stop(nil)
	})

	t.Run("Stop with notify", func(t *testing.T) {
		startStopper := startstopper.StartStopper{}
		ready := make(chan error, 1)
		closingCh, err := startStopper.Start(ready)
		<-ready

		if closingCh == nil || err != nil {
			t.Errorf("want channel and nil error, got: %v, %v", closingCh, err)
		}

		go func() {
			done := <-closingCh
			done(nil)
		}()

		errStop := make(chan error, 1)
		startStopper.Stop(errStop)

		select {
		case err := <-errStop:
			if err != nil {
				t.Errorf("want stop without error, got: %v", err)
			}
			return
		case <-time.After(time.Second):
			t.Errorf("failed to stop in 1 sec")
		}
	})

	t.Run("Stop with notify with error", func(t *testing.T) {
		startStopper := startstopper.StartStopper{}
		ready := make(chan error, 1)
		closingCh, err := startStopper.Start(ready)
		<-ready

		if closingCh == nil || err != nil {
			t.Errorf("want channel and nil error, got: %v, %v", closingCh, err)
		}

		go func() {
			done := <-closingCh
			done(someError)
		}()

		errStop := make(chan error, 1)
		startStopper.Stop(errStop)

		select {
		case err := <-errStop:
			if err != someError {
				t.Errorf("want stop with someError error, got: %v", err)
			}
			return
		case <-time.After(time.Second):
			t.Errorf("failed to stop in 1 sec")
		}
	})

	t.Run("Stop stopped", func(t *testing.T) {
		startStopper := startstopper.StartStopper{}

		errStop := make(chan error, 1)
		startStopper.Stop(errStop)
		err := <-errStop

		if err != startstopper.ErrAlreadyClosed {
			t.Errorf("want ErrAlreadyClosed, got: %v", err)
		}
	})
}
