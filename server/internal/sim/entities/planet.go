package entities

// Planet represents a planet (gravity source and collision obstacle) in the game.
// ID is the unique identifier for the planet in multiplayer matches.
// Multiple planets can exist per match (3–5 planets, static positions).
type Planet struct {
	ID     uint32  // Unique planet identifier
	Pos    Vec2    // Position in world coordinates (meters)
	Radius float32 // Collision radius (meters, typically 30–80 m)
	Mass   float64 // Mass for gravity calculations (game units, typically 500–2000)
}

// NewPlanet creates a new Planet with the given values.
func NewPlanet(id uint32, pos Vec2, radius float32, mass float64) Planet {
	return Planet{
		ID:     id,
		Pos:    pos,
		Radius: radius,
		Mass:   mass,
	}
}

