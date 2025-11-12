package rules

import (
	"math"
	"testing"

	"github.com/gorbit/orbitalrush/internal/sim/entities"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInput(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Input Processing Suite")
}

var _ = Describe("Input Processing", Label("scope:unit", "loop:g2-rules", "layer:sim", "dep:none", "b:input-processing", "r:high", "double:fake"), func() {
	const epsilon = 1e-6
	const dt = 1.0 / 30.0 // 30Hz tick rate

	Describe("ClampInput", func() {
		It("clamps thrust to valid range [0.0, 1.0]", func() {
			input := InputCommand{Thrust: 1.5, Turn: 0.0}
			clamped := ClampInput(input)
			Expect(clamped.Thrust).To(Equal(float32(1.0)))
			Expect(clamped.Turn).To(Equal(float32(0.0)))
		})

		It("clamps negative thrust to zero", func() {
			input := InputCommand{Thrust: -0.5, Turn: 0.0}
			clamped := ClampInput(input)
			Expect(clamped.Thrust).To(Equal(float32(0.0)))
		})

		It("clamps turn to valid range [-1.0, 1.0]", func() {
			input := InputCommand{Thrust: 0.0, Turn: 1.5}
			clamped := ClampInput(input)
			Expect(clamped.Turn).To(Equal(float32(1.0)))
		})

		It("clamps negative turn to -1.0", func() {
			input := InputCommand{Thrust: 0.0, Turn: -1.5}
			clamped := ClampInput(input)
			Expect(clamped.Turn).To(Equal(float32(-1.0)))
		})

		It("does not clamp values within valid range", func() {
			input := InputCommand{Thrust: 0.5, Turn: 0.3}
			clamped := ClampInput(input)
			Expect(clamped.Thrust).To(Equal(float32(0.5)))
			Expect(clamped.Turn).To(Equal(float32(0.3)))
		})

		It("clamps both values when both are out of range", func() {
			input := InputCommand{Thrust: 2.0, Turn: -2.0}
			clamped := ClampInput(input)
			Expect(clamped.Thrust).To(Equal(float32(1.0)))
			Expect(clamped.Turn).To(Equal(float32(-1.0)))
		})
	})

	Describe("UpdateRotation", func() {
		It("updates rotation when turning right", func() {
			currentRot := 0.0
			turnInput := 1.0
			newRot := UpdateRotation(currentRot, turnInput, dt)
			expectedRot := currentRot + TurnRate*float64(turnInput)*dt
			Expect(newRot).To(BeNumerically("~", expectedRot, epsilon))
		})

		It("updates rotation when turning left", func() {
			currentRot := math.Pi / 2.0
			turnInput := -1.0
			newRot := UpdateRotation(currentRot, turnInput, dt)
			expectedRot := currentRot + TurnRate*float64(turnInput)*dt
			Expect(newRot).To(BeNumerically("~", expectedRot, epsilon))
		})

		It("does not change rotation when turn input is zero", func() {
			currentRot := math.Pi / 4.0
			turnInput := 0.0
			newRot := UpdateRotation(currentRot, turnInput, dt)
			Expect(newRot).To(BeNumerically("~", currentRot, epsilon))
		})

		It("normalizes rotation to [0, 2π) when exceeding 2π", func() {
			currentRot := 2*math.Pi - 0.1
			turnInput := 1.0
			newRot := UpdateRotation(currentRot, turnInput, dt)
			// Should wrap around to [0, 2π)
			Expect(newRot).To(BeNumerically(">=", 0.0))
			Expect(newRot).To(BeNumerically("<", 2*math.Pi))
		})

		It("normalizes rotation to [0, 2π) when going negative", func() {
			currentRot := 0.1
			turnInput := -1.0
			newRot := UpdateRotation(currentRot, turnInput, dt)
			// Should wrap around to [0, 2π)
			Expect(newRot).To(BeNumerically(">=", 0.0))
			Expect(newRot).To(BeNumerically("<", 2*math.Pi))
		})

		It("handles multiple sequential turns", func() {
			rot := 0.0
			rot = UpdateRotation(rot, 1.0, dt)
			rot = UpdateRotation(rot, 1.0, dt)
			rot = UpdateRotation(rot, 1.0, dt)
			expectedRot := 3.0 * TurnRate * dt
			Expect(rot).To(BeNumerically("~", expectedRot, epsilon))
		})

		It("handles rotation at exactly 2π", func() {
			currentRot := 2 * math.Pi
			turnInput := 0.1
			newRot := UpdateRotation(currentRot, turnInput, dt)
			// Should normalize to [0, 2π)
			Expect(newRot).To(BeNumerically(">=", 0.0))
			Expect(newRot).To(BeNumerically("<", 2*math.Pi))
		})

		It("handles partial turn input", func() {
			currentRot := 0.0
			turnInput := 0.5
			newRot := UpdateRotation(currentRot, turnInput, dt)
			expectedRot := currentRot + TurnRate*0.5*dt
			Expect(newRot).To(BeNumerically("~", expectedRot, epsilon))
		})
	})

	Describe("CalculateThrustAcceleration", func() {
		It("calculates thrust acceleration in forward direction (rotation = 0)", func() {
			rotation := 0.0
			thrustInput := float32(1.0)
			acc := CalculateThrustAcceleration(rotation, thrustInput)
			// At rotation 0, should point in +X direction
			Expect(acc.X).To(BeNumerically("~", ThrustAcceleration, epsilon))
			Expect(acc.Y).To(BeNumerically("~", 0.0, epsilon))
		})

		It("calculates thrust acceleration at 90 degrees (rotation = π/2)", func() {
			rotation := math.Pi / 2.0
			thrustInput := float32(1.0)
			acc := CalculateThrustAcceleration(rotation, thrustInput)
			// At rotation π/2, should point in +Y direction
			Expect(acc.X).To(BeNumerically("~", 0.0, epsilon))
			Expect(acc.Y).To(BeNumerically("~", ThrustAcceleration, epsilon))
		})

		It("calculates thrust acceleration at 180 degrees (rotation = π)", func() {
			rotation := math.Pi
			thrustInput := float32(1.0)
			acc := CalculateThrustAcceleration(rotation, thrustInput)
			// At rotation π, should point in -X direction
			Expect(acc.X).To(BeNumerically("~", -ThrustAcceleration, epsilon))
			Expect(acc.Y).To(BeNumerically("~", 0.0, epsilon))
		})

		It("scales acceleration by thrust input", func() {
			rotation := 0.0
			thrustInput := float32(0.5)
			acc := CalculateThrustAcceleration(rotation, thrustInput)
			// Should be half of full thrust
			Expect(acc.X).To(BeNumerically("~", ThrustAcceleration*0.5, epsilon))
			Expect(acc.Y).To(BeNumerically("~", 0.0, epsilon))
		})

		It("returns zero acceleration when thrust input is zero", func() {
			rotation := math.Pi / 4.0
			thrustInput := float32(0.0)
			acc := CalculateThrustAcceleration(rotation, thrustInput)
			Expect(acc.X).To(BeNumerically("~", 0.0, epsilon))
			Expect(acc.Y).To(BeNumerically("~", 0.0, epsilon))
		})

		It("calculates correct direction for arbitrary rotation", func() {
			rotation := math.Pi / 4.0
			thrustInput := float32(1.0)
			acc := CalculateThrustAcceleration(rotation, thrustInput)
			// At 45 degrees, both X and Y should be positive and equal
			expectedX := ThrustAcceleration * math.Cos(rotation)
			expectedY := ThrustAcceleration * math.Sin(rotation)
			Expect(acc.X).To(BeNumerically("~", expectedX, epsilon))
			Expect(acc.Y).To(BeNumerically("~", expectedY, epsilon))
		})

		It("maintains correct magnitude for all rotations", func() {
			thrustInput := float32(1.0)
			for _, rot := range []float64{0.0, math.Pi / 4.0, math.Pi / 2.0, math.Pi, 3 * math.Pi / 2.0} {
				acc := CalculateThrustAcceleration(rot, thrustInput)
				magnitude := acc.Length()
				Expect(magnitude).To(BeNumerically("~", ThrustAcceleration, epsilon))
			}
		})
	})

	Describe("ApplyInput", func() {
		It("applies thrust when energy > 0", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}
			updatedShip := ApplyInput(ship, input, dt)

			// Velocity should increase in forward direction
			Expect(updatedShip.Vel.X).To(BeNumerically(">", 0.0))
			Expect(updatedShip.Vel.Y).To(BeNumerically("~", 0.0, epsilon))
			// Energy should be drained
			Expect(updatedShip.Energy).To(BeNumerically("~", 100.0-ThrustDrainRate, epsilon))
		})

		It("does not apply thrust when energy = 0", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				0.0,
			)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}
			updatedShip := ApplyInput(ship, input, dt)

			// Velocity should remain unchanged
			Expect(updatedShip.Vel.X).To(BeNumerically("~", 0.0, epsilon))
			Expect(updatedShip.Vel.Y).To(BeNumerically("~", 0.0, epsilon))
			// Energy should remain at 0
			Expect(updatedShip.Energy).To(Equal(float32(0.0)))
		})

		It("does not apply thrust when energy < 0 (edge case)", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				-10.0, // Invalid but should be handled
			)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}
			updatedShip := ApplyInput(ship, input, dt)

			// Velocity should remain unchanged
			Expect(updatedShip.Vel.X).To(BeNumerically("~", 0.0, epsilon))
			Expect(updatedShip.Vel.Y).To(BeNumerically("~", 0.0, epsilon))
		})

		It("applies turn regardless of energy", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				0.0,
			)
			input := InputCommand{Thrust: 0.0, Turn: 1.0}
			updatedShip := ApplyInput(ship, input, dt)

			// Rotation should change
			Expect(updatedShip.Rot).To(BeNumerically(">", 0.0))
			// Energy should remain unchanged
			Expect(updatedShip.Energy).To(Equal(float32(0.0)))
		})

		It("drains energy when thrusting with energy > 0", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				50.0,
			)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}
			updatedShip := ApplyInput(ship, input, dt)

			// Energy should be drained
			Expect(updatedShip.Energy).To(BeNumerically("~", 50.0-ThrustDrainRate, epsilon))
		})

		It("does not drain energy when not thrusting", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				50.0,
			)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}
			updatedShip := ApplyInput(ship, input, dt)

			// Energy should remain unchanged
			Expect(updatedShip.Energy).To(Equal(float32(50.0)))
		})

		It("applies both thrust and turn together", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			input := InputCommand{Thrust: 1.0, Turn: 1.0}
			updatedShip := ApplyInput(ship, input, dt)

			// Both rotation and velocity should change
			Expect(updatedShip.Rot).To(BeNumerically(">", 0.0))
			Expect(updatedShip.Vel.Length()).To(BeNumerically(">", 0.0))
			// Energy should be drained
			Expect(updatedShip.Energy).To(BeNumerically("~", 100.0-ThrustDrainRate, epsilon))
		})

		It("clamps input values to valid ranges", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			input := InputCommand{Thrust: 2.0, Turn: -2.0} // Out of range
			updatedShip := ApplyInput(ship, input, dt)

			// Should still work (input is clamped internally)
			// Turning left from 0 wraps to near 2π (normalized to [0, 2π))
			Expect(updatedShip.Rot).NotTo(Equal(0.0)) // Rotation changed
			Expect(updatedShip.Rot).To(BeNumerically(">=", 0.0))  // In valid range
			Expect(updatedShip.Rot).To(BeNumerically("<", 2*math.Pi)) // In valid range
			Expect(updatedShip.Vel.Length()).To(BeNumerically(">", 0.0))
		})

		It("updates velocity correctly with thrust", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(1.0, 2.0), // Initial velocity
				0.0,
				100.0,
			)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}
			updatedShip := ApplyInput(ship, input, dt)

			// Velocity should increase in forward direction
			expectedVelX := 1.0 + ThrustAcceleration*dt
			Expect(updatedShip.Vel.X).To(BeNumerically("~", expectedVelX, epsilon))
			Expect(updatedShip.Vel.Y).To(BeNumerically("~", 2.0, epsilon))
		})

		It("handles partial thrust input", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			input := InputCommand{Thrust: 0.5, Turn: 0.0}
			updatedShip := ApplyInput(ship, input, dt)

			// Velocity should increase by half the acceleration
			expectedVelX := 0.5 * ThrustAcceleration * dt
			Expect(updatedShip.Vel.X).To(BeNumerically("~", expectedVelX, epsilon))
		})

		It("preserves position (position is not updated by ApplyInput)", func() {
			ship := entities.NewShip(
				entities.NewVec2(10.0, 20.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			input := InputCommand{Thrust: 1.0, Turn: 1.0}
			updatedShip := ApplyInput(ship, input, dt)

			// Position should remain unchanged (position updates happen in physics step)
			Expect(updatedShip.Pos.X).To(Equal(10.0))
			Expect(updatedShip.Pos.Y).To(Equal(20.0))
		})
	})

	Describe("Input Processing Integration", func() {
		It("handles sequence of thrust then turn", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)

			// Thrust first
			input1 := InputCommand{Thrust: 1.0, Turn: 0.0}
			ship = ApplyInput(ship, input1, dt)
			initialVel := ship.Vel.Length()
			Expect(initialVel).To(BeNumerically(">", 0.0))
			Expect(ship.Energy).To(BeNumerically("~", 100.0-ThrustDrainRate, epsilon))

			// Then turn
			input2 := InputCommand{Thrust: 0.0, Turn: 1.0}
			ship = ApplyInput(ship, input2, dt)
			Expect(ship.Rot).To(BeNumerically(">", 0.0))
			// Velocity should remain (not reset)
			Expect(ship.Vel.Length()).To(BeNumerically("~", initialVel, epsilon))
		})

		It("handles sequence of turn then thrust", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)

			// Turn first
			input1 := InputCommand{Thrust: 0.0, Turn: 1.0}
			ship = ApplyInput(ship, input1, dt)
			initialRot := ship.Rot
			Expect(initialRot).To(BeNumerically(">", 0.0))

			// Then thrust (should be in new direction)
			input2 := InputCommand{Thrust: 1.0, Turn: 0.0}
			ship = ApplyInput(ship, input2, dt)
			// Velocity should be in the direction of rotation
			velDirection := math.Atan2(ship.Vel.Y, ship.Vel.X)
			Expect(velDirection).To(BeNumerically("~", initialRot, epsilon))
		})

		It("handles multiple ticks of input processing", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)

			input := InputCommand{Thrust: 1.0, Turn: 0.0}
			// Process 10 ticks
			for i := 0; i < 10; i++ {
				ship = ApplyInput(ship, input, dt)
			}

			// Velocity should accumulate
			expectedVel := 10.0 * ThrustAcceleration * dt
			Expect(ship.Vel.Length()).To(BeNumerically("~", expectedVel, epsilon))
			// Energy should be drained 10 times
			Expect(ship.Energy).To(BeNumerically("~", 100.0-10.0*ThrustDrainRate, epsilon))
		})

		It("stops thrusting when energy depletes", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				ThrustDrainRate*2.0, // Enough for 2 ticks
			)

			input := InputCommand{Thrust: 1.0, Turn: 0.0}
			// First tick: should thrust
			ship = ApplyInput(ship, input, dt)
			velAfterFirst := ship.Vel.Length()
			Expect(velAfterFirst).To(BeNumerically(">", 0.0))

			// Second tick: should still thrust
			ship = ApplyInput(ship, input, dt)
			velAfterSecond := ship.Vel.Length()
			Expect(velAfterSecond).To(BeNumerically(">", velAfterFirst))

			// Third tick: should NOT thrust (energy depleted)
			ship = ApplyInput(ship, input, dt)
			velAfterThird := ship.Vel.Length()
			Expect(velAfterThird).To(BeNumerically("~", velAfterSecond, epsilon))
			Expect(ship.Energy).To(Equal(float32(0.0)))
		})

		It("maintains correct state across complex input sequence", func() {
			ship := entities.NewShip(
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)

			// Complex sequence: turn, thrust, turn, thrust, no input
			ship = ApplyInput(ship, InputCommand{Thrust: 0.0, Turn: 1.0}, dt)
			ship = ApplyInput(ship, InputCommand{Thrust: 1.0, Turn: 0.0}, dt)
			ship = ApplyInput(ship, InputCommand{Thrust: 0.0, Turn: -1.0}, dt)
			ship = ApplyInput(ship, InputCommand{Thrust: 0.5, Turn: 0.0}, dt)
			ship = ApplyInput(ship, InputCommand{Thrust: 0.0, Turn: 0.0}, dt)

			// Final state should be valid
			Expect(ship.Energy).To(BeNumerically(">=", 0.0))
			Expect(ship.Energy).To(BeNumerically("<=", MaxEnergy))
			Expect(ship.Rot).To(BeNumerically(">=", 0.0))
			Expect(ship.Rot).To(BeNumerically("<", 2*math.Pi))
		})
	})
})

