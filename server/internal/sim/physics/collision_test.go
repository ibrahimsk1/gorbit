package physics

import (
	"math"
	"testing"

	"github.com/gorbit/orbitalrush/internal/sim/entities"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCollision(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Collision Suite")
}

var _ = Describe("Collision", Label("scope:unit", "loop:g2-physics", "layer:physics", "dep:none", "b:collision-detection", "r:medium", "double:fake"), func() {
	const epsilon = 1e-9
	const pickupRadius = 1.2 // Pickup radius for pallets

	Describe("ShipPlanetCollision", func() {
		Describe("Determinism", func() {
			It("produces identical results for identical inputs across multiple runs", func() {
				shipPos := entities.NewVec2(10.0, 0.0)
				planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

				// Run collision detection multiple times
				var firstResult bool
				for i := 0; i < 100; i++ {
					result := ShipPlanetCollision(shipPos, planet)
					if i == 0 {
						firstResult = result
					} else {
						// Verify bit-exact results
						Expect(result).To(Equal(firstResult))
					}
				}
			})

			It("produces identical results when called with same inputs in different order", func() {
				shipPos := entities.NewVec2(5.0, 5.0)
				planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 10.0, 1000.0)

				result1 := ShipPlanetCollision(shipPos, planet)
				result2 := ShipPlanetCollision(shipPos, planet)

				Expect(result1).To(Equal(result2))
			})
		})

		Describe("Collision detection", func() {
			It("detects collision when ship is at planet center", func() {
				shipPos := entities.NewVec2(0.0, 0.0)
				planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

				result := ShipPlanetCollision(shipPos, planet)

				Expect(result).To(BeTrue())
			})

			It("detects collision when ship is exactly at planet radius", func() {
				planetRadius := float32(50.0)
				planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), planetRadius, 1000.0)
				// Place ship exactly at planet radius
				shipPos := entities.NewVec2(float64(planetRadius), 0.0)

				result := ShipPlanetCollision(shipPos, planet)

				Expect(result).To(BeTrue())
			})

			It("detects collision when ship is within planet radius", func() {
				planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
				// Place ship inside planet
				shipPos := entities.NewVec2(25.0, 0.0)

				result := ShipPlanetCollision(shipPos, planet)

				Expect(result).To(BeTrue())
			})

			It("does not detect collision when ship is outside planet radius", func() {
				planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
				// Place ship outside planet
				shipPos := entities.NewVec2(100.0, 0.0)

				result := ShipPlanetCollision(shipPos, planet)

				Expect(result).To(BeFalse())
			})

			It("does not detect collision when ship is just outside planet radius", func() {
				planetRadius := float32(50.0)
				planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), planetRadius, 1000.0)
				// Place ship just outside planet radius
				shipPos := entities.NewVec2(float64(planetRadius)+epsilon, 0.0)

				result := ShipPlanetCollision(shipPos, planet)

				Expect(result).To(BeFalse())
			})

			It("detects collision at various positions around planet", func() {
				planetRadius := float32(50.0)
				planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), planetRadius, 1000.0)

				// Test at different angles
				angles := []float64{0.0, math.Pi / 4, math.Pi / 2, math.Pi, 3 * math.Pi / 2}
				for _, angle := range angles {
					// Place ship at planet radius
					shipPos := entities.NewVec2(
						float64(planetRadius)*math.Cos(angle),
						float64(planetRadius)*math.Sin(angle),
					)

					result := ShipPlanetCollision(shipPos, planet)
					Expect(result).To(BeTrue(), "should collide at angle %v", angle)
				}
			})

			It("handles zero radius planet", func() {
				shipPos := entities.NewVec2(0.0, 0.0)
				planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 0.0, 1000.0)

				result := ShipPlanetCollision(shipPos, planet)

				// Ship at same position as zero-radius planet should collide
				Expect(result).To(BeTrue())
			})

			It("handles zero radius planet with ship at different position", func() {
				shipPos := entities.NewVec2(10.0, 10.0)
				planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 0.0, 1000.0)

				result := ShipPlanetCollision(shipPos, planet)

				Expect(result).To(BeFalse())
			})
		})

		Describe("Edge cases", func() {
			It("handles negative coordinates", func() {
				shipPos := entities.NewVec2(-10.0, -10.0)
				planet := entities.NewPlanet(1, entities.NewVec2(-5.0, -5.0), 10.0, 1000.0)

				result := ShipPlanetCollision(shipPos, planet)

				// Distance is sqrt((10-5)^2 + (10-5)^2) = sqrt(50) ≈ 7.07 < 10, so should collide
				Expect(result).To(BeTrue())
			})

			It("handles very large distances", func() {
				shipPos := entities.NewVec2(1000000.0, 0.0)
				planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

				result := ShipPlanetCollision(shipPos, planet)

				Expect(result).To(BeFalse())
			})

			It("handles very small planet radius", func() {
				shipPos := entities.NewVec2(0.001, 0.0)
				planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 0.0005, 1000.0)

				result := ShipPlanetCollision(shipPos, planet)

				Expect(result).To(BeFalse())
			})
		})
	})

	var _ = Describe("CheckShipPlanetCollisions", Label("scope:unit", "loop:g2-physics", "layer:physics", "dep:none", "b:collision-multiplanet", "r:medium", "double:fake"), func() {
		const epsilon = 1e-9

		Describe("Single planet collision", func() {
			It("returns true and planet ID when ship collides with planet", func() {
				shipPos := entities.NewVec2(10.0, 0.0)
				planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
				planets := []entities.Planet{planet}

				colliding, planetID := CheckShipPlanetCollisions(shipPos, planets)

				Expect(colliding).To(BeTrue())
				Expect(planetID).To(Equal(uint32(1)))
			})

			It("returns false and 0 when ship does not collide with planet", func() {
				shipPos := entities.NewVec2(100.0, 0.0)
				planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
				planets := []entities.Planet{planet}

				colliding, planetID := CheckShipPlanetCollisions(shipPos, planets)

				Expect(colliding).To(BeFalse())
				Expect(planetID).To(Equal(uint32(0)))
			})
		})

		Describe("Multiple planets", func() {
			It("returns correct planet ID when ship collides with one planet", func() {
				shipPos := entities.NewVec2(10.0, 0.0)
				planet1 := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
				planet2 := entities.NewPlanet(2, entities.NewVec2(200.0, 0.0), 50.0, 1000.0)
				planet3 := entities.NewPlanet(3, entities.NewVec2(0.0, 200.0), 50.0, 1000.0)
				planets := []entities.Planet{planet1, planet2, planet3}

				colliding, planetID := CheckShipPlanetCollisions(shipPos, planets)

				Expect(colliding).To(BeTrue())
				Expect(planetID).To(Equal(uint32(1))) // First planet in array
			})

			It("returns first colliding planet ID when ship collides with multiple planets", func() {
				// Ship at origin, collides with planet1 and planet2 (both at origin)
				shipPos := entities.NewVec2(0.0, 0.0)
				planet1 := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
				planet2 := entities.NewPlanet(2, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
				planet3 := entities.NewPlanet(3, entities.NewVec2(200.0, 200.0), 50.0, 1000.0)
				planets := []entities.Planet{planet1, planet2, planet3}

				colliding, planetID := CheckShipPlanetCollisions(shipPos, planets)

				Expect(colliding).To(BeTrue())
				Expect(planetID).To(Equal(uint32(1))) // First colliding planet in array
			})

			It("returns false and 0 when ship doesn't collide with any planet", func() {
				shipPos := entities.NewVec2(1000.0, 1000.0)
				planet1 := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
				planet2 := entities.NewPlanet(2, entities.NewVec2(200.0, 0.0), 50.0, 1000.0)
				planet3 := entities.NewPlanet(3, entities.NewVec2(0.0, 200.0), 50.0, 1000.0)
				planets := []entities.Planet{planet1, planet2, planet3}

				colliding, planetID := CheckShipPlanetCollisions(shipPos, planets)

				Expect(colliding).To(BeFalse())
				Expect(planetID).To(Equal(uint32(0)))
			})
		})

		Describe("Edge cases", func() {
			It("returns false and 0 for empty planets array", func() {
				shipPos := entities.NewVec2(10.0, 0.0)
				planets := []entities.Planet{}

				colliding, planetID := CheckShipPlanetCollisions(shipPos, planets)

				Expect(colliding).To(BeFalse())
				Expect(planetID).To(Equal(uint32(0)))
			})

			It("handles ship at planet center", func() {
				shipPos := entities.NewVec2(0.0, 0.0)
				planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
				planets := []entities.Planet{planet}

				colliding, planetID := CheckShipPlanetCollisions(shipPos, planets)

				Expect(colliding).To(BeTrue())
				Expect(planetID).To(Equal(uint32(1)))
			})

			It("handles ship exactly at planet radius", func() {
				planetRadius := float32(50.0)
				planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), planetRadius, 1000.0)
				shipPos := entities.NewVec2(float64(planetRadius), 0.0)
				planets := []entities.Planet{planet}

				colliding, planetID := CheckShipPlanetCollisions(shipPos, planets)

				Expect(colliding).To(BeTrue())
				Expect(planetID).To(Equal(uint32(1)))
			})

			It("handles ship between multiple planets", func() {
				// Ship positioned between two planets, closer to planet1
				shipPos := entities.NewVec2(15.0, 0.0)
				planet1 := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
				planet2 := entities.NewPlanet(2, entities.NewVec2(100.0, 0.0), 50.0, 1000.0)
				planets := []entities.Planet{planet1, planet2}

				colliding, planetID := CheckShipPlanetCollisions(shipPos, planets)

				Expect(colliding).To(BeTrue())
				Expect(planetID).To(Equal(uint32(1))) // First planet in array
			})

			It("handles ship between multiple planets, colliding with second planet", func() {
				// Ship positioned between two planets, colliding with planet2
				shipPos := entities.NewVec2(85.0, 0.0)
				planet1 := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
				planet2 := entities.NewPlanet(2, entities.NewVec2(100.0, 0.0), 50.0, 1000.0)
				planets := []entities.Planet{planet1, planet2}

				colliding, planetID := CheckShipPlanetCollisions(shipPos, planets)

				Expect(colliding).To(BeTrue())
				Expect(planetID).To(Equal(uint32(2))) // Second planet in array
			})
		})

		Describe("Determinism", func() {
			It("produces identical results for identical inputs across multiple runs", func() {
				shipPos := entities.NewVec2(10.0, 5.0)
				planet1 := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
				planet2 := entities.NewPlanet(2, entities.NewVec2(200.0, 0.0), 50.0, 1000.0)
				planets := []entities.Planet{planet1, planet2}

				// Run collision detection multiple times
				var firstColliding bool
				var firstPlanetID uint32
				for i := 0; i < 100; i++ {
					colliding, planetID := CheckShipPlanetCollisions(shipPos, planets)
					if i == 0 {
						firstColliding = colliding
						firstPlanetID = planetID
					} else {
						// Verify bit-exact results
						Expect(colliding).To(Equal(firstColliding))
						Expect(planetID).To(Equal(firstPlanetID))
					}
				}
			})

			It("produces identical results when called with same inputs in different order", func() {
				shipPos := entities.NewVec2(5.0, 5.0)
				planet1 := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
				planet2 := entities.NewPlanet(2, entities.NewVec2(200.0, 0.0), 50.0, 1000.0)
				planets := []entities.Planet{planet1, planet2}

				colliding1, planetID1 := CheckShipPlanetCollisions(shipPos, planets)
				colliding2, planetID2 := CheckShipPlanetCollisions(shipPos, planets)

				Expect(colliding1).To(Equal(colliding2))
				Expect(planetID1).To(Equal(planetID2))
			})
		})
	})

	Describe("ShipPalletCollision", func() {
		Describe("Determinism", func() {
			It("produces identical results for identical inputs across multiple runs", func() {
				shipPos := entities.NewVec2(10.0, 0.0)
				palletPos := entities.NewVec2(10.0, 0.0)

				// Run collision detection multiple times
				var firstResult bool
				for i := 0; i < 100; i++ {
					result := ShipPalletCollision(shipPos, palletPos, pickupRadius)
					if i == 0 {
						firstResult = result
					} else {
						// Verify bit-exact results
						Expect(result).To(Equal(firstResult))
					}
				}
			})

			It("produces identical results when called with same inputs in different order", func() {
				shipPos := entities.NewVec2(5.0, 5.0)
				palletPos := entities.NewVec2(5.0, 5.0)

				result1 := ShipPalletCollision(shipPos, palletPos, pickupRadius)
				result2 := ShipPalletCollision(shipPos, palletPos, pickupRadius)

				Expect(result1).To(Equal(result2))
			})
		})

		Describe("Collision detection", func() {
			It("detects collision when ship is at pallet center", func() {
				shipPos := entities.NewVec2(10.0, 10.0)
				palletPos := entities.NewVec2(10.0, 10.0)

				result := ShipPalletCollision(shipPos, palletPos, pickupRadius)

				Expect(result).To(BeTrue())
			})

			It("detects collision when ship is exactly at pickup radius", func() {
				palletPos := entities.NewVec2(0.0, 0.0)
				// Place ship exactly at pickup radius
				shipPos := entities.NewVec2(pickupRadius, 0.0)

				result := ShipPalletCollision(shipPos, palletPos, pickupRadius)

				Expect(result).To(BeTrue())
			})

			It("detects collision when ship is within pickup radius", func() {
				palletPos := entities.NewVec2(0.0, 0.0)
				// Place ship inside pickup radius
				shipPos := entities.NewVec2(0.6, 0.0)

				result := ShipPalletCollision(shipPos, palletPos, pickupRadius)

				Expect(result).To(BeTrue())
			})

			It("does not detect collision when ship is outside pickup radius", func() {
				palletPos := entities.NewVec2(0.0, 0.0)
				// Place ship outside pickup radius
				shipPos := entities.NewVec2(10.0, 0.0)

				result := ShipPalletCollision(shipPos, palletPos, pickupRadius)

				Expect(result).To(BeFalse())
			})

			It("does not detect collision when ship is just outside pickup radius", func() {
				palletPos := entities.NewVec2(0.0, 0.0)
				// Place ship just outside pickup radius
				shipPos := entities.NewVec2(pickupRadius+epsilon, 0.0)

				result := ShipPalletCollision(shipPos, palletPos, pickupRadius)

				Expect(result).To(BeFalse())
			})

			It("detects collision at various positions around pallet", func() {
				palletPos := entities.NewVec2(0.0, 0.0)

				// Test at different angles
				angles := []float64{0.0, math.Pi / 4, math.Pi / 2, math.Pi, 3 * math.Pi / 2}
				for _, angle := range angles {
					// Place ship at pickup radius
					shipPos := entities.NewVec2(
						pickupRadius*math.Cos(angle),
						pickupRadius*math.Sin(angle),
					)

					result := ShipPalletCollision(shipPos, palletPos, pickupRadius)
					Expect(result).To(BeTrue(), "should collide at angle %v", angle)
				}
			})

			It("handles zero pickup radius", func() {
				shipPos := entities.NewVec2(0.0, 0.0)
				palletPos := entities.NewVec2(0.0, 0.0)

				result := ShipPalletCollision(shipPos, palletPos, 0.0)

				// Ship at same position as pallet with zero radius should collide
				Expect(result).To(BeTrue())
			})

			It("handles zero pickup radius with ship at different position", func() {
				shipPos := entities.NewVec2(10.0, 10.0)
				palletPos := entities.NewVec2(0.0, 0.0)

				result := ShipPalletCollision(shipPos, palletPos, 0.0)

				Expect(result).To(BeFalse())
			})
		})

		Describe("Edge cases", func() {
			It("handles negative coordinates", func() {
				shipPos := entities.NewVec2(-1.0, -1.0)
				palletPos := entities.NewVec2(-0.5, -0.5)

				result := ShipPalletCollision(shipPos, palletPos, pickupRadius)

				// Distance is sqrt((1-0.5)^2 + (1-0.5)^2) = sqrt(0.5) ≈ 0.707 < 1.2, so should collide
				Expect(result).To(BeTrue())
			})

			It("handles very large distances", func() {
				shipPos := entities.NewVec2(1000000.0, 0.0)
				palletPos := entities.NewVec2(0.0, 0.0)

				result := ShipPalletCollision(shipPos, palletPos, pickupRadius)

				Expect(result).To(BeFalse())
			})

			It("handles very small pickup radius", func() {
				shipPos := entities.NewVec2(0.001, 0.0)
				palletPos := entities.NewVec2(0.0, 0.0)

				result := ShipPalletCollision(shipPos, palletPos, 0.0005)

				Expect(result).To(BeFalse())
			})

			It("handles multiple pallets independently", func() {
				shipPos := entities.NewVec2(5.0, 5.0)
				pallet1Pos := entities.NewVec2(5.0, 5.0) // Ship at pallet 1
				pallet2Pos := entities.NewVec2(100.0, 100.0) // Ship far from pallet 2

				result1 := ShipPalletCollision(shipPos, pallet1Pos, pickupRadius)
				result2 := ShipPalletCollision(shipPos, pallet2Pos, pickupRadius)

				Expect(result1).To(BeTrue())
				Expect(result2).To(BeFalse())
			})
		})
	})
})

