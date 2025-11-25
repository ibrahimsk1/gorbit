package adapter

import (
	"github.com/gorbit/orbitalrush/internal/room"
	"github.com/gorbit/orbitalrush/internal/session"
	"github.com/gorbit/orbitalrush/internal/sim/entities"
	"github.com/gorbit/orbitalrush/internal/sim/rules"
	"github.com/gorbit/orbitalrush/internal/transport"
)

// WireRoomToTransport wires RoomManager to Transport layer by setting up
// the RoomOperations adapter. This function should be called during server
// initialization, before any WebSocket connections are established.
//
// This adapter layer follows clean architecture principles:
// - Transport depends on RoomOperations interface (abstraction)
// - Room package doesn't depend on Transport (avoids circular dependency)
// - Adapter package imports both and connects them (adapter pattern)
//
// Labels: scope:integration loop:g9-scenario layer:adapter
func WireRoomToTransport(roomManager *room.RoomManager) {
	transport.SetRoomOperationsAdapter(func() transport.RoomOperations {
		return transport.RoomOperations{
			CreateRoomFunc: func(conn *transport.Connection) (transport.RoomData, uint32, error) {
				rm, playerID, err := roomManager.CreateRoom(conn)
				if err != nil {
					return transport.RoomData{}, 0, err
				}
				return convertRoomToRoomData(rm), playerID, nil
			},
			JoinRoomFunc: func(roomCode string, conn *transport.Connection) (transport.RoomData, uint32, error) {
				rm, playerID, err := roomManager.JoinRoom(roomCode, conn)
				if err != nil {
					return transport.RoomData{}, 0, err
				}
				return convertRoomToRoomData(rm), playerID, nil
			},
			LeaveRoomFunc: func(roomCode string, playerID uint32) error {
				return roomManager.LeaveRoom(roomCode, playerID)
			},
			GetRoomFunc: func(roomCode string) (transport.RoomData, error) {
				rm, err := roomManager.GetRoom(roomCode)
				if err != nil {
					return transport.RoomData{}, err
				}
				return convertRoomToRoomData(rm), nil
			},
			StartMatchFunc: func(roomCode string, hostPlayerID uint32, clock session.Clock) error {
				return roomManager.StartMatch(roomCode, hostPlayerID, clock)
			},
			EnqueueCommandToRoomFunc: func(roomCode string, playerID uint32, seq uint32, cmd rules.InputCommand) error {
				return roomManager.EnqueueCommandToRoom(roomCode, playerID, seq, cmd)
			},
			GetWorldFromRoomFunc: func(roomCode string) (entities.World, error) {
				return roomManager.GetWorldFromRoom(roomCode)
			},
		}
	})
}

// convertRoomToRoomData converts a room.Room to transport.RoomData.
// This helper function extracts thread-safe data from the room.
func convertRoomToRoomData(rm *room.Room) transport.RoomData {
	// Get thread-safe copies of room data
	players := rm.GetPlayers()
	state := rm.GetState()
	hostPlayerID := rm.GetHostPlayerID()

	// Convert room.PlayerConnection to transport.PlayerData
	playerData := make([]transport.PlayerData, len(players))
	for i, p := range players {
		playerData[i] = transport.PlayerData{
			PlayerID: p.PlayerID,
			Name:     p.Name,
			Conn:     p.Conn,
		}
	}

	return transport.RoomData{
		RoomCode:     rm.RoomCode,
		Players:      playerData,
		State:        state.String(), // Convert RoomState to string
		HostPlayerID: hostPlayerID,
	}
}
