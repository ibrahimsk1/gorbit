package entities

import "math/rand"

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

// GeneratePlanets generates multiple planets with varied sizes, masses, and positions.
// Planets are distributed across the world with minimum spacing (200m) and no overlaps.
// Parameters:
//   - count: Number of planets to generate (3–5 per match)
//   - worldWidth: World width in meters (typically WORLD_WIDTH)
//   - worldHeight: World height in meters (typically WORLD_HEIGHT)
//
// Returns a slice of planets with:
//   - Radius in [30, 80] m range
//   - Mass in [500, 2000] game units range
//   - Positions within world bounds
//   - Minimum spacing of 200m between planets
//   - No overlapping planets (distance >= sum of radii)
//   - Unique IDs starting from 1
func GeneratePlanets(count int, worldWidth, worldHeight float64) []Planet {
	if count <= 0 {
		return []Planet{}
	}

	planets := make([]Planet, 0, count)
	const minSpacing = 200.0
	const maxRetries = 50

	for i := 0; i < count; i++ {
		id := uint32(i + 1)
		var planet Planet
		valid := false

		for retry := 0; retry < maxRetries; retry++ {
			// Generate random radius: [30, 80] m
			radius := 30.0 + rand.Float64()*50.0

			// Generate random mass: [500, 2000] game units
			mass := 500.0 + rand.Float64()*1500.0

			// Generate random position within world bounds
			posX := (rand.Float64() - 0.5) * worldWidth
			posY := (rand.Float64() - 0.5) * worldHeight
			pos := NewVec2(posX, posY)

			// Check if position meets spacing and overlap constraints
			meetsConstraints := true
			for _, existing := range planets {
				distance := pos.Sub(existing.Pos).Length()
				sumRadii := float64(radius) + float64(existing.Radius)

				// Check minimum spacing
				if distance < minSpacing {
					meetsConstraints = false
					break
				}

				// Check no overlap
				if distance < sumRadii {
					meetsConstraints = false
					break
				}
			}

			if meetsConstraints {
				planet = NewPlanet(id, pos, float32(radius), mass)
				valid = true
				break
			}
		}

		// If we couldn't find a valid position after max retries, use the last generated planet
		// This should rarely happen with reasonable world size and planet count
		if !valid {
			radius := 30.0 + rand.Float64()*50.0
			mass := 500.0 + rand.Float64()*1500.0
			posX := (rand.Float64() - 0.5) * worldWidth
			posY := (rand.Float64() - 0.5) * worldHeight
			pos := NewVec2(posX, posY)
			planet = NewPlanet(id, pos, float32(radius), mass)
		}

		planets = append(planets, planet)
	}

	return planets
}
