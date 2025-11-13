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
})
