package startstopper_test

import (
	"testing"
	"time"

	"github.com/Darigaaz/startstopper/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func waitValue[T any](v *T, ch chan T) func() bool {
	return func() bool {
		select {
		case *v = <-ch:
			return true
		default:
			return false
		}
	}
}

func TestStartStopper_StartNotify(t *testing.T) {
	t.Run("Start", func(t *testing.T) {
		startStopper := startstopper.StartStopper{}
		err := startStopper.Init(t.Context(), nil)
		require.NoError(t, err)
		t.Cleanup(startStopper.Close)

		signalCh := make(chan error, 1)
		cleanupDoneChan, cleanupDoneFunc := startstopper.ChanCloser(nil)
		t.Cleanup(cleanupDoneFunc)

		// SUT
		ctx, killCtx, done, err := startStopper.Start(t.Context(), cleanupDoneChan, signalCh, nil)
		require.NoError(t, err)
		assert.NotNil(t, ctx)
		assert.NotNil(t, killCtx)
		assert.NotNil(t, done)

		var signaledErr error
		assert.Eventually(t, waitValue(&signaledErr, signalCh), time.Second, time.Millisecond)
		require.NoError(t, signaledErr)
	})

	t.Run("error on multiple Start", func(t *testing.T) {
		startStopper := startstopper.StartStopper{}
		err := startStopper.Init(t.Context(), nil)
		require.NoError(t, err)
		t.Cleanup(startStopper.Close)

		signalCh1 := make(chan error, 1)
		cleanupDoneChan1, cleanupDoneFunc1 := startstopper.ChanCloser(nil)
		t.Cleanup(cleanupDoneFunc1)

		// SUT
		ctx, killCtx, done, err := startStopper.Start(t.Context(), cleanupDoneChan1, signalCh1, nil)
		require.NoError(t, err)
		assert.NotNil(t, ctx)
		assert.NotNil(t, killCtx)
		assert.NotNil(t, done)

		var signaledErr error
		assert.Eventually(t, waitValue(&signaledErr, signalCh1), time.Second, time.Millisecond)
		require.NoError(t, signaledErr)

		signalCh2 := make(chan error, 1)
		cleanupDoneChan2, cleanupDoneFunc2 := startstopper.ChanCloser(nil)
		t.Cleanup(cleanupDoneFunc2)

		// SUT
		ctx2, killCtx2, done2, err2 := startStopper.Start(t.Context(), cleanupDoneChan2, signalCh2, nil)
		require.ErrorIs(t, err2, startstopper.ErrStart)
		assert.Nil(t, ctx2)
		assert.Nil(t, killCtx2)
		assert.Nil(t, done2)

		assert.Eventually(t, waitValue(&signaledErr, signalCh2), time.Second, time.Millisecond)
		require.ErrorIs(t, signaledErr, startstopper.ErrStart)
	})
}

func TestStartStopper_StartStopStartStop(t *testing.T) {
	t.Run("multiple Start Stop", func(t *testing.T) {
		startStopper := startstopper.StartStopper{}
		err := startStopper.Init(t.Context(), nil)
		require.NoError(t, err)
		t.Cleanup(startStopper.Close)

		for i := 0; i < 5; i++ {
			signalCh := make(chan error, 1)
			cleanupDoneChan, cleanupDoneFunc := startstopper.ChanCloser(nil)
			// t.Cleanup(cleanupDoneFunc)

			// SUT
			ctx, killCtx, done, err := startStopper.Start(t.Context(), cleanupDoneChan, signalCh, nil)
			require.NoError(t, err)
			assert.NotNil(t, ctx)
			assert.NotNil(t, killCtx)
			assert.NotNil(t, done)

			var signaledErr error
			assert.Eventually(t, waitValue(&signaledErr, signalCh), time.Second, time.Millisecond)
			require.NoError(t, signaledErr)

			// We are done, wait for completion
			cleanupDoneFunc()
			<-done
		}
	})
}
