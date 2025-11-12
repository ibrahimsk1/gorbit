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

var _ = Describe("Collision", Label("scope:unit", "loop:g1-physics", "layer:sim", "dep:none", "b:collision-detection", "r:medium", "double:fake"), func() {
	const epsilon = 1e-9
	const pickupRadius = 1.2 // Pickup radius for pallets

	Describe("ShipSunCollision", func() {
		Describe("Determinism", func() {
			It("produces identical results for identical inputs across multiple runs", func() {
				shipPos := entities.NewVec2(10.0, 0.0)
				sunPos := entities.NewVec2(0.0, 0.0)
				sunRadius := float32(50.0)

				// Run collision detection multiple times
				var firstResult bool
				for i := 0; i < 100; i++ {
					result := ShipSunCollision(shipPos, sunPos, sunRadius)
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
				sunPos := entities.NewVec2(0.0, 0.0)
				sunRadius := float32(10.0)

				result1 := ShipSunCollision(shipPos, sunPos, sunRadius)
				result2 := ShipSunCollision(shipPos, sunPos, sunRadius)

				Expect(result1).To(Equal(result2))
			})
		})

		Describe("Collision detection", func() {
			It("detects collision when ship is at sun center", func() {
				shipPos := entities.NewVec2(0.0, 0.0)
				sunPos := entities.NewVec2(0.0, 0.0)
				sunRadius := float32(50.0)

				result := ShipSunCollision(shipPos, sunPos, sunRadius)

				Expect(result).To(BeTrue())
			})

			It("detects collision when ship is exactly at sun radius", func() {
				sunPos := entities.NewVec2(0.0, 0.0)
				sunRadius := float32(50.0)
				// Place ship exactly at sun radius
				shipPos := entities.NewVec2(float64(sunRadius), 0.0)

				result := ShipSunCollision(shipPos, sunPos, sunRadius)

				Expect(result).To(BeTrue())
			})

			It("detects collision when ship is within sun radius", func() {
				sunPos := entities.NewVec2(0.0, 0.0)
				sunRadius := float32(50.0)
				// Place ship inside sun
				shipPos := entities.NewVec2(25.0, 0.0)

				result := ShipSunCollision(shipPos, sunPos, sunRadius)

				Expect(result).To(BeTrue())
			})

			It("does not detect collision when ship is outside sun radius", func() {
				sunPos := entities.NewVec2(0.0, 0.0)
				sunRadius := float32(50.0)
				// Place ship outside sun
				shipPos := entities.NewVec2(100.0, 0.0)

				result := ShipSunCollision(shipPos, sunPos, sunRadius)

				Expect(result).To(BeFalse())
			})

			It("does not detect collision when ship is just outside sun radius", func() {
				sunPos := entities.NewVec2(0.0, 0.0)
				sunRadius := float32(50.0)
				// Place ship just outside sun radius
				shipPos := entities.NewVec2(float64(sunRadius)+epsilon, 0.0)

				result := ShipSunCollision(shipPos, sunPos, sunRadius)

				Expect(result).To(BeFalse())
			})

			It("detects collision at various positions around sun", func() {
				sunPos := entities.NewVec2(0.0, 0.0)
				sunRadius := float32(50.0)

				// Test at different angles
				angles := []float64{0.0, math.Pi / 4, math.Pi / 2, math.Pi, 3 * math.Pi / 2}
				for _, angle := range angles {
					// Place ship at sun radius
					shipPos := entities.NewVec2(
						float64(sunRadius)*math.Cos(angle),
						float64(sunRadius)*math.Sin(angle),
					)

					result := ShipSunCollision(shipPos, sunPos, sunRadius)
					Expect(result).To(BeTrue(), "should collide at angle %v", angle)
				}
			})

			It("handles zero radius sun", func() {
				shipPos := entities.NewVec2(0.0, 0.0)
				sunPos := entities.NewVec2(0.0, 0.0)
				sunRadius := float32(0.0)

				result := ShipSunCollision(shipPos, sunPos, sunRadius)

				// Ship at same position as zero-radius sun should collide
				Expect(result).To(BeTrue())
			})

			It("handles zero radius sun with ship at different position", func() {
				shipPos := entities.NewVec2(10.0, 10.0)
				sunPos := entities.NewVec2(0.0, 0.0)
				sunRadius := float32(0.0)

				result := ShipSunCollision(shipPos, sunPos, sunRadius)

				Expect(result).To(BeFalse())
			})
		})

		Describe("Edge cases", func() {
			It("handles negative coordinates", func() {
				shipPos := entities.NewVec2(-10.0, -10.0)
				sunPos := entities.NewVec2(-5.0, -5.0)
				sunRadius := float32(10.0)

				result := ShipSunCollision(shipPos, sunPos, sunRadius)

				// Distance is sqrt((10-5)^2 + (10-5)^2) = sqrt(50) ≈ 7.07 < 10, so should collide
				Expect(result).To(BeTrue())
			})

			It("handles very large distances", func() {
				shipPos := entities.NewVec2(1000000.0, 0.0)
				sunPos := entities.NewVec2(0.0, 0.0)
				sunRadius := float32(50.0)

				result := ShipSunCollision(shipPos, sunPos, sunRadius)

				Expect(result).To(BeFalse())
			})

			It("handles very small sun radius", func() {
				shipPos := entities.NewVec2(0.001, 0.0)
				sunPos := entities.NewVec2(0.0, 0.0)
				sunRadius := float32(0.0005)

				result := ShipSunCollision(shipPos, sunPos, sunRadius)

				Expect(result).To(BeFalse())
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

