package session

import (
	"time"

	"github.com/gorbit/orbitalrush/internal/sim/entities"
)

// Snapshot represents a captured state of the game world at a specific point in time.
type Snapshot struct {
	World    entities.World
	Tick     uint32
	Time     time.Time
}

// RollbackHook is an interface for components that need to react to rollback events.
type RollbackHook interface {
	// BeforeSnapshot is called before a snapshot is taken.
	// This allows hooks to prepare for snapshot capture.
	BeforeSnapshot(snapshot *Snapshot)

	// AfterRestore is called after a snapshot is restored.
	// This allows hooks to react to state restoration.
	AfterRestore(snapshot *Snapshot)
}

// SnapshotManager manages state snapshots for rollback functionality.
type SnapshotManager struct {
	snapshots map[uint32]*Snapshot
	hooks     []RollbackHook
}

// NewSnapshotManager creates a new snapshot manager.
func NewSnapshotManager() *SnapshotManager {
	return &SnapshotManager{
		snapshots: make(map[uint32]*Snapshot),
		hooks:     make([]RollbackHook, 0),
	}
}

// RegisterHook registers a rollback hook that will be called during snapshot operations.
func (sm *SnapshotManager) RegisterHook(hook RollbackHook) {
	sm.hooks = append(sm.hooks, hook)
}

// CaptureSnapshot captures a snapshot of the world state at the current tick.
// Returns a snapshot that can be used to restore the world state later.
func (sm *SnapshotManager) CaptureSnapshot(world entities.World, tick uint32, clock Clock) *Snapshot {
	snapshot := &Snapshot{
		World: copyWorld(world),
		Tick:   tick,
		Time:   clock.Now(),
	}

	// Call hooks before snapshot
	for _, hook := range sm.hooks {
		hook.BeforeSnapshot(snapshot)
	}

	// Store snapshot
	sm.snapshots[tick] = snapshot

	return snapshot
}

// RestoreSnapshot restores the world state from a snapshot.
// Returns the restored world state.
func (sm *SnapshotManager) RestoreSnapshot(snapshot *Snapshot) entities.World {
	// Call hooks after restore
	for _, hook := range sm.hooks {
		hook.AfterRestore(snapshot)
	}

	// Return a copy of the snapshot's world state
	return copyWorld(snapshot.World)
}

// GetSnapshot retrieves a snapshot by tick number.
// Returns the snapshot and true if found, nil and false otherwise.
func (sm *SnapshotManager) GetSnapshot(tick uint32) (*Snapshot, bool) {
	snapshot, exists := sm.snapshots[tick]
	return snapshot, exists
}

// ClearSnapshots removes all stored snapshots.
func (sm *SnapshotManager) ClearSnapshots() {
	sm.snapshots = make(map[uint32]*Snapshot)
}

// copyWorld creates a deep copy of a World struct.
// This ensures that modifying the restored state doesn't affect the snapshot.
func copyWorld(world entities.World) entities.World {
	// Copy pallets slice
	palletsCopy := make([]entities.Pallet, len(world.Pallets))
	copy(palletsCopy, world.Pallets)

	return entities.World{
		Ship:    world.Ship,    // Ship is a struct, so this is a copy
		Sun:     world.Sun,     // Sun is a struct, so this is a copy
		Pallets: palletsCopy,   // Explicitly copy the slice
		Tick:    world.Tick,
		Done:    world.Done,
		Win:     world.Win,
	}
}

