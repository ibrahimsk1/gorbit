package proto

import (
	"fmt"
	"math"
)

// ValidateInputMessage validates an InputMessage.
// Returns an error if the message is invalid.
func ValidateInputMessage(msg *InputMessage) error {
	if msg == nil {
		return fmt.Errorf("input message is nil")
	}

	if msg.Type != "input" {
		return fmt.Errorf("invalid type: expected 'input', got '%s'", msg.Type)
	}

	if msg.Seq == 0 {
		return fmt.Errorf("invalid seq: must be greater than 0")
	}

	if msg.Thrust < 0.0 || msg.Thrust > 1.0 {
		return fmt.Errorf("invalid thrust: must be in range [0.0, 1.0], got %f", msg.Thrust)
	}

	if msg.Turn < -1.0 || msg.Turn > 1.0 {
		return fmt.Errorf("invalid turn: must be in range [-1.0, 1.0], got %f", msg.Turn)
	}

	return nil
}

// ValidateRestartMessage validates a RestartMessage.
// Returns an error if the message is invalid.
func ValidateRestartMessage(msg *RestartMessage) error {
	if msg == nil {
		return fmt.Errorf("restart message is nil")
	}

	if msg.Type != "restart" {
		return fmt.Errorf("invalid type: expected 'restart', got '%s'", msg.Type)
	}

	return nil
}

// ValidateSnapshotMessage validates a SnapshotMessage.
// Returns an error if the message is invalid.
func ValidateSnapshotMessage(msg *SnapshotMessage) error {
	if msg == nil {
		return fmt.Errorf("snapshot message is nil")
	}

	if msg.Type != "snapshot" {
		return fmt.Errorf("invalid type: expected 'snapshot', got '%s'", msg.Type)
	}

	for i, ship := range msg.Ships {
		if err := ValidateShipSnapshot(&ship); err != nil {
			return fmt.Errorf("invalid ship at index %d: %w", i, err)
		}
	}

	for i, planet := range msg.Planets {
		if err := ValidatePlanetSnapshot(&planet); err != nil {
			return fmt.Errorf("invalid planet at index %d: %w", i, err)
		}
	}

	for i, pallet := range msg.Pallets {
		if err := ValidatePalletSnapshot(&pallet); err != nil {
			return fmt.Errorf("invalid pallet at index %d: %w", i, err)
		}
	}

	return nil
}

// ValidateShipSnapshot validates a ShipSnapshot.
// Returns an error if the snapshot is invalid.
func ValidateShipSnapshot(ship *ShipSnapshot) error {
	if ship == nil {
		return fmt.Errorf("ship snapshot is nil")
	}

	if err := ValidateVec2Snapshot(&ship.Pos); err != nil {
		return fmt.Errorf("invalid pos: %w", err)
	}

	if err := ValidateVec2Snapshot(&ship.Vel); err != nil {
		return fmt.Errorf("invalid vel: %w", err)
	}

	if ship.Energy < 0.0 {
		return fmt.Errorf("invalid energy: must be >= 0.0, got %f", ship.Energy)
	}

	return nil
}

// ValidateSunSnapshot validates a SunSnapshot.
// Returns an error if the snapshot is invalid.
func ValidateSunSnapshot(sun *SunSnapshot) error {
	if sun == nil {
		return fmt.Errorf("sun snapshot is nil")
	}

	if err := ValidateVec2Snapshot(&sun.Pos); err != nil {
		return fmt.Errorf("invalid pos: %w", err)
	}

	if sun.Radius <= 0.0 {
		return fmt.Errorf("invalid radius: must be > 0.0, got %f", sun.Radius)
	}

	return nil
}

// ValidatePlanetSnapshot validates a PlanetSnapshot.
// Returns an error if the snapshot is invalid.
func ValidatePlanetSnapshot(planet *PlanetSnapshot) error {
	if planet == nil {
		return fmt.Errorf("planet snapshot is nil")
	}

	if planet.ID == 0 {
		return fmt.Errorf("invalid id: must be greater than 0")
	}

	if err := ValidateVec2Snapshot(&planet.Pos); err != nil {
		return fmt.Errorf("invalid pos: %w", err)
	}

	if planet.Radius <= 0.0 {
		return fmt.Errorf("invalid radius: must be > 0.0, got %f", planet.Radius)
	}

	return nil
}

