// Package timelatch provides a lightweight, goroutine-free, one-shot time trigger.
//
// TimeLatch is a simple struct that allows you to detect the exact moment when
// the current time passes a predefined target time. It is edge-triggered: the
// Triggered() or TriggeredAt() methods return true only once, on the first check
// where the current time has moved from before the target to at-or-after the target.
//
// This makes it ideal for scenarios such as:
//
//   - Implementing delayed actions in a single-threaded loop
//   - Scheduling one-time events without spawning timers or goroutines
//   - Detecting timeouts in long-running operations
//   - Building simple state machines with timed transitions
//
// The latch does not use any background timers or channels. It relies entirely
// on periodic manual checks (e.g., in a game loop, poll loop, or main event loop).
//
// Example usage:
//
//	latch := timelatch.New(time.Now().Add(5 * time.Second))
//	for {
//	    if latch.Triggered() {
//	        fmt.Println("Five seconds have passed!")
//	        break
//	    }
//	    // do other work...
//	    time.Sleep(100 * time.Millisecond)
//	}
//
// The target time can be adjusted dynamically with Add(), AddAt(), SetTime(), or
// SetTimeAt(). All four methods return a boolean indicating whether — at the
// moment of the update and using the reference time for that call — the current
// time was before the new target time (true) or not (false).
//
// Changing the target time resets the edge-trigger detection appropriately.
//
// TimeLatch is safe for concurrent use only if all accesses are serialized
// (e.g., protected by a mutex), as it contains mutable state.
package timelatch

import (
	"time"
)

// TimeLatch is a simple one-shot timer trigger that tracks whether a specific
// moment in time has been reached. It remembers if the target time was in the
// future ("before" the trigger) and can detect the exact moment when current
// time passes the target time.
//
// It is useful for implementing delayed actions, timeouts, or scheduled events
// without requiring a separate goroutine or timer.
//
// The trigger is edge-triggered: Triggered() or TriggeredAt() will return true
// only once, during the call where the current time first becomes >= target time.
// Subsequent calls will return false unless the target time is changed.
type TimeLatch struct {
	before bool      // true if the last observed "now" was before the target time
	t      time.Time // the target time to trigger at
}

// New creates a new TimeLatch that triggers at the given time t.
// The comparison uses the current wall-clock time (time.Now()) as the reference.
func New(t time.Time) TimeLatch {
	return NewAt(t, time.Now())
}

// NewAt creates a new TimeLatch that triggers at the given time t,
// using the provided now as the reference point for determining whether
// the trigger time is in the future or past.
//
// This is useful for testing or when the current time is mocked.
func NewAt(t, now time.Time) TimeLatch {
	return TimeLatch{
		before: now.Before(t),
		t:      t,
	}
}

// Add advances the target trigger time by the given duration d.
//
// It updates the internal edge-detection state based on the current wall-clock time.
//
// Returns whether the reference time (time.Now()) was before the new target time
// at the moment of the update.
func (trig *TimeLatch) Add(d time.Duration) bool {
	return trig.SetTime(trig.t.Add(d))
}

// AddAt advances the target trigger time by the given duration d,
// using the provided now as the reference time.
//
// Returns whether the provided now was before the new target time.
// Useful for deterministic testing or replay scenarios.
func (trig *TimeLatch) AddAt(d time.Duration, now time.Time) bool {
	return trig.SetTimeAt(trig.t.Add(d), now)
}

// SetTime changes the target trigger time to t.
//
// The internal state is updated based on the current wall-clock time (time.Now()).
//
// Returns whether the reference time (time.Now()) was before the new target time
// at the moment of the update. This can be useful for determining whether the
// change crossed the trigger boundary.
func (trig *TimeLatch) SetTime(t time.Time) bool {
	return trig.SetTimeAt(t, time.Now())
}

// SetTimeAt changes the target trigger time to t and updates the internal
// state using the provided now as the reference time.
//
// Returns whether now was before t at the time of the call.
// Primarily useful for deterministic testing or when time is mocked.
func (trig *TimeLatch) SetTimeAt(t time.Time, now time.Time) bool {
	trig.before = now.Before(t)
	trig.t = t
	return trig.before
}

// Time returns the current target time that the latch will trigger at.
func (trig *TimeLatch) Time() time.Time {
	return trig.t
}

// TriggeredAt checks whether the latch has just triggered as of the provided now time.
// It returns true only on the first call where now >= trig.t after previously
// having now < trig.t. Subsequent calls return false until the target time
// is changed again.
//
// The internal state is updated to reflect the new comparison.
func (trig *TimeLatch) TriggeredAt(now time.Time) bool {
	wasBefore := trig.before
	trig.before = now.Before(trig.t)
	return wasBefore && !trig.before
}

// Triggered checks whether the latch has just triggered as of the current
// wall-clock time (time.Now()).
//
// It behaves the same as TriggeredAt(time.Now()).
func (trig *TimeLatch) Triggered() bool {
	return trig.TriggeredAt(time.Now())
}

// AdvanceUntilFuture adds multiples of dur to the target time until it is
// once again strictly after the current wall-clock time (time.Now()).
//
// It is equivalent to:
//
//	for !trig.Add(dur) {
//	    // keep adding until now < target
//	}
//
// Returns true if the final target time is after now (always true unless dur ≤ 0).
// Returns false only in degenerate cases (dur ≤ 0, in which case no change occurs).
//
// This is useful when you want to push the trigger into the future by at least
// one dur step, even if the current target is already in the past.
func (trig *TimeLatch) AdvanceUntilFuture(dur time.Duration) bool {
	if dur <= 0 {
		return time.Now().Before(trig.t) // unchanged state
	}

	now := time.Now()
	for !now.Before(trig.t) {
		trig.t = trig.t.Add(dur)
	}
	trig.before = true // since we exited when now < t
	return true
}

// AdvanceUntilFutureAt is the time-controllable version of AdvanceUntilFuture.
// It adds multiples of dur until the target is strictly after the provided now.
//
// Returns true if the final target is after now, false if dur ≤ 0 (no change).
func (trig *TimeLatch) AdvanceUntilFutureAt(dur time.Duration, now time.Time) bool {
	if dur <= 0 {
		return now.Before(trig.t)
	}

	for !now.Before(trig.t) {
		trig.t = trig.t.Add(dur)
	}
	trig.before = true
	return true
}
