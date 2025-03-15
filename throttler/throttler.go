// Package throttler provides a generic Throttler type that throttles the execution of a callback
// function while accumulating input values. The Throttler ensures that the callback is invoked
// at most once per specified duration, processing all input values accumulated during the
// throttling period in batches.
package throttler

import (
	"sync"
	"time"
)

// Throttler is a generic type that limits the frequency of a callback function's execution.
// It accumulates input values of type T into a buffer and invokes the callback with batches
// of these values, ensuring the callback is not called more often than the specified duration.
//
// The Throttler is safe for concurrent use by multiple goroutines. However, care must be taken
// when designing the callback function, as it is executed while holding an internal mutex.
type Throttler[T any] struct {
	mutex       sync.Mutex
	buffer      []T
	resumeTimer *time.Timer
	duration    time.Duration
	callback    func(buf []T)
}

// New creates a new Throttler with the specified duration and callback function.
// The duration defines the minimum time interval between consecutive callback invocations.
// The callback function receives a slice of accumulated values of type T, in the order they
// were added via Push.
//
// Example:
//
//	throttler := throttler.New(time.Second, func(values []string) {
//	    fmt.Println("Processing:", values)
//	})
func New[T any](dur time.Duration, callback func(buf []T)) *Throttler[T] {
	return &Throttler[T]{duration: dur, callback: callback}
}

func (t *Throttler[T]) resume() {
	<-t.resumeTimer.C
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if len(t.buffer) != 0 {
		// If there are values in the buffer, dispatch them and start a new throttling period.
		t.callback(t.buffer)
		t.buffer = t.buffer[:0]
		t.resumeTimer = time.NewTimer(t.duration)
		go t.resume()
	} else {
		// If the buffer is empty, clear the timer.
		t.resumeTimer = nil
	}
}

// Push adds a new value to the Throttler for processing.
// The behavior depends on the Throttler's state:
//   - If no throttling period is active (i.e., no timer is running), the callback is invoked
//     immediately with a slice containing only the pushed value, and a new throttling period
//     begins.
//   - If a throttling period is active, the value is appended to an internal buffer. When the
//     period ends, the callback is invoked with all accumulated values in the order they were
//     pushed.
//
// Note: The callback function is executed while holding an internal mutex, so it should be
// designed to execute quickly to avoid delaying other Push calls. Calling Push from within
// the callback will result in a deadlock, as it attempts to acquire the same mutex already
// held by the callback's execution. Avoid calls to Push in the callback to prevent this issue.
func (t *Throttler[T]) Push(v T) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.resumeTimer == nil {
		// If not in a throttling period, process the value immediately and start throttling.
		t.buffer = append(t.buffer, v)
		t.callback(t.buffer)
		t.buffer = t.buffer[:0]
		t.resumeTimer = time.NewTimer(t.duration)
		go t.resume()
	} else {
		// If in a throttling period, accumulate the value in the buffer.
		t.buffer = append(t.buffer, v)
	}
}