// ValidatePalletSnapshot validates a PalletSnapshot.
// Returns an error if the snapshot is invalid.
func ValidatePalletSnapshot(pallet *PalletSnapshot) error {
	if pallet == nil {
		return fmt.Errorf("pallet snapshot is nil")
	}

	if pallet.ID == 0 {
		return fmt.Errorf("invalid id: must be greater than 0")
	}

	if err := ValidateVec2Snapshot(&pallet.Pos); err != nil {
		return fmt.Errorf("invalid pos: %w", err)
	}

	return nil
}

// ValidateVec2Snapshot validates a Vec2Snapshot.
// Returns an error if the vector is invalid (contains NaN or Inf).
func ValidateVec2Snapshot(vec *Vec2Snapshot) error {
	if vec == nil {
		return fmt.Errorf("vec2 snapshot is nil")
	}

	if math.IsNaN(vec.X) {
		return fmt.Errorf("invalid x: must be finite, got NaN")
	}

	if math.IsInf(vec.X, 0) {
		return fmt.Errorf("invalid x: must be finite, got Inf")
	}

	if math.IsNaN(vec.Y) {
		return fmt.Errorf("invalid y: must be finite, got NaN")
	}

	if math.IsInf(vec.Y, 0) {
		return fmt.Errorf("invalid y: must be finite, got Inf")
	}

	return nil
}

// ValidatePlayerInfo validates a PlayerInfo.
// Returns an error if the player info is invalid.
func ValidatePlayerInfo(player *PlayerInfo) error {
	if player == nil {
		return fmt.Errorf("player info is nil")
	}

	if player.ID == 0 {
		return fmt.Errorf("invalid id: must be greater than 0")
	}

	if player.Name == "" {
		return fmt.Errorf("invalid name: must not be empty")
	}

	return nil
}

// ValidateCreateRoomMessage validates a CreateRoomMessage.
// Returns an error if the message is invalid.
func ValidateCreateRoomMessage(msg *CreateRoomMessage) error {
	if msg == nil {
		return fmt.Errorf("create room message is nil")
	}

	if msg.Type != "createRoom" {
		return fmt.Errorf("invalid type: expected 'createRoom', got '%s'", msg.Type)
	}

	return nil
}

// ValidateJoinRoomMessage validates a JoinRoomMessage.
// Returns an error if the message is invalid.
func ValidateJoinRoomMessage(msg *JoinRoomMessage) error {
	if msg == nil {
		return fmt.Errorf("join room message is nil")
	}

	if msg.Type != "joinRoom" {
		return fmt.Errorf("invalid type: expected 'joinRoom', got '%s'", msg.Type)
	}

	if err := validateRoomCode(msg.RoomCode); err != nil {
		return fmt.Errorf("invalid roomCode: %w", err)
	}

	return nil
}

// ValidateLeaveRoomMessage validates a LeaveRoomMessage.
// Returns an error if the message is invalid.
func ValidateLeaveRoomMessage(msg *LeaveRoomMessage) error {
	if msg == nil {
		return fmt.Errorf("leave room message is nil")
	}

	if msg.Type != "leaveRoom" {
		return fmt.Errorf("invalid type: expected 'leaveRoom', got '%s'", msg.Type)
	}

	return nil
}

// ValidateStartMatchMessage validates a StartMatchMessage.
// Returns an error if the message is invalid.
func ValidateStartMatchMessage(msg *StartMatchMessage) error {
	if msg == nil {
		return fmt.Errorf("start match message is nil")
	}

	if msg.Type != "startMatch" {
		return fmt.Errorf("invalid type: expected 'startMatch', got '%s'", msg.Type)
	}

	return nil
}

