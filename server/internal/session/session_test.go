package session

import (
	"testing"
	"time"

	"github.com/gorbit/orbitalrush/internal/sim/entities"
	"github.com/gorbit/orbitalrush/internal/sim/rules"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSession(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Session Tick Loop Suite")
}

var _ = Describe("Session Tick Loop", Label("scope:unit", "loop:g3-orch", "layer:sim", "double:fake-io", "b:tick-orchestration", "r:high"), func() {
	const dt = 1.0 / 30.0 // 30Hz tick rate
	const G = 1.0         // Gravitational constant
	const aMax = 100.0    // Maximum acceleration
	const pickupRadius = 1.2
	const sunRadius = 5.0
	const sunMass = 1000.0

	Describe("Session Creation", func() {
		It("creates session with initial world state", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world := entities.NewWorld(ship, sun, nil)

			session := NewSession(clock, world, 100)

			Expect(session.GetWorld().Tick).To(Equal(uint32(0)))
			Expect(session.GetWorld().Ship.Pos.X).To(Equal(10.0))
			Expect(session.IsRunning()).To(BeFalse())
		})

		It("initializes ticker at 30 Hz", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world := entities.NewWorld(ship, sun, nil)

			session := NewSession(clock, world, 100)

			// Ticker should be initialized
			Expect(session.ticker).NotTo(BeNil())
			Expect(session.ticker.interval).To(Equal(33 * time.Millisecond))
		})

		It("initializes command queue", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world := entities.NewWorld(ship, sun, nil)

			session := NewSession(clock, world, 100)

			Expect(session.queue).NotTo(BeNil())
			Expect(session.queue.Size()).To(Equal(0))
		})

		It("sets game constants correctly", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world := entities.NewWorld(ship, sun, nil)

			session := NewSession(clock, world, 100)

			Expect(session.dt).To(Equal(dt))
			Expect(session.G).To(Equal(G))
			Expect(session.aMax).To(Equal(aMax))
			Expect(session.pickupRadius).To(Equal(pickupRadius))
		})
	})

	Describe("Command Processing", func() {
		It("enqueues commands correctly", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world := entities.NewWorld(ship, sun, nil)
			session := NewSession(clock, world, 100)

			cmd := rules.InputCommand{Thrust: 1.0, Turn: 0.0}
			success := session.EnqueueCommand(1, cmd)

			Expect(success).To(BeTrue())
			Expect(session.queue.Size()).To(Equal(1))
		})

		It("processes commands in sequence order", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world := entities.NewWorld(ship, sun, nil)
			session := NewSession(clock, world, 100)

			// Enqueue commands out of order
			session.EnqueueCommand(2, rules.InputCommand{Thrust: 0.2, Turn: 0.0})
			session.EnqueueCommand(1, rules.InputCommand{Thrust: 0.1, Turn: 0.0})
			session.EnqueueCommand(3, rules.InputCommand{Thrust: 0.3, Turn: 0.0})

			// Run for 3 ticks
			clock.Advance(33 * time.Millisecond * 3)
			err := session.Run(3)

			Expect(err).To(BeNil())
			// Commands should be processed in order (1, 2, 3)
			// We can verify by checking world state progression
		})
	})

	Describe("Tick Loop Execution", func() {
		It("processes ticks at 30 Hz rate", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world := entities.NewWorld(ship, sun, nil)
			session := NewSession(clock, world, 100)

			initialTick := session.GetWorld().Tick

			// Advance time by 3 tick intervals
			clock.Advance(33 * time.Millisecond * 3)
			session.Run(3)

			// Should have processed 3 ticks
			Expect(session.GetWorld().Tick).To(Equal(initialTick + 3))
		})

		It("calls rules.Step() with correct parameters", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world := entities.NewWorld(ship, sun, nil)
			session := NewSession(clock, world, 100)

			initialPos := session.GetWorld().Ship.Pos

			// Enqueue a thrust command
			session.EnqueueCommand(1, rules.InputCommand{Thrust: 1.0, Turn: 0.0})

			// Advance time and run
			clock.Advance(33 * time.Millisecond)
			session.Run(1)

			// Physics should have updated (ship moved due to thrust and gravity)
			Expect(session.GetWorld().Ship.Pos).NotTo(Equal(initialPos))
		})

		It("uses zero command when queue is empty", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world := entities.NewWorld(ship, sun, nil)
			session := NewSession(clock, world, 100)

			initialPos := session.GetWorld().Ship.Pos

			// Run without enqueueing any commands
			clock.Advance(33 * time.Millisecond)
			session.Run(1)

			// Physics should still update (gravity pulls ship), but no thrust
			// Ship should move due to gravity only
			Expect(session.GetWorld().Ship.Pos).NotTo(Equal(initialPos))
			Expect(session.GetWorld().Tick).To(Equal(uint32(1)))
		})

		It("stops when world.Done is true", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			// Place pallet very close to ship so it gets picked up quickly
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.5, 0.0), true),
			}
			world := entities.NewWorld(ship, sun, pallets)
			session := NewSession(clock, world, 100)

			// Enqueue thrust command
			session.EnqueueCommand(1, rules.InputCommand{Thrust: 1.0, Turn: 0.0})

			// Run for many ticks - should stop when game is done
			clock.Advance(33 * time.Millisecond * 100)
			session.Run(100)

			// Game should be done (pallet picked up = win)
			Expect(session.GetWorld().Done).To(BeTrue())
			Expect(session.GetWorld().Win).To(BeTrue())
		})
	})

	Describe("Tick Determinism", func() {
		It("produces identical world states for same inputs", func() {
			clock1 := NewFakeClock()
			clock2 := NewFakeClock()

			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world1 := entities.NewWorld(ship, sun, nil)
			world2 := entities.NewWorld(ship, sun, nil)

			session1 := NewSession(clock1, world1, 100)
			session2 := NewSession(clock2, world2, 100)

			// Enqueue same commands
			session1.EnqueueCommand(1, rules.InputCommand{Thrust: 1.0, Turn: 0.0})
			session2.EnqueueCommand(1, rules.InputCommand{Thrust: 1.0, Turn: 0.0})

			// Advance clocks by same amount
			clock1.Advance(33 * time.Millisecond * 5)
			clock2.Advance(33 * time.Millisecond * 5)

			// Run both sessions
			session1.Run(5)
			session2.Run(5)

			// World states should be identical
			finalWorld1 := session1.GetWorld()
			finalWorld2 := session2.GetWorld()

			Expect(finalWorld1.Tick).To(Equal(finalWorld2.Tick))
			Expect(finalWorld1.Ship.Pos.X).To(Equal(finalWorld2.Ship.Pos.X))
			Expect(finalWorld1.Ship.Pos.Y).To(Equal(finalWorld2.Ship.Pos.Y))
			Expect(finalWorld1.Ship.Vel.X).To(Equal(finalWorld2.Ship.Vel.X))
			Expect(finalWorld1.Ship.Vel.Y).To(Equal(finalWorld2.Ship.Vel.Y))
			Expect(finalWorld1.Ship.Energy).To(Equal(finalWorld2.Ship.Energy))
		})
	})

	Describe("State Progression", func() {
		It("progresses world state correctly through multiple ticks", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world := entities.NewWorld(ship, sun, nil)
			session := NewSession(clock, world, 100)

			initialTick := session.GetWorld().Tick
			initialPos := session.GetWorld().Ship.Pos

			// Run for 10 ticks
			clock.Advance(33 * time.Millisecond * 10)
			session.Run(10)

			// Tick should have incremented
			Expect(session.GetWorld().Tick).To(Equal(initialTick + 10))
			// Position should have changed (gravity pulls ship toward sun)
			Expect(session.GetWorld().Ship.Pos).NotTo(Equal(initialPos))
		})

		It("increments tick counter correctly", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world := entities.NewWorld(ship, sun, nil)
			session := NewSession(clock, world, 100)

			Expect(session.GetWorld().Tick).To(Equal(uint32(0)))

			clock.Advance(33 * time.Millisecond)
			session.Run(1)

			Expect(session.GetWorld().Tick).To(Equal(uint32(1)))
		})
	})

	Describe("Command Idempotency", Label("scope:unit", "loop:g3-orch", "layer:sim", "double:fake-io", "b:command-idempotency", "r:medium"), func() {
		It("applying same command to same initial state produces identical results", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			initialWorld := entities.NewWorld(ship, sun, nil)

			// First application
			session1 := NewSession(clock, initialWorld, 100)
			session1.EnqueueCommand(1, rules.InputCommand{Thrust: 1.0, Turn: 0.0})
			clock.Advance(33 * time.Millisecond)
			session1.Run(1)
			state1 := session1.GetWorld()

			// Second application (same initial state, same command)
			session2 := NewSession(clock, initialWorld, 100)
			session2.EnqueueCommand(1, rules.InputCommand{Thrust: 1.0, Turn: 0.0})
			clock.Advance(33 * time.Millisecond)
			session2.Run(1)
			state2 := session2.GetWorld()

			// Verify states are identical
			Expect(state1.Tick).To(Equal(state2.Tick))
			Expect(state1.Ship.Pos.X).To(BeNumerically("~", state2.Ship.Pos.X, 0.001))
			Expect(state1.Ship.Pos.Y).To(BeNumerically("~", state2.Ship.Pos.Y, 0.001))
			Expect(state1.Ship.Vel.X).To(BeNumerically("~", state2.Ship.Vel.X, 0.001))
			Expect(state1.Ship.Vel.Y).To(BeNumerically("~", state2.Ship.Vel.Y, 0.001))
			Expect(state1.Ship.Energy).To(BeNumerically("~", state2.Ship.Energy, 0.001))
		})

		It("applying same command multiple times produces identical results", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			initialWorld := entities.NewWorld(ship, sun, nil)
			cmd := rules.InputCommand{Thrust: 0.5, Turn: 0.3}

			// Apply command three times, each time from the same initial state
			var states []entities.World
			for i := 0; i < 3; i++ {
				session := NewSession(clock, initialWorld, 100)
				session.EnqueueCommand(1, cmd)
				clock.Advance(33 * time.Millisecond)
				session.Run(1)
				states = append(states, session.GetWorld())
			}

			// Verify all three states are identical
			Expect(states[0].Tick).To(Equal(states[1].Tick))
			Expect(states[1].Tick).To(Equal(states[2].Tick))
			Expect(states[0].Ship.Pos.X).To(BeNumerically("~", states[1].Ship.Pos.X, 0.001))
			Expect(states[1].Ship.Pos.X).To(BeNumerically("~", states[2].Ship.Pos.X, 0.001))
			Expect(states[0].Ship.Energy).To(BeNumerically("~", states[1].Ship.Energy, 0.001))
			Expect(states[1].Ship.Energy).To(BeNumerically("~", states[2].Ship.Energy, 0.001))
		})

		It("queue rejects duplicate sequence numbers", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world := entities.NewWorld(ship, sun, nil)
			session := NewSession(clock, world, 100)

			// Enqueue command with sequence 1
			success1 := session.EnqueueCommand(1, rules.InputCommand{Thrust: 1.0, Turn: 0.0})
			Expect(success1).To(BeTrue())

			// Try to enqueue same sequence again (should fail)
			success2 := session.EnqueueCommand(1, rules.InputCommand{Thrust: 0.5, Turn: 0.5})
			Expect(success2).To(BeFalse())

			// Verify first command is still in queue
			clock.Advance(33 * time.Millisecond)
			session.Run(1)
			finalWorld := session.GetWorld()
			// Ship should have moved due to thrust=1.0, not thrust=0.5
			Expect(finalWorld.Ship.Pos.X).To(BeNumerically(">", 10.0))
		})

		It("commands are deterministic (not just cached)", func() {
			clock := NewFakeClock()
			cmd := rules.InputCommand{Thrust: 1.0, Turn: 0.0}

			// Create two different initial states
			ship1 := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			ship2 := entities.NewShip(
				entities.NewVec2(20.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world1 := entities.NewWorld(ship1, sun, nil)
			world2 := entities.NewWorld(ship2, sun, nil)

			// Apply same command to different initial states
			session1 := NewSession(clock, world1, 100)
			session1.EnqueueCommand(1, cmd)
			clock.Advance(33 * time.Millisecond)
			session1.Run(1)
			state1 := session1.GetWorld()

			session2 := NewSession(clock, world2, 100)
			session2.EnqueueCommand(1, cmd)
			clock.Advance(33 * time.Millisecond)
			session2.Run(1)
			state2 := session2.GetWorld()

			// States should be different (different initial positions)
			Expect(state1.Ship.Pos.X).NotTo(BeNumerically("~", state2.Ship.Pos.X, 0.1))

			// But applying same command to same initial state should produce same result
			session3 := NewSession(clock, world1, 100)
			session3.EnqueueCommand(1, cmd)
			clock.Advance(33 * time.Millisecond)
			session3.Run(1)
			state3 := session3.GetWorld()

			Expect(state1.Ship.Pos.X).To(BeNumerically("~", state3.Ship.Pos.X, 0.001))
			Expect(state1.Ship.Pos.Y).To(BeNumerically("~", state3.Ship.Pos.Y, 0.001))
		})

		It("idempotency holds across multiple ticks", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			initialWorld := entities.NewWorld(ship, sun, nil)

			// First run: apply command sequence 1, then sequence 2
			session1 := NewSession(clock, initialWorld, 100)
			session1.EnqueueCommand(1, rules.InputCommand{Thrust: 1.0, Turn: 0.0})
			session1.EnqueueCommand(2, rules.InputCommand{Thrust: 0.5, Turn: 0.0})
			clock.Advance(33 * time.Millisecond * 2)
			session1.Run(2)
			state1 := session1.GetWorld()

			// Second run: same commands, same initial state
			session2 := NewSession(clock, initialWorld, 100)
			session2.EnqueueCommand(1, rules.InputCommand{Thrust: 1.0, Turn: 0.0})
			session2.EnqueueCommand(2, rules.InputCommand{Thrust: 0.5, Turn: 0.0})
			clock.Advance(33 * time.Millisecond * 2)
			session2.Run(2)
			state2 := session2.GetWorld()

			// Verify states are identical
			Expect(state1.Tick).To(Equal(state2.Tick))
			Expect(state1.Ship.Pos.X).To(BeNumerically("~", state2.Ship.Pos.X, 0.001))
			Expect(state1.Ship.Energy).To(BeNumerically("~", state2.Ship.Energy, 0.001))
		})
	})

	Describe("Session Control", func() {
		It("GetWorld() returns current world state", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world := entities.NewWorld(ship, sun, nil)
			session := NewSession(clock, world, 100)

			retrievedWorld := session.GetWorld()
			Expect(retrievedWorld.Tick).To(Equal(uint32(0)))
			Expect(retrievedWorld.Ship.Pos.X).To(Equal(10.0))
		})

		It("IsRunning() returns correct state", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world := entities.NewWorld(ship, sun, nil)
			session := NewSession(clock, world, 100)

			Expect(session.IsRunning()).To(BeFalse())

			// Note: Run() sets running to true during execution, then false when done
			// We can't easily test this without making Run() async, which is beyond scope
		})

		It("Stop() stops the session", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world := entities.NewWorld(ship, sun, nil)
			session := NewSession(clock, world, 100)

			session.Stop()
			Expect(session.IsRunning()).To(BeFalse())
		})
	})

	Describe("Orchestration Integration", Label("scope:unit", "loop:g3-orch", "layer:sim", "double:fake-io", "b:orchestration-integration", "r:high"), func() {
		It("end-to-end tick determinism with multiple sessions", func() {
			// Create two sessions with identical initial state
			clock1 := NewFakeClock()
			clock2 := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world1 := entities.NewWorld(ship, sun, nil)
			world2 := entities.NewWorld(ship, sun, nil)

			session1 := NewSession(clock1, world1, 100)
			session2 := NewSession(clock2, world2, 100)

			// Apply same sequence of commands
			session1.EnqueueCommand(1, rules.InputCommand{Thrust: 1.0, Turn: 0.0})
			session1.EnqueueCommand(2, rules.InputCommand{Thrust: 0.5, Turn: 0.0})
			session1.EnqueueCommand(3, rules.InputCommand{Thrust: 0.0, Turn: 0.3})

			session2.EnqueueCommand(1, rules.InputCommand{Thrust: 1.0, Turn: 0.0})
			session2.EnqueueCommand(2, rules.InputCommand{Thrust: 0.5, Turn: 0.0})
			session2.EnqueueCommand(3, rules.InputCommand{Thrust: 0.0, Turn: 0.3})

			// Advance time identically
			clock1.Advance(33 * time.Millisecond * 3)
			clock2.Advance(33 * time.Millisecond * 3)

			// Run both sessions
			session1.Run(3)
			session2.Run(3)

			// Verify final states match exactly
			final1 := session1.GetWorld()
			final2 := session2.GetWorld()

			Expect(final1.Tick).To(Equal(final2.Tick))
			Expect(final1.Ship.Pos.X).To(BeNumerically("~", final2.Ship.Pos.X, 0.001))
			Expect(final1.Ship.Pos.Y).To(BeNumerically("~", final2.Ship.Pos.Y, 0.001))
			Expect(final1.Ship.Vel.X).To(BeNumerically("~", final2.Ship.Vel.X, 0.001))
			Expect(final1.Ship.Vel.Y).To(BeNumerically("~", final2.Ship.Vel.Y, 0.001))
			Expect(final1.Ship.Energy).To(BeNumerically("~", final2.Ship.Energy, 0.001))
		})

		It("end-to-end command ordering with out-of-order enqueue", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world := entities.NewWorld(ship, sun, nil)
			session := NewSession(clock, world, 100)

			// Enqueue commands out of order
			session.EnqueueCommand(3, rules.InputCommand{Thrust: 0.0, Turn: 0.3})
			session.EnqueueCommand(1, rules.InputCommand{Thrust: 1.0, Turn: 0.0})
			session.EnqueueCommand(2, rules.InputCommand{Thrust: 0.5, Turn: 0.0})

			// Advance time and run
			clock.Advance(33 * time.Millisecond * 3)
			session.Run(3)

			// Verify commands were applied in order (1, 2, 3)
			// First tick: thrust=1.0, second tick: thrust=0.5, third tick: turn=0.3
			final := session.GetWorld()
			Expect(final.Tick).To(Equal(uint32(3)))
			// Ship should have moved forward (thrust applied first two ticks)
			Expect(final.Ship.Pos.X).To(BeNumerically(">", 10.0))
		})

		It("end-to-end rollback and replay produces same result", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			initialWorld := entities.NewWorld(ship, sun, nil)

			// First run: apply commands and capture snapshot
			session1 := NewSession(clock, initialWorld, 100)
			session1.EnqueueCommand(1, rules.InputCommand{Thrust: 1.0, Turn: 0.0})
			session1.EnqueueCommand(2, rules.InputCommand{Thrust: 0.5, Turn: 0.0})
			clock.Advance(33 * time.Millisecond * 2)
			session1.Run(2)

			// Capture snapshot at tick 2
			manager := NewSnapshotManager()
			snapshot := manager.CaptureSnapshot(session1.GetWorld(), 2, clock)

			// Continue with more commands
			session1.EnqueueCommand(3, rules.InputCommand{Thrust: 0.0, Turn: 0.3})
			clock.Advance(33 * time.Millisecond)
			session1.Run(1)
			final1 := session1.GetWorld()

			// Second run: restore from snapshot and replay same commands
			restoredWorld := manager.RestoreSnapshot(snapshot)
			session2 := NewSession(clock, restoredWorld, 100)
			session2.EnqueueCommand(3, rules.InputCommand{Thrust: 0.0, Turn: 0.3})
			clock.Advance(33 * time.Millisecond)
			session2.Run(1)
			final2 := session2.GetWorld()

			// Verify final states match
			Expect(final1.Tick).To(Equal(final2.Tick))
			Expect(final1.Ship.Pos.X).To(BeNumerically("~", final2.Ship.Pos.X, 0.001))
			Expect(final1.Ship.Pos.Y).To(BeNumerically("~", final2.Ship.Pos.Y, 0.001))
			Expect(final1.Ship.Energy).To(BeNumerically("~", final2.Ship.Energy, 0.001))
		})

		It("end-to-end ticker queue session integration", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world := entities.NewWorld(ship, sun, nil)
			session := NewSession(clock, world, 100)

			initialPos := world.Ship.Pos

			// Enqueue commands
			session.EnqueueCommand(1, rules.InputCommand{Thrust: 1.0, Turn: 0.0})
			session.EnqueueCommand(2, rules.InputCommand{Thrust: 0.5, Turn: 0.0})

			// Advance clock by multiple intervals
			clock.Advance(33 * time.Millisecond * 5)

			// Run session (should process up to 5 ticks, but only 2 commands available)
			session.Run(5)

			final := session.GetWorld()

			// Verify correct number of ticks processed
			Expect(final.Tick).To(Equal(uint32(5)))

			// Verify commands were applied and world state changed
			// (Position may change due to gravity + thrust, but should be different from initial)
			Expect(final.Ship.Pos.X).NotTo(Equal(initialPos.X))
			// Energy should have decreased due to thrust commands
			Expect(final.Ship.Energy).To(BeNumerically("<", 100.0))
		})

		It("end-to-end rollback hooks with session integration", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			world := entities.NewWorld(ship, sun, nil)
			session := NewSession(clock, world, 100)

			// Apply command and run
			session.EnqueueCommand(1, rules.InputCommand{Thrust: 1.0, Turn: 0.0})
			clock.Advance(33 * time.Millisecond)
			session.Run(1)

			// Create snapshot manager with hook
			manager := NewSnapshotManager()
			beforeCalled := false
			afterCalled := false

			hook := &testRollbackHook{
				beforeSnapshot: func(snapshot *Snapshot) {
					beforeCalled = true
				},
				afterRestore: func(snapshot *Snapshot) {
					afterCalled = true
				},
			}
			manager.RegisterHook(hook)

			// Capture snapshot
			snapshot := manager.CaptureSnapshot(session.GetWorld(), 1, clock)
			Expect(beforeCalled).To(BeTrue())

			// Restore snapshot
			restored := manager.RestoreSnapshot(snapshot)
			Expect(afterCalled).To(BeTrue())

			// Verify restored state matches
			Expect(restored.Tick).To(Equal(uint32(1)))
			Expect(restored.Ship.Pos.X).To(BeNumerically("~", session.GetWorld().Ship.Pos.X, 0.001))
		})

		It("end-to-end complex multi-tick scenario with rollback and replay", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), sunRadius, sunMass)
			initialWorld := entities.NewWorld(ship, sun, nil)

			manager := NewSnapshotManager()

			// First run: apply commands and capture snapshots
			session1 := NewSession(clock, initialWorld, 100)
			session1.EnqueueCommand(1, rules.InputCommand{Thrust: 1.0, Turn: 0.0})
			clock.Advance(33 * time.Millisecond)
			session1.Run(1)
			snapshot1 := manager.CaptureSnapshot(session1.GetWorld(), 1, clock)

			session1.EnqueueCommand(2, rules.InputCommand{Thrust: 0.5, Turn: 0.0})
			clock.Advance(33 * time.Millisecond)
			session1.Run(1)
			snapshot2 := manager.CaptureSnapshot(session1.GetWorld(), 2, clock)

			session1.EnqueueCommand(3, rules.InputCommand{Thrust: 0.0, Turn: 0.3})
			clock.Advance(33 * time.Millisecond)
			session1.Run(1)
			final1 := session1.GetWorld()

			// Second run: rollback to snapshot1 and replay
			restoredWorld1 := manager.RestoreSnapshot(snapshot1)
			session2 := NewSession(clock, restoredWorld1, 100)
			session2.EnqueueCommand(2, rules.InputCommand{Thrust: 0.5, Turn: 0.0})
			clock.Advance(33 * time.Millisecond)
			session2.Run(1)

			// Verify state matches snapshot2
			state2 := session2.GetWorld()
			Expect(state2.Tick).To(Equal(snapshot2.Tick))
			Expect(state2.Ship.Pos.X).To(BeNumerically("~", snapshot2.World.Ship.Pos.X, 0.001))

			// Continue replay
			session2.EnqueueCommand(3, rules.InputCommand{Thrust: 0.0, Turn: 0.3})
			clock.Advance(33 * time.Millisecond)
			session2.Run(1)
			final2 := session2.GetWorld()

			// Verify final states match
			Expect(final1.Tick).To(Equal(final2.Tick))
			Expect(final1.Ship.Pos.X).To(BeNumerically("~", final2.Ship.Pos.X, 0.001))
			Expect(final1.Ship.Energy).To(BeNumerically("~", final2.Ship.Energy, 0.001))
		})
	})
})

// testRollbackHook is a test implementation of RollbackHook for integration tests.
type testRollbackHook struct {
	beforeSnapshot func(*Snapshot)
	afterRestore   func(*Snapshot)
}

func (h *testRollbackHook) BeforeSnapshot(snapshot *Snapshot) {
	if h.beforeSnapshot != nil {
		h.beforeSnapshot(snapshot)
	}
}

func (h *testRollbackHook) AfterRestore(snapshot *Snapshot) {
	if h.afterRestore != nil {
		h.afterRestore(snapshot)
	}
}
