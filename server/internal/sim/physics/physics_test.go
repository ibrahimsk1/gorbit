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

var _ = Describe("Physics Integration", Label("scope:unit", "loop:g2-physics", "layer:physics", "dep:none", "b:physics-integration", "r:high", "double:fake"), func() {
	const epsilon = 1e-9
	const dt = 1.0 / 30.0 // 30Hz tick rate
	const G = 1.0         // Gravitational constant
	const aMax = 100.0    // Maximum acceleration
	const pickupRadius = 1.2

	Describe("Determinism", func() {
		It("produces identical world states for identical initial conditions", func() {
			// Create initial world state with multiplayer model
			ships := []entities.Ship{
				entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 1.0), 0.0, 100.0),
			}
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(20.0, 0.0), true),
			}
			world1 := entities.NewWorld(ships, planets, pallets)
			world2 := entities.NewWorld(ships, planets, pallets)

			// Run simulation for multiple ticks
			numTicks := 50
			for i := 0; i < numTicks; i++ {
				// Simulate world1
				acc := CalculateTotalGravity(world1.Ships[0].Pos, world1.Planets, G, aMax)
				newPos, newVel := SemiImplicitEuler(world1.Ships[0].Pos, world1.Ships[0].Vel, acc, dt)
				world1.Ships[0].Pos = newPos
				world1.Ships[0].Vel = newVel
				world1.Tick++

				// Simulate world2 (same initial conditions)
				acc2 := CalculateTotalGravity(world2.Ships[0].Pos, world2.Planets, G, aMax)
				newPos2, newVel2 := SemiImplicitEuler(world2.Ships[0].Pos, world2.Ships[0].Vel, acc2, dt)
				world2.Ships[0].Pos = newPos2
				world2.Ships[0].Vel = newVel2
				world2.Tick++

				// Verify states are identical
				Expect(world1.Ships[0].Pos.X).To(Equal(world2.Ships[0].Pos.X))
				Expect(world1.Ships[0].Pos.Y).To(Equal(world2.Ships[0].Pos.Y))
				Expect(world1.Ships[0].Vel.X).To(Equal(world2.Ships[0].Vel.X))
				Expect(world1.Ships[0].Vel.Y).To(Equal(world2.Ships[0].Vel.Y))
				Expect(world1.Tick).To(Equal(world2.Tick))
			}
		})

		It("produces identical results when called with same inputs in different order", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 1.0), 0.0, 100.0)
			planets1 := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}

			ship2 := entities.NewShip(2, entities.NewVec2(5.0, 5.0), entities.NewVec2(1.0, 0.0), 0.0, 100.0)
			planets2 := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}

			// Run simulation for ship1
			acc1 := CalculateTotalGravity(ship1.Pos, planets1, G, aMax)
			pos1, vel1 := SemiImplicitEuler(ship1.Pos, ship1.Vel, acc1, dt)

			// Run simulation for ship2
			acc2 := CalculateTotalGravity(ship2.Pos, planets2, G, aMax)
			_, _ = SemiImplicitEuler(ship2.Pos, ship2.Vel, acc2, dt)

			// Run simulation for ship1 again
			acc1Again := CalculateTotalGravity(ship1.Pos, planets1, G, aMax)
			pos1Again, vel1Again := SemiImplicitEuler(ship1.Pos, ship1.Vel, acc1Again, dt)

			// First and third calls should be identical
			Expect(pos1.X).To(Equal(pos1Again.X))
			Expect(pos1.Y).To(Equal(pos1Again.Y))
			Expect(vel1.X).To(Equal(vel1Again.X))
			Expect(vel1.Y).To(Equal(vel1Again.Y))
		})
	})

	Describe("Gravity and Integration", func() {
		It("ship falls toward planets under gravity", func() {
			ship := entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}

			initialDistance := ship.Pos.Sub(planets[0].Pos).Length()

			// Run simulation for multiple ticks
			numTicks := 30
			for i := 0; i < numTicks; i++ {
				acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel
			}

			finalDistance := ship.Pos.Sub(planets[0].Pos).Length()

			// Ship should move closer to planet
			Expect(finalDistance).To(BeNumerically("<", initialDistance))

			// Velocity should point toward planet
			directionToPlanet := planets[0].Pos.Sub(ship.Pos).Normalize()
			velDirection := ship.Vel.Normalize()
			// Velocity should be in the same direction as direction to planet
			Expect(velDirection.Dot(directionToPlanet)).To(BeNumerically(">", 0.0))
		})

		It("ship follows orbital path around planet", func() {
			// Set up circular orbit: ship at distance r with tangential velocity
			r := 10.0
			// For circular orbit: v = sqrt(G*M/r)
			orbitalVel := math.Sqrt(G * 1000.0 / r)
			ship := entities.NewShip(1, entities.NewVec2(r, 0.0), entities.NewVec2(0.0, orbitalVel), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}

			initialDistance := ship.Pos.Sub(planets[0].Pos).Length()

			// Run simulation for multiple ticks
			numTicks := 100
			for i := 0; i < numTicks; i++ {
				acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel
			}

			finalDistance := ship.Pos.Sub(planets[0].Pos).Length()

			// Distance should remain approximately constant (within 20% for this simple test)
			distanceRatio := finalDistance / initialDistance
			Expect(distanceRatio).To(BeNumerically("~", 1.0, 0.2))
		})

		It("gravity acceleration affects velocity through integrator", func() {
			ship := entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}

			initialVel := ship.Vel.Length()

			// Run one step
			acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)
			newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
			ship.Pos = newPos
			ship.Vel = newVel

			finalVel := ship.Vel.Length()

			// Velocity should increase due to gravity
			Expect(finalVel).To(BeNumerically(">", initialVel))

			// Acceleration should be non-zero (unless at planet center)
			accMag := acc.Length()
			Expect(accMag).To(BeNumerically(">", 0.0))
		})

		It("simulation runs correctly for multiple steps", func() {
			ship := entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 1.0), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}

			// Run simulation for many ticks
			numTicks := 200
			for i := 0; i < numTicks; i++ {
				acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)
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
		It("detects ship-planet collision when ship reaches planet", func() {
			ship := entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(-1.0, 0.0), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}

			// Run simulation until collision or max ticks
			maxTicks := 100
			collided := false
			var collidedPlanetID uint32
			for i := 0; i < maxTicks; i++ {
				// Check collision
				var hasCollision bool
				hasCollision, collidedPlanetID = CheckShipPlanetCollisions(ship.Pos, planets)
				if hasCollision {
					collided = true
					break
				}

				// Update physics
				acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel
			}

			Expect(collided).To(BeTrue())
			Expect(collidedPlanetID).To(Equal(uint32(1)))
		})

		It("detects ship-pallet pickup when ship reaches pallet", func() {
			ship := entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(-1.0, 0.0), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}
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
				acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel
			}

			Expect(pickedUp).To(BeTrue())
		})

		It("detects collision at boundary", func() {
			planet := entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0)
			planets := []entities.Planet{planet}
			// Place ship exactly at planet radius
			ship := entities.NewShip(1, entities.NewVec2(float64(planet.Radius), 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)

			collided, planetID := CheckShipPlanetCollisions(ship.Pos, planets)
			Expect(collided).To(BeTrue())
			Expect(planetID).To(Equal(uint32(1)))
		})

		It("does not detect false collisions when ship is far from objects", func() {
			ship := entities.NewShip(1, entities.NewVec2(1000.0, 1000.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}
			pallet := entities.NewPallet(1, entities.NewVec2(0.0, 0.0), true)

			planetCollision, _ := CheckShipPlanetCollisions(ship.Pos, planets)
			palletCollision := ShipPalletCollision(ship.Pos, pallet.Pos, pickupRadius)

			Expect(planetCollision).To(BeFalse())
			Expect(palletCollision).To(BeFalse())
		})
	})

	Describe("End-to-End Scenarios", func() {
		It("maintains stable orbit over multiple ticks", func() {
			// Set up circular orbit
			r := 10.0
			orbitalVel := math.Sqrt(G * 1000.0 / r)
			ship := entities.NewShip(1, entities.NewVec2(r, 0.0), entities.NewVec2(0.0, orbitalVel), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}

			initialDistance := ship.Pos.Sub(planets[0].Pos).Length()

			// Run for 100 ticks
			numTicks := 100
			for i := 0; i < numTicks; i++ {
				acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel

				// Verify no numerical instability
				Expect(math.IsNaN(ship.Pos.X)).To(BeFalse())
				Expect(math.IsNaN(ship.Pos.Y)).To(BeFalse())
			}

			finalDistance := ship.Pos.Sub(planets[0].Pos).Length()
			// Distance should remain approximately constant
			distanceRatio := finalDistance / initialDistance
			Expect(distanceRatio).To(BeNumerically("~", 1.0, 0.3)) // Allow 30% variation for numerical integration
		})

		It("ship can approach and pick up pallet while under gravity", func() {
			ship := entities.NewShip(1, entities.NewVec2(15.0, 0.0), entities.NewVec2(-0.5, 0.0), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}
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
				acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel
			}

			Expect(pickedUp).To(BeTrue())

			// Ship should have moved closer to pallet
			finalDistanceToPallet := ship.Pos.Sub(pallet.Pos).Length()
			Expect(finalDistanceToPallet).To(BeNumerically("<=", pickupRadius))
		})

		It("ship collides with planet when trajectory intersects", func() {
			ship := entities.NewShip(1, entities.NewVec2(20.0, 0.0), entities.NewVec2(-2.0, 0.0), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}

			// Run simulation
			maxTicks := 200
			collided := false
			var collidedPlanetID uint32
			for i := 0; i < maxTicks; i++ {
				var hasCollision bool
				hasCollision, collidedPlanetID = CheckShipPlanetCollisions(ship.Pos, planets)
				if hasCollision {
					collided = true
					break
				}

				acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel
			}

			Expect(collided).To(BeTrue())
			Expect(collidedPlanetID).To(Equal(uint32(1)))
		})

		It("runs extended simulation without numerical instability", func() {
			ship := entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 1.0), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}

			// Run for extended period
			numTicks := 1000
			for i := 0; i < numTicks; i++ {
				acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)
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
		It("handles ship at planet center", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}

			// Should detect collision
			collided, planetID := CheckShipPlanetCollisions(ship.Pos, planets)
			Expect(collided).To(BeTrue())
			Expect(planetID).To(Equal(uint32(1)))

			// Gravity should return zero (handled by CalculateTotalGravity)
			acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)
			Expect(acc.X).To(BeNumerically("~", 0.0, epsilon))
			Expect(acc.Y).To(BeNumerically("~", 0.0, epsilon))
		})

		It("handles ship at pallet position", func() {
			ship := entities.NewShip(1, entities.NewVec2(5.0, 5.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			pallet := entities.NewPallet(1, entities.NewVec2(5.0, 5.0), true)

			// Should detect pickup
			pickedUp := ShipPalletCollision(ship.Pos, pallet.Pos, pickupRadius)
			Expect(pickedUp).To(BeTrue())
		})

		It("handles zero gravity (zero planet mass)", func() {
			ship := entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(1.0, 1.0), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 0.0), // Zero mass
			}

			initialPos := ship.Pos
			initialVel := ship.Vel

			// Run simulation
			numTicks := 10
			for i := 0; i < numTicks; i++ {
				acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)
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

		It("handles very close to planet (clamped acceleration)", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.1, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 10000.0), // Large mass
			}

			acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)
			accMag := acc.Length()

			// Acceleration should be clamped
			Expect(accMag).To(BeNumerically("<=", aMax+epsilon))
		})

		It("handles very far from planet (minimal gravity)", func() {
			ship := entities.NewShip(1, entities.NewVec2(10000.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}

			acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)
			accMag := acc.Length()

			// Acceleration should be very small
			Expect(accMag).To(BeNumerically("<", 1e-3))
		})
	})

	Describe("Conservation and Stability", func() {
		It("maintains energy within reasonable bounds", func() {
			ship := entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 1.0), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}

			// Calculate initial energy (kinetic + potential)
			// Kinetic: 0.5 * m * v² (assuming m=1)
			initialKinetic := 0.5 * ship.Vel.LengthSq()
			// Potential: -G*M/r (simplified, assuming m=1)
			initialDistance := ship.Pos.Sub(planets[0].Pos).Length()
			initialPotential := -G * planets[0].Mass / initialDistance
			initialEnergy := initialKinetic + initialPotential

			// Run simulation
			numTicks := 100
			for i := 0; i < numTicks; i++ {
				acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel
			}

			// Calculate final energy
			finalKinetic := 0.5 * ship.Vel.LengthSq()
			finalDistance := ship.Pos.Sub(planets[0].Pos).Length()
			finalPotential := -G * planets[0].Mass / finalDistance
			finalEnergy := finalKinetic + finalPotential

			// Energy should remain within reasonable bounds (within 50% for this simple test)
			energyRatio := finalEnergy / initialEnergy
			Expect(energyRatio).To(BeNumerically("~", 1.0, 0.5))
		})

		It("produces deterministic replay for same initial conditions", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 1.0), 0.0, 100.0)
			planets1 := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}

			ship2 := entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 1.0), 0.0, 100.0)
			planets2 := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}

			// Run both simulations
			numTicks := 50
			for i := 0; i < numTicks; i++ {
				acc1 := CalculateTotalGravity(ship1.Pos, planets1, G, aMax)
				pos1, vel1 := SemiImplicitEuler(ship1.Pos, ship1.Vel, acc1, dt)
				ship1.Pos = pos1
				ship1.Vel = vel1

				acc2 := CalculateTotalGravity(ship2.Pos, planets2, G, aMax)
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
			ship := entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 1.0), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}

			// Run for very long simulation
			numTicks := 5000
			for i := 0; i < numTicks; i++ {
				acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)
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

	Describe("Wraparound", func() {
		It("wraps ship position when exiting right boundary", func() {
			ship := entities.NewShip(1, entities.NewVec2(entities.WORLD_WIDTH/2+10.0, 0.0), entities.NewVec2(1.0, 0.0), 0.0, 100.0)
			wrappedPos := ApplyWraparound(ship.Pos, entities.WORLD_WIDTH, entities.WORLD_HEIGHT)

			// Should wrap to left side
			Expect(wrappedPos.X).To(BeNumerically("<", entities.WORLD_WIDTH/2))
			Expect(wrappedPos.X).To(BeNumerically(">", -entities.WORLD_WIDTH/2))
		})

		It("wraps ship position when exiting left boundary", func() {
			ship := entities.NewShip(1, entities.NewVec2(-entities.WORLD_WIDTH/2-10.0, 0.0), entities.NewVec2(-1.0, 0.0), 0.0, 100.0)
			wrappedPos := ApplyWraparound(ship.Pos, entities.WORLD_WIDTH, entities.WORLD_HEIGHT)

			// Should wrap to right side
			Expect(wrappedPos.X).To(BeNumerically(">", -entities.WORLD_WIDTH/2))
			Expect(wrappedPos.X).To(BeNumerically("<", entities.WORLD_WIDTH/2))
		})

		It("wraps ship position when exiting top boundary", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, entities.WORLD_HEIGHT/2+10.0), entities.NewVec2(0.0, 1.0), 0.0, 100.0)
			wrappedPos := ApplyWraparound(ship.Pos, entities.WORLD_WIDTH, entities.WORLD_HEIGHT)

			// Should wrap to bottom side
			Expect(wrappedPos.Y).To(BeNumerically("<", entities.WORLD_HEIGHT/2))
			Expect(wrappedPos.Y).To(BeNumerically(">", -entities.WORLD_HEIGHT/2))
		})

		It("wraps ship position when exiting bottom boundary", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, -entities.WORLD_HEIGHT/2-10.0), entities.NewVec2(0.0, -1.0), 0.0, 100.0)
			wrappedPos := ApplyWraparound(ship.Pos, entities.WORLD_WIDTH, entities.WORLD_HEIGHT)

			// Should wrap to top side
			Expect(wrappedPos.Y).To(BeNumerically(">", -entities.WORLD_HEIGHT/2))
			Expect(wrappedPos.Y).To(BeNumerically("<", entities.WORLD_HEIGHT/2))
		})

		It("applies wraparound during simulation", func() {
			ship := entities.NewShip(1, entities.NewVec2(entities.WORLD_WIDTH/2-5.0, 0.0), entities.NewVec2(10.0, 0.0), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}

			// Run simulation for a few ticks
			numTicks := 10
			for i := 0; i < numTicks; i++ {
				acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = ApplyWraparound(newPos, entities.WORLD_WIDTH, entities.WORLD_HEIGHT)
				ship.Vel = newVel
			}

			// Position should be within bounds
			Expect(ship.Pos.X).To(BeNumerically(">=", -entities.WORLD_WIDTH/2))
			Expect(ship.Pos.X).To(BeNumerically("<=", entities.WORLD_WIDTH/2))
			Expect(ship.Pos.Y).To(BeNumerically(">=", -entities.WORLD_HEIGHT/2))
			Expect(ship.Pos.Y).To(BeNumerically("<=", entities.WORLD_HEIGHT/2))
		})
	})

	Describe("Multiple Planets", func() {
		It("ship is affected by gravity from multiple planets", func() {
			ship := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(10.0, 0.0), 30.0, 1000.0),
				entities.NewPlanet(2, entities.NewVec2(-10.0, 0.0), 30.0, 1000.0),
				entities.NewPlanet(3, entities.NewVec2(0.0, 10.0), 30.0, 1000.0),
			}

			// Calculate gravity from all planets
			acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)

			// Acceleration should be non-zero (sum of all planet gravities)
			accMag := acc.Length()
			Expect(accMag).To(BeNumerically(">", 0.0))

			// Run simulation
			numTicks := 30
			for i := 0; i < numTicks; i++ {
				acc := CalculateTotalGravity(ship.Pos, planets, G, aMax)
				newPos, newVel := SemiImplicitEuler(ship.Pos, ship.Vel, acc, dt)
				ship.Pos = newPos
				ship.Vel = newVel
			}

			// Ship should have moved due to combined gravity
			Expect(ship.Pos.Length()).To(BeNumerically(">", 0.0))
			Expect(ship.Vel.Length()).To(BeNumerically(">", 0.0))
		})

		It("detects collision with correct planet when multiple planets present", func() {
			ship := entities.NewShip(1, entities.NewVec2(5.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(10.0, 0.0), 30.0, 1000.0),
				entities.NewPlanet(2, entities.NewVec2(0.0, 0.0), 50.0, 1000.0), // Ship is inside this planet
				entities.NewPlanet(3, entities.NewVec2(-10.0, 0.0), 30.0, 1000.0),
			}

			collided, planetID := CheckShipPlanetCollisions(ship.Pos, planets)
			Expect(collided).To(BeTrue())
			Expect(planetID).To(Equal(uint32(2))) // Should detect planet 2
		})

		It("handles multiple ships with multiple planets", func() {
			ships := []entities.Ship{
				entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 1.0), 0.0, 100.0),
				entities.NewShip(2, entities.NewVec2(-10.0, 0.0), entities.NewVec2(0.0, -1.0), 0.0, 100.0),
			}
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
				entities.NewPlanet(2, entities.NewVec2(20.0, 0.0), 30.0, 500.0),
				entities.NewPlanet(3, entities.NewVec2(-20.0, 0.0), 30.0, 500.0),
			}

			// Run simulation for multiple ticks
			numTicks := 50
			for i := 0; i < numTicks; i++ {
				for j := range ships {
					acc := CalculateTotalGravity(ships[j].Pos, planets, G, aMax)
					newPos, newVel := SemiImplicitEuler(ships[j].Pos, ships[j].Vel, acc, dt)
					ships[j].Pos = ApplyWraparound(newPos, entities.WORLD_WIDTH, entities.WORLD_HEIGHT)
					ships[j].Vel = newVel
				}
			}

			// Both ships should have moved
			Expect(ships[0].Pos.Length()).To(BeNumerically(">", 0.0))
			Expect(ships[1].Pos.Length()).To(BeNumerically(">", 0.0))
		})
	})

	Describe("Physics Determinism with Multiplayer Model", Label("scope:unit", "loop:g2-physics", "layer:physics", "dep:none", "b:determinism", "r:high", "double:fake"), func() {
		const epsilon = 1e-9
		const dt = 1.0 / 30.0
		const G = 1.0
		const aMax = 100.0

		It("produces identical results for complete physics simulation with multiple ships and planets across multiple runs", func() {
			// Create initial world state with multiple ships and planets
			ships1 := []entities.Ship{
				entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 1.0), 0.0, 100.0),
				entities.NewShip(2, entities.NewVec2(-10.0, 0.0), entities.NewVec2(0.0, -1.0), 0.0, 100.0),
				entities.NewShip(3, entities.NewVec2(0.0, 10.0), entities.NewVec2(1.0, 0.0), 0.0, 100.0),
			}
			planets1 := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
				entities.NewPlanet(2, entities.NewVec2(20.0, 0.0), 30.0, 800.0),
				entities.NewPlanet(3, entities.NewVec2(-20.0, 0.0), 30.0, 800.0),
				entities.NewPlanet(4, entities.NewVec2(0.0, 20.0), 30.0, 600.0),
			}

			// Create identical second world state
			ships2 := []entities.Ship{
				entities.NewShip(1, entities.NewVec2(10.0, 0.0), entities.NewVec2(0.0, 1.0), 0.0, 100.0),
				entities.NewShip(2, entities.NewVec2(-10.0, 0.0), entities.NewVec2(0.0, -1.0), 0.0, 100.0),
				entities.NewShip(3, entities.NewVec2(0.0, 10.0), entities.NewVec2(1.0, 0.0), 0.0, 100.0),
			}
			planets2 := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
				entities.NewPlanet(2, entities.NewVec2(20.0, 0.0), 30.0, 800.0),
				entities.NewPlanet(3, entities.NewVec2(-20.0, 0.0), 30.0, 800.0),
				entities.NewPlanet(4, entities.NewVec2(0.0, 20.0), 30.0, 600.0),
			}

			// Run simulation for multiple ticks
			numTicks := 100
			for i := 0; i < numTicks; i++ {
				// Simulate world1
				for j := range ships1 {
					acc := CalculateTotalGravity(ships1[j].Pos, planets1, G, aMax)
					newPos, newVel := SemiImplicitEuler(ships1[j].Pos, ships1[j].Vel, acc, dt)
					ships1[j].Pos = ApplyWraparound(newPos, entities.WORLD_WIDTH, entities.WORLD_HEIGHT)
					ships1[j].Vel = newVel
				}

				// Simulate world2 (same initial conditions)
				for j := range ships2 {
					acc := CalculateTotalGravity(ships2[j].Pos, planets2, G, aMax)
					newPos, newVel := SemiImplicitEuler(ships2[j].Pos, ships2[j].Vel, acc, dt)
					ships2[j].Pos = ApplyWraparound(newPos, entities.WORLD_WIDTH, entities.WORLD_HEIGHT)
					ships2[j].Vel = newVel
				}

				// Verify states are identical after each tick
				for j := range ships1 {
					Expect(ships1[j].Pos.X).To(Equal(ships2[j].Pos.X))
					Expect(ships1[j].Pos.Y).To(Equal(ships2[j].Pos.Y))
					Expect(ships1[j].Vel.X).To(Equal(ships2[j].Vel.X))
					Expect(ships1[j].Vel.Y).To(Equal(ships2[j].Vel.Y))
				}
			}
		})

		It("produces identical results when physics functions are called multiple times with same inputs", func() {
			// Test that all physics functions are pure (no side effects)
			shipPos := entities.NewVec2(10.0, 5.0)
			shipVel := entities.NewVec2(1.0, 2.0)
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
				entities.NewPlanet(2, entities.NewVec2(20.0, 0.0), 30.0, 800.0),
				entities.NewPlanet(3, entities.NewVec2(0.0, 20.0), 30.0, 600.0),
			}

			// Run complete physics step multiple times
			var firstPos, firstVel entities.Vec2
			var firstColliding bool
			for i := 0; i < 50; i++ {
				acc := CalculateTotalGravity(shipPos, planets, G, aMax)
				newPos, newVel := SemiImplicitEuler(shipPos, shipVel, acc, dt)
				wrappedPos := ApplyWraparound(newPos, entities.WORLD_WIDTH, entities.WORLD_HEIGHT)
				colliding, _ := CheckShipPlanetCollisions(wrappedPos, planets)

				if i == 0 {
					firstPos = wrappedPos
					firstVel = newVel
					firstColliding = colliding
				} else {
					// Verify bit-exact results
					Expect(wrappedPos.X).To(Equal(firstPos.X))
					Expect(wrappedPos.Y).To(Equal(firstPos.Y))
					Expect(newVel.X).To(Equal(firstVel.X))
					Expect(newVel.Y).To(Equal(firstVel.Y))
					// Collision detection should also be deterministic
					Expect(colliding).To(Equal(firstColliding))
				}
			}
		})

		It("maintains determinism with multiple ships interacting with multiple planets over extended simulation", func() {
			// Create 4 ships and 5 planets
			ships1 := []entities.Ship{
				entities.NewShip(1, entities.NewVec2(15.0, 0.0), entities.NewVec2(0.0, 1.5), 0.0, 100.0),
				entities.NewShip(2, entities.NewVec2(-15.0, 0.0), entities.NewVec2(0.0, -1.5), 0.0, 100.0),
				entities.NewShip(3, entities.NewVec2(0.0, 15.0), entities.NewVec2(1.5, 0.0), 0.0, 100.0),
				entities.NewShip(4, entities.NewVec2(0.0, -15.0), entities.NewVec2(-1.5, 0.0), 0.0, 100.0),
			}
			planets1 := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
				entities.NewPlanet(2, entities.NewVec2(30.0, 0.0), 30.0, 800.0),
				entities.NewPlanet(3, entities.NewVec2(-30.0, 0.0), 30.0, 800.0),
				entities.NewPlanet(4, entities.NewVec2(0.0, 30.0), 30.0, 600.0),
				entities.NewPlanet(5, entities.NewVec2(0.0, -30.0), 30.0, 600.0),
			}

			ships2 := []entities.Ship{
				entities.NewShip(1, entities.NewVec2(15.0, 0.0), entities.NewVec2(0.0, 1.5), 0.0, 100.0),
				entities.NewShip(2, entities.NewVec2(-15.0, 0.0), entities.NewVec2(0.0, -1.5), 0.0, 100.0),
				entities.NewShip(3, entities.NewVec2(0.0, 15.0), entities.NewVec2(1.5, 0.0), 0.0, 100.0),
				entities.NewShip(4, entities.NewVec2(0.0, -15.0), entities.NewVec2(-1.5, 0.0), 0.0, 100.0),
			}
			planets2 := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
				entities.NewPlanet(2, entities.NewVec2(30.0, 0.0), 30.0, 800.0),
				entities.NewPlanet(3, entities.NewVec2(-30.0, 0.0), 30.0, 800.0),
				entities.NewPlanet(4, entities.NewVec2(0.0, 30.0), 30.0, 600.0),
				entities.NewPlanet(5, entities.NewVec2(0.0, -30.0), 30.0, 600.0),
			}

			// Run extended simulation
			numTicks := 200
			for i := 0; i < numTicks; i++ {
				// Simulate world1
				for j := range ships1 {
					acc := CalculateTotalGravity(ships1[j].Pos, planets1, G, aMax)
					newPos, newVel := SemiImplicitEuler(ships1[j].Pos, ships1[j].Vel, acc, dt)
					ships1[j].Pos = ApplyWraparound(newPos, entities.WORLD_WIDTH, entities.WORLD_HEIGHT)
					ships1[j].Vel = newVel
				}

				// Simulate world2
				for j := range ships2 {
					acc := CalculateTotalGravity(ships2[j].Pos, planets2, G, aMax)
					newPos, newVel := SemiImplicitEuler(ships2[j].Pos, ships2[j].Vel, acc, dt)
					ships2[j].Pos = ApplyWraparound(newPos, entities.WORLD_WIDTH, entities.WORLD_HEIGHT)
					ships2[j].Vel = newVel
				}

				// Verify all ships have identical states
				for j := range ships1 {
					Expect(ships1[j].Pos.X).To(Equal(ships2[j].Pos.X))
					Expect(ships1[j].Pos.Y).To(Equal(ships2[j].Pos.Y))
					Expect(ships1[j].Vel.X).To(Equal(ships2[j].Vel.X))
					Expect(ships1[j].Vel.Y).To(Equal(ships2[j].Vel.Y))
				}
			}
		})

		It("verifies no side effects in physics functions (pure functions)", func() {
			// Test that physics functions don't modify their input parameters
			originalPos := entities.NewVec2(10.0, 5.0)
			originalVel := entities.NewVec2(1.0, 2.0)
			originalPlanets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
				entities.NewPlanet(2, entities.NewVec2(20.0, 0.0), 30.0, 800.0),
			}

			pos := originalPos
			vel := originalVel
			planets := make([]entities.Planet, len(originalPlanets))
			copy(planets, originalPlanets)

			// Call all physics functions
			acc := CalculateTotalGravity(pos, planets, G, aMax)
			newPos, newVel := SemiImplicitEuler(pos, vel, acc, dt)
			wrappedPos := ApplyWraparound(newPos, entities.WORLD_WIDTH, entities.WORLD_HEIGHT)
			_, _ = CheckShipPlanetCollisions(wrappedPos, planets)

			// Verify functions produced valid outputs (not NaN/Inf)
			Expect(math.IsNaN(newPos.X)).To(BeFalse())
			Expect(math.IsNaN(newPos.Y)).To(BeFalse())
			Expect(math.IsNaN(newVel.X)).To(BeFalse())
			Expect(math.IsNaN(newVel.Y)).To(BeFalse())

			// Verify input parameters were not modified
			Expect(pos.X).To(Equal(originalPos.X))
			Expect(pos.Y).To(Equal(originalPos.Y))
			Expect(vel.X).To(Equal(originalVel.X))
			Expect(vel.Y).To(Equal(originalVel.Y))
			Expect(planets[0].Pos.X).To(Equal(originalPlanets[0].Pos.X))
			Expect(planets[0].Pos.Y).To(Equal(originalPlanets[0].Pos.Y))
			Expect(planets[1].Pos.X).To(Equal(originalPlanets[1].Pos.X))
			Expect(planets[1].Pos.Y).To(Equal(originalPlanets[1].Pos.Y))
		})
	})
})

