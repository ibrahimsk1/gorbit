package room

import (
	"crypto/rand"
	"errors"
	"fmt"
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

