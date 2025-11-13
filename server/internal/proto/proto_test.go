package proto

import (
	"encoding/json"
	"math"
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

	Describe("Message Validation", Label("scope:contract", "loop:g4-proto", "layer:contract"), func() {
		Describe("ValidateInputMessage", func() {
			It("accepts valid messages", func() {
				msg := &InputMessage{
					Type:   "input",
					Seq:    1,
					Thrust: 0.5,
					Turn:   0.3,
				}
				err := ValidateInputMessage(msg)
				Expect(err).NotTo(HaveOccurred())
			})

			It("accepts boundary values", func() {
				msg1 := &InputMessage{Type: "input", Seq: 1, Thrust: 0.0, Turn: -1.0}
				err1 := ValidateInputMessage(msg1)
				Expect(err1).NotTo(HaveOccurred())

				msg2 := &InputMessage{Type: "input", Seq: 1, Thrust: 1.0, Turn: 1.0}
				err2 := ValidateInputMessage(msg2)
				Expect(err2).NotTo(HaveOccurred())
			})

			It("rejects invalid type", func() {
				msg := &InputMessage{Type: "invalid", Seq: 1, Thrust: 0.5, Turn: 0.3}
				err := ValidateInputMessage(msg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("type"))
			})

			It("rejects seq = 0", func() {
				msg := &InputMessage{Type: "input", Seq: 0, Thrust: 0.5, Turn: 0.3}
				err := ValidateInputMessage(msg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("seq"))
			})

			It("rejects thrust < 0.0", func() {
				msg := &InputMessage{Type: "input", Seq: 1, Thrust: -0.1, Turn: 0.3}
				err := ValidateInputMessage(msg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("thrust"))
			})

			It("rejects thrust > 1.0", func() {
				msg := &InputMessage{Type: "input", Seq: 1, Thrust: 1.1, Turn: 0.3}
				err := ValidateInputMessage(msg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("thrust"))
			})

			It("rejects turn < -1.0", func() {
				msg := &InputMessage{Type: "input", Seq: 1, Thrust: 0.5, Turn: -1.1}
				err := ValidateInputMessage(msg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("turn"))
			})

			It("rejects turn > 1.0", func() {
				msg := &InputMessage{Type: "input", Seq: 1, Thrust: 0.5, Turn: 1.1}
				err := ValidateInputMessage(msg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("turn"))
			})
		})

		Describe("ValidateRestartMessage", func() {
			It("accepts valid messages", func() {
				msg := &RestartMessage{Type: "restart"}
				err := ValidateRestartMessage(msg)
				Expect(err).NotTo(HaveOccurred())
			})

			It("rejects invalid type", func() {
				msg := &RestartMessage{Type: "invalid"}
				err := ValidateRestartMessage(msg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("type"))
			})
		})

		Describe("ValidateSnapshotMessage", func() {
			It("accepts valid messages", func() {
				msg := &SnapshotMessage{
					Type: "snapshot",
					Tick:  1,
					Ship: ShipSnapshot{
						Pos:    Vec2Snapshot{X: 0.0, Y: 0.0},
						Vel:    Vec2Snapshot{X: 0.0, Y: 0.0},
						Rot:    0.0,
						Energy: 100.0,
					},
					Sun: SunSnapshot{
						Pos:    Vec2Snapshot{X: 0.0, Y: 0.0},
						Radius: 5.0,
					},
					Pallets: []PalletSnapshot{},
					Done:    false,
					Win:     false,
				}
				err := ValidateSnapshotMessage(msg)
				Expect(err).NotTo(HaveOccurred())
			})

			It("rejects invalid type", func() {
				msg := &SnapshotMessage{Type: "invalid"}
				err := ValidateSnapshotMessage(msg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("type"))
			})

			It("validates nested ship structure", func() {
				msg := &SnapshotMessage{
					Type: "snapshot",
					Tick:  1,
					Ship: ShipSnapshot{
						Pos:    Vec2Snapshot{X: math.NaN(), Y: 0.0},
						Vel:    Vec2Snapshot{X: 0.0, Y: 0.0},
						Rot:    0.0,
						Energy: 100.0,
					},
					Sun: SunSnapshot{
						Pos:    Vec2Snapshot{X: 0.0, Y: 0.0},
						Radius: 5.0,
					},
					Pallets: []PalletSnapshot{},
				}
				err := ValidateSnapshotMessage(msg)
				Expect(err).To(HaveOccurred())
			})

			It("validates nested sun structure", func() {
				msg := &SnapshotMessage{
					Type: "snapshot",
					Tick:  1,
					Ship: ShipSnapshot{
						Pos:    Vec2Snapshot{X: 0.0, Y: 0.0},
						Vel:    Vec2Snapshot{X: 0.0, Y: 0.0},
						Rot:    0.0,
						Energy: 100.0,
					},
					Sun: SunSnapshot{
						Pos:    Vec2Snapshot{X: 0.0, Y: 0.0},
						Radius: 0.0, // Invalid: radius must be > 0
					},
					Pallets: []PalletSnapshot{},
				}
				err := ValidateSnapshotMessage(msg)
				Expect(err).To(HaveOccurred())
			})

			It("validates pallets array", func() {
				msg := &SnapshotMessage{
					Type: "snapshot",
					Tick:  1,
					Ship: ShipSnapshot{
						Pos:    Vec2Snapshot{X: 0.0, Y: 0.0},
						Vel:    Vec2Snapshot{X: 0.0, Y: 0.0},
						Rot:    0.0,
						Energy: 100.0,
					},
					Sun: SunSnapshot{
						Pos:    Vec2Snapshot{X: 0.0, Y: 0.0},
						Radius: 5.0,
					},
					Pallets: []PalletSnapshot{
						{ID: 0, Pos: Vec2Snapshot{X: 10.0, Y: 10.0}, Active: true}, // Invalid: ID must be > 0
					},
				}
				err := ValidateSnapshotMessage(msg)
				Expect(err).To(HaveOccurred())
			})
		})

		Describe("ValidateShipSnapshot", func() {
			It("accepts valid ship snapshots", func() {
				ship := &ShipSnapshot{
					Pos:    Vec2Snapshot{X: 10.0, Y: 20.0},
					Vel:    Vec2Snapshot{X: 1.0, Y: -1.0},
					Rot:    1.57,
					Energy: 75.5,
				}
				err := ValidateShipSnapshot(ship)
				Expect(err).NotTo(HaveOccurred())
			})

			It("rejects invalid position (NaN)", func() {
				ship := &ShipSnapshot{
					Pos:    Vec2Snapshot{X: math.NaN(), Y: 20.0},
					Vel:    Vec2Snapshot{X: 1.0, Y: -1.0},
					Rot:    1.57,
					Energy: 75.5,
				}
				err := ValidateShipSnapshot(ship)
				Expect(err).To(HaveOccurred())
			})

			It("rejects invalid velocity (Inf)", func() {
				ship := &ShipSnapshot{
					Pos:    Vec2Snapshot{X: 10.0, Y: 20.0},
					Vel:    Vec2Snapshot{X: math.Inf(1), Y: -1.0},
					Rot:    1.57,
					Energy: 75.5,
				}
				err := ValidateShipSnapshot(ship)
				Expect(err).To(HaveOccurred())
			})

			It("rejects negative energy", func() {
				ship := &ShipSnapshot{
					Pos:    Vec2Snapshot{X: 10.0, Y: 20.0},
					Vel:    Vec2Snapshot{X: 1.0, Y: -1.0},
					Rot:    1.57,
					Energy: -10.0,
				}
				err := ValidateShipSnapshot(ship)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("energy"))
			})

			It("accepts zero energy", func() {
				ship := &ShipSnapshot{
					Pos:    Vec2Snapshot{X: 10.0, Y: 20.0},
					Vel:    Vec2Snapshot{X: 1.0, Y: -1.0},
					Rot:    1.57,
					Energy: 0.0,
				}
				err := ValidateShipSnapshot(ship)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("ValidateSunSnapshot", func() {
			It("accepts valid sun snapshots", func() {
				sun := &SunSnapshot{
					Pos:    Vec2Snapshot{X: 0.0, Y: 0.0},
					Radius: 5.0,
				}
				err := ValidateSunSnapshot(sun)
				Expect(err).NotTo(HaveOccurred())
			})

			It("rejects invalid position (NaN)", func() {
				sun := &SunSnapshot{
					Pos:    Vec2Snapshot{X: math.NaN(), Y: 0.0},
					Radius: 5.0,
				}
				err := ValidateSunSnapshot(sun)
				Expect(err).To(HaveOccurred())
			})

			It("rejects zero radius", func() {
				sun := &SunSnapshot{
					Pos:    Vec2Snapshot{X: 0.0, Y: 0.0},
					Radius: 0.0,
				}
				err := ValidateSunSnapshot(sun)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("radius"))
			})

			It("rejects negative radius", func() {
				sun := &SunSnapshot{
					Pos:    Vec2Snapshot{X: 0.0, Y: 0.0},
					Radius: -1.0,
				}
				err := ValidateSunSnapshot(sun)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("radius"))
			})
		})

		Describe("ValidatePalletSnapshot", func() {
			It("accepts valid pallet snapshots", func() {
				pallet := &PalletSnapshot{
					ID:     1,
					Pos:    Vec2Snapshot{X: 10.0, Y: 20.0},
					Active: true,
				}
				err := ValidatePalletSnapshot(pallet)
				Expect(err).NotTo(HaveOccurred())
			})

			It("rejects ID = 0", func() {
				pallet := &PalletSnapshot{
					ID:     0,
					Pos:    Vec2Snapshot{X: 10.0, Y: 20.0},
					Active: true,
				}
				err := ValidatePalletSnapshot(pallet)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("id"))
			})

			It("rejects invalid position (NaN)", func() {
				pallet := &PalletSnapshot{
					ID:     1,
					Pos:    Vec2Snapshot{X: math.NaN(), Y: 20.0},
					Active: true,
				}
				err := ValidatePalletSnapshot(pallet)
				Expect(err).To(HaveOccurred())
			})

			It("rejects invalid position (Inf)", func() {
				pallet := &PalletSnapshot{
					ID:     1,
					Pos:    Vec2Snapshot{X: 10.0, Y: math.Inf(-1)},
					Active: true,
				}
				err := ValidatePalletSnapshot(pallet)
				Expect(err).To(HaveOccurred())
			})
		})

		Describe("ValidateVec2Snapshot", func() {
			It("accepts valid vectors", func() {
				vec := &Vec2Snapshot{X: 10.5, Y: -20.3}
				err := ValidateVec2Snapshot(vec)
				Expect(err).NotTo(HaveOccurred())
			})

			It("accepts zero vector", func() {
				vec := &Vec2Snapshot{X: 0.0, Y: 0.0}
				err := ValidateVec2Snapshot(vec)
				Expect(err).NotTo(HaveOccurred())
			})

			It("rejects NaN in X", func() {
				vec := &Vec2Snapshot{X: math.NaN(), Y: 10.0}
				err := ValidateVec2Snapshot(vec)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("x"))
			})

			It("rejects NaN in Y", func() {
				vec := &Vec2Snapshot{X: 10.0, Y: math.NaN()}
				err := ValidateVec2Snapshot(vec)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("y"))
			})

			It("rejects Inf in X", func() {
				vec := &Vec2Snapshot{X: math.Inf(1), Y: 10.0}
				err := ValidateVec2Snapshot(vec)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("x"))
			})

			It("rejects Inf in Y", func() {
				vec := &Vec2Snapshot{X: 10.0, Y: math.Inf(-1)}
				err := ValidateVec2Snapshot(vec)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("y"))
			})
		})
	})

	Describe("Protocol Versioning", Label("scope:contract", "loop:g4-proto", "layer:contract", "net:proto:v1"), func() {
		Describe("Version Constants", func() {
			It("defines ProtocolVersionV1 constant", func() {
				Expect(ProtocolVersionV1).NotTo(BeEmpty())
				Expect(string(ProtocolVersionV1)).To(Equal("v1"))
			})

			It("exports version constant for adapters", func() {
				// Verify constant is accessible
				version := ProtocolVersionV1
				Expect(string(version)).To(Equal("v1"))
			})
		})

		Describe("ParseVersion", func() {
			It("parses valid version strings", func() {
				version, err := ParseVersion("v1")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(version)).To(Equal("v1"))

				version2, err := ParseVersion("v2")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(version2)).To(Equal("v2"))
			})

			It("rejects invalid version strings", func() {
				_, err := ParseVersion("invalid")
				Expect(err).To(HaveOccurred())

				_, err2 := ParseVersion("1")
				Expect(err2).To(HaveOccurred())

				_, err3 := ParseVersion("version1")
				Expect(err3).To(HaveOccurred())
			})

			It("rejects empty string", func() {
				_, err := ParseVersion("")
				Expect(err).To(HaveOccurred())
			})

			It("rejects malformed version strings", func() {
				_, err := ParseVersion("v")
				Expect(err).To(HaveOccurred())

				_, err2 := ParseVersion("v-1")
				Expect(err2).To(HaveOccurred())

				_, err3 := ParseVersion("v1.0")
				Expect(err3).To(HaveOccurred())
			})
		})

		Describe("IsCompatible", func() {
			It("returns true for same major version", func() {
				compatible := IsCompatible("v1", "v1")
				Expect(compatible).To(BeTrue())
			})

			It("returns false for different major versions", func() {
				compatible := IsCompatible("v1", "v2")
				Expect(compatible).To(BeFalse())

				compatible2 := IsCompatible("v2", "v1")
				Expect(compatible2).To(BeFalse())
			})

			It("handles version comparison correctly", func() {
				// v1 should be compatible with v1
				Expect(IsCompatible("v1", "v1")).To(BeTrue())

				// v1 should not be compatible with v2
				Expect(IsCompatible("v1", "v2")).To(BeFalse())

				// v2 should not be compatible with v1
				Expect(IsCompatible("v2", "v1")).To(BeFalse())
			})
		})

		Describe("CompareVersion", func() {
			It("returns 0 for equal versions", func() {
				result := CompareVersion("v1", "v1")
				Expect(result).To(Equal(0))
			})

			It("returns -1 when v1 < v2", func() {
				result := CompareVersion("v1", "v2")
				Expect(result).To(Equal(-1))

				result2 := CompareVersion("v2", "v3")
				Expect(result2).To(Equal(-1))
			})

			It("returns 1 when v1 > v2", func() {
				result := CompareVersion("v2", "v1")
				Expect(result).To(Equal(1))

				result2 := CompareVersion("v3", "v2")
				Expect(result2).To(Equal(1))
			})

			It("compares versions correctly", func() {
				Expect(CompareVersion("v1", "v1")).To(Equal(0))
				Expect(CompareVersion("v1", "v2")).To(Equal(-1))
				Expect(CompareVersion("v2", "v1")).To(Equal(1))
				Expect(CompareVersion("v5", "v10")).To(Equal(-1))
				Expect(CompareVersion("v10", "v5")).To(Equal(1))
			})
		})
	})

	Describe("Schema Compatibility", Label("scope:contract", "loop:g4-proto", "layer:contract"), func() {
		Describe("Forward Compatibility", func() {
			It("handles extra JSON fields gracefully in InputMessage", func() {
				// Messages with extra fields should be ignored (forward compatibility)
				jsonStr := `{"t":"input","seq":1,"thrust":0.5,"turn":0.3,"extra":"field"}`
				var msg InputMessage

				err := json.Unmarshal([]byte(jsonStr), &msg)
				Expect(err).NotTo(HaveOccurred())
				Expect(msg.Type).To(Equal("input"))
				Expect(msg.Seq).To(Equal(uint32(1)))
				Expect(msg.Thrust).To(Equal(float32(0.5)))
				Expect(msg.Turn).To(Equal(float32(0.3)))
			})

			It("handles extra JSON fields gracefully in SnapshotMessage", func() {
				jsonStr := `{
					"t": "snapshot",
					"tick": 1,
					"ship": {"pos":{"x":0,"y":0},"vel":{"x":0,"y":0},"rot":0,"energy":100},
					"sun": {"pos":{"x":0,"y":0},"radius":5},
					"pallets": [],
					"done": false,
					"win": false,
					"extra": "field"
				}`
				var msg SnapshotMessage

				err := json.Unmarshal([]byte(jsonStr), &msg)
				Expect(err).NotTo(HaveOccurred())
				Expect(msg.Type).To(Equal("snapshot"))
			})
		})

		Describe("Backward Compatibility", func() {
			It("handles missing optional fields in nested structures", func() {
				// Note: All current fields are required, but this test ensures
				// that if we add optional fields in the future, they're handled correctly
				jsonStr := `{"t":"input","seq":1,"thrust":0.5,"turn":0.3}`
				var msg InputMessage

				err := json.Unmarshal([]byte(jsonStr), &msg)
				Expect(err).NotTo(HaveOccurred())
				// All required fields should be present
				Expect(msg.Type).NotTo(BeEmpty())
				Expect(msg.Seq).NotTo(Equal(uint32(0)))
			})
		})

		Describe("Field Name Stability", func() {
			It("enforces exact JSON field names for InputMessage", func() {
				// Test that field names match TDD spec exactly
				msg := InputMessage{
					Type:   "input",
					Seq:    1,
					Thrust: 0.5,
					Turn:   0.3,
				}

				data, err := json.Marshal(msg)
				Expect(err).NotTo(HaveOccurred())

				var unmarshaled map[string]interface{}
				err = json.Unmarshal(data, &unmarshaled)
				Expect(err).NotTo(HaveOccurred())

				// Verify exact field names from TDD spec
				Expect(unmarshaled).To(HaveKey("t"))
				Expect(unmarshaled).To(HaveKey("seq"))
				Expect(unmarshaled).To(HaveKey("thrust"))
				Expect(unmarshaled).To(HaveKey("turn"))
				Expect(unmarshaled).NotTo(HaveKey("type")) // Should be "t", not "type"
				Expect(unmarshaled).NotTo(HaveKey("sequence")) // Should be "seq", not "sequence"
			})

			It("enforces exact JSON field names for SnapshotMessage", func() {
				msg := SnapshotMessage{
					Type: "snapshot",
					Tick:  1,
					Ship: ShipSnapshot{
						Pos:    Vec2Snapshot{X: 0, Y: 0},
						Vel:    Vec2Snapshot{X: 0, Y: 0},
						Rot:    0,
						Energy: 100,
					},
					Sun: SunSnapshot{
						Pos:    Vec2Snapshot{X: 0, Y: 0},
						Radius: 5,
					},
					Pallets: []PalletSnapshot{},
					Done:    false,
					Win:     false,
				}

				data, err := json.Marshal(msg)
				Expect(err).NotTo(HaveOccurred())

				var unmarshaled map[string]interface{}
				err = json.Unmarshal(data, &unmarshaled)
				Expect(err).NotTo(HaveOccurred())

				// Verify exact field names from TDD spec
				Expect(unmarshaled).To(HaveKey("t"))
				Expect(unmarshaled).To(HaveKey("tick"))
				Expect(unmarshaled).To(HaveKey("ship"))
				Expect(unmarshaled).To(HaveKey("sun"))
				Expect(unmarshaled).To(HaveKey("pallets"))
				Expect(unmarshaled).To(HaveKey("done"))
				Expect(unmarshaled).To(HaveKey("win"))
			})
		})

		Describe("Field Type Stability", func() {
			It("enforces correct field types for InputMessage", func() {
				msg := InputMessage{
					Type:   "input",
					Seq:    1,
					Thrust: 0.5,
					Turn:   0.3,
				}

				data, err := json.Marshal(msg)
				Expect(err).NotTo(HaveOccurred())

				var unmarshaled map[string]interface{}
				err = json.Unmarshal(data, &unmarshaled)
				Expect(err).NotTo(HaveOccurred())

				// Verify types match TDD spec
				Expect(unmarshaled["t"]).To(BeAssignableToTypeOf(""))
				Expect(unmarshaled["seq"]).To(BeNumerically(">=", 0)) // uint32
				Expect(unmarshaled["thrust"]).To(BeAssignableToTypeOf(float64(0))) // float32 -> float64 in JSON
				Expect(unmarshaled["turn"]).To(BeAssignableToTypeOf(float64(0))) // float32 -> float64 in JSON
			})
		})
	})

	Describe("Breaking Change Detection", Label("scope:contract", "loop:g4-proto", "layer:contract"), func() {
		Describe("Message Structure", func() {
			It("detects if required fields are missing from InputMessage", func() {
				// Missing "t" field
				jsonStr := `{"seq":1,"thrust":0.5,"turn":0.3}`
				var msg InputMessage
				err := json.Unmarshal([]byte(jsonStr), &msg)
				// Should unmarshal but Type will be empty, validation should catch it
				Expect(err).NotTo(HaveOccurred())
				err = ValidateInputMessage(&msg)
				Expect(err).To(HaveOccurred())
			})

			It("detects if required fields are missing from SnapshotMessage", func() {
				// Missing "sun" field - will have zero radius which should fail validation
				jsonStr := `{"t":"snapshot","tick":1,"ship":{"pos":{"x":0,"y":0},"vel":{"x":0,"y":0},"rot":0,"energy":100},"pallets":[],"done":false,"win":false}`
				var msg SnapshotMessage
				err := json.Unmarshal([]byte(jsonStr), &msg)
				// Should unmarshal but Sun will be zero value with radius=0, validation should catch it
				Expect(err).NotTo(HaveOccurred())
				err = ValidateSnapshotMessage(&msg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("radius"))
			})
		})

		Describe("Contract Stability", func() {
			It("maintains consistent JSON schema for InputMessage", func() {
				// This test ensures the schema doesn't change unexpectedly
				msg := InputMessage{
					Type:   "input",
					Seq:    42,
					Thrust: 0.75,
					Turn:   -0.5,
				}

				data, err := json.Marshal(msg)
				Expect(err).NotTo(HaveOccurred())

				// Parse and verify structure
				var parsed map[string]interface{}
				err = json.Unmarshal(data, &parsed)
				Expect(err).NotTo(HaveOccurred())

				// Verify all expected fields are present
				Expect(parsed).To(HaveKey("t"))
				Expect(parsed).To(HaveKey("seq"))
				Expect(parsed).To(HaveKey("thrust"))
				Expect(parsed).To(HaveKey("turn"))

				// Verify no unexpected fields
				Expect(len(parsed)).To(Equal(4))
			})

			It("maintains consistent JSON schema for SnapshotMessage", func() {
				msg := SnapshotMessage{
					Type: "snapshot",
					Tick:  100,
					Ship: ShipSnapshot{
						Pos:    Vec2Snapshot{X: 10, Y: 20},
						Vel:    Vec2Snapshot{X: 1, Y: -1},
						Rot:    1.57,
						Energy: 75,
					},
					Sun: SunSnapshot{
						Pos:    Vec2Snapshot{X: 0, Y: 0},
						Radius: 5,
					},
					Pallets: []PalletSnapshot{
						{ID: 1, Pos: Vec2Snapshot{X: 15, Y: 15}, Active: true},
					},
					Done: false,
					Win:  false,
				}

				data, err := json.Marshal(msg)
				Expect(err).NotTo(HaveOccurred())

				// Parse and verify structure
				var parsed map[string]interface{}
				err = json.Unmarshal(data, &parsed)
				Expect(err).NotTo(HaveOccurred())

				// Verify all expected top-level fields are present
				Expect(parsed).To(HaveKey("t"))
				Expect(parsed).To(HaveKey("tick"))
				Expect(parsed).To(HaveKey("ship"))
				Expect(parsed).To(HaveKey("sun"))
				Expect(parsed).To(HaveKey("pallets"))
				Expect(parsed).To(HaveKey("done"))
				Expect(parsed).To(HaveKey("win"))
			})
		})
	})

	Describe("Edge Cases", Label("scope:contract", "loop:g4-proto", "layer:contract"), func() {
		It("handles maximum sequence numbers", func() {
			msg := InputMessage{
				Type:   "input",
				Seq:    ^uint32(0), // Maximum uint32
				Thrust: 1.0,
				Turn:   1.0,
			}

			data, err := json.Marshal(msg)
			Expect(err).NotTo(HaveOccurred())

			var unmarshaled InputMessage
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).NotTo(HaveOccurred())
			Expect(unmarshaled.Seq).To(Equal(^uint32(0)))
		})

		It("handles large tick values", func() {
			msg := SnapshotMessage{
				Type: "snapshot",
				Tick:  ^uint32(0), // Maximum uint32
				Ship: ShipSnapshot{
					Pos:    Vec2Snapshot{X: 0, Y: 0},
					Vel:    Vec2Snapshot{X: 0, Y: 0},
					Rot:    0,
					Energy: 100,
				},
				Sun: SunSnapshot{
					Pos:    Vec2Snapshot{X: 0, Y: 0},
					Radius: 5,
				},
				Pallets: []PalletSnapshot{},
				Done:    false,
				Win:     false,
			}

			data, err := json.Marshal(msg)
			Expect(err).NotTo(HaveOccurred())

			var unmarshaled SnapshotMessage
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).NotTo(HaveOccurred())
			Expect(unmarshaled.Tick).To(Equal(^uint32(0)))
		})

		It("handles very large floating point values", func() {
			msg := SnapshotMessage{
				Type: "snapshot",
				Tick:  1,
				Ship: ShipSnapshot{
					Pos:    Vec2Snapshot{X: 1e10, Y: -1e10},
					Vel:    Vec2Snapshot{X: 1e5, Y: -1e5},
					Rot:    6.283185307179586, // 2*pi
					Energy: 1e6,
				},
				Sun: SunSnapshot{
					Pos:    Vec2Snapshot{X: 0, Y: 0},
					Radius: 1e5,
				},
				Pallets: []PalletSnapshot{},
				Done:    false,
				Win:     false,
			}

			data, err := json.Marshal(msg)
			Expect(err).NotTo(HaveOccurred())

			var unmarshaled SnapshotMessage
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).NotTo(HaveOccurred())
			Expect(unmarshaled.Ship.Pos.X).To(BeNumerically("~", 1e10, 1e-5))
		})

		It("handles empty pallets array", func() {
			msg := SnapshotMessage{
				Type: "snapshot",
				Tick:  1,
				Ship: ShipSnapshot{
					Pos:    Vec2Snapshot{X: 0, Y: 0},
					Vel:    Vec2Snapshot{X: 0, Y: 0},
					Rot:    0,
					Energy: 100,
				},
				Sun: SunSnapshot{
					Pos:    Vec2Snapshot{X: 0, Y: 0},
					Radius: 5,
				},
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

		It("handles many pallets", func() {
			pallets := make([]PalletSnapshot, 100)
			for i := range pallets {
				pallets[i] = PalletSnapshot{
					ID:     uint32(i + 1),
					Pos:    Vec2Snapshot{X: float64(i), Y: float64(i)},
					Active: i%2 == 0,
				}
			}

			msg := SnapshotMessage{
				Type: "snapshot",
				Tick:  1,
				Ship: ShipSnapshot{
					Pos:    Vec2Snapshot{X: 0, Y: 0},
					Vel:    Vec2Snapshot{X: 0, Y: 0},
					Rot:    0,
					Energy: 100,
				},
				Sun: SunSnapshot{
					Pos:    Vec2Snapshot{X: 0, Y: 0},
					Radius: 5,
				},
				Pallets: pallets,
				Done:    false,
				Win:     false,
			}

			data, err := json.Marshal(msg)
			Expect(err).NotTo(HaveOccurred())

			var unmarshaled SnapshotMessage
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(unmarshaled.Pallets)).To(Equal(100))
		})
	})
})

