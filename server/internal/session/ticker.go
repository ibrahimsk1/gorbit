package session

import (
	"time"
)

// Clock is an interface for time abstraction, allowing deterministic testing
// with fake clocks.
type Clock interface {
	Now() time.Time
}

// FakeClock is a deterministic clock implementation for testing.
// It allows precise control over time advancement.
type FakeClock struct {
	startTime time.Time
	currentTime time.Time
}

// NewFakeClock creates a new fake clock starting at the current time.
func NewFakeClock() *FakeClock {
	now := time.Now()
	return &FakeClock{
		startTime:   now,
		currentTime: now,
	}
}

// Now returns the current fake time.
func (f *FakeClock) Now() time.Time {
	return f.currentTime
}

// Advance moves the fake clock forward by the specified duration.
func (f *FakeClock) Advance(d time.Duration) {
	f.currentTime = f.currentTime.Add(d)
}

// SetTime sets the fake clock to a specific time.
func (f *FakeClock) SetTime(t time.Time) {
	f.currentTime = t
}

// RealClock is a wrapper around the real time package for production use.
type RealClock struct{}

// NewRealClock creates a new real clock.
func NewRealClock() *RealClock {
	return &RealClock{}
}

// Now returns the current real time.
func (r *RealClock) Now() time.Time {
	return time.Now()
}

// Ticker generates ticks at a fixed rate.
// It uses a clock interface to allow deterministic testing.
type Ticker struct {
	clock    Clock
	interval time.Duration
	lastTick time.Time
}

// NewTicker creates a new ticker with the specified clock and interval.
func NewTicker(clock Clock, interval time.Duration) *Ticker {
	return &Ticker{
		clock:    clock,
		interval: interval,
		lastTick: clock.Now(),
	}
}

// NewFixedRateTicker creates a new ticker at 30 Hz (33.33ms intervals).
func NewFixedRateTicker(clock Clock) *Ticker {
	// 30 Hz = 1/30 seconds = 33.333... milliseconds
	// Using 33ms is close enough (actual would be 33.333...ms)
	interval := 33 * time.Millisecond
	return NewTicker(clock, interval)
}

// ShouldTick returns true if enough time has passed since the last tick.
func (t *Ticker) ShouldTick(now time.Time) bool {
	elapsed := now.Sub(t.lastTick)
	return elapsed >= t.interval
}

// Tick advances the ticker if enough time has passed.
// Returns true if a tick occurred, false otherwise.
func (t *Ticker) Tick(now time.Time) bool {
	if !t.ShouldTick(now) {
		return false
	}

	// Update lastTick to the current time
	// This ensures we maintain the correct interval even if time jumps
	t.lastTick = now
	return true
}

// Reset resets the ticker state, setting lastTick to the current time.
func (t *Ticker) Reset() {
	t.lastTick = t.clock.Now()
}

