package physics

import "github.com/gorbit/orbitalrush/internal/sim/entities"

// GravityAcceleration calculates the acceleration due to gravity using the inverse-square law.
// The acceleration is clamped to a maximum value to prevent extreme accelerations at close distances.
//
// Formula: a = G * M / r² * direction
// Where:
//   - G is the gravitational constant
//   - M is the mass of the sun
//   - r is the distance from ship to sun
//   - direction is the normalized vector from ship to sun
//
// The acceleration magnitude is clamped to aMax: |a|_clamped = min(|a|, aMax)
//
// Parameters:
//   - shipPos: Position of the ship
//   - sunPos: Position of the sun (gravity source)
//   - sunMass: Mass of the sun
//   - G: Gravitational constant (game-scale, typically 1.0)
//   - aMax: Maximum acceleration magnitude
//
// Returns:
//   - acc: Acceleration vector pointing toward the sun
func GravityAcceleration(shipPos, sunPos entities.Vec2, sunMass, G, aMax float64) entities.Vec2 {
	// Handle zero mass (early return - no calculations needed)
	if sunMass == 0 {
		return entities.Zero()
	}

	// Calculate direction vector from ship to sun
	direction := sunPos.Sub(shipPos)
	distanceSq := direction.LengthSq()

	// Handle zero distance (ship exactly at sun position)
	if distanceSq == 0 {
		return entities.Zero()
	}

	// Calculate acceleration magnitude using inverse-square law: |a| = G * M / r²
	accMagnitude := G * sunMass / distanceSq

	// Clamp acceleration magnitude to aMax
	if accMagnitude > aMax {
		accMagnitude = aMax
	}

	// Normalize direction and scale by acceleration magnitude
	directionNormalized := direction.Normalize()
	return directionNormalized.Scale(accMagnitude)
}
