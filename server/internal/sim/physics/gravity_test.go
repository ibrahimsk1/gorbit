package physics

import (
	"math"
	"testing"

	"github.com/gorbit/orbitalrush/internal/sim/entities"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGravity(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gravity Suite")
}

var _ = Describe("Gravity", Label("scope:unit", "loop:g1-physics", "layer:sim", "dep:none", "b:gravity-field", "r:high", "double:fake"), func() {
	const epsilon = 1e-9
	const G = 1.0 // Gravitational constant (game-scale)
	const aMax = 100.0 // Maximum acceleration

	Describe("GravityAcceleration", func() {
		Describe("Determinism", func() {
			It("produces identical results for identical inputs across multiple runs", func() {
				shipPos := entities.NewVec2(10.0, 0.0)
				sunPos := entities.NewVec2(0.0, 0.0)
				sunMass := 1000.0

				// Run calculation multiple times
				var firstAcc entities.Vec2
				for i := 0; i < 100; i++ {
					acc := GravityAcceleration(shipPos, sunPos, sunMass, G, aMax)
					if i == 0 {
						firstAcc = acc
					} else {
						// Verify bit-exact results
						Expect(acc.X).To(Equal(firstAcc.X))
						Expect(acc.Y).To(Equal(firstAcc.Y))
					}
				}
			})

			It("produces identical results when called with same inputs in different order", func() {
				shipPos := entities.NewVec2(5.0, 5.0)
				sunPos := entities.NewVec2(0.0, 0.0)
				sunMass := 500.0

				result1 := GravityAcceleration(shipPos, sunPos, sunMass, G, aMax)
				result2 := GravityAcceleration(shipPos, sunPos, sunMass, G, aMax)

				Expect(result1.X).To(Equal(result2.X))
				Expect(result1.Y).To(Equal(result2.Y))
			})
		})

		Describe("Inverse-square law", func() {
			It("follows inverse-square law at various distances", func() {
				sunPos := entities.NewVec2(0.0, 0.0)
				sunMass := 1000.0

				// Test at distance r
				r := 10.0
				shipPos1 := entities.NewVec2(r, 0.0)
				acc1 := GravityAcceleration(shipPos1, sunPos, sunMass, G, aMax)
				mag1 := acc1.Length()

				// Test at distance 2r (should be 1/4 of original)
				shipPos2 := entities.NewVec2(2*r, 0.0)
				acc2 := GravityAcceleration(shipPos2, sunPos, sunMass, G, aMax)
				mag2 := acc2.Length()

				// Test at distance 3r (should be 1/9 of original)
				shipPos3 := entities.NewVec2(3*r, 0.0)
				acc3 := GravityAcceleration(shipPos3, sunPos, sunMass, G, aMax)
				mag3 := acc3.Length()

				// Verify inverse-square relationship (within numerical precision)
				Expect(mag2).To(BeNumerically("~", mag1/4.0, epsilon*10))
				Expect(mag3).To(BeNumerically("~", mag1/9.0, epsilon*10))
			})

			It("scales linearly with sun mass", func() {
				shipPos := entities.NewVec2(10.0, 0.0)
				sunPos := entities.NewVec2(0.0, 0.0)

				// Test with mass M
				mass1 := 1000.0
				acc1 := GravityAcceleration(shipPos, sunPos, mass1, G, aMax)
				mag1 := acc1.Length()

				// Test with mass 2M (should double acceleration)
				mass2 := 2000.0
				acc2 := GravityAcceleration(shipPos, sunPos, mass2, G, aMax)
				mag2 := acc2.Length()

				// Verify linear scaling
				Expect(mag2).To(BeNumerically("~", 2.0*mag1, epsilon*10))
			})

			It("points acceleration toward the sun", func() {
				shipPos := entities.NewVec2(10.0, 5.0)
				sunPos := entities.NewVec2(0.0, 0.0)
				sunMass := 1000.0

				acc := GravityAcceleration(shipPos, sunPos, sunMass, G, aMax)

				// Direction from ship to sun
				direction := sunPos.Sub(shipPos).Normalize()

				// Acceleration should point in the same direction (toward sun)
				accNormalized := acc.Normalize()
				Expect(accNormalized.Dot(direction)).To(BeNumerically("~", 1.0, epsilon))
			})

			It("has correct magnitude for inverse-square law at medium distances", func() {
				shipPos := entities.NewVec2(10.0, 0.0)
				sunPos := entities.NewVec2(0.0, 0.0)
				sunMass := 1000.0

				acc := GravityAcceleration(shipPos, sunPos, sunMass, G, aMax)
				mag := acc.Length()

				// Expected: |a| = G * M / r²
				// r = 10.0, so |a| = 1.0 * 1000.0 / 100.0 = 10.0
				expectedMag := G * sunMass / (10.0 * 10.0)

				// Should not be clamped at this distance
				Expect(mag).To(BeNumerically("<", aMax))
				Expect(mag).To(BeNumerically("~", expectedMag, epsilon*10))
			})
		})

		Describe("Acceleration clamping", func() {
			It("clamps acceleration to a_max at close distances", func() {
				sunPos := entities.NewVec2(0.0, 0.0)
				sunMass := 10000.0 // Large mass to ensure clamping

				// Place ship very close to sun
				shipPos := entities.NewVec2(0.1, 0.0)
				acc := GravityAcceleration(shipPos, sunPos, sunMass, G, aMax)
				mag := acc.Length()

				// Acceleration should be clamped
				Expect(mag).To(BeNumerically("<=", aMax+epsilon))
			})

			It("does not clamp acceleration at far distances", func() {
				shipPos := entities.NewVec2(100.0, 0.0)
				sunPos := entities.NewVec2(0.0, 0.0)
				sunMass := 1000.0

				acc := GravityAcceleration(shipPos, sunPos, sunMass, G, aMax)
				mag := acc.Length()

				// Expected: |a| = G * M / r² = 1.0 * 1000.0 / 10000.0 = 0.1
				expectedMag := G * sunMass / (100.0 * 100.0)

				// Should not be clamped
				Expect(mag).To(BeNumerically("<", aMax))
				Expect(mag).To(BeNumerically("~", expectedMag, epsilon*10))
			})

			It("preserves direction when clamping", func() {
				sunPos := entities.NewVec2(0.0, 0.0)
				sunMass := 10000.0 // Large mass to ensure clamping

				// Place ship at an angle
				shipPos := entities.NewVec2(0.1, 0.1)
				acc := GravityAcceleration(shipPos, sunPos, sunMass, G, aMax)

				// Direction from ship to sun
				direction := sunPos.Sub(shipPos).Normalize()

				// Clamped acceleration should still point toward sun
				accNormalized := acc.Normalize()
				Expect(accNormalized.Dot(direction)).To(BeNumerically("~", 1.0, epsilon))
			})

			It("handles boundary condition where |a| = a_max", func() {
				sunPos := entities.NewVec2(0.0, 0.0)
				sunMass := 1000.0

				// Find distance where |a| = a_max
				// a_max = G * M / r²
				// r = sqrt(G * M / a_max)
				r := math.Sqrt(G * sunMass / aMax)
				shipPos := entities.NewVec2(r, 0.0)

				acc := GravityAcceleration(shipPos, sunPos, sunMass, G, aMax)
				mag := acc.Length()

				// Should be approximately a_max (may be slightly less due to clamping logic)
				Expect(mag).To(BeNumerically("<=", aMax+epsilon))
			})
		})

		Describe("Edge cases", func() {
			It("handles zero distance gracefully", func() {
				shipPos := entities.NewVec2(0.0, 0.0)
				sunPos := entities.NewVec2(0.0, 0.0)
				sunMass := 1000.0

				// Should not panic
				acc := GravityAcceleration(shipPos, sunPos, sunMass, G, aMax)

				// Should return zero or clamped value
				// At zero distance, direction is undefined, so should return zero
				Expect(acc.X).To(BeNumerically("~", 0.0, epsilon))
				Expect(acc.Y).To(BeNumerically("~", 0.0, epsilon))
			})

			It("handles very small distance", func() {
				sunPos := entities.NewVec2(0.0, 0.0)
				sunMass := 10000.0

				// Very small distance
				shipPos := entities.NewVec2(1e-6, 0.0)
				acc := GravityAcceleration(shipPos, sunPos, sunMass, G, aMax)
				mag := acc.Length()

				// Should be clamped
				Expect(mag).To(BeNumerically("<=", aMax+epsilon))
			})

			It("handles very large distance", func() {
				shipPos := entities.NewVec2(1e6, 0.0)
				sunPos := entities.NewVec2(0.0, 0.0)
				sunMass := 1000.0

				acc := GravityAcceleration(shipPos, sunPos, sunMass, G, aMax)
				mag := acc.Length()

				// Acceleration should approach zero
				Expect(mag).To(BeNumerically("<", 1e-6))
			})

			It("handles zero mass", func() {
				shipPos := entities.NewVec2(10.0, 0.0)
				sunPos := entities.NewVec2(0.0, 0.0)
				sunMass := 0.0

				acc := GravityAcceleration(shipPos, sunPos, sunMass, G, aMax)

				// Should return zero acceleration
				Expect(acc.X).To(BeNumerically("~", 0.0, epsilon))
				Expect(acc.Y).To(BeNumerically("~", 0.0, epsilon))
			})

			It("handles identical positions", func() {
				pos := entities.NewVec2(5.0, 5.0)
				sunMass := 1000.0

				acc := GravityAcceleration(pos, pos, sunMass, G, aMax)

				// Should return zero acceleration
				Expect(acc.X).To(BeNumerically("~", 0.0, epsilon))
				Expect(acc.Y).To(BeNumerically("~", 0.0, epsilon))
			})
		})

		Describe("Field behavior at various distances", func() {
			It("behaves correctly at close range", func() {
				sunPos := entities.NewVec2(0.0, 0.0)
				sunMass := 5000.0

				// Test multiple close distances
				distances := []float64{0.5, 1.0, 2.0, 5.0}
				for _, d := range distances {
					shipPos := entities.NewVec2(d, 0.0)
					acc := GravityAcceleration(shipPos, sunPos, sunMass, G, aMax)
					mag := acc.Length()

					// All should be clamped or close to clamping
					Expect(mag).To(BeNumerically("<=", aMax+epsilon))
				}
			})

			It("behaves correctly at medium range", func() {
				shipPos := entities.NewVec2(20.0, 0.0)
				sunPos := entities.NewVec2(0.0, 0.0)
				sunMass := 1000.0

				acc := GravityAcceleration(shipPos, sunPos, sunMass, G, aMax)
				mag := acc.Length()

				// Expected: |a| = G * M / r² = 1.0 * 1000.0 / 400.0 = 2.5
				expectedMag := G * sunMass / (20.0 * 20.0)

				// Should not be clamped
				Expect(mag).To(BeNumerically("<", aMax))
				Expect(mag).To(BeNumerically("~", expectedMag, epsilon*10))
			})

			It("behaves correctly at far range", func() {
				shipPos := entities.NewVec2(1000.0, 0.0)
				sunPos := entities.NewVec2(0.0, 0.0)
				sunMass := 1000.0

				acc := GravityAcceleration(shipPos, sunPos, sunMass, G, aMax)
				mag := acc.Length()

				// Expected: |a| = G * M / r² = 1.0 * 1000.0 / 1000000.0 = 0.001
				expectedMag := G * sunMass / (1000.0 * 1000.0)

				// Should be very small
				Expect(mag).To(BeNumerically("<", aMax))
				Expect(mag).To(BeNumerically("~", expectedMag, epsilon*10))
			})

			It("transitions smoothly between clamped and unclamped regions", func() {
				sunPos := entities.NewVec2(0.0, 0.0)
				sunMass := 1000.0

				// Find the transition point
				rTransition := math.Sqrt(G * sunMass / aMax)

				// Test just before transition (should not be clamped)
				shipPos1 := entities.NewVec2(rTransition*1.1, 0.0)
				acc1 := GravityAcceleration(shipPos1, sunPos, sunMass, G, aMax)
				mag1 := acc1.Length()
				Expect(mag1).To(BeNumerically("<", aMax))

				// Test just after transition (should be clamped)
				shipPos2 := entities.NewVec2(rTransition*0.9, 0.0)
				acc2 := GravityAcceleration(shipPos2, sunPos, sunMass, G, aMax)
				mag2 := acc2.Length()
				Expect(mag2).To(BeNumerically("<=", aMax+epsilon))
			})
		})
	})
})

