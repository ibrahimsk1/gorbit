package transport

import (
	"testing"

	"github.com/gorbit/orbitalrush/internal/proto"
	"github.com/gorbit/orbitalrush/internal/sim/entities"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConvert(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Entity-to-Protocol Conversion Suite")
}

var _ = Describe("Entity-to-Protocol Conversion", Label("scope:unit", "loop:g5-adapter", "layer:server", "b:entity-conversion", "r:medium"), func() {
	Describe("Vec2ToSnapshot", func() {
		It("converts zero vector correctly", func() {
			v := entities.Zero()
			result := Vec2ToSnapshot(v)

			Expect(result.X).To(Equal(0.0))
			Expect(result.Y).To(Equal(0.0))
		})

		It("converts positive coordinates correctly", func() {
			v := entities.NewVec2(10.5, 20.3)
			result := Vec2ToSnapshot(v)

			Expect(result.X).To(Equal(10.5))
			Expect(result.Y).To(Equal(20.3))
		})

		It("converts negative coordinates correctly", func() {
			v := entities.NewVec2(-5.2, -15.7)
			result := Vec2ToSnapshot(v)

			Expect(result.X).To(Equal(-5.2))
			Expect(result.Y).To(Equal(-15.7))
		})

		It("converts large values correctly", func() {
			v := entities.NewVec2(1e10, -1e10)
			result := Vec2ToSnapshot(v)

			Expect(result.X).To(Equal(1e10))
			Expect(result.Y).To(Equal(-1e10))
		})

		It("preserves precision (float64 to float64)", func() {
			v := entities.NewVec2(0.123456789012345, 0.987654321098765)
			result := Vec2ToSnapshot(v)

			Expect(result.X).To(Equal(0.123456789012345))
			Expect(result.Y).To(Equal(0.987654321098765))
		})
	})

	Describe("ShipToSnapshot", func() {
		It("converts ship with all fields set correctly", func() {
			ship := entities.NewShip(
				entities.NewVec2(10.5, 20.3),
				entities.NewVec2(1.0, -2.0),
				1.57,
				75.5,
			)
			result := ShipToSnapshot(ship)

			Expect(result.Pos.X).To(Equal(10.5))
			Expect(result.Pos.Y).To(Equal(20.3))
			Expect(result.Vel.X).To(Equal(1.0))
			Expect(result.Vel.Y).To(Equal(-2.0))
			Expect(result.Rot).To(Equal(1.57))
			Expect(result.Energy).To(Equal(float32(75.5)))
		})

		It("converts ship with zero values correctly", func() {
			ship := entities.NewShip(
				entities.Zero(),
				entities.Zero(),
				0.0,
				0.0,
			)
			result := ShipToSnapshot(ship)

			Expect(result.Pos.X).To(Equal(0.0))
			Expect(result.Pos.Y).To(Equal(0.0))
			Expect(result.Vel.X).To(Equal(0.0))
			Expect(result.Vel.Y).To(Equal(0.0))
			Expect(result.Rot).To(Equal(0.0))
			Expect(result.Energy).To(Equal(float32(0.0)))
		})

		It("verifies Pos and Vel are converted using Vec2ToSnapshot", func() {
			ship := entities.NewShip(
				entities.NewVec2(100.0, 200.0),
				entities.NewVec2(-50.0, 25.0),
				0.785,
				50.0,
			)
			result := ShipToSnapshot(ship)

			// Verify Pos conversion
			posSnapshot := Vec2ToSnapshot(ship.Pos)
			Expect(result.Pos).To(Equal(posSnapshot))

			// Verify Vel conversion
			velSnapshot := Vec2ToSnapshot(ship.Vel)
			Expect(result.Vel).To(Equal(velSnapshot))
		})

		It("verifies Rot and Energy are mapped correctly", func() {
			ship := entities.NewShip(
				entities.Zero(),
				entities.Zero(),
				3.14159,
				99.99,
			)
			result := ShipToSnapshot(ship)

			Expect(result.Rot).To(Equal(3.14159))
			Expect(result.Energy).To(Equal(float32(99.99)))
		})
	})

	Describe("SunToSnapshot", func() {
		It("converts sun with all fields set correctly", func() {
			sun := entities.NewSun(
				entities.NewVec2(0.0, 0.0),
				50.0,
				1000.0,
			)
			result := SunToSnapshot(sun)

			Expect(result.Pos.X).To(Equal(0.0))
			Expect(result.Pos.Y).To(Equal(0.0))
			Expect(result.Radius).To(Equal(float32(50.0)))
		})

		It("converts sun at origin correctly", func() {
			sun := entities.NewSun(
				entities.Zero(),
				25.5,
				500.0,
			)
			result := SunToSnapshot(sun)

			Expect(result.Pos.X).To(Equal(0.0))
			Expect(result.Pos.Y).To(Equal(0.0))
			Expect(result.Radius).To(Equal(float32(25.5)))
		})

		It("verifies Pos is converted using Vec2ToSnapshot", func() {
			sun := entities.NewSun(
				entities.NewVec2(100.0, -200.0),
				30.0,
				750.0,
			)
			result := SunToSnapshot(sun)

			posSnapshot := Vec2ToSnapshot(sun.Pos)
			Expect(result.Pos).To(Equal(posSnapshot))
		})

		It("verifies Radius is mapped correctly", func() {
			sun := entities.NewSun(
				entities.Zero(),
				42.5,
				1000.0,
			)
			result := SunToSnapshot(sun)

			Expect(result.Radius).To(Equal(float32(42.5)))
		})

		It("verifies Mass is not included (not in proto)", func() {
			// This test verifies that Mass field from entities.Sun
			// is not part of the proto.SunSnapshot structure
			// The proto.SunSnapshot only has Pos and Radius
			sun := entities.NewSun(
				entities.NewVec2(0.0, 0.0),
				50.0,
				1000.0,
			)
			result := SunToSnapshot(sun)

			// Verify that result only contains Pos and Radius
			// (Mass is not accessible in proto.SunSnapshot)
			Expect(result.Pos).ToNot(BeNil())
			Expect(result.Radius).To(Equal(float32(50.0)))
		})
	})

	Describe("PalletToSnapshot", func() {
		It("converts active pallet correctly", func() {
			pallet := entities.NewPallet(
				1,
				entities.NewVec2(15.0, 15.0),
				true,
			)
			result := PalletToSnapshot(pallet)

			Expect(result.ID).To(Equal(uint32(1)))
			Expect(result.Pos.X).To(Equal(15.0))
			Expect(result.Pos.Y).To(Equal(15.0))
			Expect(result.Active).To(BeTrue())
		})

		It("converts inactive pallet correctly", func() {
			pallet := entities.NewPallet(
				42,
				entities.NewVec2(-10.0, 10.0),
				false,
			)
			result := PalletToSnapshot(pallet)

			Expect(result.ID).To(Equal(uint32(42)))
			Expect(result.Pos.X).To(Equal(-10.0))
			Expect(result.Pos.Y).To(Equal(10.0))
			Expect(result.Active).To(BeFalse())
		})

		It("verifies Pos is converted using Vec2ToSnapshot", func() {
			pallet := entities.NewPallet(
				5,
				entities.NewVec2(100.0, 200.0),
				true,
			)
			result := PalletToSnapshot(pallet)

			posSnapshot := Vec2ToSnapshot(pallet.Pos)
			Expect(result.Pos).To(Equal(posSnapshot))
		})

		It("verifies ID and Active are mapped correctly", func() {
			pallet := entities.NewPallet(
				999,
				entities.Zero(),
				false,
			)
			result := PalletToSnapshot(pallet)

			Expect(result.ID).To(Equal(uint32(999)))
			Expect(result.Active).To(BeFalse())
		})
	})

	Describe("WorldToSnapshot", func() {
		It("converts complete world with all entities correctly", func() {
			ship := entities.NewShip(
				entities.NewVec2(10.5, 20.3),
				entities.NewVec2(1.0, -2.0),
				1.57,
				75.5,
			)
			sun := entities.NewSun(
				entities.NewVec2(0.0, 0.0),
				50.0,
				1000.0,
			)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(15.0, 15.0), true),
				entities.NewPallet(2, entities.NewVec2(-10.0, 10.0), false),
			}
			world := entities.NewWorld(ship, sun, pallets)
			world.Tick = 100
			world.Done = false
			world.Win = false

			result := WorldToSnapshot(world)

			Expect(result.Type).To(Equal("snapshot"))
			Expect(result.Tick).To(Equal(uint32(100)))
			Expect(result.Done).To(BeFalse())
			Expect(result.Win).To(BeFalse())

			// Verify Ship
			Expect(result.Ship.Pos.X).To(Equal(10.5))
			Expect(result.Ship.Pos.Y).To(Equal(20.3))
			Expect(result.Ship.Energy).To(Equal(float32(75.5)))

			// Verify Sun
			Expect(result.Sun.Pos.X).To(Equal(0.0))
			Expect(result.Sun.Pos.Y).To(Equal(0.0))
			Expect(result.Sun.Radius).To(Equal(float32(50.0)))

			// Verify Pallets
			Expect(result.Pallets).To(HaveLen(2))
			Expect(result.Pallets[0].ID).To(Equal(uint32(1)))
			Expect(result.Pallets[0].Pos.X).To(Equal(15.0))
			Expect(result.Pallets[0].Active).To(BeTrue())
			Expect(result.Pallets[1].ID).To(Equal(uint32(2)))
			Expect(result.Pallets[1].Pos.X).To(Equal(-10.0))
			Expect(result.Pallets[1].Active).To(BeFalse())
		})

		It("converts world with empty pallets slice correctly", func() {
			ship := entities.NewShip(
				entities.Zero(),
				entities.Zero(),
				0.0,
				100.0,
			)
			sun := entities.NewSun(
				entities.Zero(),
				50.0,
				1000.0,
			)
			world := entities.NewWorld(ship, sun, []entities.Pallet{})

			result := WorldToSnapshot(world)

			Expect(result.Pallets).To(BeEmpty())
			Expect(result.Pallets).ToNot(BeNil()) // Should be empty slice, not nil
		})

		It("converts world with multiple pallets correctly", func() {
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), true),
				entities.NewPallet(2, entities.NewVec2(20.0, 20.0), true),
				entities.NewPallet(3, entities.NewVec2(30.0, 30.0), false),
			}
			world := entities.NewWorld(
				entities.NewShip(entities.Zero(), entities.Zero(), 0.0, 100.0),
				entities.NewSun(entities.Zero(), 50.0, 1000.0),
				pallets,
			)

			result := WorldToSnapshot(world)

			Expect(result.Pallets).To(HaveLen(3))
			Expect(result.Pallets[0].ID).To(Equal(uint32(1)))
			Expect(result.Pallets[1].ID).To(Equal(uint32(2)))
			Expect(result.Pallets[2].ID).To(Equal(uint32(3)))
		})

		It("verifies Type is set to 'snapshot'", func() {
			world := entities.NewWorld(
				entities.NewShip(entities.Zero(), entities.Zero(), 0.0, 100.0),
				entities.NewSun(entities.Zero(), 50.0, 1000.0),
				nil,
			)

			result := WorldToSnapshot(world)

			Expect(result.Type).To(Equal("snapshot"))
		})

		It("verifies Tick, Done, Win are mapped correctly", func() {
			world := entities.NewWorld(
				entities.NewShip(entities.Zero(), entities.Zero(), 0.0, 100.0),
				entities.NewSun(entities.Zero(), 50.0, 1000.0),
				nil,
			)
			world.Tick = 42
			world.Done = true
			world.Win = true

			result := WorldToSnapshot(world)

			Expect(result.Tick).To(Equal(uint32(42)))
			Expect(result.Done).To(BeTrue())
			Expect(result.Win).To(BeTrue())
		})

		It("verifies all nested entities are converted correctly", func() {
			ship := entities.NewShip(
				entities.NewVec2(5.0, 10.0),
				entities.NewVec2(0.5, 1.0),
				1.0,
				75.0,
			)
			sun := entities.NewSun(
				entities.NewVec2(0.0, 0.0),
				25.0,
				500.0,
			)
			pallet := entities.NewPallet(
				1,
				entities.NewVec2(50.0, 50.0),
				true,
			)
			world := entities.NewWorld(ship, sun, []entities.Pallet{pallet})

			result := WorldToSnapshot(world)

			// Verify Ship conversion
			shipSnapshot := ShipToSnapshot(ship)
			Expect(result.Ship).To(Equal(shipSnapshot))

			// Verify Sun conversion
			sunSnapshot := SunToSnapshot(sun)
			Expect(result.Sun).To(Equal(sunSnapshot))

			// Verify Pallet conversion
			Expect(result.Pallets).To(HaveLen(1))
			palletSnapshot := PalletToSnapshot(pallet)
			Expect(result.Pallets[0]).To(Equal(palletSnapshot))
		})
	})

	Describe("Round-trip Validation", func() {
		It("converted SnapshotMessage should pass proto.ValidateSnapshotMessage", func() {
			world := entities.NewWorld(
				entities.NewShip(
					entities.NewVec2(10.5, 20.3),
					entities.NewVec2(1.0, -2.0),
					1.57,
					75.5,
				),
				entities.NewSun(
					entities.NewVec2(0.0, 0.0),
					50.0,
					1000.0,
				),
				[]entities.Pallet{
					entities.NewPallet(1, entities.NewVec2(15.0, 15.0), true),
				},
			)
			world.Tick = 100

			result := WorldToSnapshot(world)

			err := proto.ValidateSnapshotMessage(&result)
			Expect(err).NotTo(HaveOccurred())
		})

		It("all nested snapshots should pass their respective validation functions", func() {
			world := entities.NewWorld(
				entities.NewShip(
					entities.NewVec2(10.5, 20.3),
					entities.NewVec2(1.0, -2.0),
					1.57,
					75.5,
				),
				entities.NewSun(
					entities.NewVec2(0.0, 0.0),
					50.0,
					1000.0,
				),
				[]entities.Pallet{
					entities.NewPallet(1, entities.NewVec2(15.0, 15.0), true),
					entities.NewPallet(2, entities.NewVec2(-10.0, 10.0), false),
				},
			)

			result := WorldToSnapshot(world)

			// Validate Ship
			err := proto.ValidateShipSnapshot(&result.Ship)
			Expect(err).NotTo(HaveOccurred())

			// Validate Sun
			err = proto.ValidateSunSnapshot(&result.Sun)
			Expect(err).NotTo(HaveOccurred())

			// Validate Pallets
			for i := range result.Pallets {
				err = proto.ValidatePalletSnapshot(&result.Pallets[i])
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})
})


