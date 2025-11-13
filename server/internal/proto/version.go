package proto

import (
	"fmt"
	"strconv"
	"strings"
)

// ProtocolVersion represents a protocol version string.
// Format: "v1", "v2", etc. where the number is the major version.
// Breaking changes increment the major version.
// Non-breaking changes stay on the same major version.
type ProtocolVersion string

// ProtocolVersionV1 is the current protocol version.
// This version defines the initial protocol contract with InputMessage,
// RestartMessage, and SnapshotMessage types.
const ProtocolVersionV1 ProtocolVersion = "v1"

// ParseVersion parses a version string and returns a ProtocolVersion.
// Valid format: "v" followed by a positive integer (e.g., "v1", "v2").
// Returns an error if the version string is invalid.
func ParseVersion(versionStr string) (ProtocolVersion, error) {
	if versionStr == "" {
		return "", fmt.Errorf("version string cannot be empty")
	}

	if !strings.HasPrefix(versionStr, "v") {
		return "", fmt.Errorf("version must start with 'v', got '%s'", versionStr)
	}

	// Extract the number part after "v"
	numStr := versionStr[1:]
	if numStr == "" {
		return "", fmt.Errorf("version must include a number after 'v', got '%s'", versionStr)
	}

	// Parse the number
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return "", fmt.Errorf("version number must be a valid integer, got '%s': %w", numStr, err)
	}

	if num <= 0 {
		return "", fmt.Errorf("version number must be positive, got %d", num)
	}

	return ProtocolVersion(versionStr), nil
}

// IsCompatible checks if two protocol versions are compatible.
// Versions are compatible if they have the same major version number.
// Different major versions indicate breaking changes and are incompatible.
//
// Examples:
//   - IsCompatible("v1", "v1") -> true (same version)
//   - IsCompatible("v1", "v2") -> false (different major versions)
//   - IsCompatible("v2", "v1") -> false (different major versions)
func IsCompatible(clientVersion, serverVersion ProtocolVersion) bool {
	return clientVersion == serverVersion
}

// CompareVersion compares two protocol versions.
// Returns:
//   - -1 if v1 < v2
//   - 0 if v1 == v2
//   - 1 if v1 > v2
//
// Comparison is based on the numeric major version.
func CompareVersion(v1, v2 ProtocolVersion) int {
	// Extract major version numbers
	num1, err1 := extractMajorVersion(v1)
	num2, err2 := extractMajorVersion(v2)

	// If either version is invalid, treat as incomparable
	// In practice, versions should be validated before comparison
	if err1 != nil || err2 != nil {
		// For invalid versions, return 0 (equal) as a safe default
		// In production, versions should be validated first
		return 0
	}

	if num1 < num2 {
		return -1
	}
	if num1 > num2 {
		return 1
	}
	return 0
}

// extractMajorVersion extracts the major version number from a ProtocolVersion.
// Returns the numeric major version and an error if the version is invalid.
func extractMajorVersion(v ProtocolVersion) (int, error) {
	versionStr := string(v)
	if !strings.HasPrefix(versionStr, "v") {
		return 0, fmt.Errorf("version must start with 'v'")
	}

	numStr := versionStr[1:]
	if numStr == "" {
		return 0, fmt.Errorf("version must include a number after 'v'")
	}

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("version number must be a valid integer: %w", err)
	}

	return num, nil
}

