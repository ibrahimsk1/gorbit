package rules

import (
	"math"

	"github.com/gorbit/orbitalrush/internal/sim/entities"
)

// Input command constants
const (
	// ThrustAcceleration is the acceleration magnitude per unit thrust input (m/s²)
	ThrustAcceleration = 20.0
	// TurnRate is the angular velocity per unit turn input (rad/s)
	TurnRate = 3.0
	// MinEnergyForThrust is the minimum energy required to thrust (thrust only when energy > 0)
	MinEnergyForThrust = 0.0
)

// InputCommand represents a player input command.
type InputCommand struct {
	// Thrust is the thrust input value, clamped to [0.0, 1.0]
	Thrust float32
	// Turn is the turn input value, clamped to [-1.0, 1.0]
	// Positive values turn right (counter-clockwise), negative values turn left (clockwise)
	Turn float32
}

// ClampInput clamps input values to valid ranges.
//
// Parameters:
//   - input: Input command to clamp
//
// Returns:
//   - Clamped input command with Thrust in [0.0, 1.0] and Turn in [-1.0, 1.0]
func ClampInput(input InputCommand) InputCommand {
	clamped := input

	// Clamp thrust to [0.0, 1.0]
	if clamped.Thrust < 0.0 {
		clamped.Thrust = 0.0
	}
	if clamped.Thrust > 1.0 {
		clamped.Thrust = 1.0
	}

	// Clamp turn to [-1.0, 1.0]
	if clamped.Turn < -1.0 {
		clamped.Turn = -1.0
	}
	if clamped.Turn > 1.0 {
		clamped.Turn = 1.0
	}

	return clamped
}

// UpdateRotation updates the ship's rotation based on turn input.
// The rotation is normalized to [0, 2π) range.
//
// Parameters:
//   - currentRot: Current rotation angle in radians
//   - turnInput: Turn input value (-1.0 to 1.0)
//   - dt: Time step in seconds
//
// Returns:
//   - Updated rotation angle in radians, normalized to [0, 2π)
func UpdateRotation(currentRot float64, turnInput float64, dt float64) float64 {
	newRot := currentRot + TurnRate*turnInput*dt
	return normalizeRotation(newRot)
}

// normalizeRotation normalizes rotation angle to [0, 2π) range.
func normalizeRotation(rot float64) float64 {
	// Normalize to [0, 2π)
	rot = math.Mod(rot, 2*math.Pi)
	if rot < 0 {
		rot += 2 * math.Pi
	}
	return rot
}

// CalculateThrustAcceleration calculates the thrust acceleration vector
// in the direction of the ship's rotation.
//
// Parameters:
//   - rotation: Ship rotation angle in radians
//   - thrustInput: Thrust input value (0.0 to 1.0)
//
// Returns:
//   - Acceleration vector in the direction of ship's rotation
func CalculateThrustAcceleration(rotation float64, thrustInput float32) entities.Vec2 {
	// Calculate direction vector from rotation angle
	// In standard math coordinates: x = cos(θ), y = sin(θ)
	directionX := math.Cos(rotation)
	directionY := math.Sin(rotation)

	// Scale by thrust input and acceleration constant
	magnitude := float64(thrustInput) * ThrustAcceleration

	return entities.NewVec2(directionX*magnitude, directionY*magnitude)
}

// ApplyInput applies input commands to the ship, updating rotation, velocity, and energy.
// Thrust is only applied when energy > 0.
//
// Parameters:
//   - ship: Current ship state
//   - input: Input command to apply
//   - dt: Time step in seconds
//
// Returns:
//   - Updated ship with new rotation, velocity, and energy
func ApplyInput(ship entities.Ship, input InputCommand, dt float64) entities.Ship {
	// Clamp input values
	clampedInput := ClampInput(input)

	// Update rotation (always works, regardless of energy)
	newRot := UpdateRotation(ship.Rot, float64(clampedInput.Turn), dt)

	// Calculate thrust acceleration
	thrustAcc := CalculateThrustAcceleration(newRot, clampedInput.Thrust)

	// Determine if thrust should be applied (only when energy > 0)
	shouldThrust := ship.Energy > MinEnergyForThrust && clampedInput.Thrust > 0.0

	// Update velocity (only if thrusting and energy available)
	newVel := ship.Vel
	if shouldThrust {
		// Apply thrust acceleration: v_new = v_old + a * dt
		newVel = newVel.Add(thrustAcc.Scale(dt))
	}

	// Update energy (drain if thrusting)
	isThrusting := shouldThrust
	newEnergy := DrainEnergyOnThrust(ship.Energy, isThrusting)

	// Return updated ship
	return entities.NewShip(
		ship.Pos, // Position is not updated by input processing (handled by physics step)
		newVel,
		newRot,
		newEnergy,
	)
}

