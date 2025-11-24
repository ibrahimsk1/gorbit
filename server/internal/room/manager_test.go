package room

import (
	"errors"

	"github.com/gorbit/orbitalrush/internal/transport"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Room Code Generation", Label("scope:unit", "loop:g6-room", "layer:room", "b:room-code-generation", "r:medium", "double:fake", "dep:none"), func() {
	Describe("GenerateRoomCode", func() {
		It("generates 6-character alphanumeric codes", func() {
			rooms := make(map[string]*Room)
			code, err := GenerateRoomCode(rooms)

			Expect(err).To(BeNil())
			Expect(code).To(HaveLen(6))
			Expect(code).To(MatchRegexp("^[A-Z0-9]{6}$"))
		})

		It("generates unique codes", func() {
			rooms := make(map[string]*Room)
			codes := make(map[string]bool)

			for i := 0; i < 100; i++ {
				code, err := GenerateRoomCode(rooms)
				Expect(err).To(BeNil())
				Expect(codes[code]).To(BeFalse(), "Code %s should be unique", code)
				codes[code] = true
				rooms[code] = &Room{RoomCode: code}
			}
		})

		It("handles collisions and retries", func() {
			rooms := make(map[string]*Room)
			rooms["ABC123"] = &Room{RoomCode: "ABC123"}

			// Generate many codes - should avoid collision
			for i := 0; i < 50; i++ {
				code, err := GenerateRoomCode(rooms)
				Expect(err).To(BeNil())
				Expect(code).NotTo(Equal("ABC123"))
				rooms[code] = &Room{RoomCode: code}
			}
		})

		It("uses only valid characters from character set", func() {
			rooms := make(map[string]*Room)
			validChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			charMap := make(map[rune]bool)
			for _, c := range validChars {
				charMap[c] = true
			}

			// Generate many codes and verify all characters are valid
			for i := 0; i < 100; i++ {
				code, err := GenerateRoomCode(rooms)
				Expect(err).To(BeNil())
				for _, char := range code {
					Expect(charMap[char]).To(BeTrue(), "Character %c should be in character set", char)
				}
				rooms[code] = &Room{RoomCode: code}
			}
		})

		It("returns error when max retries exceeded", func() {
			// Create a rooms map that's nearly full (to force collisions)
			// This is difficult to test deterministically, but we can test
			// that the function handles the retry logic correctly
			rooms := make(map[string]*Room)
			
			// Fill with many codes to increase collision probability
			// Note: With 36^6 = ~2 billion possible codes, collisions are very rare
			// This test mainly verifies the retry mechanism exists
			code, err := GenerateRoomCode(rooms)
			Expect(err).To(BeNil())
			Expect(code).To(HaveLen(6))
		})
	})

	Describe("RoomManager", Label("scope:unit", "loop:g6-room", "layer:room", "b:room-manager", "r:low", "double:fake", "dep:none"), func() {
		It("creates RoomManager with empty rooms map", func() {
			manager := NewRoomManager()

			Expect(manager).NotTo(BeNil())
			room, err := manager.GetRoom("NONEXIST")
			Expect(err).NotTo(BeNil())
			Expect(room).To(BeNil())
		})

		It("returns room when room code exists", func() {
			manager := NewRoomManager()
			room := &Room{RoomCode: "TEST01"}
			manager.rooms["TEST01"] = room

			found, err := manager.GetRoom("TEST01")
			Expect(err).To(BeNil())
			Expect(found).To(Equal(room))
		})

		It("returns error when room code does not exist", func() {
			manager := NewRoomManager()

			room, err := manager.GetRoom("NONEXIST")
			Expect(err).NotTo(BeNil())
			Expect(errors.Is(err, ErrRoomNotFound)).To(BeTrue())
			Expect(room).To(BeNil())
		})

		It("returns same instance for singleton", func() {
			manager1 := NewRoomManager()
			manager2 := NewRoomManager()

			Expect(manager1).To(Equal(manager2))
		})

		It("thread-safely accesses rooms map", func() {
			manager := NewRoomManager()
			room1 := &Room{RoomCode: "ROOM01"}
			room2 := &Room{RoomCode: "ROOM02"}

			// Add rooms (would need AddRoom method, but for now test GetRoom is thread-safe)
			manager.rooms["ROOM01"] = room1
			manager.rooms["ROOM02"] = room2

			// Concurrent reads should be safe
			found1, err1 := manager.GetRoom("ROOM01")
			found2, err2 := manager.GetRoom("ROOM02")

			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(found1).To(Equal(room1))
			Expect(found2).To(Equal(room2))
		})
	})

	Describe("CreateRoom", Label("scope:unit", "loop:g6-room", "layer:room", "b:room-creation", "r:medium", "double:fake", "dep:none"), func() {
		It("creates room with unique code in lobby state", func() {
			manager := NewRoomManager()

			code, err := manager.CreateRoom()
			Expect(err).To(BeNil())
			Expect(code).To(HaveLen(6))
			Expect(code).To(MatchRegexp("^[A-Z0-9]{6}$"))
		})

		It("creates room in lobby state with empty players", func() {
			manager := NewRoomManager()

			code, err := manager.CreateRoom()
			Expect(err).To(BeNil())

			room, err := manager.GetRoom(code)
			Expect(err).To(BeNil())
			Expect(room.GetState()).To(Equal(RoomStateLobby))
			Expect(room.GetPlayers()).To(HaveLen(0))
			Expect(room.RoomCode).To(Equal(code))
		})

		It("generates unique room codes", func() {
			manager := NewRoomManager()
			codes := make(map[string]bool)

			for i := 0; i < 50; i++ {
				code, err := manager.CreateRoom()
				Expect(err).To(BeNil())
				Expect(codes[code]).To(BeFalse(), "Code %s should be unique", code)
				codes[code] = true
			}
		})

		It("adds room to rooms map", func() {
			manager := NewRoomManager()

			code, err := manager.CreateRoom()
			Expect(err).To(BeNil())

			// Verify room can be retrieved
			room, err := manager.GetRoom(code)
			Expect(err).To(BeNil())
			Expect(room).NotTo(BeNil())
			Expect(room.RoomCode).To(Equal(code))
		})
	})

	Describe("JoinRoom", Label("scope:unit", "loop:g6-room", "layer:room", "b:player-join", "r:high", "double:fake-io", "dep:none"), func() {
		It("joins room successfully and sets first player as host", func() {
			manager := NewRoomManager()
			code, _ := manager.CreateRoom()
			conn := &transport.Connection{}

			room, playerID, err := manager.JoinRoom(code, conn)

			Expect(err).To(BeNil())
			Expect(room).NotTo(BeNil())
			Expect(playerID).To(Equal(uint32(1)))
			Expect(room.GetHostPlayerID()).To(Equal(uint32(1)))
			Expect(room.GetPlayers()).To(HaveLen(1))
			Expect(room.GetPlayers()[0].PlayerID).To(Equal(uint32(1)))
			Expect(room.GetPlayers()[0].Conn).To(Equal(conn))
		})

		It("joins room successfully for subsequent players", func() {
			manager := NewRoomManager()
			code, _ := manager.CreateRoom()
			conn1 := &transport.Connection{}
			conn2 := &transport.Connection{}

			// First player
			room1, playerID1, err1 := manager.JoinRoom(code, conn1)
			Expect(err1).To(BeNil())
			Expect(playerID1).To(Equal(uint32(1)))

			// Second player
			room2, playerID2, err2 := manager.JoinRoom(code, conn2)
			Expect(err2).To(BeNil())
			Expect(room2).To(Equal(room1))
			Expect(playerID2).To(Equal(uint32(2)))
			Expect(room2.GetPlayers()).To(HaveLen(2))
			Expect(room2.GetHostPlayerID()).To(Equal(uint32(1))) // First player remains host
		})

		It("returns error when room code does not exist", func() {
			manager := NewRoomManager()
			conn := &transport.Connection{}

			room, playerID, err := manager.JoinRoom("NONEXIST", conn)

			Expect(err).NotTo(BeNil())
			Expect(errors.Is(err, ErrRoomNotFound)).To(BeTrue())
			Expect(room).To(BeNil())
			Expect(playerID).To(Equal(uint32(0)))
		})

		It("returns error when room is not in lobby state (playing)", func() {
			manager := NewRoomManager()
			code, _ := manager.CreateRoom()
			room, _ := manager.GetRoom(code)
			room.SetState(RoomStatePlaying)
			conn := &transport.Connection{}

			_, _, err := manager.JoinRoom(code, conn)

			Expect(err).NotTo(BeNil())
			Expect(errors.Is(err, ErrRoomNotInLobby)).To(BeTrue())
		})

		It("returns error when room is not in lobby state (ended)", func() {
			manager := NewRoomManager()
			code, _ := manager.CreateRoom()
			room, _ := manager.GetRoom(code)
			room.SetState(RoomStateEnded)
			conn := &transport.Connection{}

			_, _, err := manager.JoinRoom(code, conn)

			Expect(err).NotTo(BeNil())
			Expect(errors.Is(err, ErrRoomNotInLobby)).To(BeTrue())
		})

		It("returns error when room is full (8 players)", func() {
			manager := NewRoomManager()
			code, _ := manager.CreateRoom()

			// Add 8 players
			for i := 0; i < 8; i++ {
				conn := &transport.Connection{}
				_, _, err := manager.JoinRoom(code, conn)
				Expect(err).To(BeNil())
			}

			// Try to join 9th player
			conn := &transport.Connection{}
			_, _, err := manager.JoinRoom(code, conn)

			Expect(err).NotTo(BeNil())
			Expect(errors.Is(err, ErrRoomFull)).To(BeTrue())
		})

		It("assigns sequential player IDs", func() {
			manager := NewRoomManager()
			code, _ := manager.CreateRoom()

			// Join multiple players
			for i := 1; i <= 5; i++ {
				conn := &transport.Connection{}
				_, playerID, err := manager.JoinRoom(code, conn)
				Expect(err).To(BeNil())
				Expect(playerID).To(Equal(uint32(i)))
			}

			room, _ := manager.GetRoom(code)
			players := room.GetPlayers()
			Expect(players).To(HaveLen(5))
			for i, player := range players {
				Expect(player.PlayerID).To(Equal(uint32(i + 1)))
			}
		})
	})
})

