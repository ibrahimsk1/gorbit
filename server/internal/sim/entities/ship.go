package entities

// Ship represents the player's ship in the game.
type Ship struct {
	Pos    Vec2    // Position
	Vel    Vec2    // Velocity
	Rot    float64 // Rotation angle in radians
	Energy float32 // Current energy level
}

// NewShip creates a new Ship with the given values.
func NewShip(pos, vel Vec2, rot float64, energy float32) Ship {
	return Ship{
		Pos:    pos,
		Vel:    vel,
		Rot:    rot,
		Energy: energy,
	}
}
