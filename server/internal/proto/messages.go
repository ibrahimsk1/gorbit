package proto

// InputMessage represents a client input command message.
// Client → Server message format: {"t":"input","seq":u32,"thrust":0..1,"turn":-1..1}
type InputMessage struct {
	Type   string  `json:"t"`     // Message type: "input"
	Seq    uint32  `json:"seq"`   // Sequence number
	Thrust float32 `json:"thrust"` // Thrust input [0.0, 1.0]
	Turn   float32 `json:"turn"`   // Turn input [-1.0, 1.0]
}

// RestartMessage represents a client restart request message.
// Client → Server message format: {"t":"restart"}
type RestartMessage struct {
	Type string `json:"t"` // Message type: "restart"
}

// SnapshotMessage represents a server state snapshot message.
// Server → Client message format with tick, ship, sun, pallets, done, win
type SnapshotMessage struct {
	Type    string          `json:"t"`      // Message type: "snapshot"
	Tick    uint32          `json:"tick"`   // Current simulation tick
	Ship    ShipSnapshot    `json:"ship"`   // Ship state
	Sun     SunSnapshot     `json:"sun"`    // Sun state
	Pallets []PalletSnapshot `json:"pallets"` // List of pallets
	Done    bool            `json:"done"`   // Whether the game is finished
	Win     bool            `json:"win"`    // Whether the player won (only valid if Done is true)
}

// ShipSnapshot represents ship state in a snapshot.
type ShipSnapshot struct {
	Pos    Vec2Snapshot `json:"pos"`    // Position
	Vel    Vec2Snapshot `json:"vel"`    // Velocity
	Rot    float64      `json:"rot"`    // Rotation angle in radians
	Energy float32      `json:"energy"`  // Current energy level
}

// SunSnapshot represents sun state in a snapshot.
type SunSnapshot struct {
	Pos    Vec2Snapshot `json:"pos"`    // Position
	Radius float32      `json:"radius"` // Radius
}

// PalletSnapshot represents a pallet state in a snapshot.
type PalletSnapshot struct {
	ID     uint32       `json:"id"`     // Unique identifier
	Pos    Vec2Snapshot `json:"pos"`    // Position
	Active bool         `json:"active"` // Whether the pallet is active/collectible
}

// Vec2Snapshot represents a 2D vector in a snapshot.
type Vec2Snapshot struct {
	X float64 `json:"x"` // X coordinate
	Y float64 `json:"y"` // Y coordinate
}

