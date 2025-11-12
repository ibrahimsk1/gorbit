package physics

import (
	"math"
	"testing"

	"github.com/gorbit/orbitalrush/internal/sim/entities"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPhysics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Physics Integration Suite")
}

var _ = Describe("Physics Integration", Label("scope:unit", "loop:g1-physics", "layer:sim", "dep:none", "b:physics-integration", "r:high", "double:fake"), func() {
	const epsilon = 1e-9
	const dt = 1.0 / 30.0 // 30Hz tick rate
	const G = 1.0         // Gravitational constant
	const aMax = 100.0    // Maximum acceleration
	const pickupRadius = 1.2

	Describe("Determinism", func() {
		It("produces identical world states for identical initial conditions", func() {
			// Create initial world state
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 1.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(
				entities.NewVec2(0.0, 0.0),
				50.0,
				1000.0,
			)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(20.0, 0.0), true),
			}
			world1 := entities.NewWorld(ship, sun, pallets)
			world2 := entities.NewWorld(ship, sun, pallets)

			// Run simulation for multiple ticks
			numTicks := 50
			for i := 0; i < numTicks; i++ {
				// Simulate world1
				acc := GravityAcceleration(world1.Ship.Pos, world1.Sun.Pos, world1.Sun.Mass, G, aMax)
				newPos, newVel := SemiImplicitEuler(world1.Ship.Pos, world1.Ship.Vel, acc, dt)
				world1.Ship.Pos = newPos
				world1.Ship.Vel = newVel
				world1.Tick++

				// Simulate world2 (same initial conditions)
				acc2 := GravityAcceleration(world2.Ship.Pos, world2.Sun.Pos, world2.Sun.Mass, G, aMax)
				newPos2, newVel2 := SemiImplicitEuler(world2.Ship.Pos, world2.Ship.Vel, acc2, dt)
				world2.Ship.Pos = newPos2
				world2.Ship.Vel = newVel2
				world2.Tick++

				// Verify states are identical
				Expect(world1.Ship.Pos.X).To(Equal(world2.Ship.Pos.X))
				Expect(world1.Ship.Pos.Y).To(Equal(world2.Ship.Pos.Y))
				Expect(world1.Ship.Vel.X).To(Equal(world2.Ship.Vel.X))
				Expect(world1.Ship.Vel.Y).To(Equal(world2.Ship.Vel.Y))
				Expect(world1.Tick).To(Equal(world2.Tick))
			}
		})

		It("produces identical results when called with same inputs in different order", func() {
			ship1 := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 1.0),
				0.0,
				100.0,
			)
			sun1 := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

			ship2 := entities.NewShip(
				entities.NewVec2(5.0, 5.0),
				entities.NewVec2(1.0, 0.0),
				0.0,
				100.0,
			)
			sun2 := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

			// Run simulation for ship1
			acc1 := GravityAcceleration(ship1.Pos, sun1.Pos, sun1.Mass, G, aMax)
			pos1, vel1 := SemiImplicitEuler(ship1.Pos, ship1.Vel, acc1, dt)

			// Run simulation for ship2
			acc2 := GravityAcceleration(ship2.Pos, sun2.Pos, sun2.Mass, G, aMax)
			_, _ = SemiImplicitEuler(ship2.Pos, ship2.Vel, acc2, dt)

			// Run simulation for ship1 again
			acc1Again := GravityAcceleration(ship1.Pos, sun1.Pos, sun1.Mass, G, aMax)
			pos1Again, vel1Again := SemiImplicitEuler(ship1.Pos, ship1.Vel, acc1Again, dt)

			// First and third calls should be identical
			Expect(pos1.X).To(Equal(pos1Again.X))
			Expect(pos1.Y).To(Equal(pos1Again.Y))
			Expect(vel1.X).To(Equal(vel1Again.X))
			Expect(vel1.Y).To(Equal(vel1Again.Y))
		})
	})

	Describe("Gravity and Integration", func() {
		It("ship falls toward sun under gravity", func() {
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0), // Zero initial velocity
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

			initialDistance := ship.Pos.Sub(sun.Pos).Length()

			// Run simulation for multiple ticks
			numTicks := 30
			for i := 0; i < numTicks; i++ {
				acc := GravityAcceleration(ship.Pos, sun.Pos, sun.Mass, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel
			}

			finalDistance := ship.Pos.Sub(sun.Pos).Length()

			// Ship should move closer to sun
			Expect(finalDistance).To(BeNumerically("<", initialDistance))

			// Velocity should point toward sun (negative direction from ship to sun)
			directionToSun := sun.Pos.Sub(ship.Pos).Normalize()
			velDirection := ship.Vel.Normalize()
			// Velocity should be in the same direction as direction to sun
			Expect(velDirection.Dot(directionToSun)).To(BeNumerically(">", 0.0))
		})

		It("ship follows orbital path", func() {
			// Set up circular orbit: ship at distance r with tangential velocity
			r := 10.0
			// For circular orbit: v = sqrt(G*M/r)
			orbitalVel := math.Sqrt(G * 1000.0 / r)
			ship := entities.NewShip(
				entities.NewVec2(r, 0.0),
				entities.NewVec2(0.0, orbitalVel), // Tangential velocity
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

			initialDistance := ship.Pos.Sub(sun.Pos).Length()

			// Run simulation for multiple ticks
			numTicks := 100
			for i := 0; i < numTicks; i++ {
				acc := GravityAcceleration(ship.Pos, sun.Pos, sun.Mass, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel
			}

			finalDistance := ship.Pos.Sub(sun.Pos).Length()

			// Distance should remain approximately constant (within 20% for this simple test)
			distanceRatio := finalDistance / initialDistance
			Expect(distanceRatio).To(BeNumerically("~", 1.0, 0.2))
		})

		It("gravity acceleration affects velocity through integrator", func() {
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0), // Zero initial velocity
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

			initialVel := ship.Vel.Length()

			// Run one step
			acc := GravityAcceleration(ship.Pos, sun.Pos, sun.Mass, G, aMax)
			newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
			ship.Pos = newPos
			ship.Vel = newVel

			finalVel := ship.Vel.Length()

			// Velocity should increase due to gravity
			Expect(finalVel).To(BeNumerically(">", initialVel))

			// Acceleration should be non-zero (unless at sun center)
			accMag := acc.Length()
			Expect(accMag).To(BeNumerically(">", 0.0))
		})

		It("simulation runs correctly for multiple steps", func() {
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 1.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

			// Run simulation for many ticks
			numTicks := 200
			for i := 0; i < numTicks; i++ {
				acc := GravityAcceleration(ship.Pos, sun.Pos, sun.Mass, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel

				// Verify no NaN or Inf values
				Expect(math.IsNaN(ship.Pos.X)).To(BeFalse())
				Expect(math.IsNaN(ship.Pos.Y)).To(BeFalse())
				Expect(math.IsNaN(ship.Vel.X)).To(BeFalse())
				Expect(math.IsNaN(ship.Vel.Y)).To(BeFalse())
				Expect(math.IsInf(ship.Pos.X, 0)).To(BeFalse())
				Expect(math.IsInf(ship.Pos.Y, 0)).To(BeFalse())
				Expect(math.IsInf(ship.Vel.X, 0)).To(BeFalse())
				Expect(math.IsInf(ship.Vel.Y, 0)).To(BeFalse())
			}
		})
	})

	Describe("Collision Detection", func() {
		It("detects ship-sun collision when ship reaches sun", func() {
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(-1.0, 0.0), // Moving toward sun
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

			// Run simulation until collision or max ticks
			maxTicks := 100
			collided := false
			for i := 0; i < maxTicks; i++ {
				// Check collision
				if ShipSunCollision(ship.Pos, sun.Pos, sun.Radius) {
					collided = true
					break
				}

				// Update physics
				acc := GravityAcceleration(ship.Pos, sun.Pos, sun.Mass, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel
			}

			Expect(collided).To(BeTrue())
		})

		It("detects ship-pallet pickup when ship reaches pallet", func() {
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(-1.0, 0.0), // Moving toward pallet
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallet := entities.NewPallet(1, entities.NewVec2(5.0, 0.0), true)

			// Run simulation until pickup or max ticks
			maxTicks := 100
			pickedUp := false
			for i := 0; i < maxTicks; i++ {
				// Check collision
				if ShipPalletCollision(ship.Pos, pallet.Pos, pickupRadius) {
					pickedUp = true
					break
				}

				// Update physics
				acc := GravityAcceleration(ship.Pos, sun.Pos, sun.Mass, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel
			}

			Expect(pickedUp).To(BeTrue())
		})

		It("detects collision at boundary", func() {
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			// Place ship exactly at sun radius
			ship := entities.NewShip(
				entities.NewVec2(float64(sun.Radius), 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)

			collided := ShipSunCollision(ship.Pos, sun.Pos, sun.Radius)
			Expect(collided).To(BeTrue())
		})

		It("does not detect false collisions when ship is far from objects", func() {
			ship := entities.NewShip(
				entities.NewVec2(1000.0, 1000.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallet := entities.NewPallet(1, entities.NewVec2(0.0, 0.0), true)

			sunCollision := ShipSunCollision(ship.Pos, sun.Pos, sun.Radius)
			palletCollision := ShipPalletCollision(ship.Pos, pallet.Pos, pickupRadius)

			Expect(sunCollision).To(BeFalse())
			Expect(palletCollision).To(BeFalse())
		})
	})

	Describe("End-to-End Scenarios", func() {
		It("maintains stable orbit over multiple ticks", func() {
			// Set up circular orbit
			r := 10.0
			orbitalVel := math.Sqrt(G * 1000.0 / r)
			ship := entities.NewShip(
				entities.NewVec2(r, 0.0),
				entities.NewVec2(0.0, orbitalVel),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

			initialDistance := ship.Pos.Sub(sun.Pos).Length()

			// Run for 100 ticks
			numTicks := 100
			for i := 0; i < numTicks; i++ {
				acc := GravityAcceleration(ship.Pos, sun.Pos, sun.Mass, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel

				// Verify no numerical instability
				Expect(math.IsNaN(ship.Pos.X)).To(BeFalse())
				Expect(math.IsNaN(ship.Pos.Y)).To(BeFalse())
			}

			finalDistance := ship.Pos.Sub(sun.Pos).Length()
			// Distance should remain approximately constant
			distanceRatio := finalDistance / initialDistance
			Expect(distanceRatio).To(BeNumerically("~", 1.0, 0.3)) // Allow 30% variation for numerical integration
		})

		It("ship can approach and pick up pallet while under gravity", func() {
			ship := entities.NewShip(
				entities.NewVec2(15.0, 0.0),
				entities.NewVec2(-0.5, 0.0), // Moving toward pallet
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			pallet := entities.NewPallet(1, entities.NewVec2(5.0, 0.0), true)

			// Run simulation
			maxTicks := 200
			pickedUp := false
			for i := 0; i < maxTicks; i++ {
				// Check pickup
				if ShipPalletCollision(ship.Pos, pallet.Pos, pickupRadius) {
					pickedUp = true
					break
				}

				// Update physics (gravity affects ship)
				acc := GravityAcceleration(ship.Pos, sun.Pos, sun.Mass, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel
			}

			Expect(pickedUp).To(BeTrue())

			// Ship should have moved closer to pallet
			finalDistanceToPallet := ship.Pos.Sub(pallet.Pos).Length()
			Expect(finalDistanceToPallet).To(BeNumerically("<=", pickupRadius))
		})

		It("ship collides with sun when trajectory intersects", func() {
			ship := entities.NewShip(
				entities.NewVec2(20.0, 0.0),
				entities.NewVec2(-2.0, 0.0), // Moving directly toward sun
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

			// Run simulation
			maxTicks := 200
			collided := false
			for i := 0; i < maxTicks; i++ {
				if ShipSunCollision(ship.Pos, sun.Pos, sun.Radius) {
					collided = true
					break
				}

				acc := GravityAcceleration(ship.Pos, sun.Pos, sun.Mass, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel
			}

			Expect(collided).To(BeTrue())
		})

		It("runs extended simulation without numerical instability", func() {
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 1.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

			// Run for extended period
			numTicks := 1000
			for i := 0; i < numTicks; i++ {
				acc := GravityAcceleration(ship.Pos, sun.Pos, sun.Mass, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel

				// Verify no numerical instability
				Expect(math.IsNaN(ship.Pos.X)).To(BeFalse())
				Expect(math.IsNaN(ship.Pos.Y)).To(BeFalse())
				Expect(math.IsNaN(ship.Vel.X)).To(BeFalse())
				Expect(math.IsNaN(ship.Vel.Y)).To(BeFalse())
				Expect(math.IsInf(ship.Pos.X, 0)).To(BeFalse())
				Expect(math.IsInf(ship.Pos.Y, 0)).To(BeFalse())
				Expect(math.IsInf(ship.Vel.X, 0)).To(BeFalse())
				Expect(math.IsInf(ship.Vel.Y, 0)).To(BeFalse())
			}
		})
	})

	Describe("Edge Cases", func() {
		It("handles ship at sun center", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

			// Should detect collision
			collided := ShipSunCollision(ship.Pos, sun.Pos, sun.Radius)
			Expect(collided).To(BeTrue())

			// Gravity should return zero (handled by GravityAcceleration)
			acc := GravityAcceleration(ship.Pos, sun.Pos, sun.Mass, G, aMax)
			Expect(acc.X).To(BeNumerically("~", 0.0, epsilon))
			Expect(acc.Y).To(BeNumerically("~", 0.0, epsilon))
		})

		It("handles ship at pallet position", func() {
			ship := entities.NewShip(
				entities.NewVec2(5.0, 5.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			pallet := entities.NewPallet(1, entities.NewVec2(5.0, 5.0), true)

			// Should detect pickup
			pickedUp := ShipPalletCollision(ship.Pos, pallet.Pos, pickupRadius)
			Expect(pickedUp).To(BeTrue())
		})

		It("handles zero gravity (zero sun mass)", func() {
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(1.0, 1.0), // Initial velocity
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 0.0) // Zero mass

			initialPos := ship.Pos
			initialVel := ship.Vel

			// Run simulation
			numTicks := 10
			for i := 0; i < numTicks; i++ {
				acc := GravityAcceleration(ship.Pos, sun.Pos, sun.Mass, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel
			}

			// Velocity should remain constant
			Expect(ship.Vel.X).To(BeNumerically("~", initialVel.X, epsilon))
			Expect(ship.Vel.Y).To(BeNumerically("~", initialVel.Y, epsilon))

			// Position should move by velocity * time
			expectedPos := initialPos.Add(initialVel.Scale(dt * float64(numTicks)))
			Expect(ship.Pos.X).To(BeNumerically("~", expectedPos.X, epsilon*10))
			Expect(ship.Pos.Y).To(BeNumerically("~", expectedPos.Y, epsilon*10))
		})

		It("handles very close to sun (clamped acceleration)", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.1, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 10000.0) // Large mass

			acc := GravityAcceleration(ship.Pos, sun.Pos, sun.Mass, G, aMax)
			accMag := acc.Length()

			// Acceleration should be clamped
			Expect(accMag).To(BeNumerically("<=", aMax+epsilon))
		})

		It("handles very far from sun (minimal gravity)", func() {
			ship := entities.NewShip(
				entities.NewVec2(10000.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

			acc := GravityAcceleration(ship.Pos, sun.Pos, sun.Mass, G, aMax)
			accMag := acc.Length()

			// Acceleration should be very small
			Expect(accMag).To(BeNumerically("<", 1e-3))
		})
	})

	Describe("Conservation and Stability", func() {
		It("maintains energy within reasonable bounds", func() {
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 1.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

			// Calculate initial energy (kinetic + potential)
			// Kinetic: 0.5 * m * v² (assuming m=1)
			initialKinetic := 0.5 * ship.Vel.LengthSq()
			// Potential: -G*M/r (simplified, assuming m=1)
			initialDistance := ship.Pos.Sub(sun.Pos).Length()
			initialPotential := -G * sun.Mass / initialDistance
			initialEnergy := initialKinetic + initialPotential

			// Run simulation
			numTicks := 100
			for i := 0; i < numTicks; i++ {
				acc := GravityAcceleration(ship.Pos, sun.Pos, sun.Mass, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel
			}

			// Calculate final energy
			finalKinetic := 0.5 * ship.Vel.LengthSq()
			finalDistance := ship.Pos.Sub(sun.Pos).Length()
			finalPotential := -G * sun.Mass / finalDistance
			finalEnergy := finalKinetic + finalPotential

			// Energy should remain within reasonable bounds (within 50% for this simple test)
			energyRatio := finalEnergy / initialEnergy
			Expect(energyRatio).To(BeNumerically("~", 1.0, 0.5))
		})

		It("produces deterministic replay for same initial conditions", func() {
			ship1 := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 1.0),
				0.0,
				100.0,
			)
			sun1 := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

			ship2 := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 1.0),
				0.0,
				100.0,
			)
			sun2 := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

			// Run both simulations
			numTicks := 50
			for i := 0; i < numTicks; i++ {
				acc1 := GravityAcceleration(ship1.Pos, sun1.Pos, sun1.Mass, G, aMax)
				pos1, vel1 := SemiImplicitEuler(ship1.Pos, ship1.Vel, acc1, dt)
				ship1.Pos = pos1
				ship1.Vel = vel1

				acc2 := GravityAcceleration(ship2.Pos, sun2.Pos, sun2.Mass, G, aMax)
				pos2, vel2 := SemiImplicitEuler(ship2.Pos, ship2.Vel, acc2, dt)
				ship2.Pos = pos2
				ship2.Vel = vel2
			}

			// Final states should be identical
			Expect(ship1.Pos.X).To(Equal(ship2.Pos.X))
			Expect(ship1.Pos.Y).To(Equal(ship2.Pos.Y))
			Expect(ship1.Vel.X).To(Equal(ship2.Vel.X))
			Expect(ship1.Vel.Y).To(Equal(ship2.Vel.Y))
		})

		It("maintains numerical stability over long simulation", func() {
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 1.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 50.0, 1000.0)

			// Run for very long simulation
			numTicks := 5000
			for i := 0; i < numTicks; i++ {
				acc := GravityAcceleration(ship.Pos, sun.Pos, sun.Mass, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel

				// Verify no numerical instability
				Expect(math.IsNaN(ship.Pos.X)).To(BeFalse())
				Expect(math.IsNaN(ship.Pos.Y)).To(BeFalse())
				Expect(math.IsNaN(ship.Vel.X)).To(BeFalse())
				Expect(math.IsNaN(ship.Vel.Y)).To(BeFalse())
				Expect(math.IsInf(ship.Pos.X, 0)).To(BeFalse())
				Expect(math.IsInf(ship.Pos.Y, 0)).To(BeFalse())
				Expect(math.IsInf(ship.Vel.X, 0)).To(BeFalse())
				Expect(math.IsInf(ship.Vel.Y, 0)).To(BeFalse())
			}
		})
	})
})

