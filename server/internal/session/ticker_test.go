package session

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTicker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fixed-Rate Ticker Suite")
}

var _ = Describe("Fixed-Rate Ticker", Label("scope:unit", "loop:g3-orch", "layer:sim", "double:fake-io", "b:tick-determinism", "r:high"), func() {
	const tickInterval = 33 * time.Millisecond // 30 Hz = 33.33ms
	const epsilon = 1 * time.Millisecond        // Allow 1ms tolerance for timing checks

	Describe("Ticker Creation", func() {
		It("creates ticker with correct 30 Hz interval", func() {
			clock := NewFakeClock()
			ticker := NewFixedRateTicker(clock)

			Expect(ticker.interval).To(Equal(tickInterval))
		})

		It("creates ticker with custom interval", func() {
			clock := NewFakeClock()
			customInterval := 50 * time.Millisecond
			ticker := NewTicker(clock, customInterval)

			Expect(ticker.interval).To(Equal(customInterval))
		})
	})

	Describe("Tick Timing Accuracy", func() {
		It("should not tick before interval elapses", func() {
			clock := NewFakeClock()
			ticker := NewFixedRateTicker(clock)

			// Advance time by less than interval
			clock.Advance(tickInterval - 1*time.Millisecond)
			shouldTick := ticker.ShouldTick(clock.Now())

			Expect(shouldTick).To(BeFalse())
		})

		It("should tick after interval elapses", func() {
			clock := NewFakeClock()
			ticker := NewFixedRateTicker(clock)

			// Advance time by exactly interval
			clock.Advance(tickInterval)
			shouldTick := ticker.ShouldTick(clock.Now())

			Expect(shouldTick).To(BeTrue())
		})

		It("should tick after more than interval elapses", func() {
			clock := NewFakeClock()
			ticker := NewFixedRateTicker(clock)

			// Advance time by more than interval
			clock.Advance(tickInterval + 10*time.Millisecond)
			shouldTick := ticker.ShouldTick(clock.Now())

			Expect(shouldTick).To(BeTrue())
		})

		It("Tick returns false if called too early", func() {
			clock := NewFakeClock()
			ticker := NewFixedRateTicker(clock)

			// Advance time by less than interval
			clock.Advance(tickInterval - 1*time.Millisecond)
			ticked := ticker.Tick(clock.Now())

			Expect(ticked).To(BeFalse())
		})

		It("Tick returns true when interval has elapsed", func() {
			clock := NewFakeClock()
			ticker := NewFixedRateTicker(clock)

			// Advance time by exactly interval
			clock.Advance(tickInterval)
			ticked := ticker.Tick(clock.Now())

			Expect(ticked).To(BeTrue())
		})

		It("produces ticks at exactly 30 Hz (33.33ms intervals)", func() {
			clock := NewFakeClock()
			ticker := NewFixedRateTicker(clock)

			// Collect tick times
			var tickTimes []time.Time
			for i := 0; i < 10; i++ {
				clock.Advance(tickInterval)
				if ticker.Tick(clock.Now()) {
					tickTimes = append(tickTimes, clock.Now())
				}
			}

			// Verify we got 10 ticks
			Expect(len(tickTimes)).To(Equal(10))

			// Verify intervals between ticks are approximately correct
			for i := 1; i < len(tickTimes); i++ {
				interval := tickTimes[i].Sub(tickTimes[i-1])
				Expect(interval).To(BeNumerically("~", tickInterval, epsilon))
			}
		})
	})

	Describe("Tick Determinism", func() {
		It("produces identical tick patterns for same time sequence", func() {
			clock1 := NewFakeClock()
			clock2 := NewFakeClock()
			ticker1 := NewFixedRateTicker(clock1)
			ticker2 := NewFixedRateTicker(clock2)

			// Same time sequence
			timeSequence := []time.Duration{
				0,
				tickInterval,
				tickInterval * 2,
				tickInterval * 3,
			}

			var ticks1 []bool
			var ticks2 []bool

			for _, offset := range timeSequence {
				clock1.SetTime(clock1.startTime.Add(offset))
				clock2.SetTime(clock2.startTime.Add(offset))
				ticks1 = append(ticks1, ticker1.Tick(clock1.Now()))
				ticks2 = append(ticks2, ticker2.Tick(clock2.Now()))
			}

			Expect(ticks1).To(Equal(ticks2))
		})

		It("maintains deterministic state across multiple ticks", func() {
			clock := NewFakeClock()
			ticker := NewFixedRateTicker(clock)

			// First tick
			clock.Advance(tickInterval)
			ticked1 := ticker.Tick(clock.Now())
			Expect(ticked1).To(BeTrue())

			// Second tick should not occur immediately
			ticked2 := ticker.Tick(clock.Now())
			Expect(ticked2).To(BeFalse())

			// Advance and second tick should occur
			clock.Advance(tickInterval)
			ticked3 := ticker.Tick(clock.Now())
			Expect(ticked3).To(BeTrue())
		})
	})

	Describe("Ticker State Management", func() {
		It("Reset clears ticker state", func() {
			clock := NewFakeClock()
			ticker := NewFixedRateTicker(clock)

			// Advance and tick
			clock.Advance(tickInterval)
			ticker.Tick(clock.Now())

			// Reset
			ticker.Reset()

			// Should be able to tick again immediately after reset
			clock.Advance(tickInterval)
			ticked := ticker.Tick(clock.Now())
			Expect(ticked).To(BeTrue())
		})

		It("handles time jumps forward correctly", func() {
			clock := NewFakeClock()
			ticker := NewFixedRateTicker(clock)

			// Jump forward by multiple intervals
			clock.Advance(tickInterval * 5)
			now := clock.Now()
			ticked := ticker.Tick(now)

			// Should tick once (not multiple times in a single call)
			Expect(ticked).To(BeTrue())

			// After tick, lastTick = now, so calling Tick() again with same now should not tick
			// This is the expected behavior - Tick() only ticks once per call
			ticked2 := ticker.Tick(now)
			Expect(ticked2).To(BeFalse())

			// Advance time by one more interval and tick should work again
			clock.Advance(tickInterval)
			ticked3 := ticker.Tick(clock.Now())
			Expect(ticked3).To(BeTrue())
		})

		It("handles time moving backward gracefully", func() {
			clock := NewFakeClock()
			ticker := NewFixedRateTicker(clock)

			// Advance and tick
			clock.Advance(tickInterval)
			ticker.Tick(clock.Now())

			// Move time backward
			clock.SetTime(clock.Now().Add(-tickInterval))
			ticked := ticker.Tick(clock.Now())

			// Should not tick (time went backward)
			Expect(ticked).To(BeFalse())

			// Advance forward again
			clock.Advance(tickInterval * 2)
			ticked2 := ticker.Tick(clock.Now())
			Expect(ticked2).To(BeTrue())
		})
	})
})