// ValidateRoomCreatedMessage validates a RoomCreatedMessage.
// Returns an error if the message is invalid.
func ValidateRoomCreatedMessage(msg *RoomCreatedMessage) error {
	if msg == nil {
		return fmt.Errorf("room created message is nil")
	}

	if msg.Type != "roomCreated" {
		return fmt.Errorf("invalid type: expected 'roomCreated', got '%s'", msg.Type)
	}

	if err := validateRoomCode(msg.RoomCode); err != nil {
		return fmt.Errorf("invalid roomCode: %w", err)
	}

	return nil
}

// ValidateRoomStateMessage validates a RoomStateMessage.
// Returns an error if the message is invalid.
func ValidateRoomStateMessage(msg *RoomStateMessage) error {
	if msg == nil {
		return fmt.Errorf("room state message is nil")
	}

	if msg.Type != "roomState" {
		return fmt.Errorf("invalid type: expected 'roomState', got '%s'", msg.Type)
	}

	if err := validateRoomCode(msg.RoomCode); err != nil {
		return fmt.Errorf("invalid roomCode: %w", err)
	}

	if msg.State != "lobby" && msg.State != "playing" {
		return fmt.Errorf("invalid state: must be 'lobby' or 'playing', got '%s'", msg.State)
	}

	if msg.HostID == 0 {
		return fmt.Errorf("invalid hostId: must be greater than 0")
	}

	for i, player := range msg.Players {
		if err := ValidatePlayerInfo(&player); err != nil {
			return fmt.Errorf("invalid player at index %d: %w", i, err)
		}
	}

	return nil
}

// ValidatePlayerJoinedMessage validates a PlayerJoinedMessage.
// Returns an error if the message is invalid.
func ValidatePlayerJoinedMessage(msg *PlayerJoinedMessage) error {
	if msg == nil {
		return fmt.Errorf("player joined message is nil")
	}

	if msg.Type != "playerJoined" {
		return fmt.Errorf("invalid type: expected 'playerJoined', got '%s'", msg.Type)
	}

	if err := ValidatePlayerInfo(&msg.Player); err != nil {
		return fmt.Errorf("invalid player: %w", err)
	}

	return nil
}

// ValidatePlayerLeftMessage validates a PlayerLeftMessage.
// Returns an error if the message is invalid.
func ValidatePlayerLeftMessage(msg *PlayerLeftMessage) error {
	if msg == nil {
		return fmt.Errorf("player left message is nil")
	}

	if msg.Type != "playerLeft" {
		return fmt.Errorf("invalid type: expected 'playerLeft', got '%s'", msg.Type)
	}

	if msg.PlayerID == 0 {
		return fmt.Errorf("invalid playerId: must be greater than 0")
	}

	return nil
}

// ValidateMatchStartedMessage validates a MatchStartedMessage.
// Returns an error if the message is invalid.
func ValidateMatchStartedMessage(msg *MatchStartedMessage) error {
	if msg == nil {
		return fmt.Errorf("match started message is nil")
	}

	if msg.Type != "matchStarted" {
		return fmt.Errorf("invalid type: expected 'matchStarted', got '%s'", msg.Type)
	}

	return nil
}

// ValidateMatchEndedMessage validates a MatchEndedMessage.
// Returns an error if the message is invalid.
func ValidateMatchEndedMessage(msg *MatchEndedMessage) error {
	if msg == nil {
		return fmt.Errorf("match ended message is nil")
	}

	if msg.Type != "matchEnded" {
		return fmt.Errorf("invalid type: expected 'matchEnded', got '%s'", msg.Type)
	}

	// WinnerID is optional (0 is valid for draw/no winner), so no validation needed

	return nil
}

// validateRoomCode validates a room code.
// Room code must be exactly 6 alphanumeric characters (A-Z, a-z, 0-9).
func validateRoomCode(code string) error {
	if len(code) != 6 {
		return fmt.Errorf("must be exactly 6 characters, got %d", len(code))
	}

	for _, char := range code {
		if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
			return fmt.Errorf("must contain only alphanumeric characters (A-Z, a-z, 0-9), got '%c'", char)
		}
	}

	return nil
}

