package rules

import (
	"github.com/gorbit/orbitalrush/internal/sim/entities"
	"github.com/gorbit/orbitalrush/internal/sim/physics"
)

// CheckWinCondition checks if the win condition is met.
// Win condition: all pallets are collected (all Pallets have Active=false).
// If there are no pallets configured, the win condition cannot trigger.
//
// Parameters:
//   - world: Current world state
//
// Returns:
//   - true if all pallets are collected, false otherwise (including when no pallets exist)
func CheckWinCondition(world entities.World) bool {
	// No pallets means there is no collection objective yet, so the win condition cannot trigger.
	if len(world.Pallets) == 0 {
		return false
	}

	// Check if all pallets are collected (Active=false)
	for _, pallet := range world.Pallets {
		if pallet.Active {
			return false
		}
	}

	return true
}

// CheckLoseCondition checks if the lose condition is met for a specific ship.
// Lose condition: ship collides with any planet (using CheckShipPlanetCollisions).
//
// Parameters:
//   - world: Current world state
//   - shipID: ID of the ship to check
//
// Returns:
//   - true if ship collides with any planet, false otherwise (including if shipID not found)
func CheckLoseCondition(world entities.World, shipID uint32) bool {
	// Find ship by shipID
	var ship *entities.Ship
	for i := range world.Ships {
		if world.Ships[i].ID == shipID {
			ship = &world.Ships[i]
			break
		}
	}

	// If ship not found, return false (graceful handling)
	if ship == nil {
		return false
	}

	// If no planets, return false
	if len(world.Planets) == 0 {
		return false
	}

	// Check ship against all planets using CheckShipPlanetCollisions
	collides, _ := physics.CheckShipPlanetCollisions(ship.Pos, world.Planets)
	return collides
}

// EvaluateGameState evaluates win/lose conditions and updates World.Done and World.Win flags.
// Win condition takes precedence over lose condition (if both are true, win is set).
// Once Done is true, state should not change (idempotent evaluation).
// Other world fields (Ship, Sun, Pallets, Tick) are not modified.
//
// Parameters:
//   - world: Current world state
//
// Returns:
//   - Updated world with Done and Win flags set appropriately
func EvaluateGameState(world entities.World) entities.World {
	// If game is already done, return unchanged (idempotent)
	if world.Done {
		return world
	}

	// Check win condition first (takes precedence)
	if CheckWinCondition(world) {
		world.Done = true
		world.Win = true
		return world
	}

	// Check lose condition
	// Check all ships against all planets
	// If any ship collides with a planet, game ends with lose
	for _, ship := range world.Ships {
		if CheckLoseCondition(world, ship.ID) {
			world.Done = true
			world.Win = false
			return world
		}
	}

	// Neither condition met, leave Done and Win unchanged
	return world
}
