package rules

import (
	"github.com/gorbit/orbitalrush/internal/sim/entities"
	"github.com/gorbit/orbitalrush/internal/sim/physics"
)

// CheckWinCondition checks if the win condition is met.
// Win condition: all pallets are collected (all Pallets have Active=false).
// If there are no pallets, the win condition is considered met.
//
// Parameters:
//   - world: Current world state
//
// Returns:
//   - true if all pallets are collected (or no pallets exist), false otherwise
func CheckWinCondition(world entities.World) bool {
	// If there are no pallets, win condition is met
	if len(world.Pallets) == 0 {
		return true
	}

	// Check if all pallets are collected (Active=false)
	for _, pallet := range world.Pallets {
		if pallet.Active {
			return false
		}
	}

	return true
}

// CheckLoseCondition checks if the lose condition is met.
// Lose condition: ship collides with sun (using ShipSunCollision).
//
// Parameters:
//   - world: Current world state
//
// Returns:
//   - true if ship collides with sun, false otherwise
func CheckLoseCondition(world entities.World) bool {
	return physics.ShipSunCollision(world.Ship.Pos, world.Sun.Pos, world.Sun.Radius)
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
	if CheckLoseCondition(world) {
		world.Done = true
		world.Win = false
		return world
	}

	// Neither condition met, leave Done and Win unchanged
	return world
}
