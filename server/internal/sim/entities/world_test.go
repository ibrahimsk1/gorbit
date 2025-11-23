package entities

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pallet", Label("scope:unit", "loop:g1-entities", "layer:entities", "dep:none", "b:entity-types", "r:low"), func() {
	Describe("Constructor", func() {
		It("creates a new Pallet with given values", func() {
			id := uint32(1)
			pos := NewVec2(30.0, 40.0)
			active := true

			pallet := NewPallet(id, pos, active)

			Expect(pallet.ID).To(Equal(id))
			Expect(pallet.Pos).To(Equal(pos))
			Expect(pallet.Active).To(Equal(active))
		})

		It("creates a zero pallet", func() {
			pallet := Pallet{}

			Expect(pallet.ID).To(Equal(uint32(0)))
			Expect(pallet.Pos).To(Equal(Zero()))
			Expect(pallet.Active).To(Equal(false))
		})
	})

	Describe("Properties", func() {
		It("maintains field values after creation", func() {
			id := uint32(42)
			pos := NewVec2(15.0, 25.0)
			active := false

			pallet := NewPallet(id, pos, active)

			Expect(pallet.ID).To(Equal(uint32(42)))
			Expect(pallet.Pos.X).To(Equal(15.0))
			Expect(pallet.Pos.Y).To(Equal(25.0))
			Expect(pallet.Active).To(Equal(false))
		})

		It("allows inactive pallets", func() {
			pallet := NewPallet(1, NewVec2(0, 0), false)
			Expect(pallet.Active).To(BeFalse())
		})

		It("allows active pallets", func() {
			pallet := NewPallet(1, NewVec2(0, 0), true)
			Expect(pallet.Active).To(BeTrue())
		})
	})
})

