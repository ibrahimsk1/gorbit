package entities

// Ship represents the player's ship in the game.
// ID is the unique identifier for player identification in multiplayer matches.
type Ship struct {
	ID     uint32  // Unique ship identifier (player ID)
	Pos    Vec2    // Position
	Vel    Vec2    // Velocity
	Rot    float64 // Rotation angle in radians
	Energy float32 // Current energy level
}

// NewShip creates a new Ship with the given values.
func NewShip(id uint32, pos, vel Vec2, rot float64, energy float32) Ship {
	return Ship{
		ID:     id,
		Pos:    pos,
		Vel:    vel,
		Rot:    rot,
		Energy: energy,
	}
}
