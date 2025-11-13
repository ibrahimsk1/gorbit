package rules

// Energy economy constants
const (
	// MaxEnergy is the maximum energy value (energy bar cap)
	MaxEnergy = float32(100.0)
	// ThrustDrainRate is the energy drained per tick when thrusting
	ThrustDrainRate = float32(0.5)
	// PalletRestoreAmount is the energy restored per pallet pickup
	PalletRestoreAmount = float32(25.0)
)

// DrainEnergyOnThrust drains energy when the ship is thrusting.
// If isThrusting is true, energy is drained by ThrustDrainRate per tick.
// Energy cannot go below 0.
//
// Parameters:
//   - currentEnergy: Current energy level
//   - isThrusting: Whether the ship is currently thrusting
//
// Returns:
//   - New energy value after draining (clamped to [0, MaxEnergy])
func DrainEnergyOnThrust(currentEnergy float32, isThrusting bool) float32 {
	if !isThrusting {
		return currentEnergy
	}
	newEnergy := currentEnergy - ThrustDrainRate
	return ClampEnergy(newEnergy)
}

// RestoreEnergyOnPickup restores energy when a pallet is collected.
// Energy is increased by PalletRestoreAmount.
// Energy cannot exceed MaxEnergy.
//
// Parameters:
//   - currentEnergy: Current energy level
//
// Returns:
//   - New energy value after restoring (clamped to [0, MaxEnergy])
func RestoreEnergyOnPickup(currentEnergy float32) float32 {
	newEnergy := currentEnergy + PalletRestoreAmount
	return ClampEnergy(newEnergy)
}

// ClampEnergy clamps energy to valid range [0, MaxEnergy].
//
// Parameters:
//   - energy: Energy value to clamp
//
// Returns:
//   - Clamped energy value (0 <= energy <= MaxEnergy)
func ClampEnergy(energy float32) float32 {
	if energy < 0 {
		return 0
	}
	if energy > MaxEnergy {
		return MaxEnergy
	}
	return energy
}

