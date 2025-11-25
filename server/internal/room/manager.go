package room

import (
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"sync"

	"github.com/gorbit/orbitalrush/internal/sim/entities"
	"github.com/gorbit/orbitalrush/internal/sim/rules"
	"github.com/gorbit/orbitalrush/internal/transport"
)

const (
	// roomCodeLength is the length of generated room codes (6 characters).
	roomCodeLength = 6
	// maxRetries is the maximum number of retries when generating a room code to avoid collisions.
	maxRetries = 10
	// characterSet contains all valid characters for room codes (A-Z, 0-9).
	characterSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	// ErrMaxRetriesExceeded is returned when GenerateRoomCode cannot generate a unique code after maxRetries attempts.
	ErrMaxRetriesExceeded = errors.New("failed to generate unique room code after maximum retries")
	// ErrRoomNotFound is returned when a room with the given code does not exist.
	ErrRoomNotFound = errors.New("room not found")
	// ErrRoomNotInLobby is returned when trying to join a room that is not in lobby state.
	ErrRoomNotInLobby = errors.New("room is not in lobby state")
	// ErrRoomFull is returned when trying to join a room that has reached maximum capacity (8 players).
	ErrRoomFull = errors.New("room is full")
	// ErrPlayerNotFound is returned when trying to leave a room with a player ID that does not exist in the room.
	ErrPlayerNotFound = errors.New("player not found in room")
	// ErrNotHost is returned when trying to start a match with a player ID that is not the room host.
	ErrNotHost = errors.New("player is not the room host")
	// ErrNotEnoughPlayers is returned when trying to start a match with less than 2 players.
	ErrNotEnoughPlayers = errors.New("room must have at least 2 players to start match")
	// ErrSessionNotFound is returned when trying to access a session that does not exist (room in lobby state).
	ErrSessionNotFound = errors.New("session not found (room is not in playing state)")

	// DefaultRoomManager is the singleton instance of RoomManager.
	DefaultRoomManager *RoomManager
	roomManagerOnce    sync.Once
)

// GenerateRoomCode generates a unique 6-character alphanumeric room code.
// It uses cryptographically secure randomness and checks for collisions against the provided rooms map.
// Returns an error if a unique code cannot be generated after maxRetries attempts.
//
// Parameters:
//   - rooms: Map of existing room codes to Room instances (used for collision detection)
//
// Returns:
//   - string: 6-character alphanumeric room code (A-Z, 0-9)
//   - error: Error if unique code cannot be generated (nil on success)
func GenerateRoomCode(rooms map[string]*Room) (string, error) {
	for attempt := 0; attempt < maxRetries; attempt++ {
		code, err := generateRandomCode()
		if err != nil {
			return "", fmt.Errorf("failed to generate random code: %w", err)
		}

		// Check for collision
		if _, exists := rooms[code]; !exists {
			return code, nil
		}
	}

	return "", ErrMaxRetriesExceeded
}

// generateRandomCode generates a single 6-character alphanumeric code using crypto/rand.
func generateRandomCode() (string, error) {
	bytes := make([]byte, roomCodeLength)
	if _, err := cryptorand.Read(bytes); err != nil {
		return "", err
	}

	code := make([]byte, roomCodeLength)
	for i := 0; i < roomCodeLength; i++ {
		// Map random byte (0-255) to character set index (0-35) using modulo
		index := int(bytes[i]) % len(characterSet)
		code[i] = characterSet[index]
	}

	return string(code), nil
}

// RoomManager manages all rooms and provides room operations.
// It uses a singleton pattern to ensure one instance per server.
type RoomManager struct {
	rooms map[string]*Room // Map from room code to room instance
	mu    sync.RWMutex     // Mutex for concurrent access
}

// NewRoomManager returns the singleton RoomManager instance.
// The singleton is initialized on first call.
func NewRoomManager() *RoomManager {
	roomManagerOnce.Do(func() {
		DefaultRoomManager = &RoomManager{
			rooms: make(map[string]*Room),
		}
	})
	return DefaultRoomManager
}

// GetRoom returns the room with the given room code (thread-safe).
// Returns ErrRoomNotFound if the room does not exist.
func (rm *RoomManager) GetRoom(roomCode string) (*Room, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomCode]
	if !exists {
		return nil, ErrRoomNotFound
	}

	return room, nil
}

// CreateRoom creates a new room with a unique room code and adds it to the rooms map.
// The room is created in lobby state with no players.
// Returns the generated room code or an error if room creation fails.
func (rm *RoomManager) CreateRoom() (string, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Generate unique room code
	code, err := GenerateRoomCode(rm.rooms)
	if err != nil {
		return "", fmt.Errorf("failed to generate room code: %w", err)
	}

	// Create new room in lobby state
	room := &Room{
		RoomCode:     code,
		Players:      []*PlayerConnection{},
		State:        RoomStateLobby,
		HostPlayerID: 0, // Will be set when first player joins
		Session:      nil,
	}

	// Add room to map
	rm.rooms[code] = room

	return code, nil
}

