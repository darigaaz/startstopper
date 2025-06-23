package startstopper

// sync.Locker
type locker = interface {
	Lock()
	Unlock()
}

// WithMutex ...
func WithMutex(mu locker, fn func()) {
	mu.Lock()
	defer mu.Unlock()

	fn()
}

// WithMutex1 ...
func WithMutex1[T any](mu locker, fn func() T) T {
	mu.Lock()
	defer mu.Unlock()

	return fn()
}

// WithMutex2 ...
func WithMutex2[T1, T2 any](mu locker, fn func() (T1, T2)) (T1, T2) {
	mu.Lock()
	defer mu.Unlock()

	return fn()
}

// WithMutex3 ...
func WithMutex3[T1, T2, T3 any](mu locker, fn func() (T1, T2, T3)) (T1, T2, T3) {
	mu.Lock()
	defer mu.Unlock()

	return fn()
}
