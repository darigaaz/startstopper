package startstopper_test

import (
	"testing"

	"github.com/Darigaaz/startstopper/v2"
)

func TestStartStopper_StartNotify(t *testing.T) {
	t.Run("StartNotify", func(t *testing.T) {
		startStopper := startstopper.StartStopper{}
		signalCh := make(chan error, 1)
		closingCh, err := startStopper.StartNotify(signalCh)

		signaledErr := <-signalCh

		if closingCh == nil || err != nil || signaledErr != nil {
			t.Errorf("want channel and nil error and nil error, got: %v, '%v', '%v'", closingCh, err, signaledErr)
		}
	})

	t.Run("StartNotify started", func(t *testing.T) {
		startStopper := startstopper.StartStopper{}
		signalCh1 := make(chan error, 1)
		closingCh1, err1 := startStopper.StartNotify(signalCh1)

		signaledErr1 := <-signalCh1

		if closingCh1 == nil || err1 != nil || signaledErr1 != nil {
			t.Errorf("want channel and nil error and nil error, got: %v, '%v', '%v'", closingCh1, err1, signaledErr1)
		}

		signalCh2 := make(chan error, 1)
		closingCh2, err2 := startStopper.StartNotify(signalCh2)

		signaledErr2 := <-signalCh2

		if closingCh2 != nil || err2 != startstopper.ErrAlreadyStarted || signaledErr2 != startstopper.ErrAlreadyStarted {
			t.Errorf(
				"want nil channel and '%v' error and '%v' error, got: %v, '%v', '%v'",
				startstopper.ErrAlreadyStarted,
				startstopper.ErrAlreadyStarted,
				closingCh2,
				err2,
				signaledErr2,
			)
		}
	})
}

func TestStartStopper_Start(t *testing.T) {
	t.Run("Start", func(t *testing.T) {
		startStopper := startstopper.StartStopper{}
		closingCh, err := startStopper.Start()
		if closingCh == nil || err != nil {
			t.Errorf("want channel and nil error, got: %v, %v", closingCh, err)
		}
	})

	t.Run("Start started", func(t *testing.T) {
		startStopper := startstopper.StartStopper{}

		closingCh1, err1 := startStopper.Start()
		if closingCh1 == nil || err1 != nil {
			t.Errorf("want channel and nil error, got: %v, %v", closingCh1, err1)
		}

		closingCh2, err2 := startStopper.Start()
		if closingCh2 != nil || err2 != startstopper.ErrAlreadyStarted {
			t.Errorf("want nil channel and '%v' error, got: %v, '%v'", startstopper.ErrAlreadyStarted, closingCh2, err2)
		}
	})
}

func TestStartStopper_Stop(t *testing.T) {
	t.Run("Stop", func(t *testing.T) {
		startStopper := startstopper.StartStopper{}
		closingCh, err := startStopper.Start()

		if closingCh == nil || err != nil {
			t.Errorf("want channel and nil error, got: %v, '%v'", closingCh, err)
		}

		go func() {
			<-closingCh
		}()

		err = startStopper.Stop()
		if err != nil {
			t.Errorf("want nil error, got: '%v'", err)
		}
	})

	t.Run("Stop stopped", func(t *testing.T) {
		startStopper := startstopper.StartStopper{}

		err := startStopper.Stop()
		if err != startstopper.ErrAlreadyStopped {
			t.Errorf("want '%v' error, got: '%v'", startstopper.ErrAlreadyStopped, err)
		}
	})
}

func TestStartStopper_StartStopStartStop(t *testing.T) {
	t.Run("multiple Start Stop", func(t *testing.T) {
		startStopper := startstopper.StartStopper{}

		for i := 0; i < 2; i++ {
			closingCh, err := startStopper.Start()

			if closingCh == nil || err != nil {
				t.Errorf("loop %d: want channel and nil error, got: %v, '%v'", i+1, closingCh, err)
			}

			go func() {
				<-closingCh
			}()

			err = startStopper.Stop()
			if err != nil {
				t.Errorf("want nil error, got: '%v'", err)
			}
		}
	})
}
