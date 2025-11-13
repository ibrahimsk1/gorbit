package rules

import (
	"github.com/gorbit/orbitalrush/internal/sim/entities"
	"github.com/gorbit/orbitalrush/internal/sim/physics"
)

// Step performs one complete game loop step, applying all rules in the correct order:
// 1. Input → Apply player input (thrust, turn)
// 2. Physics → Update position and velocity (gravity + integrator)
// 3. Collisions → Process pallet pickups (deactivate pallet, restore energy)
// 4. Rules → Evaluate win/lose conditions (update Done/Win flags)
// 5. State → Increment tick counter
//
// If the game is already done (world.Done == true), most processing is skipped
// and only the tick counter is incremented.
//
// Parameters:
//   - world: Current world state
//   - input: Player input command (thrust, turn)
//   - dt: Time step in seconds
//   - G: Gravitational constant (game-scale)
//   - aMax: Maximum acceleration magnitude
//   - pickupRadius: Pallet pickup radius
//
// Returns:
//   - Updated world state after one game loop step
func Step(world entities.World, input InputCommand, dt float64, G float64, aMax float64, pickupRadius float64) entities.World {
	// If game is already done, skip processing and only increment tick
	if world.Done {
		world.Tick++
		return world
	}

	// Step 1: Apply Input
	// Process player input (thrust, turn) - updates rotation, velocity, and energy
	world.Ship = ApplyInput(world.Ship, input, dt)

	// Step 2: Update Physics
	// Calculate gravity acceleration and integrate position and velocity
	acc := physics.GravityAcceleration(world.Ship.Pos, world.Sun.Pos, world.Sun.Mass, G, aMax)
	newPos, newVel := physics.SemiImplicitEuler(world.Ship.Pos, world.Ship.Vel, acc, dt)
	world.Ship.Pos = newPos
	world.Ship.Vel = newVel

	// Step 3: Process Collisions
	// Check for pallet pickups and process them (deactivate pallet, restore energy)
	for i := range world.Pallets {
		if world.Pallets[i].Active && physics.ShipPalletCollision(world.Ship.Pos, world.Pallets[i].Pos, pickupRadius) {
			// Deactivate pallet
			world.Pallets[i].Active = false
			// Restore energy
			world.Ship.Energy = RestoreEnergyOnPickup(world.Ship.Energy)
		}
	}

	// Step 4: Evaluate Rules
	// Check win/lose conditions and update Done/Win flags
	world = EvaluateGameState(world)

	// Step 5: Update State
	// Increment tick counter
	world.Tick++

	return world
}

