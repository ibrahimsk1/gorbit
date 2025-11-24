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
				1, // Player ID
				entities.NewVec2(10.5, 20.3),
				entities.NewVec2(1.0, -2.0),
				1.57,
				75.5,
			)
			result := ShipToSnapshot(ship)

			Expect(result.ID).To(Equal(uint32(1)))
			Expect(result.Pos.X).To(Equal(10.5))
			Expect(result.Pos.Y).To(Equal(20.3))
			Expect(result.Vel.X).To(Equal(1.0))
			Expect(result.Vel.Y).To(Equal(-2.0))
			Expect(result.Rot).To(Equal(1.57))
			Expect(result.Energy).To(Equal(float32(75.5)))
		})

		It("converts ship with zero values correctly", func() {
			ship := entities.NewShip(
				1, // Player ID
				entities.Zero(),
				entities.Zero(),
				0.0,
				0.0,
			)
			result := ShipToSnapshot(ship)

			Expect(result.ID).To(Equal(uint32(1)))
			Expect(result.Pos.X).To(Equal(0.0))
			Expect(result.Pos.Y).To(Equal(0.0))
			Expect(result.Vel.X).To(Equal(0.0))
			Expect(result.Vel.Y).To(Equal(0.0))
			Expect(result.Rot).To(Equal(0.0))
			Expect(result.Energy).To(Equal(float32(0.0)))
		})

		It("verifies Pos and Vel are converted using Vec2ToSnapshot", func() {
			ship := entities.NewShip(
				1, // Player ID
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
				1, // Player ID
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

	Describe("PlanetToSnapshot", func() {
		It("converts planet with all fields set correctly", func() {
			planet := entities.NewPlanet(
				1, // Planet ID
				entities.NewVec2(0.0, 0.0),
				50.0,
				1000.0,
			)
			result := PlanetToSnapshot(planet)

			Expect(result.ID).To(Equal(uint32(1)))
			Expect(result.Pos.X).To(Equal(0.0))
			Expect(result.Pos.Y).To(Equal(0.0))
			Expect(result.Radius).To(Equal(float32(50.0)))
		})

		It("converts planet at origin correctly", func() {
			planet := entities.NewPlanet(
				1, // Planet ID
				entities.Zero(),
				25.5,
				500.0,
			)
			result := PlanetToSnapshot(planet)

			Expect(result.ID).To(Equal(uint32(1)))
			Expect(result.Pos.X).To(Equal(0.0))
			Expect(result.Pos.Y).To(Equal(0.0))
			Expect(result.Radius).To(Equal(float32(25.5)))
		})

		It("verifies Pos is converted using Vec2ToSnapshot", func() {
			planet := entities.NewPlanet(
				1, // Planet ID
				entities.NewVec2(100.0, -200.0),
				30.0,
				750.0,
			)
			result := PlanetToSnapshot(planet)

			posSnapshot := Vec2ToSnapshot(planet.Pos)
			Expect(result.Pos).To(Equal(posSnapshot))
		})

		It("verifies Radius is mapped correctly", func() {
			planet := entities.NewPlanet(
				1, // Planet ID
				entities.Zero(),
				42.5,
				1000.0,
			)
			result := PlanetToSnapshot(planet)

			Expect(result.Radius).To(Equal(float32(42.5)))
		})

		It("verifies Mass is not included (not in proto)", func() {
			// This test verifies that Mass field from entities.Planet
			// is not part of the proto.PlanetSnapshot structure
			// The proto.PlanetSnapshot only has ID, Pos, and Radius
			planet := entities.NewPlanet(
				1, // Planet ID
				entities.NewVec2(0.0, 0.0),
				50.0,
				1000.0,
			)
			result := PlanetToSnapshot(planet)

			// Verify that result only contains ID, Pos, and Radius
			// (Mass is not accessible in proto.PlanetSnapshot)
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
			ships := []entities.Ship{
				entities.NewShip(1, entities.NewVec2(10.5, 20.3), entities.NewVec2(1.0, -2.0), 1.57, 75.5),
				entities.NewShip(2, entities.NewVec2(-10.5, -20.3), entities.NewVec2(-1.0, 2.0), 0.0, 50.0),
			}
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
				entities.NewPlanet(2, entities.NewVec2(100.0, 100.0), 30.0, 500.0),
			}
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(15.0, 15.0), true),
				entities.NewPallet(2, entities.NewVec2(-10.0, 10.0), false),
			}
			world := entities.NewWorld(ships, planets, pallets)
			world.Tick = 100

			playerID := uint32(1)
			result := WorldToSnapshot(world, playerID)

			Expect(result.Type).To(Equal("snapshot"))
			Expect(result.Tick).To(Equal(uint32(100)))
			Expect(result.MyShipId).To(Equal(playerID))

			// Verify Ships array
			Expect(result.Ships).To(HaveLen(2))
			Expect(result.Ships[0].ID).To(Equal(uint32(1)))
			Expect(result.Ships[0].Pos.X).To(Equal(10.5))
			Expect(result.Ships[0].Pos.Y).To(Equal(20.3))
			Expect(result.Ships[0].Energy).To(Equal(float32(75.5)))
			Expect(result.Ships[1].ID).To(Equal(uint32(2)))

			// Verify Planets array
			Expect(result.Planets).To(HaveLen(2))
			Expect(result.Planets[0].ID).To(Equal(uint32(1)))
			Expect(result.Planets[0].Pos.X).To(Equal(0.0))
			Expect(result.Planets[0].Pos.Y).To(Equal(0.0))
			Expect(result.Planets[0].Radius).To(Equal(float32(50.0)))

			// Verify Pallets
			Expect(result.Pallets).To(HaveLen(2))
			Expect(result.Pallets[0].ID).To(Equal(uint32(1)))
			Expect(result.Pallets[0].Pos.X).To(Equal(15.0))
			Expect(result.Pallets[0].Active).To(BeTrue())
			Expect(result.Pallets[1].ID).To(Equal(uint32(2)))
			Expect(result.Pallets[1].Pos.X).To(Equal(-10.0))
			Expect(result.Pallets[1].Active).To(BeFalse())

			// Verify WorldBounds
			Expect(result.WorldBounds.Width).To(Equal(entities.WORLD_WIDTH))
			Expect(result.WorldBounds.Height).To(Equal(entities.WORLD_HEIGHT))
		})

		It("converts world with empty pallets slice correctly", func() {
			ships := []entities.Ship{
				entities.NewShip(1, entities.Zero(), entities.Zero(), 0.0, 100.0),
			}
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.Zero(), 50.0, 1000.0),
			}
			world := entities.NewWorld(ships, planets, []entities.Pallet{})

			playerID := uint32(1)
			result := WorldToSnapshot(world, playerID)

			Expect(result.Pallets).To(BeEmpty())
			Expect(result.Pallets).ToNot(BeNil()) // Should be empty slice, not nil
		})

		It("converts world with multiple pallets correctly", func() {
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(10.0, 10.0), true),
				entities.NewPallet(2, entities.NewVec2(20.0, 20.0), true),
				entities.NewPallet(3, entities.NewVec2(30.0, 30.0), false),
			}
			ships := []entities.Ship{
				entities.NewShip(1, entities.Zero(), entities.Zero(), 0.0, 100.0),
			}
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.Zero(), 50.0, 1000.0),
			}
			world := entities.NewWorld(ships, planets, pallets)

			playerID := uint32(1)
			result := WorldToSnapshot(world, playerID)

			Expect(result.Pallets).To(HaveLen(3))
			Expect(result.Pallets[0].ID).To(Equal(uint32(1)))
			Expect(result.Pallets[1].ID).To(Equal(uint32(2)))
			Expect(result.Pallets[2].ID).To(Equal(uint32(3)))
		})

		It("verifies Type is set to 'snapshot'", func() {
			ships := []entities.Ship{
				entities.NewShip(1, entities.Zero(), entities.Zero(), 0.0, 100.0),
			}
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.Zero(), 50.0, 1000.0),
			}
			world := entities.NewWorld(ships, planets, nil)

			playerID := uint32(1)
			result := WorldToSnapshot(world, playerID)

			Expect(result.Type).To(Equal("snapshot"))
		})

		It("verifies Tick and MyShipId are mapped correctly", func() {
			ships := []entities.Ship{
				entities.NewShip(1, entities.Zero(), entities.Zero(), 0.0, 100.0),
			}
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.Zero(), 50.0, 1000.0),
			}
			world := entities.NewWorld(ships, planets, nil)
			world.Tick = 42

			playerID := uint32(1)
			result := WorldToSnapshot(world, playerID)

			Expect(result.Tick).To(Equal(uint32(42)))
			Expect(result.MyShipId).To(Equal(playerID))
		})

		It("verifies all nested entities are converted correctly", func() {
			ships := []entities.Ship{
				entities.NewShip(1, entities.NewVec2(5.0, 10.0), entities.NewVec2(0.5, 1.0), 1.0, 75.0),
			}
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 25.0, 500.0),
			}
			pallet := entities.NewPallet(
				1,
				entities.NewVec2(50.0, 50.0),
				true,
			)
			world := entities.NewWorld(ships, planets, []entities.Pallet{pallet})

			playerID := uint32(1)
			result := WorldToSnapshot(world, playerID)

			// Verify Ship conversion
			shipSnapshot := ShipToSnapshot(ships[0])
			Expect(result.Ships).To(HaveLen(1))
			Expect(result.Ships[0]).To(Equal(shipSnapshot))

			// Verify Planet conversion
			planetSnapshot := PlanetToSnapshot(planets[0])
			Expect(result.Planets).To(HaveLen(1))
			Expect(result.Planets[0]).To(Equal(planetSnapshot))

			// Verify Pallet conversion
			Expect(result.Pallets).To(HaveLen(1))
			palletSnapshot := PalletToSnapshot(pallet)
			Expect(result.Pallets[0]).To(Equal(palletSnapshot))
		})
	})

	Describe("Round-trip Validation", func() {
		It("converted SnapshotMessage should pass proto.ValidateSnapshotMessage", func() {
			ships := []entities.Ship{
				entities.NewShip(1, entities.NewVec2(10.5, 20.3), entities.NewVec2(1.0, -2.0), 1.57, 75.5),
			}
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}
			world := entities.NewWorld(
				ships,
				planets,
				[]entities.Pallet{
					entities.NewPallet(1, entities.NewVec2(15.0, 15.0), true),
				},
			)
			world.Tick = 100

			playerID := uint32(1)
			result := WorldToSnapshot(world, playerID)

			err := proto.ValidateSnapshotMessage(&result)
			Expect(err).NotTo(HaveOccurred())
		})

		It("all nested snapshots should pass their respective validation functions", func() {
			ships := []entities.Ship{
				entities.NewShip(1, entities.NewVec2(10.5, 20.3), entities.NewVec2(1.0, -2.0), 1.57, 75.5),
			}
			planets := []entities.Planet{
				entities.NewPlanet(1, entities.NewVec2(0.0, 0.0), 50.0, 1000.0),
			}
			world := entities.NewWorld(
				ships,
				planets,
				[]entities.Pallet{
					entities.NewPallet(1, entities.NewVec2(15.0, 15.0), true),
					entities.NewPallet(2, entities.NewVec2(-10.0, 10.0), false),
				},
			)

			playerID := uint32(1)
			result := WorldToSnapshot(world, playerID)

			// Validate Ships
			for i := range result.Ships {
				err := proto.ValidateShipSnapshot(&result.Ships[i])
				Expect(err).NotTo(HaveOccurred())
			}

			// Validate Planets
			for i := range result.Planets {
				err := proto.ValidatePlanetSnapshot(&result.Planets[i])
				Expect(err).NotTo(HaveOccurred())
			}

			// Validate Pallets
			for i := range result.Pallets {
				err := proto.ValidatePalletSnapshot(&result.Pallets[i])
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})
})


