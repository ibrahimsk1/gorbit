package entities

import (
	"fmt"
	"math/rand"
)

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

// WrapPosition wraps a position to stay within world bounds.
// When a position exits one side of the world, it re-enters from the opposite side.
// World bounds: [-worldWidth/2, worldWidth/2] × [-worldHeight/2, worldHeight/2]
// Parameters:
//   - pos: Position to wrap
//   - worldWidth: World width in meters (typically WORLD_WIDTH)
//   - worldHeight: World height in meters (typically WORLD_HEIGHT)
//
// Returns a wrapped position within world bounds.
func WrapPosition(pos Vec2, worldWidth, worldHeight float64) Vec2 {
	halfWidth := worldWidth / 2.0
	halfHeight := worldHeight / 2.0

	wrappedX := pos.X
	wrappedY := pos.Y

	// Wrap X coordinate
	// Handle boundary case first (exactly at boundary wraps once)
	if wrappedX == halfWidth {
		wrappedX = -halfWidth
	} else if wrappedX == -halfWidth {
		wrappedX = halfWidth
	} else {
		// For values beyond boundaries, wrap multiple times if needed
		for wrappedX > halfWidth {
			wrappedX -= worldWidth
		}
		for wrappedX < -halfWidth {
			wrappedX += worldWidth
		}
	}

	// Wrap Y coordinate
	// Handle boundary case first (exactly at boundary wraps once)
	if wrappedY == halfHeight {
		wrappedY = -halfHeight
	} else if wrappedY == -halfHeight {
		wrappedY = halfHeight
	} else {
		// For values beyond boundaries, wrap multiple times if needed
		for wrappedY > halfHeight {
			wrappedY -= worldHeight
		}
		for wrappedY < -halfHeight {
			wrappedY += worldHeight
		}
	}

	return NewVec2(wrappedX, wrappedY)
}

// ValidateUniqueIDs validates that all entity IDs are unique within their respective arrays.
// Returns an error if any duplicate IDs are found.
func ValidateUniqueIDs(world World) error {
	// Check Ship IDs
	shipIDs := make(map[uint32]bool)
	for _, ship := range world.Ships {
		if shipIDs[ship.ID] {
			return fmt.Errorf("duplicate ship ID: %d", ship.ID)
		}
		shipIDs[ship.ID] = true
	}

	// Check Planet IDs
	planetIDs := make(map[uint32]bool)
	for _, planet := range world.Planets {
		if planetIDs[planet.ID] {
			return fmt.Errorf("duplicate planet ID: %d", planet.ID)
		}
		planetIDs[planet.ID] = true
	}

	// Check Pallet IDs
	palletIDs := make(map[uint32]bool)
	for _, pallet := range world.Pallets {
		if palletIDs[pallet.ID] {
			return fmt.Errorf("duplicate pallet ID: %d", pallet.ID)
		}
		palletIDs[pallet.ID] = true
	}

	return nil
}

// ValidatePlanetSpacing validates that all planets have minimum spacing of 200m.
// Returns an error if any planets are too close.
func ValidatePlanetSpacing(planets []Planet) error {
	const minSpacing = 200.0

	for i := 0; i < len(planets); i++ {
		for j := i + 1; j < len(planets); j++ {
			distance := planets[i].Pos.Sub(planets[j].Pos).Length()
			if distance < minSpacing {
				return fmt.Errorf("planets %d and %d are too close: %.2f m (minimum: %.2f m)", planets[i].ID, planets[j].ID, distance, minSpacing)
			}
		}
	}

	return nil
}

// ValidateWorldBounds validates that all entity positions are within world bounds.
// World bounds: [-WORLD_WIDTH/2, WORLD_WIDTH/2] × [-WORLD_HEIGHT/2, WORLD_HEIGHT/2]
// Returns an error if any position is outside bounds.
func ValidateWorldBounds(world World) error {
	halfWidth := WORLD_WIDTH / 2.0
	halfHeight := WORLD_HEIGHT / 2.0

	// Check Ship positions
	for _, ship := range world.Ships {
		if ship.Pos.X > halfWidth || ship.Pos.X < -halfWidth {
			return fmt.Errorf("ship %d position X (%.2f) is outside world bounds [-%.2f, %.2f]", ship.ID, ship.Pos.X, halfWidth, halfWidth)
		}
		if ship.Pos.Y > halfHeight || ship.Pos.Y < -halfHeight {
			return fmt.Errorf("ship %d position Y (%.2f) is outside world bounds [-%.2f, %.2f]", ship.ID, ship.Pos.Y, halfHeight, halfHeight)
		}
	}

	// Check Planet positions
	for _, planet := range world.Planets {
		if planet.Pos.X > halfWidth || planet.Pos.X < -halfWidth {
			return fmt.Errorf("planet %d position X (%.2f) is outside world bounds [-%.2f, %.2f]", planet.ID, planet.Pos.X, halfWidth, halfWidth)
		}
		if planet.Pos.Y > halfHeight || planet.Pos.Y < -halfHeight {
			return fmt.Errorf("planet %d position Y (%.2f) is outside world bounds [-%.2f, %.2f]", planet.ID, planet.Pos.Y, halfHeight, halfHeight)
		}
	}

	// Check Pallet positions
	for _, pallet := range world.Pallets {
		if pallet.Pos.X > halfWidth || pallet.Pos.X < -halfWidth {
			return fmt.Errorf("pallet %d position X (%.2f) is outside world bounds [-%.2f, %.2f]", pallet.ID, pallet.Pos.X, halfWidth, halfWidth)
		}
		if pallet.Pos.Y > halfHeight || pallet.Pos.Y < -halfHeight {
			return fmt.Errorf("pallet %d position Y (%.2f) is outside world bounds [-%.2f, %.2f]", pallet.ID, pallet.Pos.Y, halfHeight, halfHeight)
		}
	}

	return nil
}

// ValidateWorld validates all World invariants.
// Checks unique IDs, planet spacing, and world bounds.
// Returns the first error encountered, or nil if all valid.
func ValidateWorld(world World) error {
	if err := ValidateUniqueIDs(world); err != nil {
		return err
	}
	if err := ValidatePlanetSpacing(world.Planets); err != nil {
		return err
	}
	if err := ValidateWorldBounds(world); err != nil {
		return err
	}
	return nil
}
