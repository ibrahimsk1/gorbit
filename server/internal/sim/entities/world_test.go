package entities

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sun", Label("scope:unit", "loop:g1-physics", "layer:sim", "dep:none", "b:entity-types", "r:low"), func() {
	Describe("Constructor", func() {
		It("creates a new Sun with given values", func() {
			pos := NewVec2(0.0, 0.0)
			radius := float32(50.0)
			mass := 1000.0

			sun := NewSun(pos, radius, mass)

			Expect(sun.Pos).To(Equal(pos))
			Expect(sun.Radius).To(Equal(radius))
			Expect(sun.Mass).To(Equal(mass))
		})

		It("creates a zero sun", func() {
			sun := Sun{}

			Expect(sun.Pos).To(Equal(Zero()))
			Expect(sun.Radius).To(Equal(float32(0.0)))
			Expect(sun.Mass).To(Equal(0.0))
		})
	})

	Describe("Properties", func() {
		It("maintains field values after creation", func() {
			pos := NewVec2(100.0, 200.0)
			radius := float32(25.5)
			mass := 5000.0

			sun := NewSun(pos, radius, mass)

			Expect(sun.Pos.X).To(Equal(100.0))
			Expect(sun.Pos.Y).To(Equal(200.0))
			Expect(sun.Radius).To(Equal(float32(25.5)))
			Expect(sun.Mass).To(Equal(5000.0))
		})

		It("allows positive radius", func() {
			sun := NewSun(NewVec2(0, 0), 10.0, 100)
			Expect(sun.Radius).To(Equal(float32(10.0)))
		})

		It("allows positive mass", func() {
			sun := NewSun(NewVec2(0, 0), 10.0, 1000.0)
			Expect(sun.Mass).To(Equal(1000.0))
		})
	})
})

var _ = Describe("Pallet", Label("scope:unit", "loop:g1-physics", "layer:sim", "dep:none", "b:entity-types", "r:low"), func() {
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
})

