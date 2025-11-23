package physics

import "github.com/gorbit/orbitalrush/internal/sim/entities"

// ShipPlanetCollision checks if the ship is colliding with a planet.
// A collision occurs when the distance from the ship to the planet center
// is less than or equal to the planet's radius.
//
// Parameters:
//   - shipPos: Position of the ship
//   - planet: Planet entity (contains Pos and Radius)
//
// Returns:
//   - true if the ship is within or at the planet's radius, false otherwise
func ShipPlanetCollision(shipPos entities.Vec2, planet entities.Planet) bool {
	// Calculate direction vector from ship to planet
	direction := planet.Pos.Sub(shipPos)
	distanceSq := direction.LengthSq()
	radiusSq := float64(planet.Radius) * float64(planet.Radius)

	// Check if distance squared <= radius squared (avoiding square root)
	return distanceSq <= radiusSq
}

// CheckShipPlanetCollisions checks if the ship is colliding with any planet
// in the planets array. Returns collision status and the ID of the first
// colliding planet.
//
// Parameters:
//   - shipPos: Position of the ship
//   - planets: Array of planets to check against
//
// Returns:
//   - colliding: true if ship collides with any planet, false otherwise
//   - planetID: ID of the first colliding planet (0 if no collision)
func CheckShipPlanetCollisions(shipPos entities.Vec2, planets []entities.Planet) (bool, uint32) {
	for _, planet := range planets {
		if ShipPlanetCollision(shipPos, planet) {
			return true, planet.ID
		}
	}
	return false, 0
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

