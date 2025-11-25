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
		ID:     s.ID,
		Pos:    Vec2ToSnapshot(s.Pos),
		Vel:    Vec2ToSnapshot(s.Vel),
		Rot:    s.Rot,
		Energy: s.Energy,
	}
}

// PlanetToSnapshot converts an entities.Planet to a proto.PlanetSnapshot.
// Note: The Mass field from entities.Planet is not included in proto.PlanetSnapshot
// as it is only used for simulation calculations.
func PlanetToSnapshot(p entities.Planet) proto.PlanetSnapshot {
	return proto.PlanetSnapshot{
		ID:     p.ID,
		Pos:    Vec2ToSnapshot(p.Pos),
		Radius: p.Radius,
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
// playerID is used to set myShipId in the snapshot (identifies which ship belongs to this player).
func WorldToSnapshot(w entities.World, playerID uint32) proto.SnapshotMessage {
	// Convert ships slice, ensuring empty slice produces empty array (not nil)
	ships := make([]proto.ShipSnapshot, len(w.Ships))
	for i, ship := range w.Ships {
		ships[i] = ShipToSnapshot(ship)
	}

	// Convert planets slice, ensuring empty slice produces empty array (not nil)
	planets := make([]proto.PlanetSnapshot, len(w.Planets))
	for i, planet := range w.Planets {
		planets[i] = PlanetToSnapshot(planet)
	}

	// Convert pallets slice, ensuring empty slice produces empty array (not nil)
	pallets := make([]proto.PalletSnapshot, len(w.Pallets))
	for i, pallet := range w.Pallets {
		pallets[i] = PalletToSnapshot(pallet)
	}

	return proto.SnapshotMessage{
		Type:        "snapshot",
		Tick:        w.Tick,
		Done:        false, // Match is not finished yet
		Win:         false, // Not applicable when done=false
		Ships:       ships,
		Planets:     planets,
		Pallets:     pallets,
		WorldBounds: proto.WorldBounds{
			Width:  entities.WORLD_WIDTH,
			Height: entities.WORLD_HEIGHT,
		},
		MyShipId: playerID,
	}
}


