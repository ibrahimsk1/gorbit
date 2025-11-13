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

	if err := ValidateShipSnapshot(&msg.Ship); err != nil {
		return fmt.Errorf("invalid ship: %w", err)
	}

	if err := ValidateSunSnapshot(&msg.Sun); err != nil {
		return fmt.Errorf("invalid sun: %w", err)
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

