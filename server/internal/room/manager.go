package room

import (
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
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
	if _, err := rand.Read(bytes); err != nil {
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

