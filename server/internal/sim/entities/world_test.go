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

var _ = Describe("World", Label("scope:unit", "loop:g1-physics", "layer:sim", "dep:none", "b:entity-types", "r:low"), func() {
	Describe("Constructor", func() {
		It("creates a new World with given values", func() {
			ship := NewShip(NewVec2(10, 20), NewVec2(1, 2), 0.5, 100)
			sun := NewSun(NewVec2(0, 0), 50.0, 1000)
			pallets := []Pallet{
				NewPallet(1, NewVec2(100, 100), true),
				NewPallet(2, NewVec2(200, 200), true),
			}

			world := NewWorld(ship, sun, pallets)

			Expect(world.Ship).To(Equal(ship))
			Expect(world.Sun).To(Equal(sun))
			Expect(world.Pallets).To(HaveLen(2))
			Expect(world.Pallets[0]).To(Equal(pallets[0]))
			Expect(world.Pallets[1]).To(Equal(pallets[1]))
			Expect(world.Tick).To(Equal(uint32(0)))
			Expect(world.Done).To(BeFalse())
			Expect(world.Win).To(BeFalse())
		})

		It("creates a zero world", func() {
			world := World{}

			Expect(world.Ship).To(Equal(Ship{}))
			Expect(world.Sun).To(Equal(Sun{}))
			Expect(world.Pallets).To(BeEmpty())
			Expect(world.Tick).To(Equal(uint32(0)))
			Expect(world.Done).To(BeFalse())
			Expect(world.Win).To(BeFalse())
		})

		It("creates a world with empty pallets", func() {
			ship := NewShip(NewVec2(0, 0), NewVec2(0, 0), 0, 100)
			sun := NewSun(NewVec2(0, 0), 50.0, 1000)
			world := NewWorld(ship, sun, nil)

			Expect(world.Pallets).To(BeEmpty())
		})
	})

	Describe("Properties", func() {
		It("maintains field values after creation", func() {
			ship := NewShip(NewVec2(5, 10), NewVec2(0.5, 1), 1.0, 75)
			sun := NewSun(NewVec2(0, 0), 25.0, 500)
			pallets := []Pallet{NewPallet(1, NewVec2(50, 50), true)}

			world := NewWorld(ship, sun, pallets)

			Expect(world.Ship.Energy).To(Equal(float32(75)))
			Expect(world.Sun.Mass).To(Equal(500.0))
			Expect(world.Pallets).To(HaveLen(1))
			Expect(world.Tick).To(Equal(uint32(0)))
		})

		It("allows multiple pallets with different IDs", func() {
			pallets := []Pallet{
				NewPallet(1, NewVec2(10, 10), true),
				NewPallet(2, NewVec2(20, 20), true),
				NewPallet(3, NewVec2(30, 30), false),
			}

			world := NewWorld(Ship{}, Sun{}, pallets)

			Expect(world.Pallets).To(HaveLen(3))
			Expect(world.Pallets[0].ID).To(Equal(uint32(1)))
			Expect(world.Pallets[1].ID).To(Equal(uint32(2)))
			Expect(world.Pallets[2].ID).To(Equal(uint32(3)))
		})
	})
})

