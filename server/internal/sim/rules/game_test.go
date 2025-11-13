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

var _ = Describe("Game State Transitions", Label("scope:unit", "loop:g2-rules", "layer:sim", "dep:none", "b:game-state-transitions", "r:high", "double:fake"), func() {
	Describe("CheckWinCondition", func() {
		It("returns true when all pallets are collected", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false), // Collected
				entities.NewPallet(2, entities.NewVec2(20.0, 20.0), false), // Collected
			}
			world := entities.NewWorld(ship, sun, pallets)

			result := CheckWinCondition(world)
			Expect(result).To(BeTrue())
		})

		It("returns false when there are no pallets", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{} // Empty
			world := entities.NewWorld(ship, sun, pallets)

			result := CheckWinCondition(world)
			Expect(result).To(BeFalse())
		})

		It("returns false when at least one pallet is active", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false), // Collected
				entities.NewPallet(2, entities.NewVec2(20.0, 20.0), true),  // Still active
			}
			world := entities.NewWorld(ship, sun, pallets)

			result := CheckWinCondition(world)
			Expect(result).To(BeFalse())
		})

		It("returns false when all pallets are active", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), true),
				entities.NewPallet(2, entities.NewVec2(20.0, 20.0), true),
			}
			world := entities.NewWorld(ship, sun, pallets)

			result := CheckWinCondition(world)
			Expect(result).To(BeFalse())
		})

		It("returns true when single pallet is collected", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false), // Collected
			}
			world := entities.NewWorld(ship, sun, pallets)

			result := CheckWinCondition(world)
			Expect(result).To(BeTrue())
		})
	})

	Describe("CheckLoseCondition", func() {
		It("returns true when ship is at sun center", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0), // At sun center
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)

			result := CheckLoseCondition(world)
			Expect(result).To(BeTrue())
		})

		It("returns true when ship is exactly at sun radius", func() {
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			ship := entities.NewShip(
				entities.NewVec2(float64(sun.Radius), 0.0), // Exactly at radius
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			world := entities.NewWorld(ship, sun, nil)

			result := CheckLoseCondition(world)
			Expect(result).To(BeTrue())
		})

		It("returns true when ship is within sun radius", func() {
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			ship := entities.NewShip(
				entities.NewVec2(25.0, 0.0), // Inside sun
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			world := entities.NewWorld(ship, sun, nil)

			result := CheckLoseCondition(world)
			Expect(result).To(BeTrue())
		})

		It("returns false when ship is outside sun radius", func() {
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			ship := entities.NewShip(
				entities.NewVec2(100.0, 0.0), // Outside sun
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			world := entities.NewWorld(ship, sun, nil)

			result := CheckLoseCondition(world)
			Expect(result).To(BeFalse())
		})

		It("returns false when ship is just outside sun radius", func() {
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			ship := entities.NewShip(
				entities.NewVec2(float64(sun.Radius)+0.1, 0.0), // Just outside
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			world := entities.NewWorld(ship, sun, nil)

			result := CheckLoseCondition(world)
			Expect(result).To(BeFalse())
		})

		It("returns false when ship is far from sun", func() {
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			ship := entities.NewShip(
				entities.NewVec2(1000.0, 1000.0), // Far away
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			world := entities.NewWorld(ship, sun, nil)

			result := CheckLoseCondition(world)
			Expect(result).To(BeFalse())
		})
	})

	Describe("EvaluateGameState", func() {
		It("sets Done=true and Win=true when all pallets collected", func() {
			ship := entities.NewShip(
				entities.NewVec2(100.0, 100.0), // Far from sun
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false), // Collected
				entities.NewPallet(2, entities.NewVec2(20.0, 20.0), false), // Collected
			}
			world := entities.NewWorld(ship, sun, pallets)

			updatedWorld := EvaluateGameState(world)

			Expect(updatedWorld.Done).To(BeTrue())
			Expect(updatedWorld.Win).To(BeTrue())
		})

		It("sets Done=true and Win=false when ship collides with sun", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0), // At sun center
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), true), // Still active
			}
			world := entities.NewWorld(ship, sun, pallets)

			updatedWorld := EvaluateGameState(world)

			Expect(updatedWorld.Done).To(BeTrue())
			Expect(updatedWorld.Win).To(BeFalse())
		})

		It("leaves Done and Win unchanged when neither condition is met", func() {
			ship := entities.NewShip(
				entities.NewVec2(100.0, 100.0), // Far from sun
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), true), // Still active
			}
			world := entities.NewWorld(ship, sun, pallets)

			updatedWorld := EvaluateGameState(world)

			Expect(updatedWorld.Done).To(BeFalse())
			Expect(updatedWorld.Win).To(BeFalse())
		})

		It("gives win precedence when both win and lose conditions are true", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0), // At sun center (lose condition)
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false), // Collected (win condition)
			}
			world := entities.NewWorld(ship, sun, pallets)

			updatedWorld := EvaluateGameState(world)

			// Win should take precedence
			Expect(updatedWorld.Done).To(BeTrue())
			Expect(updatedWorld.Win).To(BeTrue())
		})

		It("is idempotent when Done is already true", func() {
			ship := entities.NewShip(
				entities.NewVec2(100.0, 100.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false), // Collected
			}
			world := entities.NewWorld(ship, sun, pallets)
			world.Done = true
			world.Win = true

			updatedWorld := EvaluateGameState(world)

			// State should remain unchanged
			Expect(updatedWorld.Done).To(BeTrue())
			Expect(updatedWorld.Win).To(BeTrue())
		})

		It("preserves other world fields (Ship, Sun, Pallets, Tick)", func() {
			ship := entities.NewShip(
				entities.NewVec2(100.0, 100.0),
				entities.NewVec2(1.0, 2.0),
				1.5,
				75.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false),
			}
			world := entities.NewWorld(ship, sun, pallets)
			world.Tick = 42

			updatedWorld := EvaluateGameState(world)

			// Other fields should be preserved
			Expect(updatedWorld.Ship.Pos.X).To(Equal(100.0))
			Expect(updatedWorld.Ship.Pos.Y).To(Equal(100.0))
			Expect(updatedWorld.Ship.Vel.X).To(Equal(1.0))
			Expect(updatedWorld.Ship.Vel.Y).To(Equal(2.0))
			Expect(updatedWorld.Ship.Rot).To(Equal(1.5))
			Expect(updatedWorld.Ship.Energy).To(Equal(float32(75.0)))
			Expect(updatedWorld.Sun.Pos.X).To(Equal(0.0))
			Expect(updatedWorld.Sun.Pos.Y).To(Equal(0.0))
			Expect(updatedWorld.Sun.Radius).To(Equal(float32(50.0)))
			Expect(len(updatedWorld.Pallets)).To(Equal(1))
			Expect(updatedWorld.Pallets[0].ID).To(Equal(uint32(1)))
			Expect(updatedWorld.Tick).To(Equal(uint32(42)))
		})
	})

	Describe("Game State Transitions Integration", func() {
		It("transitions from ongoing to win state", func() {
			ship := entities.NewShip(
				entities.NewVec2(100.0, 100.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), true),
			}
			world := entities.NewWorld(ship, sun, pallets)

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
			ship := entities.NewShip(
				entities.NewVec2(100.0, 100.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), true),
			}
			world := entities.NewWorld(ship, sun, pallets)

			// Initially ongoing
			world = EvaluateGameState(world)
			Expect(world.Done).To(BeFalse())

			// Move ship to sun
			world.Ship = entities.NewShip(
				entities.NewVec2(0.0, 0.0), // At sun center
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)

			// Should transition to lose
			world = EvaluateGameState(world)
			Expect(world.Done).To(BeTrue())
			Expect(world.Win).To(BeFalse())
		})

		It("produces consistent results on multiple evaluations", func() {
			ship := entities.NewShip(
				entities.NewVec2(100.0, 100.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), false),
			}
			world := entities.NewWorld(ship, sun, pallets)

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
			ship := entities.NewShip(
				entities.NewVec2(100.0, 100.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallets := []entities.Pallet{} // Empty
			world := entities.NewWorld(ship, sun, pallets)

			updatedWorld := EvaluateGameState(world)

			// No pallets means no win condition can trigger, so game continues
			Expect(updatedWorld.Done).To(BeFalse())
		})
	})
})

