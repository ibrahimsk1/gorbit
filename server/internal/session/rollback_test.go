package session

import (
	"testing"
	"time"

	"github.com/gorbit/orbitalrush/internal/sim/entities"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRollback(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rollback Infrastructure Suite")
}

var _ = Describe("Rollback Infrastructure", Label("scope:unit", "loop:g3-orch", "layer:sim", "double:fake-io", "b:rollback-infrastructure", "r:medium"), func() {
	Describe("Snapshot Capture and Restore", func() {
		It("captures snapshot and restores state correctly", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(1.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 5.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)
			world.Tick = 5

			manager := NewSnapshotManager()
			snapshot := manager.CaptureSnapshot(world, 5, clock)

			// Modify world
			world.Ship.Pos = entities.NewVec2(20.0, 10.0)
			world.Tick = 10

			// Restore from snapshot
			restored := manager.RestoreSnapshot(snapshot)

			// Verify restored state matches original
			Expect(restored.Tick).To(Equal(uint32(5)))
			Expect(restored.Ship.Pos.X).To(Equal(10.0))
			Expect(restored.Ship.Pos.Y).To(Equal(0.0))
			Expect(restored.Ship.Vel.X).To(Equal(1.0))
			Expect(restored.Ship.Energy).To(Equal(float32(100.0)))
		})

		It("snapshot preserves all world state fields", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 5.0),
				entities.NewVec2(1.0, 2.0),
				1.5,
				75.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 5.0, 1000.0)
			pallets := []entities.Pallet{
				entities.NewPallet(1, entities.NewVec2(5.0, 5.0), true),
				entities.NewPallet(2, entities.NewVec2(15.0, 15.0), false),
			}
			world := entities.NewWorld(ship, sun, pallets)
			world.Tick = 42
			world.Done = true
			world.Win = true

			manager := NewSnapshotManager()
			snapshot := manager.CaptureSnapshot(world, 42, clock)
			restored := manager.RestoreSnapshot(snapshot)

			// Verify all fields
			Expect(restored.Tick).To(Equal(uint32(42)))
			Expect(restored.Done).To(BeTrue())
			Expect(restored.Win).To(BeTrue())
			Expect(restored.Ship.Pos.X).To(Equal(10.0))
			Expect(restored.Ship.Pos.Y).To(Equal(5.0))
			Expect(restored.Ship.Vel.X).To(Equal(1.0))
			Expect(restored.Ship.Vel.Y).To(Equal(2.0))
			Expect(restored.Ship.Rot).To(Equal(1.5))
			Expect(restored.Ship.Energy).To(Equal(float32(75.0)))
			Expect(restored.Sun.Pos.X).To(Equal(0.0))
			Expect(restored.Sun.Pos.Y).To(Equal(0.0))
			Expect(restored.Sun.Radius).To(Equal(float32(5.0)))
			Expect(restored.Sun.Mass).To(Equal(1000.0))
			Expect(len(restored.Pallets)).To(Equal(2))
			Expect(restored.Pallets[0].ID).To(Equal(uint32(1)))
			Expect(restored.Pallets[0].Active).To(BeTrue())
			Expect(restored.Pallets[1].ID).To(Equal(uint32(2)))
			Expect(restored.Pallets[1].Active).To(BeFalse())
		})

		It("snapshot isolation - modifying restored state doesn't affect snapshot", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 5.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)

			manager := NewSnapshotManager()
			snapshot := manager.CaptureSnapshot(world, 0, clock)

			// Restore and modify
			restored := manager.RestoreSnapshot(snapshot)
			restored.Ship.Pos = entities.NewVec2(999.0, 999.0)
			restored.Tick = 999

			// Restore again - should still have original values
			restored2 := manager.RestoreSnapshot(snapshot)
			Expect(restored2.Ship.Pos.X).To(Equal(10.0))
			Expect(restored2.Tick).To(Equal(uint32(0)))
		})
	})

	Describe("Multiple Snapshots", func() {
		It("can store and retrieve multiple snapshots", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 5.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)

			manager := NewSnapshotManager()

			// Capture snapshot at tick 0
			world.Tick = 0
			snapshot0 := manager.CaptureSnapshot(world, 0, clock)

			// Capture snapshot at tick 5
			world.Tick = 5
			world.Ship.Pos = entities.NewVec2(15.0, 0.0)
			snapshot5 := manager.CaptureSnapshot(world, 5, clock)

			// Capture snapshot at tick 10
			world.Tick = 10
			world.Ship.Pos = entities.NewVec2(20.0, 0.0)
			snapshot10 := manager.CaptureSnapshot(world, 10, clock)

			// Retrieve and verify snapshots
			retrieved0, exists0 := manager.GetSnapshot(0)
			Expect(exists0).To(BeTrue())
			Expect(retrieved0.Tick).To(Equal(uint32(0)))
			Expect(retrieved0.World.Ship.Pos.X).To(Equal(10.0))

			retrieved5, exists5 := manager.GetSnapshot(5)
			Expect(exists5).To(BeTrue())
			Expect(retrieved5.Tick).To(Equal(uint32(5)))
			Expect(retrieved5.World.Ship.Pos.X).To(Equal(15.0))

			retrieved10, exists10 := manager.GetSnapshot(10)
			Expect(exists10).To(BeTrue())
			Expect(retrieved10.Tick).To(Equal(uint32(10)))
			Expect(retrieved10.World.Ship.Pos.X).To(Equal(20.0))

			// Verify snapshots are independent
			Expect(snapshot0).To(Equal(retrieved0))
			Expect(snapshot5).To(Equal(retrieved5))
			Expect(snapshot10).To(Equal(retrieved10))
		})

		It("can restore to earlier snapshot", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 5.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)

			manager := NewSnapshotManager()

			// Capture snapshot at tick 5
			world.Tick = 5
			world.Ship.Pos = entities.NewVec2(15.0, 0.0)
			manager.CaptureSnapshot(world, 5, clock)

			// Advance world
			world.Tick = 10
			world.Ship.Pos = entities.NewVec2(20.0, 0.0)

			// Restore to tick 5
			snapshot5, _ := manager.GetSnapshot(5)
			restored := manager.RestoreSnapshot(snapshot5)

			Expect(restored.Tick).To(Equal(uint32(5)))
			Expect(restored.Ship.Pos.X).To(Equal(15.0))
		})

		It("GetSnapshot returns false for non-existent snapshot", func() {
			manager := NewSnapshotManager()
			_, exists := manager.GetSnapshot(42)
			Expect(exists).To(BeFalse())
		})

		It("ClearSnapshots removes all snapshots", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 5.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)

			manager := NewSnapshotManager()
			manager.CaptureSnapshot(world, 0, clock)
			manager.CaptureSnapshot(world, 5, clock)
			manager.CaptureSnapshot(world, 10, clock)

			// Verify snapshots exist
			_, exists0 := manager.GetSnapshot(0)
			_, exists5 := manager.GetSnapshot(5)
			_, exists10 := manager.GetSnapshot(10)
			Expect(exists0).To(BeTrue())
			Expect(exists5).To(BeTrue())
			Expect(exists10).To(BeTrue())

			// Clear snapshots
			manager.ClearSnapshots()

			// Verify snapshots are gone
			_, exists0 = manager.GetSnapshot(0)
			_, exists5 = manager.GetSnapshot(5)
			_, exists10 = manager.GetSnapshot(10)
			Expect(exists0).To(BeFalse())
			Expect(exists5).To(BeFalse())
			Expect(exists10).To(BeFalse())
		})
	})

	Describe("Rollback Hooks", func() {
		It("calls BeforeSnapshot hook when capturing snapshot", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 5.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)

			manager := NewSnapshotManager()
			called := false
			var capturedSnapshot *Snapshot

			hook := &testHook{
				beforeSnapshot: func(snapshot *Snapshot) {
					called = true
					capturedSnapshot = snapshot
				},
			}
			manager.RegisterHook(hook)

			snapshot := manager.CaptureSnapshot(world, 0, clock)

			Expect(called).To(BeTrue())
			Expect(capturedSnapshot).To(Equal(snapshot))
		})

		It("calls AfterRestore hook when restoring snapshot", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 5.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)

			manager := NewSnapshotManager()
			called := false
			var restoredSnapshot *Snapshot

			hook := &testHook{
				afterRestore: func(snapshot *Snapshot) {
					called = true
					restoredSnapshot = snapshot
				},
			}
			manager.RegisterHook(hook)

			snapshot := manager.CaptureSnapshot(world, 0, clock)
			manager.RestoreSnapshot(snapshot)

			Expect(called).To(BeTrue())
			Expect(restoredSnapshot).To(Equal(snapshot))
		})

		It("calls multiple hooks in order", func() {
			clock := NewFakeClock()
			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 5.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)

			manager := NewSnapshotManager()
			callOrder := []int{}

			hook1 := &testHook{
				beforeSnapshot: func(snapshot *Snapshot) {
					callOrder = append(callOrder, 1)
				},
			}
			hook2 := &testHook{
				beforeSnapshot: func(snapshot *Snapshot) {
					callOrder = append(callOrder, 2)
				},
			}
			hook3 := &testHook{
				beforeSnapshot: func(snapshot *Snapshot) {
					callOrder = append(callOrder, 3)
				},
			}

			manager.RegisterHook(hook1)
			manager.RegisterHook(hook2)
			manager.RegisterHook(hook3)

			manager.CaptureSnapshot(world, 0, clock)

			Expect(callOrder).To(Equal([]int{1, 2, 3}))
		})
	})

	Describe("Snapshot Timestamp", func() {
		It("captures timestamp from clock", func() {
			clock := NewFakeClock()
			expectedTime := clock.Now()

			ship := entities.NewShip(
				entities.NewVec2(10.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			sun := entities.NewSun(entities.NewVec2(0.0, 0.0), 5.0, 1000.0)
			world := entities.NewWorld(ship, sun, nil)

			manager := NewSnapshotManager()
			snapshot := manager.CaptureSnapshot(world, 0, clock)

			Expect(snapshot.Time).To(Equal(expectedTime))

			// Advance clock and capture another snapshot
			clock.Advance(100 * time.Millisecond)
			snapshot2 := manager.CaptureSnapshot(world, 1, clock)

			Expect(snapshot2.Time).To(Equal(clock.Now()))
			Expect(snapshot2.Time).To(BeTemporally(">", snapshot.Time))
		})
	})
})

// testHook is a test implementation of RollbackHook for testing.
type testHook struct {
	beforeSnapshot func(*Snapshot)
	afterRestore   func(*Snapshot)
}

func (h *testHook) BeforeSnapshot(snapshot *Snapshot) {
	if h.beforeSnapshot != nil {
		h.beforeSnapshot(snapshot)
	}
}

func (h *testHook) AfterRestore(snapshot *Snapshot) {
	if h.afterRestore != nil {
		h.afterRestore(snapshot)
	}
}

