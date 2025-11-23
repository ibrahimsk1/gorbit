package entities

// World bounds constants
const (
	// WORLD_WIDTH is the width of the game world in meters.
	// World center is at origin (0, 0), bounds extend from -1000 to +1000 on X axis.
	WORLD_WIDTH = 2000.0

	// WORLD_HEIGHT is the height of the game world in meters.
	// World center is at origin (0, 0), bounds extend from -1000 to +1000 on Y axis.
	WORLD_HEIGHT = 2000.0
)

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

// World represents the complete game world state (multiplayer).
type World struct {
	Ships   []Ship   // All ships in the match (2–8 ships, one per player)
	Planets []Planet // All planets in the match (3–5 planets)
	Pallets []Pallet // All pallets in the match (8–12 pallets)
	Tick    uint32   // Current simulation tick
	Done    bool     // Whether match has ended (per-player or global)
	Win     bool     // Whether match ended in victory (per-player or global, only valid if Done is true)
}

// NewWorld creates a new World with the given values.
// If any array is nil, it will be initialized as an empty slice.
func NewWorld(ships []Ship, planets []Planet, pallets []Pallet) World {
	if ships == nil {
		ships = []Ship{}
	}
	if planets == nil {
		planets = []Planet{}
	}
	if pallets == nil {
		pallets = []Pallet{}
	}
	return World{
		Ships:   ships,
		Planets: planets,
		Pallets: pallets,
		Tick:    0,
		Done:    false,
		Win:     false,
	}
}
