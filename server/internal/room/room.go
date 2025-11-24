package room

import (
	"sync"

	"github.com/gorbit/orbitalrush/internal/session"
	"github.com/gorbit/orbitalrush/internal/transport"
)

// RoomState represents the current state of a room.
type RoomState int

const (
	// RoomStateLobby indicates the room is in lobby state (waiting for players).
	RoomStateLobby RoomState = iota
	// RoomStatePlaying indicates the room is in playing state (match in progress).
	RoomStatePlaying
	// RoomStateEnded indicates the room is in ended state (match finished).
	RoomStateEnded
)

// String returns the string representation of the room state.
func (rs RoomState) String() string {
	switch rs {
	case RoomStateLobby:
		return "lobby"
	case RoomStatePlaying:
		return "playing"
	case RoomStateEnded:
		return "ended"
	default:
		return "unknown"
	}
}

// PlayerConnection represents a player's connection to a room.
type PlayerConnection struct {
	Conn     *transport.Connection // WebSocket connection
	PlayerID uint32                // Unique player identifier
	Name     string                 // Player name (optional, may be empty in v1)
}

// Room represents a game room with players, state, and associated session.
type Room struct {
	RoomCode     string              // 6-character alphanumeric room code (unique identifier)
	Players      []*PlayerConnection // List of connected players
	State        RoomState           // Current room state (lobby, playing, ended)
	HostPlayerID uint32              // Player ID of room host (first player to join)
	Session      *session.Session    // Simulation session (created when match starts, nil in lobby)
	mu           sync.RWMutex        // Mutex for concurrent access
}

// GetState returns the current room state (thread-safe).
func (r *Room) GetState() RoomState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.State
}

// SetState sets the room state (thread-safe).
func (r *Room) SetState(state RoomState) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.State = state
}

// GetPlayers returns a copy of the players list (thread-safe).
func (r *Room) GetPlayers() []*PlayerConnection {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Return a copy to prevent external modification
	players := make([]*PlayerConnection, len(r.Players))
	copy(players, r.Players)
	return players
}

// AddPlayer adds a player to the room (thread-safe).
func (r *Room) AddPlayer(player *PlayerConnection) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Players = append(r.Players, player)
}

// RemovePlayer removes a player from the room by player ID (thread-safe).
func (r *Room) RemovePlayer(playerID uint32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, player := range r.Players {
		if player.PlayerID == playerID {
			// Remove player by swapping with last element and truncating
			r.Players[i] = r.Players[len(r.Players)-1]
			r.Players = r.Players[:len(r.Players)-1]
			return
		}
	}
}

// GetHostPlayerID returns the host player ID (thread-safe).
func (r *Room) GetHostPlayerID() uint32 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.HostPlayerID
}

// SetHostPlayerID sets the host player ID (thread-safe).
func (r *Room) SetHostPlayerID(playerID uint32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.HostPlayerID = playerID
}

// GetSession returns the room's session (thread-safe).
func (r *Room) GetSession() *session.Session {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Session
}

// SetSession sets the room's session (thread-safe).
func (r *Room) SetSession(session *session.Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Session = session
}

