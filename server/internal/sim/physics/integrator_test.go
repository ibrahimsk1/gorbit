package physics

import (
	"math"
	"testing"

	"github.com/gorbit/orbitalrush/internal/sim/entities"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIntegrator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integrator Suite")
}

var _ = Describe("Integrator", Label("scope:unit", "loop:g1-physics", "layer:sim", "dep:none", "b:integration", "r:high", "double:fake"), func() {
	const epsilon = 1e-9
	const dt = 1.0 / 30.0 // 30Hz tick rate

	Describe("SemiImplicitEuler", func() {
		Describe("Determinism", func() {
			It("produces identical results for identical inputs across multiple runs", func() {
				pos := entities.NewVec2(0.0, 0.0)
				vel := entities.NewVec2(1.0, 2.0)
				acc := entities.NewVec2(0.5, -0.3)
				dt := 0.1

				// Run integration multiple times
				var firstPos, firstVel entities.Vec2
				for i := 0; i < 100; i++ {
					newPos, newVel := SemiImplicitEuler(pos, vel, acc, dt)
					if i == 0 {
						firstPos = newPos
						firstVel = newVel
					} else {
						// Verify bit-exact results
						Expect(newPos.X).To(Equal(firstPos.X))
						Expect(newPos.Y).To(Equal(firstPos.Y))
						Expect(newVel.X).To(Equal(firstVel.X))
						Expect(newVel.Y).To(Equal(firstVel.Y))
					}
				}
			})

			It("produces identical results when called with same inputs in different order", func() {
				// This test verifies that the function is pure and doesn't depend on external state
				pos1 := entities.NewVec2(0.0, 0.0)
				vel1 := entities.NewVec2(1.0, 0.0)
				acc1 := entities.NewVec2(0.0, 0.0)

				pos2 := entities.NewVec2(5.0, 5.0)
				vel2 := entities.NewVec2(2.0, 2.0)
				acc2 := entities.NewVec2(1.0, 1.0)

				// Call with first set
				result1Pos, result1Vel := SemiImplicitEuler(pos1, vel1, acc1, dt)
				// Call with second set (to verify function doesn't depend on external state)
				_, _ = SemiImplicitEuler(pos2, vel2, acc2, dt)
				// Call with first set again
				result1Pos2, result1Vel2 := SemiImplicitEuler(pos1, vel1, acc1, dt)

				// First and third calls should be identical
				Expect(result1Pos.X).To(Equal(result1Pos2.X))
				Expect(result1Pos.Y).To(Equal(result1Pos2.Y))
				Expect(result1Vel.X).To(Equal(result1Vel2.X))
				Expect(result1Vel.Y).To(Equal(result1Vel2.Y))
			})
		})

		Describe("Integration correctness", func() {
			It("moves object linearly with constant velocity", func() {
				pos := entities.NewVec2(0.0, 0.0)
				vel := entities.NewVec2(1.0, 2.0)
				acc := entities.NewVec2(0.0, 0.0) // No acceleration

				newPos, newVel := SemiImplicitEuler(pos, vel, acc, dt)

				// Position should be: p(t) = p(0) + v * t
				expectedPos := entities.NewVec2(vel.X*dt, vel.Y*dt)
				Expect(newPos.X).To(BeNumerically("~", expectedPos.X, epsilon))
				Expect(newPos.Y).To(BeNumerically("~", expectedPos.Y, epsilon))

				// Velocity should remain unchanged
				Expect(newVel.X).To(BeNumerically("~", vel.X, epsilon))
				Expect(newVel.Y).To(BeNumerically("~", vel.Y, epsilon))
			})

			It("follows parabolic path with constant acceleration", func() {
				pos := entities.NewVec2(0.0, 0.0)
				vel := entities.NewVec2(0.0, 0.0) // Starting from rest
				acc := entities.NewVec2(1.0, 0.0) // Constant acceleration in x direction

				newPos, newVel := SemiImplicitEuler(pos, vel, acc, dt)

				// Semi-implicit Euler: v_new = v_old + a*dt, p_new = p_old + v_new*dt
				// For v(0) = 0: v_new = a*dt, p_new = (a*dt)*dt = a*dt²
				expectedPos := entities.NewVec2(acc.X*dt*dt, acc.Y*dt*dt)
				Expect(newPos.X).To(BeNumerically("~", expectedPos.X, epsilon))
				Expect(newPos.Y).To(BeNumerically("~", expectedPos.Y, epsilon))

				// Velocity should be: v(t) = v(0) + a*t = a*t
				expectedVel := entities.NewVec2(acc.X*dt, acc.Y*dt)
				Expect(newVel.X).To(BeNumerically("~", expectedVel.X, epsilon))
				Expect(newVel.Y).To(BeNumerically("~", expectedVel.Y, epsilon))
			})

			It("handles zero acceleration correctly", func() {
				pos := entities.NewVec2(10.0, 20.0)
				vel := entities.NewVec2(5.0, -3.0)
				acc := entities.NewVec2(0.0, 0.0)

				newPos, newVel := SemiImplicitEuler(pos, vel, acc, dt)

				// Position should move by velocity * dt
				expectedPos := pos.Add(vel.Scale(dt))
				Expect(newPos.X).To(BeNumerically("~", expectedPos.X, epsilon))
				Expect(newPos.Y).To(BeNumerically("~", expectedPos.Y, epsilon))

				// Velocity should remain unchanged
				Expect(newVel.X).To(BeNumerically("~", vel.X, epsilon))
				Expect(newVel.Y).To(BeNumerically("~", vel.Y, epsilon))
			})

			It("handles zero velocity correctly", func() {
				pos := entities.NewVec2(5.0, 10.0)
				vel := entities.NewVec2(0.0, 0.0)
				acc := entities.NewVec2(1.0, 2.0)

				newPos, newVel := SemiImplicitEuler(pos, vel, acc, dt)

				// Semi-implicit Euler: v_new = v_old + a*dt = a*dt, p_new = p_old + v_new*dt
				// So position changes by: p_new = p_old + (a*dt)*dt = p_old + a*dt²
				expectedPos := pos.Add(entities.NewVec2(acc.X*dt*dt, acc.Y*dt*dt))
				Expect(newPos.X).To(BeNumerically("~", expectedPos.X, epsilon))
				Expect(newPos.Y).To(BeNumerically("~", expectedPos.Y, epsilon))

				// Velocity should be updated: v_new = v_old + a*dt = a*dt
				expectedVel := entities.NewVec2(acc.X*dt, acc.Y*dt)
				Expect(newVel.X).To(BeNumerically("~", expectedVel.X, epsilon))
				Expect(newVel.Y).To(BeNumerically("~", expectedVel.Y, epsilon))
			})

			It("handles zero velocity and zero acceleration", func() {
				pos := entities.NewVec2(7.0, 8.0)
				vel := entities.NewVec2(0.0, 0.0)
				acc := entities.NewVec2(0.0, 0.0)

				newPos, newVel := SemiImplicitEuler(pos, vel, acc, dt)

				// Position should remain unchanged
				Expect(newPos.X).To(BeNumerically("~", pos.X, epsilon))
				Expect(newPos.Y).To(BeNumerically("~", pos.Y, epsilon))

				// Velocity should remain zero
				Expect(newVel.X).To(BeNumerically("~", 0.0, epsilon))
				Expect(newVel.Y).To(BeNumerically("~", 0.0, epsilon))
			})
		})

		Describe("Fixed-step behavior", func() {
			It("produces consistent results for step size scaling", func() {
				// Semi-implicit Euler is not path-independent, but should produce
				// consistent and predictable results for different step sizes
				pos := entities.NewVec2(0.0, 0.0)
				vel := entities.NewVec2(1.0, 2.0)
				acc := entities.NewVec2(0.5, -0.3) // Constant acceleration
				totalDt := 1.0

				// Single large step
				_, largeVel := SemiImplicitEuler(pos, vel, acc, totalDt)

				// Multiple small steps
				smallPos := pos
				smallVel := vel
				numSteps := 10
				smallDt := totalDt / float64(numSteps)
				for i := 0; i < numSteps; i++ {
					smallPos, smallVel = SemiImplicitEuler(smallPos, smallVel, acc, smallDt)
				}

				// Velocity should match exactly (it's path-independent for constant acceleration)
				Expect(smallVel.X).To(BeNumerically("~", largeVel.X, 1e-10))
				Expect(smallVel.Y).To(BeNumerically("~", largeVel.Y, 1e-10))

				// Position will differ due to path-dependence, but should be reasonable
				// (small steps should produce a result, not NaN or Inf)
				Expect(math.IsNaN(smallPos.X)).To(BeFalse())
				Expect(math.IsNaN(smallPos.Y)).To(BeFalse())
				Expect(math.IsInf(smallPos.X, 0)).To(BeFalse())
				Expect(math.IsInf(smallPos.Y, 0)).To(BeFalse())
			})

			It("handles very small time steps", func() {
				pos := entities.NewVec2(0.0, 0.0)
				vel := entities.NewVec2(1.0, 1.0)
				acc := entities.NewVec2(0.1, 0.1)
				verySmallDt := 1e-10

				// Should not panic or produce invalid results
				newPos, newVel := SemiImplicitEuler(pos, vel, acc, verySmallDt)

				// Results should be very close to original (small change)
				Expect(newPos.X).To(BeNumerically(">=", pos.X))
				Expect(newPos.Y).To(BeNumerically(">=", pos.Y))
				Expect(math.IsNaN(newPos.X)).To(BeFalse())
				Expect(math.IsNaN(newPos.Y)).To(BeFalse())
				Expect(math.IsNaN(newVel.X)).To(BeFalse())
				Expect(math.IsNaN(newVel.Y)).To(BeFalse())
			})

			It("handles typical game time step (30Hz)", func() {
				pos := entities.NewVec2(0.0, 0.0)
				vel := entities.NewVec2(10.0, 20.0)
				acc := entities.NewVec2(5.0, -5.0)
				gameDt := 1.0 / 30.0

				newPos, newVel := SemiImplicitEuler(pos, vel, acc, gameDt)

				// Should produce reasonable results
				Expect(math.IsNaN(newPos.X)).To(BeFalse())
				Expect(math.IsNaN(newPos.Y)).To(BeFalse())
				Expect(math.IsNaN(newVel.X)).To(BeFalse())
				Expect(math.IsNaN(newVel.Y)).To(BeFalse())
				Expect(math.IsInf(newPos.X, 0)).To(BeFalse())
				Expect(math.IsInf(newPos.Y, 0)).To(BeFalse())
			})
		})

		Describe("Conservation properties", func() {
			It("approximately conserves energy for circular orbit scenario", func() {
				// Set up a circular orbit: object at distance r from origin
				// with tangential velocity v = sqrt(G*M/r)
				// For simplicity, use a = -G*M/r² * (r/r) = -v²/r * (r/r)
				r := 10.0
				v := 1.0
				pos := entities.NewVec2(r, 0.0)
				vel := entities.NewVec2(0.0, v)
				// Centripetal acceleration: a = -v²/r * (pos/r)
				acc := pos.Scale(-v * v / (r * r))

				// Calculate initial energy: E = 0.5*m*v² - G*M*m/r
				// For simplicity, assume m=1, and use kinetic energy only
				initialKineticEnergy := 0.5 * vel.LengthSq()

				// Integrate for multiple steps
				currentPos := pos
				currentVel := vel
				numSteps := 100
				dt := 0.01

				for i := 0; i < numSteps; i++ {
					// Update acceleration based on new position (simplified gravity)
					acc = currentPos.Scale(-v * v / (currentPos.LengthSq()))
					currentPos, currentVel = SemiImplicitEuler(currentPos, currentVel, acc, dt)
				}

				// Calculate final kinetic energy
				finalKineticEnergy := 0.5 * currentVel.LengthSq()

				// Energy should be approximately conserved (within 10% for this simple test)
				energyRatio := finalKineticEnergy / initialKineticEnergy
				Expect(energyRatio).To(BeNumerically("~", 1.0, 0.1))
			})

			It("conserves momentum for isolated system with zero net force", func() {
				pos := entities.NewVec2(0.0, 0.0)
				vel := entities.NewVec2(5.0, 10.0)
				acc := entities.NewVec2(0.0, 0.0) // No net force

				initialMomentum := vel

				// Integrate multiple steps
				currentPos := pos
				currentVel := vel
				numSteps := 50
				dt := 0.1

				for i := 0; i < numSteps; i++ {
					currentPos, currentVel = SemiImplicitEuler(currentPos, currentVel, acc, dt)
				}

				// Momentum should be exactly conserved (velocity unchanged)
				Expect(currentVel.X).To(BeNumerically("~", initialMomentum.X, epsilon))
				Expect(currentVel.Y).To(BeNumerically("~", initialMomentum.Y, epsilon))
			})
		})

		Describe("Edge cases", func() {
			It("handles zero time step", func() {
				pos := entities.NewVec2(1.0, 2.0)
				vel := entities.NewVec2(3.0, 4.0)
				acc := entities.NewVec2(5.0, 6.0)

				newPos, newVel := SemiImplicitEuler(pos, vel, acc, 0.0)

				// Position should remain unchanged (v_new = v_old, p_new = p_old + v_new*0 = p_old)
				Expect(newPos.X).To(BeNumerically("~", pos.X, epsilon))
				Expect(newPos.Y).To(BeNumerically("~", pos.Y, epsilon))

				// Velocity should remain unchanged (v_new = v_old + a*0 = v_old)
				Expect(newVel.X).To(BeNumerically("~", vel.X, epsilon))
				Expect(newVel.Y).To(BeNumerically("~", vel.Y, epsilon))
			})

			It("handles very large acceleration", func() {
				pos := entities.NewVec2(0.0, 0.0)
				vel := entities.NewVec2(0.0, 0.0)
				acc := entities.NewVec2(1e10, 1e10) // Very large acceleration
				dt := 1e-6                          // Small time step

				// Should not panic
				newPos, newVel := SemiImplicitEuler(pos, vel, acc, dt)

				// Results should be finite
				Expect(math.IsNaN(newPos.X)).To(BeFalse())
				Expect(math.IsNaN(newPos.Y)).To(BeFalse())
				Expect(math.IsNaN(newVel.X)).To(BeFalse())
				Expect(math.IsNaN(newVel.Y)).To(BeFalse())
			})

			It("handles negative time step", func() {
				pos := entities.NewVec2(1.0, 2.0)
				vel := entities.NewVec2(3.0, 4.0)
				acc := entities.NewVec2(5.0, 6.0)

				// Negative time step should work (backward integration)
				newPos, newVel := SemiImplicitEuler(pos, vel, acc, -0.1)

				// Should produce valid results (position should decrease, velocity should decrease)
				Expect(math.IsNaN(newPos.X)).To(BeFalse())
				Expect(math.IsNaN(newPos.Y)).To(BeFalse())
				Expect(math.IsNaN(newVel.X)).To(BeFalse())
				Expect(math.IsNaN(newVel.Y)).To(BeFalse())
			})
		})
	})
})
