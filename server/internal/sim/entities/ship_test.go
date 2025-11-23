package entities

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ship", Label("scope:unit", "loop:g1-entities", "layer:entities", "dep:none", "b:ship-id", "r:low"), func() {
	Describe("Constructor", func() {
		It("creates a new Ship with given values", func() {
			id := uint32(1)
			pos := NewVec2(10.0, 20.0)
			vel := NewVec2(1.0, 2.0)
			rot := 1.5
			energy := float32(100.0)

			ship := NewShip(id, pos, vel, rot, energy)

			Expect(ship.ID).To(Equal(id))
			Expect(ship.Pos).To(Equal(pos))
			Expect(ship.Vel).To(Equal(vel))
			Expect(ship.Rot).To(Equal(rot))
			Expect(ship.Energy).To(Equal(energy))
		})

		It("creates a zero ship", func() {
			ship := Ship{}

			Expect(ship.ID).To(Equal(uint32(0)))
			Expect(ship.Pos).To(Equal(Zero()))
			Expect(ship.Vel).To(Equal(Zero()))
			Expect(ship.Rot).To(Equal(0.0))
			Expect(ship.Energy).To(Equal(float32(0.0)))
		})
	})

	Describe("Properties", func() {
		It("maintains field values after creation", func() {
			id := uint32(42)
			pos := NewVec2(5.0, 10.0)
			vel := NewVec2(0.5, 1.0)
			rot := 0.785
			energy := float32(50.0)

			ship := NewShip(id, pos, vel, rot, energy)

			Expect(ship.ID).To(Equal(uint32(42)))
			Expect(ship.Pos.X).To(Equal(5.0))
			Expect(ship.Pos.Y).To(Equal(10.0))
			Expect(ship.Vel.X).To(Equal(0.5))
			Expect(ship.Vel.Y).To(Equal(1.0))
			Expect(ship.Rot).To(Equal(0.785))
			Expect(ship.Energy).To(Equal(float32(50.0)))
		})

		It("allows zero energy", func() {
			ship := NewShip(1, NewVec2(0, 0), NewVec2(0, 0), 0, 0)
			Expect(ship.Energy).To(Equal(float32(0.0)))
		})

		It("allows negative rotation", func() {
			ship := NewShip(2, NewVec2(0, 0), NewVec2(0, 0), -1.5, 100)
			Expect(ship.Rot).To(Equal(-1.5))
		})
	})

	Describe("ID Field", func() {
		It("Ship can be created with a valid ID", func() {
			id := uint32(123)
			ship := NewShip(id, NewVec2(0, 0), NewVec2(0, 0), 0, 100)
			Expect(ship.ID).To(Equal(uint32(123)))
		})

		It("Ship can be created with zero ID", func() {
			ship := NewShip(0, NewVec2(0, 0), NewVec2(0, 0), 0, 100)
			Expect(ship.ID).To(Equal(uint32(0)))
		})

		It("Ship maintains ID value after creation", func() {
			id := uint32(999)
			ship := NewShip(id, NewVec2(10, 20), NewVec2(1, 2), 1.5, 50)
			Expect(ship.ID).To(Equal(uint32(999)))
			// Verify ID persists even after other operations
			ship.Pos = NewVec2(100, 200)
			Expect(ship.ID).To(Equal(uint32(999)))
		})

		It("Multiple ships can have different IDs", func() {
			ship1 := NewShip(1, NewVec2(0, 0), NewVec2(0, 0), 0, 100)
			ship2 := NewShip(2, NewVec2(0, 0), NewVec2(0, 0), 0, 100)
			ship3 := NewShip(3, NewVec2(0, 0), NewVec2(0, 0), 0, 100)

			Expect(ship1.ID).To(Equal(uint32(1)))
			Expect(ship2.ID).To(Equal(uint32(2)))
			Expect(ship3.ID).To(Equal(uint32(3)))
			Expect(ship1.ID).NotTo(Equal(ship2.ID))
			Expect(ship2.ID).NotTo(Equal(ship3.ID))
			Expect(ship1.ID).NotTo(Equal(ship3.ID))
		})
	})
})

