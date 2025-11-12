package entities

// Sun represents the sun (gravity source) in the game.
type Sun struct {
	Pos    Vec2    // Position
	Radius float32 // Radius
	Mass   float64 // Mass (for gravity calculations)
}

// NewSun creates a new Sun with the given values.
func NewSun(pos Vec2, radius float32, mass float64) Sun {
	return Sun{
		Pos:    pos,
		Radius: radius,
		Mass:   mass,
	}
}

// Pallet represents an energy pallet in the game.
type Pallet struct {
	ID     uint32 // Unique identifier
	Pos    Vec2   // Position
	Active bool   // Whether the pallet is active/collectible
}

// NewPallet creates a new Pallet with the given values.
func NewPallet(id uint32, pos Vec2, active bool) Pallet {
	return Pallet{
		ID:     id,
		Pos:    pos,
		Active: active,
	}
}

// World represents the complete game world state.
type World struct {
	Ship    Ship     // The player's ship
	Sun     Sun      // The sun (gravity source)
	Pallets []Pallet // List of energy pallets
	Tick    uint32   // Current simulation tick
	Done    bool     // Whether the game is finished
	Win     bool     // Whether the player won (only valid if Done is true)
}

// NewWorld creates a new World with the given values.
// If pallets is nil, it will be initialized as an empty slice.
func NewWorld(ship Ship, sun Sun, pallets []Pallet) World {
	if pallets == nil {
		pallets = []Pallet{}
	}
	return World{
		Ship:    ship,
		Sun:     sun,
		Pallets: pallets,
		Tick:    0,
		Done:    false,
		Win:     false,
	}
}
