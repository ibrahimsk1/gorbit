package physics

import (
	"math"

	"github.com/gorbit/orbitalrush/internal/sim/entities"
)

// ApplyWraparound wraps a position to stay within world bounds using modulo-based formula.
// When a position exits one side of the world, it re-enters from the opposite side.
// World bounds: [-worldWidth/2, worldWidth/2] × [-worldHeight/2, worldHeight/2]
//
// Formula:
//   - X: pos.x = mod(pos.x + WORLD_WIDTH/2, WORLD_WIDTH) - WORLD_WIDTH/2
//   - Y: pos.y = mod(pos.y + WORLD_HEIGHT/2, WORLD_HEIGHT) - WORLD_HEIGHT/2
//
// Uses math.Mod for modulo operation. Handles negative modulo results by adding
// world size to bring the result back into bounds.
//
// Parameters:
//   - pos: Position to wrap
//   - worldWidth: World width in meters (typically 2000.0)
//   - worldHeight: World height in meters (typically 2000.0)
//
// Returns a wrapped position within world bounds.
func ApplyWraparound(pos entities.Vec2, worldWidth, worldHeight float64) entities.Vec2 {
	halfWidth := worldWidth / 2.0
	halfHeight := worldHeight / 2.0

	// Apply modulo-based wraparound formula
	wrappedX := math.Mod(pos.X+halfWidth, worldWidth) - halfWidth
	wrappedY := math.Mod(pos.Y+halfHeight, worldHeight) - halfHeight

	// Handle negative modulo results (math.Mod can return negative values when dividend is negative)
	// If result is less than -halfWidth/Height, add worldWidth/Height to bring it into bounds
	if wrappedX < -halfWidth {
		wrappedX += worldWidth
	}
	if wrappedY < -halfHeight {
		wrappedY += worldHeight
	}

	return entities.NewVec2(wrappedX, wrappedY)
}