// JoinRoom adds a player to a room by room code.
// Validates that the room exists, is in lobby state, and has capacity.
// Assigns a player ID and creates a PlayerConnection.
// Sets the player as host if they are the first player to join.
// Returns the room and assigned player ID, or an error if join fails.
func (rm *RoomManager) JoinRoom(roomCode string, conn *transport.Connection) (*Room, uint32, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Find room
	room, exists := rm.rooms[roomCode]
	if !exists {
		return nil, 0, ErrRoomNotFound
	}

	// Validate room state (must be lobby)
	room.mu.RLock()
	state := room.State
	playerCount := len(room.Players)
	room.mu.RUnlock()

	if state != RoomStateLobby {
		return nil, 0, ErrRoomNotInLobby
	}

	// Check capacity (max 8 players)
	if playerCount >= 8 {
		return nil, 0, ErrRoomFull
	}

	// Assign player ID (incrementing from 1)
	playerID := uint32(playerCount + 1)

	// Create PlayerConnection
	player := &PlayerConnection{
		Conn:     conn,
		PlayerID: playerID,
		Name:     "", // Empty name in v1
	}

	// Add player to room
	room.AddPlayer(player)

	// Set host if first player
	if playerCount == 0 {
		room.SetHostPlayerID(playerID)
	}

	return room, playerID, nil
}

// LeaveRoom removes a player from a room by room code and player ID.
// Closes the player's connection, stops the session if the room becomes empty,
// and removes the room from the rooms map if it becomes empty.
// Returns an error if the room or player is not found.
func (rm *RoomManager) LeaveRoom(roomCode string, playerID uint32) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Find room
	room, exists := rm.rooms[roomCode]
	if !exists {
		return ErrRoomNotFound
	}

	// Find player and get their connection before removing
	room.mu.RLock()
	var playerConn *transport.Connection
	playerFound := false
	for _, player := range room.Players {
		if player.PlayerID == playerID {
			playerConn = player.Conn
			playerFound = true
			break
		}
	}
	roomSession := room.Session
	room.mu.RUnlock()

	if !playerFound {
		return ErrPlayerNotFound
	}

	// Remove player from room
	room.RemovePlayer(playerID)

	// Close player's connection
	if playerConn != nil {
		_ = playerConn.Close() // Ignore close errors
	}

	// Check if room is now empty
	room.mu.RLock()
	isEmpty := len(room.Players) == 0
	room.mu.RUnlock()

	if isEmpty {
		// Stop session if running
		if roomSession != nil {
			roomSession.Stop()
		}

		// Remove room from map
		delete(rm.rooms, roomCode)
	}

	return nil
}

// EnqueueCommandToRoom enqueues a command to a room's session.
// Returns an error if the room or session does not exist.
func (rm *RoomManager) EnqueueCommandToRoom(roomCode string, playerID uint32, seq uint32, cmd rules.InputCommand) error {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Find room
	room, exists := rm.rooms[roomCode]
	if !exists {
		return ErrRoomNotFound
	}

	// Get session
	sess := room.GetSession()
	if sess == nil {
		return ErrSessionNotFound
	}

	// Enqueue command
	_ = sess.EnqueueCommand(seq, playerID, cmd)
	return nil
}

// GetWorldFromRoom gets the current world state from a room's session.
// Returns an error if the room or session does not exist.
func (rm *RoomManager) GetWorldFromRoom(roomCode string) (entities.World, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Find room
	room, exists := rm.rooms[roomCode]
	if !exists {
		return entities.World{}, ErrRoomNotFound
	}

	// Get session
	sess := room.GetSession()
	if sess == nil {
		return entities.World{}, ErrSessionNotFound
	}

	// Get world
	world := sess.GetWorld()
	return world, nil
}

// CleanupEmptyRooms removes all empty rooms (rooms with no players) from the rooms map.
// Stops sessions for empty rooms before removing them.
// Returns the number of rooms cleaned up.
func (rm *RoomManager) CleanupEmptyRooms() int {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	cleanedCount := 0
	roomsToClean := make([]string, 0)

	// First pass: identify empty rooms
	for code, room := range rm.rooms {
		room.mu.RLock()
		isEmpty := len(room.Players) == 0
		sess := room.Session
		room.mu.RUnlock()

		if isEmpty {
			roomsToClean = append(roomsToClean, code)
			// Stop session if exists
			if sess != nil {
				sess.Stop()
			}
		}
	}

	// Second pass: remove empty rooms
	for _, code := range roomsToClean {
		delete(rm.rooms, code)
		cleanedCount++
	}

	return cleanedCount
}

// CheckMatchEnd checks if a match has ended (world.Done == true) and transitions the room to ended state.
// Returns true if the match has ended, false otherwise.
// Returns an error if the room does not exist.
func (rm *RoomManager) CheckMatchEnd(roomCode string) (bool, error) {
	rm.mu.RLock()
	room, exists := rm.rooms[roomCode]
	rm.mu.RUnlock()

	if !exists {
		return false, ErrRoomNotFound
	}

	// Check if room is in playing state
	room.mu.RLock()
	state := room.State
	sess := room.Session
	room.mu.RUnlock()

	if state != RoomStatePlaying || sess == nil {
		return false, nil
	}

	// Check world.Done
	world := sess.GetWorld()
	if !world.Done {
		return false, nil
	}

	// Match has ended, transition to ended state
	room.mu.Lock()
	room.State = RoomStateEnded
	room.mu.Unlock()

	return true, nil
}
