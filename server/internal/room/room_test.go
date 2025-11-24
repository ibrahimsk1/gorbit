package room

import (
	"testing"

	"github.com/gorbit/orbitalrush/internal/session"
	"github.com/gorbit/orbitalrush/internal/transport"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRoom(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Room Types Suite")
}

var _ = Describe("Room Types", Label("scope:unit", "loop:g6-room", "layer:room", "b:room-types", "r:low", "double:fake", "dep:none"), func() {
	Describe("RoomState", func() {
		It("defines RoomState enum values", func() {
			Expect(RoomStateLobby).To(Equal(RoomState(0)))
			Expect(RoomStatePlaying).To(Equal(RoomState(1)))
			Expect(RoomStateEnded).To(Equal(RoomState(2)))
		})

		It("implements String() method correctly", func() {
			Expect(RoomStateLobby.String()).To(Equal("lobby"))
			Expect(RoomStatePlaying.String()).To(Equal("playing"))
			Expect(RoomStateEnded.String()).To(Equal("ended"))
		})
	})

	Describe("PlayerConnection", func() {
		It("creates PlayerConnection with all fields", func() {
			conn := &transport.Connection{}
			player := &PlayerConnection{
				Conn:     conn,
				PlayerID: 1,
				Name:     "Player1",
			}

			Expect(player.Conn).To(Equal(conn))
			Expect(player.PlayerID).To(Equal(uint32(1)))
			Expect(player.Name).To(Equal("Player1"))
		})

		It("allows empty name", func() {
			player := &PlayerConnection{
				Conn:     &transport.Connection{},
				PlayerID: 2,
				Name:     "",
			}

			Expect(player.Name).To(Equal(""))
		})
	})

	Describe("Room", func() {
		It("creates Room with initial state", func() {
			room := &Room{
				RoomCode: "ABC123",
				Players:  []*PlayerConnection{},
				State:    RoomStateLobby,
			}

			Expect(room.RoomCode).To(Equal("ABC123"))
			Expect(room.Players).To(HaveLen(0))
			Expect(room.GetState()).To(Equal(RoomStateLobby))
			Expect(room.GetSession()).To(BeNil())
		})

		It("thread-safely updates and reads state", func() {
			room := &Room{RoomCode: "TEST01", State: RoomStateLobby}

			room.SetState(RoomStatePlaying)
			Expect(room.GetState()).To(Equal(RoomStatePlaying))

			room.SetState(RoomStateEnded)
			Expect(room.GetState()).To(Equal(RoomStateEnded))
		})

		It("thread-safely manages players", func() {
			room := &Room{
				RoomCode: "TEST02",
				Players:  []*PlayerConnection{},
			}

			player1 := &PlayerConnection{
				Conn:     &transport.Connection{},
				PlayerID: 1,
				Name:     "Player1",
			}
			player2 := &PlayerConnection{
				Conn:     &transport.Connection{},
				PlayerID: 2,
				Name:     "Player2",
			}

			room.AddPlayer(player1)
			players := room.GetPlayers()
			Expect(players).To(HaveLen(1))
			Expect(players[0].PlayerID).To(Equal(uint32(1)))

			room.AddPlayer(player2)
			players = room.GetPlayers()
			Expect(players).To(HaveLen(2))

			room.RemovePlayer(1)
			players = room.GetPlayers()
			Expect(players).To(HaveLen(1))
			Expect(players[0].PlayerID).To(Equal(uint32(2)))
		})

		It("thread-safely manages host player ID", func() {
			room := &Room{RoomCode: "TEST03"}

			room.SetHostPlayerID(5)
			Expect(room.GetHostPlayerID()).To(Equal(uint32(5)))

			room.SetHostPlayerID(10)
			Expect(room.GetHostPlayerID()).To(Equal(uint32(10)))
		})

		It("thread-safely manages session", func() {
			room := &Room{RoomCode: "TEST04"}
			sess := &session.Session{}

			Expect(room.GetSession()).To(BeNil())

			room.SetSession(sess)
			Expect(room.GetSession()).To(Equal(sess))

			room.SetSession(nil)
			Expect(room.GetSession()).To(BeNil())
		})

		It("returns copy of players slice to prevent external modification", func() {
			room := &Room{
				RoomCode: "TEST05",
				Players:  []*PlayerConnection{},
			}

			player := &PlayerConnection{
				Conn:     &transport.Connection{},
				PlayerID: 1,
			}
			room.AddPlayer(player)

			players := room.GetPlayers()
			players = append(players, &PlayerConnection{Conn: &transport.Connection{}, PlayerID: 999})

			// Original room should not be modified
			Expect(room.GetPlayers()).To(HaveLen(1))
			Expect(room.GetPlayers()[0].PlayerID).To(Equal(uint32(1)))
		})
	})
})

