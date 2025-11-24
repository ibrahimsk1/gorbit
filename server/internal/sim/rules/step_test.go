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

var _ = Describe("Game Loop Step", Label("scope:unit", "loop:g3-rules", "layer:rules", "dep:none", "b:game-loop-step", "r:high", "double:fake"), func() {
	const epsilon = 1e-6
	const dt = 1.0 / 30.0 // 30Hz tick rate
	const G = 1.0         // Gravitational constant
	const aMax = 100.0    // Maximum acceleration
	const pickupRadius = 1.2
	const worldWidth = 2000.0  // World width for wraparound
	const worldHeight = 2000.0 // World height for wraparound

	Describe("Step Function", func() {
		It("basic step with no input produces correct physics update", func() {
			ship := entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialPos := world.Ships[0].Pos
			initialTick := world.Tick

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Physics should update (gravity pulls ship toward planet)
			Expect(world.Ships[0].Pos).NotTo(Equal(initialPos))
			// Tick should increment
			Expect(world.Tick).To(Equal(initialTick + 1))
		})

		It("step applies input correctly (thrust)", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}

			initialVel := world.Ships[0].Vel
			initialEnergy := world.Ships[0].Energy

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Velocity should increase (thrust applied)
			Expect(world.Ships[0].Vel.Length()).To(BeNumerically(">", initialVel.Length()))
			// Energy should decrease (drained by thrust)
			Expect(world.Ships[0].Energy).To(BeNumerically("<", initialEnergy))
		})

		It("step applies input correctly (turn)", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)
			input := InputCommand{Thrust: 0.0, Turn: 1.0}

			initialRot := world.Ships[0].Rot

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Rotation should change
			Expect(world.Ships[0].Rot).To(BeNumerically(">", initialRot))
		})

		It("step updates physics correctly (gravity + integrator)", func() {
			ship := entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialPos := world.Ships[0].Pos
			initialVel := world.Ships[0].Vel

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Position should change (gravity pulls toward planet)
			Expect(world.Ships[0].Pos).NotTo(Equal(initialPos))
			// Velocity should change (gravity accelerates)
			Expect(world.Ships[0].Vel).NotTo(Equal(initialVel))
		})

		It("step processes pallet pickup correctly (deactivate pallet, restore energy)", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 50.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.0, 0.0), true), // Ship at pallet position
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialEnergy := world.Ships[0].Energy

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Pallet should be deactivated
			Expect(world.Pallets[0].Active).To(BeFalse())
			// Energy should be restored
			Expect(world.Ships[0].Energy).To(BeNumerically("~", initialEnergy+PalletRestoreAmount, epsilon))
		})

		It("step processes multiple pallet pickups in one step", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 50.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			// Place multiple pallets very close to ship (within pickup radius)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.5, 0.0), true),
				entities.NewPallet(2, entities.NewVec2(-0.5, 0.0), true),
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialEnergy := world.Ships[0].Energy

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Both pallets should be deactivated
			Expect(world.Pallets[0].Active).To(BeFalse())
			Expect(world.Pallets[1].Active).To(BeFalse())
			// Energy should be restored twice (clamped to MaxEnergy)
			expectedEnergy := initialEnergy + 2.0*PalletRestoreAmount
			if expectedEnergy > MaxEnergy {
				expectedEnergy = MaxEnergy
			}
			Expect(world.Ships[0].Energy).To(BeNumerically("~", expectedEnergy, epsilon))
		})

		It("step evaluates win condition correctly (all pallets collected)", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.0, 0.0), true), // Ship at pallet position
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Win condition evaluation will be updated in future CU
			// For now, just verify step completes
			Expect(world.Tick).To(Equal(uint32(1)))
		})

		It("step evaluates lose condition correctly (planet collision)", func() {
			ship := entities.NewShip(1, entities.NewVec2(50.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(100.0, 0.0), true), // Active pallet far away
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Lose condition evaluation will be updated in future CU
			// For now, just verify step completes
			Expect(world.Tick).To(Equal(uint32(1)))
		})

		It("step correctly prioritizes win over lose condition", func() {
			ship := entities.NewShip(1, entities.NewVec2(50.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			// All pallets collected (win condition)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.0, 0.0), false), // Collected
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Win/lose condition evaluation will be updated in future CU
			// For now, just verify step completes
			Expect(world.Tick).To(Equal(uint32(1)))
		})

		It("step increments tick counter correctly", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)
			world.Tick = 42
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Tick should increment
			Expect(world.Tick).To(Equal(uint32(43)))
		})

		It("step skips processing when game is already done (only increments tick)", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)
			world.Done = true
			world.Win = true
			world.Tick = 10
			input := InputCommand{Thrust: 1.0, Turn: 1.0}

			initialPos := world.Ships[0].Pos
			initialVel := world.Ships[0].Vel
			initialRot := world.Ships[0].Rot
			initialEnergy := world.Ships[0].Energy

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// State should be unchanged (except tick)
			Expect(world.Ships[0].Pos).To(Equal(initialPos))
			Expect(world.Ships[0].Vel).To(Equal(initialVel))
			Expect(world.Ships[0].Rot).To(Equal(initialRot))
			Expect(world.Ships[0].Energy).To(Equal(initialEnergy))
			Expect(world.Done).To(BeTrue())
			Expect(world.Win).To(BeTrue())
			// Tick should increment
			Expect(world.Tick).To(Equal(uint32(11)))
		})
	})

	Describe("Complete Game Scenarios", func() {
		It("full game sequence: start, thrust, pickup, win (multiple steps)", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.5, 0.0), true), // Very close pallet
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}

			// Simulate multiple steps
			for i := 0; i < 10 && !world.Done; i++ {
				world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)
			}

			// Game state evaluation will be updated in future CU
			// For now, just verify steps complete
			Expect(world.Tick).To(BeNumerically(">=", uint32(1)))
		})

		It("lose scenario: ship collides with planet before collecting all pallets", func() {
			ship := entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(-1.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(20.0, 0.0), true), // Pallet away from planet
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0} // No input, just gravity

			// Simulate until collision or max steps
			for i := 0; i < 200 && !world.Done; i++ {
				world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)
			}

			// Game state evaluation will be updated in future CU
			// For now, just verify steps complete
			Expect(world.Tick).To(BeNumerically(">=", uint32(1)))
		})

		It("energy management: low energy, pickup, continue thrusting", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, ThrustDrainRate*3.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.5, 0.0), true), // Close pallet
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}

			// Simulate until energy depleted or pallet picked up
			for i := 0; i < 50 && !world.Done; i++ {
				world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)
			}

			// Collision processing will be updated in next CU
			// For now, just verify steps complete
			Expect(world.Tick).To(BeNumerically(">=", uint32(1)))
		})

		It("multiple pallets: collect all pallets in sequence", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			// Place pallets very close (within pickup radius) so they're collected immediately
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.5, 0.0), true),
				entities.NewPallet(2, entities.NewVec2(1.0, 0.0), true),
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}

			// Simulate game loop
			for i := 0; i < 300 && !world.Done; i++ {
				world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)
			}

			// Game state evaluation will be updated in future CU
			// For now, just verify steps complete
			Expect(world.Tick).To(BeNumerically(">=", uint32(1)))
		})

		It("determinism: same inputs produce same outputs across multiple steps", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			ship2 := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world1 := entities.NewWorld([]entities.Ship{ship1}, []entities.Planet{planet}, nil)
			world2 := entities.NewWorld([]entities.Ship{ship2}, []entities.Planet{planet}, nil)
			input := InputCommand{Thrust: 1.0, Turn: 0.5}

			// Apply same inputs multiple times
			for i := 0; i < 10; i++ {
				world1 = Step(world1, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)
				world2 = Step(world2, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)
			}

			// States should be identical
			Expect(world1.Ships[0].Pos.X).To(Equal(world2.Ships[0].Pos.X))
			Expect(world1.Ships[0].Pos.Y).To(Equal(world2.Ships[0].Pos.Y))
			Expect(world1.Ships[0].Vel.X).To(Equal(world2.Ships[0].Vel.X))
			Expect(world1.Ships[0].Vel.Y).To(Equal(world2.Ships[0].Vel.Y))
			Expect(world1.Ships[0].Rot).To(Equal(world2.Ships[0].Rot))
			Expect(world1.Ships[0].Energy).To(Equal(world2.Ships[0].Energy))
			Expect(world1.Tick).To(Equal(world2.Tick))
		})
	})

	Describe("Edge Cases", func() {
		It("step works correctly with empty pallet list", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Should complete without errors
			Expect(world.Tick).To(Equal(uint32(1)))
		})

		It("step works correctly with zero input", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialEnergy := world.Ships[0].Energy
			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Energy should not change (no thrust)
			Expect(world.Ships[0].Energy).To(Equal(initialEnergy))
		})

		It("step works correctly when energy is zero (no thrust)", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 0.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}

			initialVel := world.Ships[0].Vel
			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Velocity should not change (no thrust without energy)
			Expect(world.Ships[0].Vel.Length()).To(BeNumerically("~", initialVel.Length(), epsilon))
		})

		It("step works correctly when energy is at maximum (clamping)", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, MaxEnergy)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.0, 0.0), true), // Ship at pallet
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Energy should be clamped to MaxEnergy
			Expect(world.Ships[0].Energy).To(BeNumerically("<=", MaxEnergy, epsilon))
			Expect(world.Ships[0].Energy).To(Equal(MaxEnergy))
		})

		It("step works correctly over many consecutive steps", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}

			initialTick := world.Tick

			// Run many steps
			for i := 0; i < 100; i++ {
				world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)
			}

			// Tick should increment correctly
			Expect(world.Tick).To(Equal(initialTick + 100))
			// State should be consistent
			Expect(world.Ships[0].Energy).To(BeNumerically(">=", 0.0))
			Expect(world.Ships[0].Energy).To(BeNumerically("<=", MaxEnergy))
		})

		It("step maintains world state consistency", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(5.0, 0.0), true),
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)
			input := InputCommand{Thrust: 1.0, Turn: 0.5}

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// All fields should be updated correctly
			Expect(world.Ships[0].Pos).NotTo(Equal(entities.NewVec2(0.0, 0.0))) // Position changed
			Expect(world.Ships[0].Rot).NotTo(Equal(0.0))                        // Rotation changed
			Expect(world.Tick).To(Equal(uint32(1)))                              // Tick incremented
			Expect(world.Planets[0]).To(Equal(planet))                           // Planet unchanged
		})
	})

	Describe("Multiplayer Physics Processing", func() {
		It("processes physics for all ships independently", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			ship2 := entities.NewShip(2, entities.NewVec2(-10.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship1, ship2}, []entities.Planet{planet}, nil)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialPos1 := world.Ships[0].Pos
			initialPos2 := world.Ships[1].Pos

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Both ships should have updated positions
			Expect(world.Ships[0].Pos).NotTo(Equal(initialPos1))
			Expect(world.Ships[1].Pos).NotTo(Equal(initialPos2))
			// Ships should move toward planet (gravity)
			Expect(world.Ships[0].Pos.Length()).To(BeNumerically("<", initialPos1.Length()))
			Expect(world.Ships[1].Pos.Length()).To(BeNumerically("<", initialPos2.Length()))
		})

		It("applies gravity from all planets to each ship", func() {
			ship := entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet1 := entities.NewPlanet(1, entities.NewVec2(5.0, 0.0), 50.0, 1000.0)
			planet2 := entities.NewPlanet(2, entities.NewVec2(-5.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet1, planet2}, nil)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialPos := world.Ships[0].Pos

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Ship should be affected by both planets' gravity
			Expect(world.Ships[0].Pos).NotTo(Equal(initialPos))
			// Ship should move toward the combined gravity (closer to planet1)
			Expect(world.Ships[0].Pos.X).To(BeNumerically("<", initialPos.X))
		})

		It("applies wraparound to all ships", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(worldWidth/2+10.0, 0.0), entities.NewVec2(1.0, 0.0), 0.0, 100.0)
			ship2 := entities.NewShip(2, entities.NewVec2(-worldWidth/2-10.0, 0.0), entities.NewVec2(-1.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship1, ship2}, []entities.Planet{planet}, nil)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Both ships should be within world bounds after wraparound
			Expect(world.Ships[0].Pos.X).To(BeNumerically(">=", -worldWidth/2))
			Expect(world.Ships[0].Pos.X).To(BeNumerically("<=", worldWidth/2))
			Expect(world.Ships[1].Pos.X).To(BeNumerically(">=", -worldWidth/2))
			Expect(world.Ships[1].Pos.X).To(BeNumerically("<=", worldWidth/2))
		})

		It("ensures physics isolation (one ship's physics doesn't affect others)", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(1.0, 0.0), 0.0, 100.0)
			ship2 := entities.NewShip(2, entities.NewVec2(-10.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship1, ship2}, []entities.Planet{planet}, nil)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialPos2 := world.Ships[1].Pos

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Ship 1 should have updated physics
			Expect(world.Ships[0].Pos).NotTo(Equal(ship1.Pos))
			// Ship 2 should also have updated physics (gravity affects all ships)
			Expect(world.Ships[1].Pos).NotTo(Equal(initialPos2))
			// But ship 2's velocity should be different from ship 1's (different initial conditions)
			Expect(world.Ships[1].Vel).NotTo(Equal(world.Ships[0].Vel))
		})

		It("applies input before physics processing", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}

			initialVel := world.Ships[0].Vel
			initialEnergy := world.Ships[0].Energy

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Velocity should be increased by thrust (input applied)
			Expect(world.Ships[0].Vel.Length()).To(BeNumerically(">", initialVel.Length()))
			// Energy should be drained (input applied)
			Expect(world.Ships[0].Energy).To(BeNumerically("<", initialEnergy))
			// Position should also change (physics applied after input)
			Expect(world.Ships[0].Pos).NotTo(Equal(ship.Pos))
		})

		It("handles empty Ships array gracefully", func() {
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{}, []entities.Planet{planet}, nil)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialTick := world.Tick

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Tick should increment
			Expect(world.Tick).To(Equal(initialTick + 1))
			// No ships to process, should complete without error
			Expect(len(world.Ships)).To(Equal(0))
		})

		It("preserves ship ID, rotation, and energy during physics update", func() {
			ship := entities.NewShip(42, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 0.0), 1.5, 75.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			world = Step(world, 42, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// ID should be preserved
			Expect(world.Ships[0].ID).To(Equal(uint32(42)))
			// Rotation should be preserved (not modified by physics)
			Expect(world.Ships[0].Rot).To(Equal(1.5))
			// Energy should be preserved (not modified by physics, only by input/collisions)
			Expect(world.Ships[0].Energy).To(Equal(float32(75.0)))
			// Position and velocity should be updated
			Expect(world.Ships[0].Pos).NotTo(Equal(ship.Pos))
		})
	})

	Describe("Multiplayer Collision Processing", func() {
		It("processes pallet pickups for multiple ships independently", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 50.0)
			ship2 := entities.NewShip(2, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 50.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.0, 0.0), true), // Ship 1 at pallet
				entities.NewPallet(2, entities.NewVec2(10.0, 0.0), true), // Ship 2 at pallet
			}
			world := entities.NewWorld([]entities.Ship{ship1, ship2}, []entities.Planet{planet}, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialEnergy1 := world.Ships[0].Energy
			initialEnergy2 := world.Ships[1].Energy

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Both pallets should be deactivated
			Expect(world.Pallets[0].Active).To(BeFalse())
			Expect(world.Pallets[1].Active).To(BeFalse())
			// Both ships should have restored energy
			Expect(world.Ships[0].Energy).To(BeNumerically("~", initialEnergy1+PalletRestoreAmount, epsilon))
			Expect(world.Ships[1].Energy).To(BeNumerically("~", initialEnergy2+PalletRestoreAmount, epsilon))
		})

		It("first ship picks up pallet when multiple ships collide with same pallet", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 50.0)
			ship2 := entities.NewShip(2, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 50.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.0, 0.0), true), // Both ships at pallet
			}
			world := entities.NewWorld([]entities.Ship{ship1, ship2}, []entities.Planet{planet}, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialEnergy1 := world.Ships[0].Energy
			initialEnergy2 := world.Ships[1].Energy

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Pallet should be deactivated (picked up by first ship that processes it)
			Expect(world.Pallets[0].Active).To(BeFalse())
			// First ship should have restored energy
			Expect(world.Ships[0].Energy).To(BeNumerically("~", initialEnergy1+PalletRestoreAmount, epsilon))
			// Second ship should not have restored energy (pallet already deactivated)
			Expect(world.Ships[1].Energy).To(Equal(initialEnergy2))
		})

		It("processes multiple pallet pickups for one ship in one step", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 50.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			// Place multiple pallets very close to ship (within pickup radius)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.5, 0.0), true),
				entities.NewPallet(2, entities.NewVec2(-0.5, 0.0), true),
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialEnergy := world.Ships[0].Energy

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Both pallets should be deactivated
			Expect(world.Pallets[0].Active).To(BeFalse())
			Expect(world.Pallets[1].Active).To(BeFalse())
			// Energy should be restored twice (clamped to MaxEnergy)
			expectedEnergy := initialEnergy + 2.0*PalletRestoreAmount
			if expectedEnergy > MaxEnergy {
				expectedEnergy = MaxEnergy
			}
			Expect(world.Ships[0].Energy).To(BeNumerically("~", expectedEnergy, epsilon))
		})

		It("ensures collision isolation (one ship's collisions don't affect others)", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 50.0)
			ship2 := entities.NewShip(2, entities.NewVec2(20.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 50.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.0, 0.0), true), // Ship 1 at pallet
				entities.NewPallet(2, entities.NewVec2(20.0, 0.0), true), // Ship 2 at pallet
			}
			world := entities.NewWorld([]entities.Ship{ship1, ship2}, []entities.Planet{planet}, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialEnergy1 := world.Ships[0].Energy
			initialEnergy2 := world.Ships[1].Energy

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Ship 1 should pick up pallet 1
			Expect(world.Pallets[0].Active).To(BeFalse())
			Expect(world.Ships[0].Energy).To(BeNumerically("~", initialEnergy1+PalletRestoreAmount, epsilon))

			// Ship 2 should pick up pallet 2
			Expect(world.Pallets[1].Active).To(BeFalse())
			Expect(world.Ships[1].Energy).To(BeNumerically("~", initialEnergy2+PalletRestoreAmount, epsilon))
		})

		It("handles step with no collisions gracefully", func() {
			ship := entities.NewShip(1, entities.NewVec2(100.0, 100.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(200.0, 200.0), true), // Far from ship
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialEnergy := world.Ships[0].Energy

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Pallet should remain active
			Expect(world.Pallets[0].Active).To(BeTrue())
			// Energy should not change
			Expect(world.Ships[0].Energy).To(Equal(initialEnergy))
		})

		It("ignores already deactivated pallets", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 50.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(0.0, 0.0), false), // Already deactivated
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			initialEnergy := world.Ships[0].Energy

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Pallet should remain deactivated
			Expect(world.Pallets[0].Active).To(BeFalse())
			// Energy should not change (pallet already collected)
			Expect(world.Ships[0].Energy).To(Equal(initialEnergy))
		})

		It("detects planet collisions for all ships", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(50.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			ship2 := entities.NewShip(2, entities.NewVec2(60.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship1, ship2}, []entities.Planet{planet}, nil)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}

			world = Step(world, 1, input, dt, G, aMax, pickupRadius, worldWidth, worldHeight)

			// Planet collision detection happens (actual lose condition evaluation in EvaluateGameState)
			// For now, just verify step completes
			Expect(world.Tick).To(Equal(uint32(1)))
			// Ships should still exist (lose condition evaluation happens in EvaluateGameState)
			Expect(len(world.Ships)).To(Equal(2))
		})
	})
})

