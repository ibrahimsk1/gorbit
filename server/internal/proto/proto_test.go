package proto

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProto(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Protocol Messages Suite")
}

var _ = Describe("Protocol Messages", Label("scope:contract", "loop:g4-proto", "layer:contract"), func() {
	Describe("InputMessage", func() {
		It("serializes to JSON matching TDD spec", func() {
			msg := InputMessage{
				Type:   "input",
				Seq:    1,
				Thrust: 0.5,
				Turn:   0.3,
			}

			data, err := json.Marshal(msg)
			Expect(err).NotTo(HaveOccurred())

			expected := `{"t":"input","seq":1,"thrust":0.5,"turn":0.3}`
			Expect(string(data)).To(MatchJSON(expected))
		})

		It("deserializes from valid JSON", func() {
			jsonStr := `{"t":"input","seq":42,"thrust":0.8,"turn":-0.5}`
			var msg InputMessage

			err := json.Unmarshal([]byte(jsonStr), &msg)
			Expect(err).NotTo(HaveOccurred())
			Expect(msg.Type).To(Equal("input"))
			Expect(msg.Seq).To(Equal(uint32(42)))
			Expect(msg.Thrust).To(Equal(float32(0.8)))
			Expect(msg.Turn).To(Equal(float32(-0.5)))
		})

		It("round-trips correctly (serialize → deserialize → serialize)", func() {
			original := InputMessage{
				Type:   "input",
				Seq:    100,
				Thrust: 0.75,
				Turn:   -0.25,
			}

			data, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())

			var roundTripped InputMessage
			err = json.Unmarshal(data, &roundTripped)
			Expect(err).NotTo(HaveOccurred())

			Expect(roundTripped).To(Equal(original))
		})

		It("handles edge case values", func() {
			msg := InputMessage{
				Type:   "input",
				Seq:    0,
				Thrust: 0.0,
				Turn:   0.0,
			}

			data, err := json.Marshal(msg)
			Expect(err).NotTo(HaveOccurred())

			var unmarshaled InputMessage
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).NotTo(HaveOccurred())
			Expect(unmarshaled).To(Equal(msg))
		})
	})

	Describe("RestartMessage", func() {
		It("serializes to JSON matching TDD spec", func() {
			msg := RestartMessage{
				Type: "restart",
			}

			data, err := json.Marshal(msg)
			Expect(err).NotTo(HaveOccurred())

			expected := `{"t":"restart"}`
			Expect(string(data)).To(MatchJSON(expected))
		})

		It("deserializes from valid JSON", func() {
			jsonStr := `{"t":"restart"}`
			var msg RestartMessage

			err := json.Unmarshal([]byte(jsonStr), &msg)
			Expect(err).NotTo(HaveOccurred())
			Expect(msg.Type).To(Equal("restart"))
		})

		It("round-trips correctly", func() {
			original := RestartMessage{
				Type: "restart",
			}

			data, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())

			var roundTripped RestartMessage
			err = json.Unmarshal(data, &roundTripped)
			Expect(err).NotTo(HaveOccurred())

			Expect(roundTripped).To(Equal(original))
		})
	})

	Describe("SnapshotMessage", func() {
		It("serializes to JSON matching TDD spec format", func() {
			msg := SnapshotMessage{
				Type: "snapshot",
				Tick:  42,
				Ship: ShipSnapshot{
					Pos:    Vec2Snapshot{X: 10.5, Y: 20.3},
					Vel:    Vec2Snapshot{X: 1.0, Y: -2.0},
					Rot:    1.57,
					Energy: 75.5,
				},
				Sun: SunSnapshot{
					Pos:    Vec2Snapshot{X: 0.0, Y: 0.0},
					Radius: 5.0,
				},
				Pallets: []PalletSnapshot{
					{ID: 1, Pos: Vec2Snapshot{X: 15.0, Y: 15.0}, Active: true},
					{ID: 2, Pos: Vec2Snapshot{X: -10.0, Y: 10.0}, Active: false},
				},
				Done: false,
				Win:  false,
			}

			data, err := json.Marshal(msg)
			Expect(err).NotTo(HaveOccurred())

			// Verify structure matches TDD spec
			var unmarshaled map[string]interface{}
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).NotTo(HaveOccurred())

			Expect(unmarshaled["t"]).To(Equal("snapshot"))
			Expect(unmarshaled["tick"]).To(BeNumerically("==", 42))
			Expect(unmarshaled["done"]).To(Equal(false))
			Expect(unmarshaled["win"]).To(Equal(false))
		})

		It("deserializes from valid JSON", func() {
			jsonStr := `{
				"t": "snapshot",
				"tick": 100,
				"ship": {
					"pos": {"x": 5.0, "y": 10.0},
					"vel": {"x": 0.5, "y": -0.5},
					"rot": 0.785,
					"energy": 50.0
				},
				"sun": {
					"pos": {"x": 0.0, "y": 0.0},
					"radius": 3.0
				},
				"pallets": [
					{"id": 1, "pos": {"x": 20.0, "y": 20.0}, "active": true}
				],
				"done": true,
				"win": true
			}`
			var msg SnapshotMessage

			err := json.Unmarshal([]byte(jsonStr), &msg)
			Expect(err).NotTo(HaveOccurred())
			Expect(msg.Type).To(Equal("snapshot"))
			Expect(msg.Tick).To(Equal(uint32(100)))
			Expect(msg.Ship.Pos.X).To(Equal(5.0))
			Expect(msg.Ship.Pos.Y).To(Equal(10.0))
			Expect(msg.Ship.Vel.X).To(Equal(0.5))
			Expect(msg.Ship.Vel.Y).To(Equal(-0.5))
			Expect(msg.Ship.Rot).To(Equal(0.785))
			Expect(msg.Ship.Energy).To(Equal(float32(50.0)))
			Expect(msg.Sun.Pos.X).To(Equal(0.0))
			Expect(msg.Sun.Pos.Y).To(Equal(0.0))
			Expect(msg.Sun.Radius).To(Equal(float32(3.0)))
			Expect(len(msg.Pallets)).To(Equal(1))
			Expect(msg.Pallets[0].ID).To(Equal(uint32(1)))
			Expect(msg.Pallets[0].Pos.X).To(Equal(20.0))
			Expect(msg.Pallets[0].Pos.Y).To(Equal(20.0))
			Expect(msg.Pallets[0].Active).To(Equal(true))
			Expect(msg.Done).To(Equal(true))
			Expect(msg.Win).To(Equal(true))
		})

		It("round-trips correctly", func() {
			original := SnapshotMessage{
				Type: "snapshot",
				Tick:  200,
				Ship: ShipSnapshot{
					Pos:    Vec2Snapshot{X: -5.5, Y: 7.3},
					Vel:    Vec2Snapshot{X: -1.2, Y: 0.8},
					Rot:    3.14,
					Energy: 25.75,
				},
				Sun: SunSnapshot{
					Pos:    Vec2Snapshot{X: 1.0, Y: -1.0},
					Radius: 4.5,
				},
				Pallets: []PalletSnapshot{
					{ID: 10, Pos: Vec2Snapshot{X: 30.0, Y: -30.0}, Active: true},
				},
				Done: false,
				Win:  false,
			}

			data, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())

			var roundTripped SnapshotMessage
			err = json.Unmarshal(data, &roundTripped)
			Expect(err).NotTo(HaveOccurred())

			Expect(roundTripped.Type).To(Equal(original.Type))
			Expect(roundTripped.Tick).To(Equal(original.Tick))
			Expect(roundTripped.Ship).To(Equal(original.Ship))
			Expect(roundTripped.Sun).To(Equal(original.Sun))
			Expect(roundTripped.Pallets).To(Equal(original.Pallets))
			Expect(roundTripped.Done).To(Equal(original.Done))
			Expect(roundTripped.Win).To(Equal(original.Win))
		})

		It("handles empty pallets array", func() {
			msg := SnapshotMessage{
				Type:    "snapshot",
				Tick:    1,
				Ship:    ShipSnapshot{},
				Sun:     SunSnapshot{},
				Pallets: []PalletSnapshot{},
				Done:    false,
				Win:     false,
			}

			data, err := json.Marshal(msg)
			Expect(err).NotTo(HaveOccurred())

			var unmarshaled SnapshotMessage
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(unmarshaled.Pallets)).To(Equal(0))
		})
	})

	Describe("Nested Structures", func() {
		It("Vec2Snapshot serializes correctly", func() {
			vec := Vec2Snapshot{X: 12.34, Y: -56.78}
			data, err := json.Marshal(vec)
			Expect(err).NotTo(HaveOccurred())

			expected := `{"x":12.34,"y":-56.78}`
			Expect(string(data)).To(MatchJSON(expected))
		})

		It("ShipSnapshot serializes correctly", func() {
			ship := ShipSnapshot{
				Pos:    Vec2Snapshot{X: 1.0, Y: 2.0},
				Vel:    Vec2Snapshot{X: 3.0, Y: 4.0},
				Rot:    1.5,
				Energy: 80.0,
			}
			data, err := json.Marshal(ship)
			Expect(err).NotTo(HaveOccurred())

			var unmarshaled map[string]interface{}
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).NotTo(HaveOccurred())
			Expect(unmarshaled).To(HaveKey("pos"))
			Expect(unmarshaled).To(HaveKey("vel"))
			Expect(unmarshaled).To(HaveKey("rot"))
			Expect(unmarshaled).To(HaveKey("energy"))
		})

		It("SunSnapshot serializes correctly", func() {
			sun := SunSnapshot{
				Pos:    Vec2Snapshot{X: 0.0, Y: 0.0},
				Radius: 10.0,
			}
			data, err := json.Marshal(sun)
			Expect(err).NotTo(HaveOccurred())

			var unmarshaled map[string]interface{}
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).NotTo(HaveOccurred())
			Expect(unmarshaled).To(HaveKey("pos"))
			Expect(unmarshaled).To(HaveKey("radius"))
		})

		It("PalletSnapshot serializes correctly", func() {
			pallet := PalletSnapshot{
				ID:     42,
				Pos:    Vec2Snapshot{X: 15.0, Y: 25.0},
				Active: true,
			}
			data, err := json.Marshal(pallet)
			Expect(err).NotTo(HaveOccurred())

			expected := `{"id":42,"pos":{"x":15.0,"y":25.0},"active":true}`
			Expect(string(data)).To(MatchJSON(expected))
		})
	})
})

