package startstopper

type NotifyCloseMode int

const (
	NotifyCloseModeOnError NotifyCloseMode = iota
	NotifyCloseModeAlways
)

// Notify ...
func Notify(signalCh chan<- error, err error, closeMode NotifyCloseMode) {
	needClose := closeMode == NotifyCloseModeAlways

	if signalCh != nil {
		if err != nil {
			signalCh <- err
			needClose = closeMode == NotifyCloseModeOnError
		}

		if needClose {
			close(signalCh)
		}
	}
}
