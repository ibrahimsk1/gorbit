package entities

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Planet", Label("scope:unit", "loop:g1-entities", "layer:entities", "dep:none", "b:planet-entity", "r:medium"), func() {
	Describe("Constructor", func() {
		It("creates a new Planet with given values", func() {
			id := uint32(1)
			pos := NewVec2(0.0, 0.0)
			radius := float32(50.0)
			mass := 1000.0

			planet := NewPlanet(id, pos, radius, mass)

			Expect(planet.ID).To(Equal(id))
			Expect(planet.Pos).To(Equal(pos))
			Expect(planet.Radius).To(Equal(radius))
			Expect(planet.Mass).To(Equal(mass))
		})

		It("creates a zero planet", func() {
			planet := Planet{}

			Expect(planet.ID).To(Equal(uint32(0)))
			Expect(planet.Pos).To(Equal(Zero()))
			Expect(planet.Radius).To(Equal(float32(0.0)))
			Expect(planet.Mass).To(Equal(0.0))
		})
	})

	Describe("Properties", func() {
		It("maintains field values after creation", func() {
			id := uint32(42)
			pos := NewVec2(100.0, 200.0)
			radius := float32(25.5)
			mass := 5000.0

			planet := NewPlanet(id, pos, radius, mass)

			Expect(planet.ID).To(Equal(uint32(42)))
			Expect(planet.Pos.X).To(Equal(100.0))
			Expect(planet.Pos.Y).To(Equal(200.0))
			Expect(planet.Radius).To(Equal(float32(25.5)))
			Expect(planet.Mass).To(Equal(5000.0))
		})

		It("allows positive radius", func() {
			planet := NewPlanet(1, NewVec2(0, 0), 10.0, 100)
			Expect(planet.Radius).To(Equal(float32(10.0)))
		})

		It("allows positive mass", func() {
			planet := NewPlanet(1, NewVec2(0, 0), 10.0, 1000.0)
			Expect(planet.Mass).To(Equal(1000.0))
		})
	})

	Describe("ID Field", func() {
		It("Planet can be created with a valid ID", func() {
			id := uint32(123)
			planet := NewPlanet(id, NewVec2(0, 0), 50.0, 1000.0)
			Expect(planet.ID).To(Equal(uint32(123)))
		})

		It("Planet can be created with zero ID", func() {
			planet := NewPlanet(0, NewVec2(0, 0), 50.0, 1000.0)
			Expect(planet.ID).To(Equal(uint32(0)))
		})

		It("Planet maintains ID value after creation", func() {
			id := uint32(999)
			planet := NewPlanet(id, NewVec2(10, 20), 30.0, 2000.0)
			Expect(planet.ID).To(Equal(uint32(999)))
			// Verify ID persists even after other operations
			planet.Pos = NewVec2(100, 200)
			Expect(planet.ID).To(Equal(uint32(999)))
		})

		It("Multiple planets can have different IDs", func() {
			planet1 := NewPlanet(1, NewVec2(0, 0), 30.0, 500.0)
			planet2 := NewPlanet(2, NewVec2(0, 0), 40.0, 1000.0)
			planet3 := NewPlanet(3, NewVec2(0, 0), 50.0, 1500.0)

			Expect(planet1.ID).To(Equal(uint32(1)))
			Expect(planet2.ID).To(Equal(uint32(2)))
			Expect(planet3.ID).To(Equal(uint32(3)))
			Expect(planet1.ID).NotTo(Equal(planet2.ID))
			Expect(planet2.ID).NotTo(Equal(planet3.ID))
			Expect(planet1.ID).NotTo(Equal(planet3.ID))
		})
	})

	Describe("Invariants", func() {
		It("Planet can have positive radius", func() {
			planet := NewPlanet(1, NewVec2(0, 0), 30.0, 500.0)
			Expect(planet.Radius).To(BeNumerically(">", float32(0)))
		})

		It("Planet can have positive mass", func() {
			planet := NewPlanet(1, NewVec2(0, 0), 30.0, 500.0)
			Expect(planet.Mass).To(BeNumerically(">", 0.0))
		})

		It("Planet position is finite Vec2", func() {
			planet := NewPlanet(1, NewVec2(100.0, 200.0), 30.0, 500.0)
			Expect(planet.Pos.X).To(BeNumerically(">", -1e10))
			Expect(planet.Pos.X).To(BeNumerically("<", 1e10))
			Expect(planet.Pos.Y).To(BeNumerically(">", -1e10))
			Expect(planet.Pos.Y).To(BeNumerically("<", 1e10))
		})
	})
})

