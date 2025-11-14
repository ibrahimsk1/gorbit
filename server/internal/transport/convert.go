package transport

import (
	"github.com/gorbit/orbitalrush/internal/proto"
	"github.com/gorbit/orbitalrush/internal/sim/entities"
)

// Vec2ToSnapshot converts an entities.Vec2 to a proto.Vec2Snapshot.
func Vec2ToSnapshot(v entities.Vec2) proto.Vec2Snapshot {
	return proto.Vec2Snapshot{
		X: v.X,
		Y: v.Y,
	}
}

// ShipToSnapshot converts an entities.Ship to a proto.ShipSnapshot.
func ShipToSnapshot(s entities.Ship) proto.ShipSnapshot {
	return proto.ShipSnapshot{
		Pos:    Vec2ToSnapshot(s.Pos),
		Vel:    Vec2ToSnapshot(s.Vel),
		Rot:    s.Rot,
		Energy: s.Energy,
	}
}

// SunToSnapshot converts an entities.Sun to a proto.SunSnapshot.
// Note: The Mass field from entities.Sun is not included in proto.SunSnapshot
// as it is only used for simulation calculations.
func SunToSnapshot(s entities.Sun) proto.SunSnapshot {
	return proto.SunSnapshot{
		Pos:    Vec2ToSnapshot(s.Pos),
		Radius: s.Radius,
	}
}

// PalletToSnapshot converts an entities.Pallet to a proto.PalletSnapshot.
func PalletToSnapshot(p entities.Pallet) proto.PalletSnapshot {
	return proto.PalletSnapshot{
		ID:     p.ID,
		Pos:    Vec2ToSnapshot(p.Pos),
		Active: p.Active,
	}
}

// WorldToSnapshot converts an entities.World to a proto.SnapshotMessage.
// This function bridges the simulation layer with the protocol layer,
// enabling the server to broadcast game state to clients.
func WorldToSnapshot(w entities.World) proto.SnapshotMessage {
	// Convert pallets slice, ensuring empty slice produces empty array (not nil)
	pallets := make([]proto.PalletSnapshot, len(w.Pallets))
	for i, pallet := range w.Pallets {
		pallets[i] = PalletToSnapshot(pallet)
	}

	return proto.SnapshotMessage{
		Type:    "snapshot",
		Tick:    w.Tick,
		Ship:    ShipToSnapshot(w.Ship),
		Sun:     SunToSnapshot(w.Sun),
		Pallets: pallets,
		Done:    w.Done,
		Win:     w.Win,
	}
}


