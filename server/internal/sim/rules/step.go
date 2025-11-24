package rules

import (
	"github.com/gorbit/orbitalrush/internal/sim/entities"
	"github.com/gorbit/orbitalrush/internal/sim/physics"
)

// Step performs one complete game loop step, applying all rules in the correct order:
// 1. Input → Apply player input (thrust, turn)
// 2. Physics → Update position and velocity (gravity + integrator + wraparound) for all ships
// 3. Collisions → Process pallet pickups (deactivate pallet, restore energy)
// 4. Rules → Evaluate win/lose conditions (update Done/Win flags)
// 5. State → Increment tick counter
//
// If the game is already done (world.Done == true), most processing is skipped
// and only the tick counter is incremented.
//
// Parameters:
//   - world: Current world state
//   - playerID: Player ID for input command (identifies which ship to update)
//   - input: Player input command (thrust, turn)
//   - dt: Time step in seconds
//   - G: Gravitational constant (game-scale)
//   - aMax: Maximum acceleration magnitude
//   - pickupRadius: Pallet pickup radius
//   - worldWidth: World width for wraparound (typically 2000.0 m)
//   - worldHeight: World height for wraparound (typically 2000.0 m)
//
// Returns:
//   - Updated world state after one game loop step
func Step(world entities.World, playerID uint32, input InputCommand, dt float64, G float64, aMax float64, pickupRadius float64, worldWidth float64, worldHeight float64) entities.World {
	// If game is already done, skip processing and only increment tick
	if world.Done {
		world.Tick++
		return world
	}

	// Step 1: Apply Input
	// Process player input (thrust, turn) - updates rotation, velocity, and energy
	world = ApplyInput(world, playerID, input, dt)

	// Step 2: Update Physics (for all ships)
	// Process each ship: calculate gravity, apply drag, integrate, apply wraparound
	const dragK = 0.12 // Linear drag coefficient (from SPEC)
	for i := range world.Ships {
		ship := world.Ships[i]

		// Calculate total gravity from all planets
		gravityAcc := physics.CalculateTotalGravity(ship.Pos, world.Planets, G, aMax)

		// Apply drag: a_drag = -k * v (linear drag)
		dragAcc := ship.Vel.Scale(-dragK)

		// Total acceleration = gravity + drag
		totalAcc := gravityAcc.Add(dragAcc)

		// Integrate position and velocity using semi-implicit Euler
		newPos, newVel := physics.SemiImplicitEuler(ship.Pos, ship.Vel, totalAcc, dt)

		// Apply wraparound to keep ship within world bounds
		wrappedPos := physics.ApplyWraparound(newPos, worldWidth, worldHeight)

		// Update ship in world array (preserve ID, rotation, and energy)
		world.Ships[i] = entities.NewShip(
			ship.ID,
			wrappedPos,
			newVel,
			ship.Rot,    // Rotation not modified by physics
			ship.Energy, // Energy not modified by physics (modified by input/collisions)
		)
	}

	// Step 3: Process Collisions (for all ships)
	// Process pallet pickups and planet collisions for each ship
	for i := range world.Ships {
		ship := world.Ships[i]

		// Process pallet pickups
		// Loop through all pallets and check for collisions
		for j := range world.Pallets {
			// Only check active pallets
			if !world.Pallets[j].Active {
				continue
			}

			// Check collision with pallet
			if physics.ShipPalletCollision(ship.Pos, world.Pallets[j].Pos, pickupRadius) {
				// Deactivate pallet
				world.Pallets[j].Active = false

				// Restore energy for the ship
				restoredEnergy := RestoreEnergyOnPickup(ship.Energy)
				world.Ships[i] = entities.NewShip(
					ship.ID,
					ship.Pos,
					ship.Vel,
					ship.Rot,
					restoredEnergy,
				)
				// Update ship reference for next iteration
				ship = world.Ships[i]
			}
		}

		// Process planet collisions
		// Check collision with any planet (detection only, lose condition evaluation in EvaluateGameState)
		// Note: Planet collision lose condition will be handled in cu/lose-condition-multiplanet
		_, _ = physics.CheckShipPlanetCollisions(ship.Pos, world.Planets)
	}

	// Step 4: Evaluate Rules
	// Check win/lose conditions and update Done/Win flags
	world = EvaluateGameState(world)

	// Step 5: Update State
	// Increment tick counter
	world.Tick++

	return world
}
