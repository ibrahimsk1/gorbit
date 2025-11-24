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

// PlayerInfo represents player information.
// Used in room state messages: {"id":u32,"name":str}
type PlayerInfo struct {
	ID   uint32 `json:"id"`   // Player identifier
	Name string `json:"name"` // Player display name
}

// CreateRoomMessage represents a client room creation request message.
// Client → Server message format: {"t":"createRoom"}
type CreateRoomMessage struct {
	Type string `json:"t"` // Message type: "createRoom"
}

// JoinRoomMessage represents a client room join request message.
// Client → Server message format: {"t":"joinRoom","roomCode":"ABC123"}
type JoinRoomMessage struct {
	Type     string `json:"t"`        // Message type: "joinRoom"
	RoomCode string `json:"roomCode"` // 6-character room code
}

// LeaveRoomMessage represents a client room leave request message.
// Client → Server message format: {"t":"leaveRoom"}
type LeaveRoomMessage struct {
	Type string `json:"t"` // Message type: "leaveRoom"
}

// StartMatchMessage represents a client match start request message.
// Client → Server message format: {"t":"startMatch"}
type StartMatchMessage struct {
	Type string `json:"t"` // Message type: "startMatch"
}

// RoomCreatedMessage represents a server room creation response message.
// Server → Client message format: {"t":"roomCreated","roomCode":"ABC123"}
type RoomCreatedMessage struct {
	Type     string `json:"t"`        // Message type: "roomCreated"
	RoomCode string `json:"roomCode"` // 6-character room code
}

// RoomStateMessage represents a server room state update message.
// Server → Client message format: {"t":"roomState","roomCode":"ABC123","players":[{"id":u32,"name":str}],"state":"lobby|playing","hostId":u32}
type RoomStateMessage struct {
	Type     string      `json:"t"`        // Message type: "roomState"
	RoomCode string      `json:"roomCode"` // 6-character room code
	Players  []PlayerInfo `json:"players"` // List of players in the room
	State    string      `json:"state"`    // Room state: "lobby" or "playing"
	HostID   uint32      `json:"hostId"`   // Host player identifier
}

// PlayerJoinedMessage represents a server player joined event message.
// Server → Client message format: {"t":"playerJoined","player":{"id":u32,"name":str}}
type PlayerJoinedMessage struct {
	Type   string    `json:"t"`      // Message type: "playerJoined"
	Player PlayerInfo `json:"player"` // Player information
}

// PlayerLeftMessage represents a server player left event message.
// Server → Client message format: {"t":"playerLeft","playerId":u32}
type PlayerLeftMessage struct {
	Type     string `json:"t"`        // Message type: "playerLeft"
	PlayerID uint32 `json:"playerId"` // Player identifier who left
}

// MatchStartedMessage represents a server match started event message.
// Server → Client message format: {"t":"matchStarted"}
type MatchStartedMessage struct {
	Type string `json:"t"` // Message type: "matchStarted"
}

// MatchEndedMessage represents a server match ended event message.
// Server → Client message format: {"t":"matchEnded","winnerId":u32}
type MatchEndedMessage struct {
	Type     string `json:"t"`        // Message type: "matchEnded"
	WinnerID uint32 `json:"winnerId"` // Winner player identifier (optional, may be 0)
}

// WorldBounds represents world boundaries in a snapshot.
// Used in SnapshotMessage: {"width":f64,"height":f64}
type WorldBounds struct {
	Width  float64 `json:"width"`  // World width
	Height float64 `json:"height"` // World height
}

// PlanetSnapshot represents planet state in a snapshot.
// Used in SnapshotMessage: {"id":u32,"pos":{"x":f64,"y":f64},"radius":f32}
type PlanetSnapshot struct {
	ID     uint32       `json:"id"`     // Planet identifier
	Pos    Vec2Snapshot `json:"pos"`    // Position
	Radius float32      `json:"radius"` // Radius
}

// SnapshotMessage represents a server state snapshot message.
// Server → Client message format with tick, ships[], planets[], pallets[], worldBounds, myShipId
type SnapshotMessage struct {
	Type        string            `json:"t"`          // Message type: "snapshot"
	Tick        uint32            `json:"tick"`       // Current simulation tick
	Ships       []ShipSnapshot    `json:"ships"`      // List of ships
	Planets     []PlanetSnapshot  `json:"planets"`    // List of planets
	Pallets     []PalletSnapshot  `json:"pallets"`   // List of pallets
	WorldBounds WorldBounds       `json:"worldBounds"` // World boundaries
	MyShipId    uint32            `json:"myShipId"`   // Player's ship identifier
}

// ShipSnapshot represents ship state in a snapshot.
// Used in SnapshotMessage: {"id":u32,"pos":{"x":f64,"y":f64},"vel":{"x":f64,"y":f64},"rot":f64,"energy":f32}
type ShipSnapshot struct {
	ID     uint32       `json:"id"`     // Ship identifier
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