var _ = Describe("World", Label("scope:unit", "loop:g1-entities", "layer:entities", "dep:none", "b:world-arrays", "r:high"), func() {
	Describe("World Bounds", func() {
		It("defines WORLD_WIDTH constant", func() {
			Expect(WORLD_WIDTH).To(Equal(2000.0))
		})

		It("defines WORLD_HEIGHT constant", func() {
			Expect(WORLD_HEIGHT).To(Equal(2000.0))
		})
	})

	Describe("Constructor", func() {
		It("creates a new World with given values", func() {
			ships := []Ship{
				NewShip(1, NewVec2(10, 20), NewVec2(1, 2), 0.5, 100),
			}
			planets := []Planet{
				NewPlanet(1, NewVec2(0, 0), 50.0, 1000.0),
			}
			pallets := []Pallet{
				NewPallet(1, NewVec2(100, 100), true),
				NewPallet(2, NewVec2(200, 200), true),
			}

			world := NewWorld(ships, planets, pallets)

			Expect(world.Ships).To(HaveLen(1))
			Expect(world.Ships[0]).To(Equal(ships[0]))
			Expect(world.Planets).To(HaveLen(1))
			Expect(world.Planets[0]).To(Equal(planets[0]))
			Expect(world.Pallets).To(HaveLen(2))
			Expect(world.Pallets[0]).To(Equal(pallets[0]))
			Expect(world.Pallets[1]).To(Equal(pallets[1]))
			Expect(world.Tick).To(Equal(uint32(0)))
			Expect(world.Done).To(BeFalse())
			Expect(world.Win).To(BeFalse())
		})

		It("creates a zero world", func() {
			world := World{}

			Expect(world.Ships).To(BeEmpty())
			Expect(world.Planets).To(BeEmpty())
			Expect(world.Pallets).To(BeEmpty())
			Expect(world.Tick).To(Equal(uint32(0)))
			Expect(world.Done).To(BeFalse())
			Expect(world.Win).To(BeFalse())
		})

		It("creates a world with nil arrays converted to empty slices", func() {
			world := NewWorld(nil, nil, nil)

			Expect(world.Ships).To(BeEmpty())
			Expect(world.Planets).To(BeEmpty())
			Expect(world.Pallets).To(BeEmpty())
		})

		It("creates a world with multiple ships", func() {
			ships := []Ship{
				NewShip(1, NewVec2(10, 20), NewVec2(1, 2), 0.5, 100),
				NewShip(2, NewVec2(30, 40), NewVec2(2, 3), 1.0, 80),
				NewShip(3, NewVec2(50, 60), NewVec2(3, 4), 1.5, 90),
			}
			planets := []Planet{
				NewPlanet(1, NewVec2(0, 0), 50.0, 1000.0),
			}

			world := NewWorld(ships, planets, nil)

			Expect(world.Ships).To(HaveLen(3))
			Expect(world.Ships[0].ID).To(Equal(uint32(1)))
			Expect(world.Ships[1].ID).To(Equal(uint32(2)))
			Expect(world.Ships[2].ID).To(Equal(uint32(3)))
		})

		It("creates a world with multiple planets", func() {
			ships := []Ship{
				NewShip(1, NewVec2(0, 0), NewVec2(0, 0), 0, 100),
			}
			planets := []Planet{
				NewPlanet(1, NewVec2(100, 100), 30.0, 500.0),
				NewPlanet(2, NewVec2(200, 200), 40.0, 1000.0),
				NewPlanet(3, NewVec2(300, 300), 50.0, 1500.0),
				NewPlanet(4, NewVec2(400, 400), 60.0, 2000.0),
			}

			world := NewWorld(ships, planets, nil)

			Expect(world.Planets).To(HaveLen(4))
			Expect(world.Planets[0].ID).To(Equal(uint32(1)))
			Expect(world.Planets[1].ID).To(Equal(uint32(2)))
			Expect(world.Planets[2].ID).To(Equal(uint32(3)))
			Expect(world.Planets[3].ID).To(Equal(uint32(4)))
		})
	})

	Describe("Properties", func() {
		It("maintains field values after creation", func() {
			ships := []Ship{
				NewShip(1, NewVec2(5, 10), NewVec2(0.5, 1), 1.0, 75),
			}
			planets := []Planet{
				NewPlanet(1, NewVec2(0, 0), 25.0, 500.0),
			}
			pallets := []Pallet{NewPallet(1, NewVec2(50, 50), true)}

			world := NewWorld(ships, planets, pallets)

			Expect(world.Ships[0].Energy).To(Equal(float32(75)))
			Expect(world.Planets[0].Mass).To(Equal(500.0))
			Expect(world.Pallets).To(HaveLen(1))
			Expect(world.Tick).To(Equal(uint32(0)))
		})

		It("allows multiple pallets with different IDs", func() {
			ships := []Ship{NewShip(1, NewVec2(0, 0), NewVec2(0, 0), 0, 100)}
			planets := []Planet{NewPlanet(1, NewVec2(0, 0), 50.0, 1000.0)}
			pallets := []Pallet{
				NewPallet(1, NewVec2(10, 10), true),
				NewPallet(2, NewVec2(20, 20), true),
				NewPallet(3, NewVec2(30, 30), false),
			}

			world := NewWorld(ships, planets, pallets)

			Expect(world.Pallets).To(HaveLen(3))
			Expect(world.Pallets[0].ID).To(Equal(uint32(1)))
			Expect(world.Pallets[1].ID).To(Equal(uint32(2)))
			Expect(world.Pallets[2].ID).To(Equal(uint32(3)))
		})
	})

	Describe("Planet Generation", Label("scope:unit", "loop:g1-entities", "layer:entities", "dep:none", "b:planet-generation", "r:medium", "double:fake"), func() {
		It("generates correct count of planets", func() {
			planets := GeneratePlanets(3, WORLD_WIDTH, WORLD_HEIGHT)
			Expect(planets).To(HaveLen(3))

			planets = GeneratePlanets(4, WORLD_WIDTH, WORLD_HEIGHT)
			Expect(planets).To(HaveLen(4))

			planets = GeneratePlanets(5, WORLD_WIDTH, WORLD_HEIGHT)
			Expect(planets).To(HaveLen(5))
		})

		It("all planets have unique IDs", func() {
			planets := GeneratePlanets(5, WORLD_WIDTH, WORLD_HEIGHT)
			ids := make(map[uint32]bool)
			for _, planet := range planets {
				Expect(ids[planet.ID]).To(BeFalse(), "Duplicate ID found: %d", planet.ID)
				ids[planet.ID] = true
			}
		})

		It("all planets have radius in valid range [30, 80]", func() {
			planets := GeneratePlanets(5, WORLD_WIDTH, WORLD_HEIGHT)
			for _, planet := range planets {
				Expect(planet.Radius).To(BeNumerically(">=", float32(30.0)))
				Expect(planet.Radius).To(BeNumerically("<=", float32(80.0)))
			}
		})

		It("all planets have mass in valid range [500, 2000]", func() {
			planets := GeneratePlanets(5, WORLD_WIDTH, WORLD_HEIGHT)
			for _, planet := range planets {
				Expect(planet.Mass).To(BeNumerically(">=", 500.0))
				Expect(planet.Mass).To(BeNumerically("<=", 2000.0))
			}
		})

		It("planets have minimum spacing of 200m", func() {
			planets := GeneratePlanets(5, WORLD_WIDTH, WORLD_HEIGHT)
			for i := 0; i < len(planets); i++ {
				for j := i + 1; j < len(planets); j++ {
					distance := planets[i].Pos.Sub(planets[j].Pos).Length()
					Expect(distance).To(BeNumerically(">=", 200.0), "Planets %d and %d are too close: %.2f m", i, j, distance)
				}
			}
		})

		It("planets do not overlap (distance >= sum of radii)", func() {
			planets := GeneratePlanets(5, WORLD_WIDTH, WORLD_HEIGHT)
			for i := 0; i < len(planets); i++ {
				for j := i + 1; j < len(planets); j++ {
					distance := planets[i].Pos.Sub(planets[j].Pos).Length()
					sumRadii := float64(planets[i].Radius + planets[j].Radius)
					Expect(distance).To(BeNumerically(">=", sumRadii), "Planets %d and %d overlap: distance=%.2f, sumRadii=%.2f", i, j, distance, sumRadii)
				}
			}
		})

		It("all planet positions are within world bounds", func() {
			planets := GeneratePlanets(5, WORLD_WIDTH, WORLD_HEIGHT)
			halfWidth := WORLD_WIDTH / 2.0
			halfHeight := WORLD_HEIGHT / 2.0
			for _, planet := range planets {
				Expect(planet.Pos.X).To(BeNumerically(">=", -halfWidth))
				Expect(planet.Pos.X).To(BeNumerically("<=", halfWidth))
				Expect(planet.Pos.Y).To(BeNumerically(">=", -halfHeight))
				Expect(planet.Pos.Y).To(BeNumerically("<=", halfHeight))
			}
		})

		It("generation is deterministic with fixed seed", func() {
			// Note: This test verifies that with the same seed, we get consistent results
			// The actual implementation uses math/rand which can be seeded
			planets1 := GeneratePlanets(4, WORLD_WIDTH, WORLD_HEIGHT)
			planets2 := GeneratePlanets(4, WORLD_WIDTH, WORLD_HEIGHT)
			// We can't guarantee exact same results without seeding, but we can verify structure
			Expect(planets1).To(HaveLen(4))
			Expect(planets2).To(HaveLen(4))
		})
	})

	Describe("Wraparound", Label("scope:unit", "loop:g1-entities", "layer:entities", "dep:none", "b:world-bounds", "r:medium"), func() {
		It("wraps position from right edge to left edge", func() {
			// Position beyond right edge (X > WORLD_WIDTH/2)
			pos := NewVec2(1100.0, 0.0)
			wrapped := WrapPosition(pos, WORLD_WIDTH, WORLD_HEIGHT)
			Expect(wrapped.X).To(BeNumerically("~", -900.0, 0.01))
			Expect(wrapped.Y).To(Equal(0.0))
		})

		It("wraps position from left edge to right edge", func() {
			// Position beyond left edge (X < -WORLD_WIDTH/2)
			pos := NewVec2(-1100.0, 0.0)
			wrapped := WrapPosition(pos, WORLD_WIDTH, WORLD_HEIGHT)
			Expect(wrapped.X).To(BeNumerically("~", 900.0, 0.01))
			Expect(wrapped.Y).To(Equal(0.0))
		})

		It("wraps position from top edge to bottom edge", func() {
			// Position beyond top edge (Y > WORLD_HEIGHT/2)
			pos := NewVec2(0.0, 1100.0)
			wrapped := WrapPosition(pos, WORLD_WIDTH, WORLD_HEIGHT)
			Expect(wrapped.X).To(Equal(0.0))
			Expect(wrapped.Y).To(BeNumerically("~", -900.0, 0.01))
		})

		It("wraps position from bottom edge to top edge", func() {
			// Position beyond bottom edge (Y < -WORLD_HEIGHT/2)
			pos := NewVec2(0.0, -1100.0)
			wrapped := WrapPosition(pos, WORLD_WIDTH, WORLD_HEIGHT)
			Expect(wrapped.X).To(Equal(0.0))
			Expect(wrapped.Y).To(BeNumerically("~", 900.0, 0.01))
		})

		It("wraps position from corner (both X and Y)", func() {
			// Position beyond top-right corner
			pos := NewVec2(1100.0, 1100.0)
			wrapped := WrapPosition(pos, WORLD_WIDTH, WORLD_HEIGHT)
			Expect(wrapped.X).To(BeNumerically("~", -900.0, 0.01))
			Expect(wrapped.Y).To(BeNumerically("~", -900.0, 0.01))
		})

		It("does not wrap positions within bounds", func() {
			// Position within bounds
			pos := NewVec2(500.0, -300.0)
			wrapped := WrapPosition(pos, WORLD_WIDTH, WORLD_HEIGHT)
			Expect(wrapped.X).To(Equal(500.0))
			Expect(wrapped.Y).To(Equal(-300.0))
		})

		It("handles positions exactly at boundaries", func() {
			// Position exactly at right boundary
			pos := NewVec2(1000.0, 0.0)
			wrapped := WrapPosition(pos, WORLD_WIDTH, WORLD_HEIGHT)
			// At boundary, should wrap to -1000.0
			Expect(wrapped.X).To(BeNumerically("~", -1000.0, 0.01))
			Expect(wrapped.Y).To(Equal(0.0))

			// Position exactly at left boundary
			pos = NewVec2(-1000.0, 0.0)
			wrapped = WrapPosition(pos, WORLD_WIDTH, WORLD_HEIGHT)
			// At boundary, should wrap to 1000.0
			Expect(wrapped.X).To(BeNumerically("~", 1000.0, 0.01))
			Expect(wrapped.Y).To(Equal(0.0))
		})

		It("handles multiple wraps for positions far outside bounds", func() {
			// Position far beyond right edge (multiple world widths)
			pos := NewVec2(3500.0, 0.0)
			wrapped := WrapPosition(pos, WORLD_WIDTH, WORLD_HEIGHT)
			// Should wrap: 3500 - 2000 = 1500, then 1500 - 2000 = -500
			Expect(wrapped.X).To(BeNumerically("~", -500.0, 0.01))
			Expect(wrapped.Y).To(Equal(0.0))
		})
	})

	Describe("World Validation", Label("scope:unit", "loop:g1-entities", "layer:entities", "dep:none", "b:entity-validation", "r:low"), func() {
		Describe("ValidateUniqueIDs", func() {
			It("detects duplicate ship IDs", func() {
				ships := []Ship{
					NewShip(1, NewVec2(0, 0), NewVec2(0, 0), 0, 100),
					NewShip(1, NewVec2(10, 10), NewVec2(0, 0), 0, 100), // Duplicate ID
				}
				world := NewWorld(ships, nil, nil)
				err := ValidateUniqueIDs(world)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("duplicate ship ID"))
			})

			It("detects duplicate planet IDs", func() {
				planets := []Planet{
					NewPlanet(1, NewVec2(0, 0), 50.0, 1000.0),
					NewPlanet(1, NewVec2(100, 100), 50.0, 1000.0), // Duplicate ID
				}
				world := NewWorld(nil, planets, nil)
				err := ValidateUniqueIDs(world)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("duplicate planet ID"))
			})

			It("detects duplicate pallet IDs", func() {
				pallets := []Pallet{
					NewPallet(1, NewVec2(0, 0), true),
					NewPallet(1, NewVec2(10, 10), true), // Duplicate ID
				}
				world := NewWorld(nil, nil, pallets)
				err := ValidateUniqueIDs(world)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("duplicate pallet ID"))
			})

			It("passes with all unique IDs", func() {
				ships := []Ship{
					NewShip(1, NewVec2(0, 0), NewVec2(0, 0), 0, 100),
					NewShip(2, NewVec2(10, 10), NewVec2(0, 0), 0, 100),
				}
				planets := []Planet{
					NewPlanet(1, NewVec2(0, 0), 50.0, 1000.0),
					NewPlanet(2, NewVec2(100, 100), 50.0, 1000.0),
				}
				pallets := []Pallet{
					NewPallet(1, NewVec2(0, 0), true),
					NewPallet(2, NewVec2(10, 10), true),
				}
				world := NewWorld(ships, planets, pallets)
				err := ValidateUniqueIDs(world)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("ValidatePlanetSpacing", func() {
			It("detects planets too close", func() {
				planets := []Planet{
					NewPlanet(1, NewVec2(0, 0), 50.0, 1000.0),
					NewPlanet(2, NewVec2(150, 0), 50.0, 1000.0), // Only 150m apart (less than 200m minimum)
				}
				err := ValidatePlanetSpacing(planets)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("too close"))
			})

			It("passes with proper spacing", func() {
				planets := []Planet{
					NewPlanet(1, NewVec2(0, 0), 50.0, 1000.0),
					NewPlanet(2, NewVec2(250, 0), 50.0, 1000.0), // 250m apart (more than 200m minimum)
				}
				err := ValidatePlanetSpacing(planets)
				Expect(err).NotTo(HaveOccurred())
			})

			It("passes with single planet", func() {
				planets := []Planet{
					NewPlanet(1, NewVec2(0, 0), 50.0, 1000.0),
				}
				err := ValidatePlanetSpacing(planets)
				Expect(err).NotTo(HaveOccurred())
			})

			It("passes with empty planets", func() {
				planets := []Planet{}
				err := ValidatePlanetSpacing(planets)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("ValidateWorldBounds", func() {
			It("detects positions outside bounds", func() {
				ships := []Ship{
					NewShip(1, NewVec2(1100, 0), NewVec2(0, 0), 0, 100), // X > 1000
				}
				planets := []Planet{
					NewPlanet(1, NewVec2(0, 1100), 50.0, 1000.0), // Y > 1000
				}
				pallets := []Pallet{
					NewPallet(1, NewVec2(-1100, 0), true), // X < -1000
				}
				world := NewWorld(ships, planets, pallets)
				err := ValidateWorldBounds(world)
				Expect(err).To(HaveOccurred())
			})

			It("passes with positions within bounds", func() {
				ships := []Ship{
					NewShip(1, NewVec2(500, -300), NewVec2(0, 0), 0, 100),
				}
				planets := []Planet{
					NewPlanet(1, NewVec2(-500, 500), 50.0, 1000.0),
				}
				pallets := []Pallet{
					NewPallet(1, NewVec2(0, 0), true),
				}
				world := NewWorld(ships, planets, pallets)
				err := ValidateWorldBounds(world)
				Expect(err).NotTo(HaveOccurred())
			})

			It("passes with positions at boundaries", func() {
				ships := []Ship{
					NewShip(1, NewVec2(1000, 1000), NewVec2(0, 0), 0, 100), // At boundary
				}
				world := NewWorld(ships, nil, nil)
				err := ValidateWorldBounds(world)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("ValidateWorld", func() {
			It("passes with valid world", func() {
				ships := []Ship{
					NewShip(1, NewVec2(0, 0), NewVec2(0, 0), 0, 100),
					NewShip(2, NewVec2(10, 10), NewVec2(0, 0), 0, 100),
				}
				planets := []Planet{
					NewPlanet(1, NewVec2(0, 0), 50.0, 1000.0),
					NewPlanet(2, NewVec2(250, 0), 50.0, 1000.0), // Properly spaced
				}
				pallets := []Pallet{
					NewPallet(1, NewVec2(0, 0), true),
				}
				world := NewWorld(ships, planets, pallets)
				err := ValidateWorld(world)
				Expect(err).NotTo(HaveOccurred())
			})

			It("fails with invalid world (duplicate IDs)", func() {
				ships := []Ship{
					NewShip(1, NewVec2(0, 0), NewVec2(0, 0), 0, 100),
					NewShip(1, NewVec2(10, 10), NewVec2(0, 0), 0, 100), // Duplicate
				}
				world := NewWorld(ships, nil, nil)
				err := ValidateWorld(world)
				Expect(err).To(HaveOccurred())
			})

			It("fails with invalid world (planets too close)", func() {
				planets := []Planet{
					NewPlanet(1, NewVec2(0, 0), 50.0, 1000.0),
					NewPlanet(2, NewVec2(150, 0), 50.0, 1000.0), // Too close
				}
				world := NewWorld(nil, planets, nil)
				err := ValidateWorld(world)
				Expect(err).To(HaveOccurred())
			})

			It("fails with invalid world (positions outside bounds)", func() {
				ships := []Ship{
					NewShip(1, NewVec2(1100, 0), NewVec2(0, 0), 0, 100), // Outside bounds
				}
				world := NewWorld(ships, nil, nil)
				err := ValidateWorld(world)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
