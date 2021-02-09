package startstopper

func Notify(signalCh chan<- error, err error, closeSignalCh bool) {
	if signalCh != nil {
		if err != nil {
			signalCh <- err
		}

		if closeSignalCh {
			close(signalCh)
		}
	}
}
