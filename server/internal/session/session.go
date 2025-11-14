package session

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/gorbit/orbitalrush/internal/observability"
	"github.com/gorbit/orbitalrush/internal/sim/entities"
	"github.com/gorbit/orbitalrush/internal/sim/rules"
)

// Session orchestrates the game loop by combining ticker, command queue, and game rules.
type Session struct {
	world        entities.World
	queue        *CommandQueue
	ticker       *Ticker
	clock        Clock
	dt           float64
	G            float64
	aMax         float64
	pickupRadius float64
	running      bool
	logger       logr.Logger // Optional logger for observability
}

// NewSession creates a new session with the given clock, initial world state, and max queue size.
func NewSession(clock Clock, world entities.World, maxQueueSize int) *Session {
	return &Session{
		world:        world,
		queue:        NewCommandQueue(maxQueueSize),
		ticker:       NewFixedRateTicker(clock),
		clock:        clock,
		dt:           1.0 / 30.0, // 30Hz tick rate
		G:            1.0,        // Gravitational constant
		aMax:         100.0,      // Maximum acceleration
		pickupRadius: 1.2,        // Pallet pickup radius
		running:      false,
	}
}

// EnqueueCommand adds a command to the queue with the specified sequence number.
// Returns true if the command was successfully enqueued, false otherwise.
func (s *Session) EnqueueCommand(seq uint32, cmd rules.InputCommand) bool {
	return s.queue.Enqueue(seq, cmd)
}

// Run executes the tick loop for up to maxTicks iterations.
// The loop processes commands and calls rules.Step() at the correct tick rate.
// Returns nil on success, or an error if something goes wrong.
func (s *Session) Run(maxTicks int) error {
	s.running = true
	defer func() {
		s.running = false
	}()

	ticksProcessed := 0
	now := s.clock.Now()

	// Calculate how many ticks should occur based on elapsed time
	// This handles the case where time was advanced by multiple intervals
	elapsed := now.Sub(s.ticker.lastTick)

	// Calculate total ticks needed - use integer division to get whole ticks
	totalTicksNeeded := int(elapsed / s.ticker.interval)
	// Ensure we process at least 1 tick if any time has passed
	// This handles edge cases where elapsed is slightly less than interval
	if totalTicksNeeded == 0 && elapsed > 0 {
		totalTicksNeeded = 1
	}

	// Limit to maxTicks (don't process more than requested)
	if totalTicksNeeded > maxTicks {
		totalTicksNeeded = maxTicks
	}

	// Process all ticks that should have occurred
	for ticksProcessed < totalTicksNeeded && !s.world.Done {
		// Measure tick execution time
		tickStart := time.Now()

		// Advance ticker - manually update lastTick to simulate time progression
		s.ticker.lastTick = s.ticker.lastTick.Add(s.ticker.interval)

		// Get next command from queue (or zero command if queue is empty)
		var input rules.InputCommand
		if queuedCmd, ok := s.queue.Dequeue(); ok {
			input = queuedCmd.Command
		} else {
			// No command available, use zero command
			input = rules.InputCommand{Thrust: 0.0, Turn: 0.0}
		}

		// Call rules.Step() to update world state
		s.world = rules.Step(s.world, input, s.dt, s.G, s.aMax, s.pickupRadius)

		ticksProcessed++

		// Measure tick duration and record to metrics
		tickDuration := time.Since(tickStart)
		tickDurationSeconds := tickDuration.Seconds()
		
		// Record to Prometheus histogram (if metrics are initialized)
		if histogram := observability.GetTickDurationHistogram(); histogram != nil {
			histogram.Observe(tickDurationSeconds)
		}

		// Log slow ticks (>10ms threshold)
		const thresholdSeconds = 0.01 // 10ms
		if tickDurationSeconds > thresholdSeconds {
			// Check if logger is enabled (zero logger will return false)
			if s.logger.Enabled() {
				tickNumber := s.world.Tick
				durationMs := tickDurationSeconds * 1000.0
				thresholdMs := thresholdSeconds * 1000.0
				s.logger.WithValues(
					"component", "session",
					"tick", tickNumber,
					"duration_ms", durationMs,
					"threshold_ms", thresholdMs,
				).Info("Tick execution exceeded threshold")
			}
		}

		// If world is done, stop processing
		if s.world.Done {
			break
		}
	}

	return nil
}

// GetWorld returns the current world state.
func (s *Session) GetWorld() entities.World {
	return s.world
}

// IsRunning returns true if the session is currently running.
func (s *Session) IsRunning() bool {
	return s.running
}

// Stop stops the session (sets running to false).
func (s *Session) Stop() {
	s.running = false
}

// SetLogger sets the logger for this session. This is optional and can be nil.
// When set, the logger will be used for structured logging of tick performance.
func (s *Session) SetLogger(logger logr.Logger) {
	s.logger = logger
}
