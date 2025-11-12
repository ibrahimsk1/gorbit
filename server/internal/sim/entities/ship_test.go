package entities

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ship", Label("scope:unit", "loop:g1-physics", "layer:sim", "dep:none", "b:entity-types", "r:low"), func() {
	Describe("Constructor", func() {
		It("creates a new Ship with given values", func() {
			pos := NewVec2(10.0, 20.0)
			vel := NewVec2(1.0, 2.0)
			rot := 1.5
			energy := float32(100.0)

			ship := NewShip(pos, vel, rot, energy)

			Expect(ship.Pos).To(Equal(pos))
			Expect(ship.Vel).To(Equal(vel))
			Expect(ship.Rot).To(Equal(rot))
			Expect(ship.Energy).To(Equal(energy))
		})

		It("creates a zero ship", func() {
			ship := Ship{}

			Expect(ship.Pos).To(Equal(Zero()))
			Expect(ship.Vel).To(Equal(Zero()))
			Expect(ship.Rot).To(Equal(0.0))
			Expect(ship.Energy).To(Equal(float32(0.0)))
		})
	})

	Describe("Properties", func() {
		It("maintains field values after creation", func() {
			pos := NewVec2(5.0, 10.0)
			vel := NewVec2(0.5, 1.0)
			rot := 0.785
			energy := float32(50.0)

			ship := NewShip(pos, vel, rot, energy)

			Expect(ship.Pos.X).To(Equal(5.0))
			Expect(ship.Pos.Y).To(Equal(10.0))
			Expect(ship.Vel.X).To(Equal(0.5))
			Expect(ship.Vel.Y).To(Equal(1.0))
			Expect(ship.Rot).To(Equal(0.785))
			Expect(ship.Energy).To(Equal(float32(50.0)))
		})

		It("allows zero energy", func() {
			ship := NewShip(NewVec2(0, 0), NewVec2(0, 0), 0, 0)
			Expect(ship.Energy).To(Equal(float32(0.0)))
		})

		It("allows negative rotation", func() {
			ship := NewShip(NewVec2(0, 0), NewVec2(0, 0), -1.5, 100)
			Expect(ship.Rot).To(Equal(-1.5))
		})
	})
})

