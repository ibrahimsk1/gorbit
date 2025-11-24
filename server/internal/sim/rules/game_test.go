package rules

import (
	"testing"

	"github.com/gorbit/orbitalrush/internal/sim/entities"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGame(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Game State Transitions Suite")
}

var _ = Describe("Game State Transitions", Label("scope:unit", "loop:g3-rules", "layer:rules", "dep:none", "b:game-state-transitions", "r:high", "double:fake"), func() {
	Describe("CheckWinCondition", func() {
		It("returns true when all pallets are collected", func() {
			// TODO: Temporary fix - will be updated in cu/evaluate-gamestate-multiplayer
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false), // Collected
				entities.NewPallet(2, entities.NewVec2(20.0, 20.0), false), // Collected
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)

			result := CheckWinCondition(world)
			Expect(result).To(BeTrue())
		})

		It("returns false when there are no pallets", func() {
			// TODO: Temporary fix - will be updated in cu/evaluate-gamestate-multiplayer
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{} // Empty
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)

			result := CheckWinCondition(world)
			Expect(result).To(BeFalse())
		})

		It("returns false when at least one pallet is active", func() {
			// TODO: Temporary fix - will be updated in cu/evaluate-gamestate-multiplayer
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false), // Collected
				entities.NewPallet(2, entities.NewVec2(20.0, 20.0), true),  // Still active
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)

			result := CheckWinCondition(world)
			Expect(result).To(BeFalse())
		})

		It("returns false when all pallets are active", func() {
			// TODO: Temporary fix - will be updated in cu/evaluate-gamestate-multiplayer
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), true),
				entities.NewPallet(2, entities.NewVec2(20.0, 20.0), true),
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)

			result := CheckWinCondition(world)
			Expect(result).To(BeFalse())
		})

		It("returns true when single pallet is collected", func() {
			// TODO: Temporary fix - will be updated in cu/evaluate-gamestate-multiplayer
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false), // Collected
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)

			result := CheckWinCondition(world)
			Expect(result).To(BeTrue())
		})
	})

	Describe("CheckLoseCondition", func() {
		It("returns true when ship collides with planet", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)

			result := CheckLoseCondition(world, 1)

			Expect(result).To(BeTrue())
		})

		It("returns true when ship is exactly at planet radius", func() {
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			ship := entities.NewShip(1, entities.NewVec2(float64(planet.Radius), 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)

			result := CheckLoseCondition(world, 1)

			Expect(result).To(BeTrue())
		})

		It("returns true when ship is within planet radius", func() {
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			ship := entities.NewShip(1, entities.NewVec2(25.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)

			result := CheckLoseCondition(world, 1)

			Expect(result).To(BeTrue())
		})

		It("returns false when ship is outside planet radius", func() {
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			ship := entities.NewShip(1, entities.NewVec2(100.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)

			result := CheckLoseCondition(world, 1)

			Expect(result).To(BeFalse())
		})

		It("returns false when ship is just outside planet radius", func() {
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			ship := entities.NewShip(1, entities.NewVec2(float64(planet.Radius)+0.1, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)

			result := CheckLoseCondition(world, 1)

			Expect(result).To(BeFalse())
		})

		It("returns false when ship is far from planet", func() {
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			ship := entities.NewShip(1, entities.NewVec2(1000.0, 1000.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)

			result := CheckLoseCondition(world, 1)

			Expect(result).To(BeFalse())
		})

		It("returns true when ship collides with any planet", func() {
			ship := entities.NewShip(1, entities.NewVec2(5.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet1 := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			planet2 := entities.NewPlanet(2, entities.NewVec2(100.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet1, planet2}, nil)

			result := CheckLoseCondition(world, 1)

			// Ship collides with planet1 (within radius)
			Expect(result).To(BeTrue())
		})

		It("returns false when ship does not collide with any planet", func() {
			ship := entities.NewShip(1, entities.NewVec2(100.0, 100.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet1 := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			planet2 := entities.NewPlanet(2, entities.NewVec2(200.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet1, planet2}, nil)

			result := CheckLoseCondition(world, 1)

			Expect(result).To(BeFalse())
		})

		It("returns false when shipID is not found", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, nil)

			result := CheckLoseCondition(world, 999) // Invalid shipID

			Expect(result).To(BeFalse())
		})

		It("checks only the specified ship by shipID", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0) // Collides
			ship2 := entities.NewShip(2, entities.NewVec2(100.0, 100.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0) // No collision
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship1, ship2}, []entities.Planet{planet}, nil)

			result1 := CheckLoseCondition(world, 1)
			result2 := CheckLoseCondition(world, 2)

			// Ship 1 collides
			Expect(result1).To(BeTrue())
			// Ship 2 does not collide
			Expect(result2).To(BeFalse())
		})

		It("detects collision with different planets correctly", func() {
			ship := entities.NewShip(1, entities.NewVec2(105.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet1 := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			planet2 := entities.NewPlanet(2, entities.NewVec2(100.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet1, planet2}, nil)

			result := CheckLoseCondition(world, 1)

			// Ship collides with planet2 (within radius)
			Expect(result).To(BeTrue())
		})

		It("returns false when there are no planets", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{}, nil)

			result := CheckLoseCondition(world, 1)

			Expect(result).To(BeFalse())
		})
	})

	Describe("EvaluateGameState", func() {
		It("sets Done=true and Win=true when all pallets collected", func() {
			ship := entities.NewShip(1, entities.NewVec2(100.0, 100.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0) // Far from planet
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false), // Collected
				entities.NewPallet(2, entities.NewVec2(20.0, 20.0), false), // Collected
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)

			updatedWorld := EvaluateGameState(world)

			Expect(updatedWorld.Done).To(BeTrue())
			Expect(updatedWorld.Win).To(BeTrue())
		})

		It("sets Done=true and Win=false when ship collides with planet", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0) // At planet center
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), true), // Still active
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)

			updatedWorld := EvaluateGameState(world)

			Expect(updatedWorld.Done).To(BeTrue())
			Expect(updatedWorld.Win).To(BeFalse())
		})

		It("leaves Done and Win unchanged when neither condition is met", func() {
			ship := entities.NewShip(1, entities.NewVec2(100.0, 100.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0) // Far from planet
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), true), // Still active
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)

			updatedWorld := EvaluateGameState(world)

			Expect(updatedWorld.Done).To(BeFalse())
			Expect(updatedWorld.Win).To(BeFalse())
		})

		It("sets Done=true and Win=false when any ship collides with planet", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0) // Collides
			ship2 := entities.NewShip(2, entities.NewVec2(100.0, 100.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0) // No collision
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), true), // Still active
			}
			world := entities.NewWorld([]entities.Ship{ship1, ship2}, []entities.Planet{planet}, pallets)

			updatedWorld := EvaluateGameState(world)

			// Ship1 collides, so game should end with lose
			Expect(updatedWorld.Done).To(BeTrue())
			Expect(updatedWorld.Win).To(BeFalse())
		})

		It("leaves Done and Win unchanged when no ships collide", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(100.0, 100.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			ship2 := entities.NewShip(2, entities.NewVec2(200.0, 200.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), true), // Still active
			}
			world := entities.NewWorld([]entities.Ship{ship1, ship2}, []entities.Planet{planet}, pallets)

			updatedWorld := EvaluateGameState(world)

			// No collisions, game continues
			Expect(updatedWorld.Done).To(BeFalse())
			Expect(updatedWorld.Win).To(BeFalse())
		})

		It("sets Done=true and Win=true when all pallets collected (multiple ships)", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(100.0, 100.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			ship2 := entities.NewShip(2, entities.NewVec2(200.0, 200.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false), // Collected
				entities.NewPallet(2, entities.NewVec2(20.0, 20.0), false), // Collected
			}
			world := entities.NewWorld([]entities.Ship{ship1, ship2}, []entities.Planet{planet}, pallets)

			updatedWorld := EvaluateGameState(world)

			// All pallets collected, win condition
			Expect(updatedWorld.Done).To(BeTrue())
			Expect(updatedWorld.Win).To(BeTrue())
		})

		It("gives win precedence when both win and lose conditions are true (multiple ships)", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0) // Collides
			ship2 := entities.NewShip(2, entities.NewVec2(100.0, 100.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false), // Collected (win condition)
			}
			world := entities.NewWorld([]entities.Ship{ship1, ship2}, []entities.Planet{planet}, pallets)

			updatedWorld := EvaluateGameState(world)

			// Win should take precedence
			Expect(updatedWorld.Done).To(BeTrue())
			Expect(updatedWorld.Win).To(BeTrue())
		})

		It("checks all ships independently for lose condition", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0) // Collides
			ship2 := entities.NewShip(2, entities.NewVec2(100.0, 100.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0) // No collision
			ship3 := entities.NewShip(3, entities.NewVec2(200.0, 200.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0) // No collision
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), true), // Still active
			}
			world := entities.NewWorld([]entities.Ship{ship1, ship2, ship3}, []entities.Planet{planet}, pallets)

			updatedWorld := EvaluateGameState(world)

			// Ship1 collides, so game should end
			Expect(updatedWorld.Done).To(BeTrue())
			Expect(updatedWorld.Win).To(BeFalse())
		})

		It("handles empty ships array gracefully", func() {
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), true), // Still active
			}
			world := entities.NewWorld([]entities.Ship{}, []entities.Planet{planet}, pallets)

			updatedWorld := EvaluateGameState(world)

			// No ships, no lose condition, game continues
			Expect(updatedWorld.Done).To(BeFalse())
			Expect(updatedWorld.Win).To(BeFalse())
		})

		It("checks all ships against all planets", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(105.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0) // Collides with planet2
			ship2 := entities.NewShip(2, entities.NewVec2(200.0, 200.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0) // No collision
			planet1 := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			planet2 := entities.NewPlanet(2, entities.NewVec2(100.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), true), // Still active
			}
			world := entities.NewWorld([]entities.Ship{ship1, ship2}, []entities.Planet{planet1, planet2}, pallets)

			updatedWorld := EvaluateGameState(world)

			// Ship1 collides with planet2, so game should end
			Expect(updatedWorld.Done).To(BeTrue())
			Expect(updatedWorld.Win).To(BeFalse())
		})

		It("gives win precedence when both win and lose conditions are true", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), // At sun center (lose condition)
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false), // Collected (win condition)
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)

			updatedWorld := EvaluateGameState(world)

			// Win should take precedence
			Expect(updatedWorld.Done).To(BeTrue())
			Expect(updatedWorld.Win).To(BeTrue())
		})

		It("is idempotent when Done is already true", func() {
			ship := entities.NewShip(1, entities.NewVec2(100.0, 100.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false), // Collected
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)
			world.Done = true
			world.Win = true

			updatedWorld := EvaluateGameState(world)

			// State should remain unchanged
			Expect(updatedWorld.Done).To(BeTrue())
			Expect(updatedWorld.Win).To(BeTrue())
		})

		It("preserves other world fields (Ship, Sun, Pallets, Tick)", func() {
			ship := entities.NewShip(1, entities.NewVec2(100.0, 100.0),
				entities.NewVec2(1.0, 2.0),
				1.5,
				75.0,
			)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false),
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)
			world.Tick = 42

			updatedWorld := EvaluateGameState(world)

			// Other fields should be preserved
			// TODO: Temporary fix - will be updated in cu/evaluate-gamestate-multiplayer
			Expect(len(updatedWorld.Ships)).To(BeNumerically(">", 0))
			Expect(updatedWorld.Ships[0].Pos.X).To(Equal(100.0))
			Expect(updatedWorld.Ships[0].Pos.Y).To(Equal(100.0))
			Expect(updatedWorld.Ships[0].Vel.X).To(Equal(1.0))
			Expect(updatedWorld.Ships[0].Vel.Y).To(Equal(2.0))
			Expect(updatedWorld.Ships[0].Rot).To(Equal(1.5))
			Expect(updatedWorld.Ships[0].Energy).To(Equal(float32(75.0)))
			Expect(len(updatedWorld.Planets)).To(BeNumerically(">", 0))
			Expect(updatedWorld.Planets[0].Pos.X).To(Equal(0.0))
			Expect(updatedWorld.Planets[0].Pos.Y).To(Equal(0.0))
			Expect(updatedWorld.Planets[0].Radius).To(Equal(float32(50.0)))
			Expect(len(updatedWorld.Pallets)).To(Equal(1))
			Expect(updatedWorld.Pallets[0].ID).To(Equal(uint32(1)))
			Expect(updatedWorld.Tick).To(Equal(uint32(42)))
		})
	})

	Describe("Game State Transitions Integration", func() {
		It("transitions from ongoing to win state", func() {
			ship := entities.NewShip(1, entities.NewVec2(100.0, 100.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), true),
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)

			// Initially ongoing
			world = EvaluateGameState(world)
			Expect(world.Done).To(BeFalse())

			// Collect all pallets
			pallets[0] = entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false)
			world.Pallets = pallets

			// Should transition to win
			world = EvaluateGameState(world)
			Expect(world.Done).To(BeTrue())
			Expect(world.Win).To(BeTrue())
		})

		It("transitions from ongoing to lose state", func() {
			ship := entities.NewShip(1, entities.NewVec2(100.0, 100.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), true),
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)

			// Initially ongoing
			world = EvaluateGameState(world)
			Expect(world.Done).To(BeFalse())

			// Move ship to planet
			// TODO: Temporary fix - will be updated in cu/evaluate-gamestate-multiplayer
			if len(world.Ships) > 0 {
				world.Ships[0] = entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0) // At planet center
			}

			// Should transition to lose
			world = EvaluateGameState(world)
			Expect(world.Done).To(BeTrue())
			Expect(world.Win).To(BeFalse())
		})

		It("produces consistent results on multiple evaluations", func() {
			ship := entities.NewShip(1, entities.NewVec2(100.0, 100.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false),
			}
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)

			// Evaluate multiple times
			world1 := EvaluateGameState(world)
			world2 := EvaluateGameState(world)
			world3 := EvaluateGameState(world)

			// All should produce same result
			Expect(world1.Done).To(Equal(world2.Done))
			Expect(world2.Done).To(Equal(world3.Done))
			Expect(world1.Win).To(Equal(world2.Win))
			Expect(world2.Win).To(Equal(world3.Win))
		})

		It("handles empty world (no pallets) as not done", func() {
			ship := entities.NewShip(1, entities.NewVec2(100.0, 100.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{} // Empty
			world := entities.NewWorld([]entities.Ship{ship}, []entities.Planet{planet}, pallets)

			updatedWorld := EvaluateGameState(world)

			// No pallets means no win condition can trigger, so game continues
			Expect(updatedWorld.Done).To(BeFalse())
		})
	})
})

