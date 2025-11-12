package physics

import "github.com/gorbit/orbitalrush/internal/sim/entities"

// ShipSunCollision checks if the ship is colliding with the sun.
// A collision occurs when the distance from the ship to the sun center
// is less than or equal to the sun's radius.
//
// Parameters:
//   - shipPos: Position of the ship
//   - sunPos: Position of the sun center
//   - sunRadius: Radius of the sun
//
// Returns:
//   - true if the ship is within or at the sun's radius, false otherwise
func ShipSunCollision(shipPos, sunPos entities.Vec2, sunRadius float32) bool {
	// Calculate direction vector from ship to sun
	direction := sunPos.Sub(shipPos)
	distanceSq := direction.LengthSq()
	radiusSq := float64(sunRadius) * float64(sunRadius)

	// Check if distance squared <= radius squared (avoiding square root)
	return distanceSq <= radiusSq
}

// ShipPalletCollision checks if the ship is colliding with a pallet.
// A collision occurs when the distance from the ship to the pallet center
// is less than or equal to the pickup radius.
//
// Parameters:
//   - shipPos: Position of the ship
//   - palletPos: Position of the pallet center
//   - pickupRadius: Pickup radius for pallets
//
// Returns:
//   - true if the ship is within or at the pickup radius, false otherwise
func ShipPalletCollision(shipPos, palletPos entities.Vec2, pickupRadius float64) bool {
	// Calculate direction vector from ship to pallet
	direction := palletPos.Sub(shipPos)
	distanceSq := direction.LengthSq()
	radiusSq := pickupRadius * pickupRadius

	// Check if distance squared <= radius squared (avoiding square root)
	return distanceSq <= radiusSq
}

