package physics

import (
	"math"
	"testing"

	"github.com/gorbit/orbitalrush/internal/sim/entities"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWraparound(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Wraparound Suite")
}

var _ = Describe("ApplyWraparound", Label("scope:unit", "loop:g2-physics", "layer:physics", "dep:none", "b:wraparound", "r:medium"), func() {
	const epsilon = 1e-9
	const worldWidth = 2000.0
	const worldHeight = 2000.0
	const halfWidth = worldWidth / 2.0
	const halfHeight = worldHeight / 2.0

	Describe("Basic wraparound", func() {
		It("wraps position from right edge to left edge", func() {
			// Position beyond right edge (X > WORLD_WIDTH/2)
			pos := entities.NewVec2(1100.0, 0.0)
			wrapped := ApplyWraparound(pos, worldWidth, worldHeight)

			// Expected: mod(1100 + 1000, 2000) - 1000 = mod(2100, 2000) - 1000 = 100 - 1000 = -900
			Expect(wrapped.X).To(BeNumerically("~", -900.0, epsilon*10))
			Expect(wrapped.Y).To(BeNumerically("~", 0.0, epsilon))
		})

		It("wraps position from left edge to right edge", func() {
			// Position beyond left edge (X < -WORLD_WIDTH/2)
			pos := entities.NewVec2(-1100.0, 0.0)
			wrapped := ApplyWraparound(pos, worldWidth, worldHeight)

			// Expected: mod(-1100 + 1000, 2000) - 1000 = mod(-100, 2000) - 1000 = 1900 - 1000 = 900
			Expect(wrapped.X).To(BeNumerically("~", 900.0, epsilon*10))
			Expect(wrapped.Y).To(BeNumerically("~", 0.0, epsilon))
		})

		It("wraps position from top edge to bottom edge", func() {
			// Position beyond top edge (Y > WORLD_HEIGHT/2)
			pos := entities.NewVec2(0.0, 1100.0)
			wrapped := ApplyWraparound(pos, worldWidth, worldHeight)

			// Expected: mod(1100 + 1000, 2000) - 1000 = mod(2100, 2000) - 1000 = 100 - 1000 = -900
			Expect(wrapped.X).To(BeNumerically("~", 0.0, epsilon))
			Expect(wrapped.Y).To(BeNumerically("~", -900.0, epsilon*10))
		})

		It("wraps position from bottom edge to top edge", func() {
			// Position beyond bottom edge (Y < -WORLD_HEIGHT/2)
			pos := entities.NewVec2(0.0, -1100.0)
			wrapped := ApplyWraparound(pos, worldWidth, worldHeight)

			// Expected: mod(-1100 + 1000, 2000) - 1000 = mod(-100, 2000) - 1000 = 1900 - 1000 = 900
			Expect(wrapped.X).To(BeNumerically("~", 0.0, epsilon))
			Expect(wrapped.Y).To(BeNumerically("~", 900.0, epsilon*10))
		})
	})

	Describe("Corner cases", func() {
		It("wraps position when both X and Y exceed bounds", func() {
			// Position beyond top-right corner
			pos := entities.NewVec2(1100.0, 1100.0)
			wrapped := ApplyWraparound(pos, worldWidth, worldHeight)

			Expect(wrapped.X).To(BeNumerically("~", -900.0, epsilon*10))
			Expect(wrapped.Y).To(BeNumerically("~", -900.0, epsilon*10))
		})

		It("wraps position when both X and Y are below bounds", func() {
			// Position beyond bottom-left corner
			pos := entities.NewVec2(-1100.0, -1100.0)
			wrapped := ApplyWraparound(pos, worldWidth, worldHeight)

			Expect(wrapped.X).To(BeNumerically("~", 900.0, epsilon*10))
			Expect(wrapped.Y).To(BeNumerically("~", 900.0, epsilon*10))
		})
	})

	Describe("Boundary conditions", func() {
		It("handles positions exactly at right boundary", func() {
			// Position exactly at right boundary
			pos := entities.NewVec2(halfWidth, 0.0)
			wrapped := ApplyWraparound(pos, worldWidth, worldHeight)

			// Expected: mod(1000 + 1000, 2000) - 1000 = mod(2000, 2000) - 1000 = 0 - 1000 = -1000
			Expect(wrapped.X).To(BeNumerically("~", -halfWidth, epsilon*10))
			Expect(wrapped.Y).To(BeNumerically("~", 0.0, epsilon))
		})

		It("handles positions exactly at left boundary", func() {
			// Position exactly at left boundary
			pos := entities.NewVec2(-halfWidth, 0.0)
			wrapped := ApplyWraparound(pos, worldWidth, worldHeight)

			// Expected: mod(-1000 + 1000, 2000) - 1000 = mod(0, 2000) - 1000 = 0 - 1000 = -1000
			// But we need to handle negative modulo: -1000 + 2000 = 1000
			Expect(wrapped.X).To(BeNumerically("~", -halfWidth, epsilon*10))
			Expect(wrapped.Y).To(BeNumerically("~", 0.0, epsilon))
		})

		It("handles positions exactly at top boundary", func() {
			// Position exactly at top boundary
			pos := entities.NewVec2(0.0, halfHeight)
			wrapped := ApplyWraparound(pos, worldWidth, worldHeight)

			Expect(wrapped.X).To(BeNumerically("~", 0.0, epsilon))
			Expect(wrapped.Y).To(BeNumerically("~", -halfHeight, epsilon*10))
		})

		It("handles positions exactly at bottom boundary", func() {
			// Position exactly at bottom boundary
			pos := entities.NewVec2(0.0, -halfHeight)
			wrapped := ApplyWraparound(pos, worldWidth, worldHeight)

			Expect(wrapped.X).To(BeNumerically("~", 0.0, epsilon))
			Expect(wrapped.Y).To(BeNumerically("~", -halfHeight, epsilon*10))
		})
	})

	Describe("Within bounds", func() {
		It("does not modify positions within bounds", func() {
			// Position within bounds
			pos := entities.NewVec2(500.0, -300.0)
			wrapped := ApplyWraparound(pos, worldWidth, worldHeight)

			Expect(wrapped.X).To(BeNumerically("~", 500.0, epsilon*10))
			Expect(wrapped.Y).To(BeNumerically("~", -300.0, epsilon*10))
		})

		It("does not modify positions at origin", func() {
			pos := entities.NewVec2(0.0, 0.0)
			wrapped := ApplyWraparound(pos, worldWidth, worldHeight)

			Expect(wrapped.X).To(BeNumerically("~", 0.0, epsilon))
			Expect(wrapped.Y).To(BeNumerically("~", 0.0, epsilon))
		})

		It("does not modify positions just inside bounds", func() {
			// Position just inside right boundary
			pos := entities.NewVec2(halfWidth-epsilon, 0.0)
			wrapped := ApplyWraparound(pos, worldWidth, worldHeight)

			Expect(wrapped.X).To(BeNumerically("~", halfWidth-epsilon, epsilon*10))
			Expect(wrapped.Y).To(BeNumerically("~", 0.0, epsilon))
		})
	})

	Describe("Multiple wraps", func() {
		It("handles positions far outside bounds (multiple world widths)", func() {
			// Position far beyond right edge (multiple world widths)
			pos := entities.NewVec2(3500.0, 0.0)
			wrapped := ApplyWraparound(pos, worldWidth, worldHeight)

			// Expected: mod(3500 + 1000, 2000) - 1000 = mod(4500, 2000) - 1000 = 500 - 1000 = -500
			Expect(wrapped.X).To(BeNumerically("~", -500.0, epsilon*10))
			Expect(wrapped.Y).To(BeNumerically("~", 0.0, epsilon))
		})

		It("handles positions far outside bounds (multiple world heights)", func() {
			// Position far beyond top edge (multiple world heights)
			pos := entities.NewVec2(0.0, 3500.0)
			wrapped := ApplyWraparound(pos, worldWidth, worldHeight)

			// Expected: mod(3500 + 1000, 2000) - 1000 = mod(4500, 2000) - 1000 = 500 - 1000 = -500
			Expect(wrapped.X).To(BeNumerically("~", 0.0, epsilon))
			Expect(wrapped.Y).To(BeNumerically("~", -500.0, epsilon*10))
		})

		It("handles positions far outside bounds (negative, multiple world widths)", func() {
			// Position far beyond left edge (negative, multiple world widths)
			pos := entities.NewVec2(-3500.0, 0.0)
			wrapped := ApplyWraparound(pos, worldWidth, worldHeight)

			// Expected: mod(-3500 + 1000, 2000) - 1000 = mod(-2500, 2000) - 1000 = -500 - 1000 = -1500
			// But math.Mod(-2500, 2000) = -500, so we need to add 2000: -500 + 2000 = 1500, then 1500 - 1000 = 500
			Expect(wrapped.X).To(BeNumerically("~", 500.0, epsilon*10))
			Expect(wrapped.Y).To(BeNumerically("~", 0.0, epsilon))
		})
	})

	Describe("Determinism", func() {
		It("produces identical results for identical inputs across multiple runs", func() {
			pos := entities.NewVec2(1100.0, -1100.0)

			// Run wraparound multiple times
			var firstWrapped entities.Vec2
			for i := 0; i < 100; i++ {
				wrapped := ApplyWraparound(pos, worldWidth, worldHeight)
				if i == 0 {
					firstWrapped = wrapped
				} else {
					// Verify bit-exact results
					Expect(wrapped.X).To(Equal(firstWrapped.X))
					Expect(wrapped.Y).To(Equal(firstWrapped.Y))
				}
			}
		})

		It("produces identical results when called with same inputs in different order", func() {
			pos := entities.NewVec2(500.0, -300.0)

			result1 := ApplyWraparound(pos, worldWidth, worldHeight)
			result2 := ApplyWraparound(pos, worldWidth, worldHeight)

			Expect(result1.X).To(Equal(result2.X))
			Expect(result1.Y).To(Equal(result2.Y))
		})
	})

	Describe("Edge cases", func() {
		It("handles zero world size", func() {
			pos := entities.NewVec2(100.0, 200.0)
			wrapped := ApplyWraparound(pos, 0.0, 0.0)

			// With zero world size, modulo behavior is undefined, but function should not panic
			// Result may be implementation-dependent
			Expect(math.IsNaN(wrapped.X)).To(BeFalse())
			Expect(math.IsNaN(wrapped.Y)).To(BeFalse())
			Expect(math.IsInf(wrapped.X, 0)).To(BeFalse())
			Expect(math.IsInf(wrapped.Y, 0)).To(BeFalse())
		})

		It("handles very small world size", func() {
			pos := entities.NewVec2(10.0, 20.0)
			smallWidth := 1.0
			smallHeight := 1.0
			wrapped := ApplyWraparound(pos, smallWidth, smallHeight)

			// Should still produce valid results
			Expect(math.IsNaN(wrapped.X)).To(BeFalse())
			Expect(math.IsNaN(wrapped.Y)).To(BeFalse())
			Expect(math.IsInf(wrapped.X, 0)).To(BeFalse())
			Expect(math.IsInf(wrapped.Y, 0)).To(BeFalse())
		})

		It("handles negative positions within bounds", func() {
			pos := entities.NewVec2(-500.0, -300.0)
			wrapped := ApplyWraparound(pos, worldWidth, worldHeight)

			// Negative positions within bounds should remain unchanged
			Expect(wrapped.X).To(BeNumerically("~", -500.0, epsilon*10))
			Expect(wrapped.Y).To(BeNumerically("~", -300.0, epsilon*10))
		})
	})
})

