package rules

import (
	"testing"

	"github.com/gorbit/orbitalrush/internal/sim/entities"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStep(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Game Loop Step Suite")
}

var _ = Describe("Game Loop Step", Label("scope:unit", "loop:g2-rules", "layer:sim", "dep:none", "b:game-loop-step", "r:high", "double:fake"), func() {
	const epsilon = 1e-6
	const dt = 1.0 / 30.0 // 30Hz tick rate
	const G = 1.0         // Gravitational constant
	const aMax = 100.0    // Maximum acceleration
	const pickupRadius = 1.2

	Describe("Step Function", func() {
		It("basic step with no input produces correct physics update", func() {
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialPos := world.Ship.Pos
			initialTick := world.Tick

			world = Step(world, input, dt, G, aMax, pickupRadius)

			// Physics should update (gravity pulls ship toward sun)
			Expect(world.Ship.Pos).NotTo(Equal(initialPos))
			// Tick should increment
			Expect(world.Tick).To(Equal(initialTick + 1))
		})

		It("step applies input correctly (thrust)", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0, // Facing right
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}

			initialVel := world.Ship.Vel
			initialEnergy := world.Ship.Energy

			world = Step(world, input, dt, G, aMax, pickupRadius)

			// Velocity should increase (thrust applied)
			Expect(world.Ship.Vel.Length()).To(BeNumerically(">", initialVel.Length()))
			// Energy should decrease (drained by thrust)
			Expect(world.Ship.Energy).To(BeNumerically("<", initialEnergy))
		})

		It("step applies input correctly (turn)", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)
			input := InputCommand{Thrust: 0.0, Turn: 1.0}

			initialRot := world.Ship.Rot

			world = Step(world, input, dt, G, aMax, pickupRadius)

			// Rotation should change
			Expect(world.Ship.Rot).To(BeNumerically(">", initialRot))
		})

		It("step updates physics correctly (gravity + integrator)", func() {
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialPos := world.Ship.Pos
			initialVel := world.Ship.Vel

			world = Step(world, input, dt, G, aMax, pickupRadius)

			// Position should change (gravity pulls toward sun)
			Expect(world.Ship.Pos).NotTo(Equal(initialPos))
			// Velocity should change (gravity accelerates)
			Expect(world.Ship.Vel).NotTo(Equal(initialVel))
		})

		It("step processes pallet pickup correctly (deactivate pallet, restore energy)", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				50.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.0, 0.0), true), // Ship at pallet position
			}
			world := entities.NewWorld(ship, sun, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialEnergy := world.Ship.Energy

			world = Step(world, input, dt, G, aMax, pickupRadius)

			// Pallet should be deactivated
			Expect(world.Pallets[0].Active).To(BeFalse())
			// Energy should be restored
			Expect(world.Ship.Energy).To(BeNumerically("~", initialEnergy+PalletRestoreAmount, epsilon))
		})

		It("step processes multiple pallet pickups in one step", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				50.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			// Place multiple pallets very close to ship (within pickup radius)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.5, 0.0), true),
				entities.NewPallet(2, entities.NewVec2(-0.5, 0.0), true),
			}
			world := entities.NewWorld(ship, sun, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialEnergy := world.Ship.Energy

			world = Step(world, input, dt, G, aMax, pickupRadius)

			// Both pallets should be deactivated
			Expect(world.Pallets[0].Active).To(BeFalse())
			Expect(world.Pallets[1].Active).To(BeFalse())
			// Energy should be restored twice (clamped to MaxEnergy)
			expectedEnergy := initialEnergy + 2.0*PalletRestoreAmount
			if expectedEnergy > MaxEnergy {
				expectedEnergy = MaxEnergy
			}
			Expect(world.Ship.Energy).To(BeNumerically("~", expectedEnergy, epsilon))
		})

		It("step evaluates win condition correctly (all pallets collected)", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.0, 0.0), true), // Ship at pallet position
			}
			world := entities.NewWorld(ship, sun, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			world = Step(world, input, dt, G, aMax, pickupRadius)

			// Win condition should be met
			Expect(world.Done).To(BeTrue())
			Expect(world.Win).To(BeTrue())
		})

		It("step evaluates lose condition correctly (sun collision)", func() {
			ship := entities.NewShip(
				entities.NewVec2(50.0, 0.0), // At sun radius
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(100.0, 0.0), true), // Active pallet far away
			}
			world := entities.NewWorld(ship, sun, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			world = Step(world, input, dt, G, aMax, pickupRadius)

			// Lose condition should be met
			Expect(world.Done).To(BeTrue())
			Expect(world.Win).To(BeFalse())
		})

		It("step correctly prioritizes win over lose condition", func() {
			ship := entities.NewShip(
				entities.NewVec2(50.0, 0.0), // At sun radius
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			// All pallets collected (win condition)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.0, 0.0), false), // Collected
			}
			world := entities.NewWorld(ship, sun, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			world = Step(world, input, dt, G, aMax, pickupRadius)

			// Win should take precedence
			Expect(world.Done).To(BeTrue())
			Expect(world.Win).To(BeTrue())
		})

		It("step increments tick counter correctly", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)
			world.Tick = 42
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			world = Step(world, input, dt, G, aMax, pickupRadius)

			// Tick should increment
			Expect(world.Tick).To(Equal(uint32(43)))
		})

		It("step skips processing when game is already done (only increments tick)", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)
			world.Done = true
			world.Win = true
			world.Tick = 10
			input := InputCommand{Thrust: 1.0, Turn: 1.0}

			initialPos := world.Ship.Pos
			initialVel := world.Ship.Vel
			initialRot := world.Ship.Rot
			initialEnergy := world.Ship.Energy

			world = Step(world, input, dt, G, aMax, pickupRadius)

			// State should be unchanged (except tick)
			Expect(world.Ship.Pos).To(Equal(initialPos))
			Expect(world.Ship.Vel).To(Equal(initialVel))
			Expect(world.Ship.Rot).To(Equal(initialRot))
			Expect(world.Ship.Energy).To(Equal(initialEnergy))
			Expect(world.Done).To(BeTrue())
			Expect(world.Win).To(BeTrue())
			// Tick should increment
			Expect(world.Tick).To(Equal(uint32(11)))
		})
	})

	Describe("Complete Game Scenarios", func() {
		It("full game sequence: start, thrust, pickup, win (multiple steps)", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.5, 0.0), true), // Very close pallet
			}
			world := entities.NewWorld(ship, sun, pallets)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}

			// Simulate multiple steps
			for i := 0; i < 10 && !world.Done; i++ {
				world = Step(world, input, dt, G, aMax, pickupRadius)
			}

			// Game should be won (pallet should be picked up)
			Expect(world.Done).To(BeTrue())
			Expect(world.Win).To(BeTrue())
			// All pallets should be collected
			for _, pallet := range world.Pallets {
				Expect(pallet.Active).To(BeFalse())
			}
		})

		It("lose scenario: ship collides with sun before collecting all pallets", func() {
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(-1.0, 0.0), // Moving toward sun
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(20.0, 0.0), true), // Pallet away from sun
			}
			world := entities.NewWorld(ship, sun, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0} // No input, just gravity

			// Simulate until collision or max steps
			for i := 0; i < 200 && !world.Done; i++ {
				world = Step(world, input, dt, G, aMax, pickupRadius)
			}

			// Game should be lost
			Expect(world.Done).To(BeTrue())
			Expect(world.Win).To(BeFalse())
			// Pallet should still be active (not collected)
			Expect(world.Pallets[0].Active).To(BeTrue())
		})

		It("energy management: low energy, pickup, continue thrusting", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				ThrustDrainRate*3.0, // Low energy
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.5, 0.0), true), // Close pallet
			}
			world := entities.NewWorld(ship, sun, pallets)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}

			initialEnergy := world.Ship.Energy

			// Simulate until energy depleted or pallet picked up
			for i := 0; i < 50 && !world.Done; i++ {
				world = Step(world, input, dt, G, aMax, pickupRadius)
			}

			// Energy should be restored after pickup
			if !world.Pallets[0].Active {
				Expect(world.Ship.Energy).To(BeNumerically(">", initialEnergy-ThrustDrainRate*10.0))
			}
		})

		It("multiple pallets: collect all pallets in sequence", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			// Place pallets very close (within pickup radius) so they're collected immediately
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.5, 0.0), true),
				entities.NewPallet(2, entities.NewVec2(1.0, 0.0), true),
			}
			world := entities.NewWorld(ship, sun, pallets)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}

			// Simulate game loop
			for i := 0; i < 300 && !world.Done; i++ {
				world = Step(world, input, dt, G, aMax, pickupRadius)
			}

			// Game should be won (pallets should be picked up)
			Expect(world.Done).To(BeTrue())
			Expect(world.Win).To(BeTrue())
			// All pallets should be collected
			for _, pallet := range world.Pallets {
				Expect(pallet.Active).To(BeFalse())
			}
		})

		It("determinism: same inputs produce same outputs across multiple steps", func() {
			ship1 := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			ship2 := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world1 := entities.NewWorld(ship1, sun, nil)
			world2 := entities.NewWorld(ship2, sun, nil)
			input := InputCommand{Thrust: 1.0, Turn: 0.5}

			// Apply same inputs multiple times
			for i := 0; i < 10; i++ {
				world1 = Step(world1, input, dt, G, aMax, pickupRadius)
				world2 = Step(world2, input, dt, G, aMax, pickupRadius)
			}

			// States should be identical
			Expect(world1.Ship.Pos.X).To(Equal(world2.Ship.Pos.X))
			Expect(world1.Ship.Pos.Y).To(Equal(world2.Ship.Pos.Y))
			Expect(world1.Ship.Vel.X).To(Equal(world2.Ship.Vel.X))
			Expect(world1.Ship.Vel.Y).To(Equal(world2.Ship.Vel.Y))
			Expect(world1.Ship.Rot).To(Equal(world2.Ship.Rot))
			Expect(world1.Ship.Energy).To(Equal(world2.Ship.Energy))
			Expect(world1.Tick).To(Equal(world2.Tick))
		})
	})

	Describe("Edge Cases", func() {
		It("step works correctly with empty pallet list", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}

			world = Step(world, input, dt, G, aMax, pickupRadius)

			// Should complete without errors
			Expect(world.Tick).To(Equal(uint32(1)))
		})

		It("step works correctly with zero input", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialEnergy := world.Ship.Energy
			world = Step(world, input, dt, G, aMax, pickupRadius)

			// Energy should not change (no thrust)
			Expect(world.Ship.Energy).To(Equal(initialEnergy))
		})

		It("step works correctly when energy is zero (no thrust)", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				0.0, // No energy
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}

			initialVel := world.Ship.Vel
			world = Step(world, input, dt, G, aMax, pickupRadius)

			// Velocity should not change (no thrust without energy)
			Expect(world.Ship.Vel.Length()).To(BeNumerically("~", initialVel.Length(), epsilon))
		})

		It("step works correctly when energy is at maximum (clamping)", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				MaxEnergy,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.0, 0.0), true), // Ship at pallet
			}
			world := entities.NewWorld(ship, sun, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			world = Step(world, input, dt, G, aMax, pickupRadius)

			// Energy should be clamped to MaxEnergy
			Expect(world.Ship.Energy).To(BeNumerically("<=", MaxEnergy, epsilon))
			Expect(world.Ship.Energy).To(Equal(MaxEnergy))
		})

		It("step works correctly over many consecutive steps", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}

			initialTick := world.Tick

			// Run many steps
			for i := 0; i < 100; i++ {
				world = Step(world, input, dt, G, aMax, pickupRadius)
			}

			// Tick should increment correctly
			Expect(world.Tick).To(Equal(initialTick + 100))
			// State should be consistent
			Expect(world.Ship.Energy).To(BeNumerically(">=", 0.0))
			Expect(world.Ship.Energy).To(BeNumerically("<=", MaxEnergy))
		})

		It("step maintains world state consistency", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(5.0, 0.0), true),
			}
			world := entities.NewWorld(ship, sun, pallets)
			input := InputCommand{Thrust: 1.0, Turn: 0.5}

			world = Step(world, input, dt, G, aMax, pickupRadius)

			// All fields should be updated correctly
			Expect(world.Ship.Pos).NotTo(Equal(entities.NewVec2(0.0, 0.0))) // Position changed
			Expect(world.Ship.Rot).NotTo(Equal(0.0))                        // Rotation changed
			Expect(world.Tick).To(Equal(uint32(1)))                        // Tick incremented
			Expect(world.Sun).To(Equal(sun))                                // Sun unchanged
		})
	})
})

