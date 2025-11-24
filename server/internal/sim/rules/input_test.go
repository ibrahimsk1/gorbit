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

var _ = Describe("Input Processing", Label("scope:unit", "loop:g3-rules", "layer:rules", "dep:none", "b:input-processing", "r:high", "double:fake"), func() {
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
			// At rotation π/2, should point in -Y direction (negated to match screen coords)
			Expect(acc.X).To(BeNumerically("~", 0.0, epsilon))
			Expect(acc.Y).To(BeNumerically("~", -ThrustAcceleration, epsilon))
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
			// At 45 degrees, X should be positive, Y should be negated (to match screen coords)
			expectedX := ThrustAcceleration * math.Cos(rotation)
			expectedY := -ThrustAcceleration * math.Sin(rotation) // Y is negated
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
				1, // playerID
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			world := entities.NewWorld([]entities.Ship{ship}, nil, nil)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}
			updatedWorld := ApplyInput(world, 1, input, dt)

			// Velocity should increase in forward direction
			Expect(updatedWorld.Ships[0].Vel.X).To(BeNumerically(">", 0.0))
			Expect(updatedWorld.Ships[0].Vel.Y).To(BeNumerically("~", 0.0, epsilon))
			// Energy should be drained
			Expect(updatedWorld.Ships[0].Energy).To(BeNumerically("~", 100.0-ThrustDrainRate, epsilon))
		})

		It("does not apply thrust when energy = 0", func() {
			ship := entities.NewShip(
				1, // playerID
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				0.0,
			)
			world := entities.NewWorld([]entities.Ship{ship}, nil, nil)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}
			updatedWorld := ApplyInput(world, 1, input, dt)

			// Velocity should remain unchanged
			Expect(updatedWorld.Ships[0].Vel.X).To(BeNumerically("~", 0.0, epsilon))
			Expect(updatedWorld.Ships[0].Vel.Y).To(BeNumerically("~", 0.0, epsilon))
			// Energy should remain at 0
			Expect(updatedWorld.Ships[0].Energy).To(Equal(float32(0.0)))
		})

		It("does not apply thrust when energy < 0 (edge case)", func() {
			ship := entities.NewShip(
				1, // playerID
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				-10.0, // Invalid but should be handled
			)
			world := entities.NewWorld([]entities.Ship{ship}, nil, nil)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}
			updatedWorld := ApplyInput(world, 1, input, dt)

			// Velocity should remain unchanged
			Expect(updatedWorld.Ships[0].Vel.X).To(BeNumerically("~", 0.0, epsilon))
			Expect(updatedWorld.Ships[0].Vel.Y).To(BeNumerically("~", 0.0, epsilon))
		})

		It("applies turn regardless of energy", func() {
			ship := entities.NewShip(
				1, // playerID
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				0.0,
			)
			world := entities.NewWorld([]entities.Ship{ship}, nil, nil)
			input := InputCommand{Thrust: 0.0, Turn: 1.0}
			updatedWorld := ApplyInput(world, 1, input, dt)

			// Rotation should change
			Expect(updatedWorld.Ships[0].Rot).To(BeNumerically(">", 0.0))
			// Energy should remain unchanged
			Expect(updatedWorld.Ships[0].Energy).To(Equal(float32(0.0)))
		})

		It("drains energy when thrusting with energy > 0", func() {
			ship := entities.NewShip(
				1, // playerID
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				50.0,
			)
			world := entities.NewWorld([]entities.Ship{ship}, nil, nil)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}
			updatedWorld := ApplyInput(world, 1, input, dt)

			// Energy should be drained
			Expect(updatedWorld.Ships[0].Energy).To(BeNumerically("~", 50.0-ThrustDrainRate, epsilon))
		})

		It("does not drain energy when not thrusting", func() {
			ship := entities.NewShip(
				1, // playerID
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				50.0,
			)
			world := entities.NewWorld([]entities.Ship{ship}, nil, nil)
			input := InputCommand{Thrust: 0.0, Turn: 0.0}
			updatedWorld := ApplyInput(world, 1, input, dt)

			// Energy should remain unchanged
			Expect(updatedWorld.Ships[0].Energy).To(Equal(float32(50.0)))
		})

		It("applies both thrust and turn together", func() {
			ship := entities.NewShip(
				1, // playerID
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			world := entities.NewWorld([]entities.Ship{ship}, nil, nil)
			input := InputCommand{Thrust: 1.0, Turn: 1.0}
			updatedWorld := ApplyInput(world, 1, input, dt)

			// Both rotation and velocity should change
			Expect(updatedWorld.Ships[0].Rot).To(BeNumerically(">", 0.0))
			Expect(updatedWorld.Ships[0].Vel.Length()).To(BeNumerically(">", 0.0))
			// Energy should be drained
			Expect(updatedWorld.Ships[0].Energy).To(BeNumerically("~", 100.0-ThrustDrainRate, epsilon))
		})

		It("clamps input values to valid ranges", func() {
			ship := entities.NewShip(
				1, // playerID
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			world := entities.NewWorld([]entities.Ship{ship}, nil, nil)
			input := InputCommand{Thrust: 2.0, Turn: -2.0} // Out of range
			updatedWorld := ApplyInput(world, 1, input, dt)

			// Should still work (input is clamped internally)
			// Turning left from 0 wraps to near 2π (normalized to [0, 2π))
			Expect(updatedWorld.Ships[0].Rot).NotTo(Equal(0.0)) // Rotation changed
			Expect(updatedWorld.Ships[0].Rot).To(BeNumerically(">=", 0.0))  // In valid range
			Expect(updatedWorld.Ships[0].Rot).To(BeNumerically("<", 2*math.Pi)) // In valid range
			Expect(updatedWorld.Ships[0].Vel.Length()).To(BeNumerically(">", 0.0))
		})

		It("updates velocity correctly with thrust", func() {
			ship := entities.NewShip(
				1, // playerID
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(1.0, 2.0), // Initial velocity
				0.0,
				100.0,
			)
			world := entities.NewWorld([]entities.Ship{ship}, nil, nil)
			input := InputCommand{Thrust: 1.0, Turn: 0.0}
			updatedWorld := ApplyInput(world, 1, input, dt)

			// Velocity should increase in forward direction
			expectedVelX := 1.0 + ThrustAcceleration*dt
			Expect(updatedWorld.Ships[0].Vel.X).To(BeNumerically("~", expectedVelX, epsilon))
			Expect(updatedWorld.Ships[0].Vel.Y).To(BeNumerically("~", 2.0, epsilon))
		})

		It("handles partial thrust input", func() {
			ship := entities.NewShip(
				1, // playerID
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			world := entities.NewWorld([]entities.Ship{ship}, nil, nil)
			input := InputCommand{Thrust: 0.5, Turn: 0.0}
			updatedWorld := ApplyInput(world, 1, input, dt)

			// Velocity should increase by half the acceleration
			expectedVelX := 0.5 * ThrustAcceleration * dt
			Expect(updatedWorld.Ships[0].Vel.X).To(BeNumerically("~", expectedVelX, epsilon))
		})

		It("preserves position (position is not updated by ApplyInput)", func() {
			ship := entities.NewShip(
				1, // playerID
				entities.NewVec2(10.0, 20.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			world := entities.NewWorld([]entities.Ship{ship}, nil, nil)
			input := InputCommand{Thrust: 1.0, Turn: 1.0}
			updatedWorld := ApplyInput(world, 1, input, dt)

			// Position should remain unchanged (position updates happen in physics step)
			Expect(updatedWorld.Ships[0].Pos.X).To(Equal(10.0))
			Expect(updatedWorld.Ships[0].Pos.Y).To(Equal(20.0))
		})
	})

	Describe("Input Processing Integration", func() {
		It("handles sequence of thrust then turn", func() {
			ship := entities.NewShip(
				1, // playerID
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			world := entities.NewWorld([]entities.Ship{ship}, nil, nil)

			// Thrust first
			input1 := InputCommand{Thrust: 1.0, Turn: 0.0}
			world = ApplyInput(world, 1, input1, dt)
			initialVel := world.Ships[0].Vel.Length()
			Expect(initialVel).To(BeNumerically(">", 0.0))
			Expect(world.Ships[0].Energy).To(BeNumerically("~", 100.0-ThrustDrainRate, epsilon))

			// Then turn
			input2 := InputCommand{Thrust: 0.0, Turn: 1.0}
			world = ApplyInput(world, 1, input2, dt)
			Expect(world.Ships[0].Rot).To(BeNumerically(">", 0.0))
			// Velocity should remain (not reset)
			Expect(world.Ships[0].Vel.Length()).To(BeNumerically("~", initialVel, epsilon))
		})

		It("handles sequence of turn then thrust", func() {
			ship := entities.NewShip(
				1, // playerID
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			world := entities.NewWorld([]entities.Ship{ship}, nil, nil)

			// Turn first
			input1 := InputCommand{Thrust: 0.0, Turn: 1.0}
			world = ApplyInput(world, 1, input1, dt)
			initialRot := world.Ships[0].Rot
			Expect(initialRot).To(BeNumerically(">", 0.0))

			// Then thrust (should be in new direction)
			input2 := InputCommand{Thrust: 1.0, Turn: 0.0}
			world = ApplyInput(world, 1, input2, dt)
			// Velocity should be in the direction of rotation
			// Note: CalculateThrustAcceleration negates Y component (line 103: -directionY)
			// So velocity vector is (cos(θ), -sin(θ)), and Atan2(-sin(θ), cos(θ)) = -θ
			// We need to compare velDirection with -initialRot (normalized to [0, 2π))
			velDirection := math.Atan2(world.Ships[0].Vel.Y, world.Ships[0].Vel.X)
			// Normalize to [0, 2π)
			if velDirection < 0 {
				velDirection += 2 * math.Pi
			}
			// Expected direction is -initialRot (because Y is negated), normalized to [0, 2π)
			normalizedRot := math.Mod(initialRot, 2*math.Pi)
			expectedDirection := 2*math.Pi - normalizedRot
			if expectedDirection >= 2*math.Pi {
				expectedDirection -= 2 * math.Pi
			}
			// Check if velocity direction matches expected direction (accounting for Y negation)
			diff := math.Abs(velDirection - expectedDirection)
			if diff > math.Pi {
				diff = 2*math.Pi - diff
			}
			Expect(diff).To(BeNumerically("<", epsilon))
		})

		It("handles multiple ticks of input processing", func() {
			ship := entities.NewShip(
				1, // playerID
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			world := entities.NewWorld([]entities.Ship{ship}, nil, nil)

			input := InputCommand{Thrust: 1.0, Turn: 0.0}
			// Process 10 ticks
			for i := 0; i < 10; i++ {
				world = ApplyInput(world, 1, input, dt)
			}

			// Velocity should accumulate
			expectedVel := 10.0 * ThrustAcceleration * dt
			Expect(world.Ships[0].Vel.Length()).To(BeNumerically("~", expectedVel, epsilon))
			// Energy should be drained 10 times
			Expect(world.Ships[0].Energy).To(BeNumerically("~", 100.0-10.0*ThrustDrainRate, epsilon))
		})

		It("stops thrusting when energy depletes", func() {
			ship := entities.NewShip(
				1, // playerID
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				ThrustDrainRate*2.0, // Enough for 2 ticks
			)
			world := entities.NewWorld([]entities.Ship{ship}, nil, nil)

			input := InputCommand{Thrust: 1.0, Turn: 0.0}
			// First tick: should thrust
			world = ApplyInput(world, 1, input, dt)
			velAfterFirst := world.Ships[0].Vel.Length()
			Expect(velAfterFirst).To(BeNumerically(">", 0.0))

			// Second tick: should still thrust
			world = ApplyInput(world, 1, input, dt)
			velAfterSecond := world.Ships[0].Vel.Length()
			Expect(velAfterSecond).To(BeNumerically(">", velAfterFirst))

			// Third tick: should NOT thrust (energy depleted)
			world = ApplyInput(world, 1, input, dt)
			velAfterThird := world.Ships[0].Vel.Length()
			Expect(velAfterThird).To(BeNumerically("~", velAfterSecond, epsilon))
			Expect(world.Ships[0].Energy).To(Equal(float32(0.0)))
		})

		It("maintains correct state across complex input sequence", func() {
			ship := entities.NewShip(
				1, // playerID
				entities.NewVec2(0.0, 0.0),
				entities.NewVec2(0.0, 0.0),
				0.0,
				100.0,
			)
			world := entities.NewWorld([]entities.Ship{ship}, nil, nil)

			// Complex sequence: turn, thrust, turn, thrust, no input
			world = ApplyInput(world, 1, InputCommand{Thrust: 0.0, Turn: 1.0}, dt)
			world = ApplyInput(world, 1, InputCommand{Thrust: 1.0, Turn: 0.0}, dt)
			world = ApplyInput(world, 1, InputCommand{Thrust: 0.0, Turn: -1.0}, dt)
			world = ApplyInput(world, 1, InputCommand{Thrust: 0.5, Turn: 0.0}, dt)
			world = ApplyInput(world, 1, InputCommand{Thrust: 0.0, Turn: 0.0}, dt)

			// Final state should be valid
			Expect(world.Ships[0].Energy).To(BeNumerically(">=", 0.0))
			Expect(world.Ships[0].Energy).To(BeNumerically("<=", MaxEnergy))
			Expect(world.Ships[0].Rot).To(BeNumerically(">=", 0.0))
			Expect(world.Ships[0].Rot).To(BeNumerically("<", 2*math.Pi))
		})
	})

	Describe("ApplyInput Multiplayer", func() {
		It("applies input to ship with valid playerID", func() {
			pos1 := entities.NewVec2(0.0, 0.0)
			vel1 := entities.NewVec2(0.0, 0.0)
			pos2 := entities.NewVec2(100.0, 100.0)
			vel2 := entities.NewVec2(1.0, 1.0)
			ship1 := entities.NewShip(1, pos1, vel1, 0.0, 100.0)
			ship2 := entities.NewShip(2, pos2, vel2, 0.0, 100.0)
			world := entities.NewWorld([]entities.Ship{ship1, ship2}, nil, nil)

			input := InputCommand{Thrust: 1.0, Turn: 0.0}
			updatedWorld := ApplyInput(world, 1, input, dt)

			// Ship 1 should be updated
			Expect(updatedWorld.Ships[0].Vel.Length()).To(BeNumerically(">", 0.0))
			Expect(updatedWorld.Ships[0].Energy).To(BeNumerically("<", 100.0))
			// Ship 2 should be unchanged
			Expect(updatedWorld.Ships[1].Vel).To(Equal(vel2))
			Expect(updatedWorld.Ships[1].Energy).To(Equal(float32(100.0)))
		})

		It("returns unchanged world when playerID not found", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			world := entities.NewWorld([]entities.Ship{ship1}, nil, nil)

			input := InputCommand{Thrust: 1.0, Turn: 0.0}
			updatedWorld := ApplyInput(world, 999, input, dt) // Invalid ID

			// World should be unchanged
			Expect(updatedWorld.Ships[0].Vel).To(Equal(entities.NewVec2(0.0, 0.0)))
			Expect(updatedWorld.Ships[0].Energy).To(Equal(float32(100.0)))
		})

		It("only affects target ship when multiple ships present", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			ship2 := entities.NewShip(2, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			world := entities.NewWorld([]entities.Ship{ship1, ship2}, nil, nil)

			input := InputCommand{Thrust: 1.0, Turn: 1.0}
			updatedWorld := ApplyInput(world, 1, input, dt)

			// Ship 1 should be updated
			Expect(updatedWorld.Ships[0].Rot).To(BeNumerically(">", 0.0))
			Expect(updatedWorld.Ships[0].Vel.Length()).To(BeNumerically(">", 0.0))
			Expect(updatedWorld.Ships[0].Energy).To(BeNumerically("<", 100.0))

			// Ship 2 should be unchanged
			Expect(updatedWorld.Ships[1].Rot).To(Equal(0.0))
			Expect(updatedWorld.Ships[1].Vel).To(Equal(entities.NewVec2(0.0, 0.0)))
			Expect(updatedWorld.Ships[1].Energy).To(Equal(float32(100.0)))
		})

		It("handles different inputs to different ships independently", func() {
			ship1 := entities.NewShip(1, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			ship2 := entities.NewShip(2, entities.NewVec2(0.0, 0.0), entities.NewVec2(0.0, 0.0), 0.0, 100.0)
			world := entities.NewWorld([]entities.Ship{ship1, ship2}, nil, nil)

			// Apply input to ship 1
			input1 := InputCommand{Thrust: 1.0, Turn: 0.0}
			world = ApplyInput(world, 1, input1, dt)

			// Apply different input to ship 2
			input2 := InputCommand{Thrust: 0.0, Turn: 1.0}
			world = ApplyInput(world, 2, input2, dt)

			// Both ships should be updated independently
			Expect(world.Ships[0].Vel.Length()).To(BeNumerically(">", 0.0))
			Expect(world.Ships[1].Rot).To(BeNumerically(">", 0.0))
		})

		It("returns unchanged world when Ships array is empty", func() {
			world := entities.NewWorld([]entities.Ship{}, nil, nil)

			input := InputCommand{Thrust: 1.0, Turn: 0.0}
			updatedWorld := ApplyInput(world, 1, input, dt)

			// World should be unchanged
			Expect(len(updatedWorld.Ships)).To(Equal(0))
		})

		It("preserves all other ships when updating one ship", func() {
			pos1 := entities.NewVec2(0.0, 0.0)
			vel1 := entities.NewVec2(1.0, 1.0)
			pos2 := entities.NewVec2(100.0, 100.0)
			vel2 := entities.NewVec2(2.0, 2.0)
			pos3 := entities.NewVec2(200.0, 200.0)
			vel3 := entities.NewVec2(3.0, 3.0)
			ship1 := entities.NewShip(1, pos1, vel1, 0.0, 100.0)
			ship2 := entities.NewShip(2, pos2, vel2, 0.0, 100.0)
			ship3 := entities.NewShip(3, pos3, vel3, 0.0, 100.0)
			world := entities.NewWorld([]entities.Ship{ship1, ship2, ship3}, nil, nil)

			input := InputCommand{Thrust: 1.0, Turn: 0.0}
			updatedWorld := ApplyInput(world, 2, input, dt)

			// Ship 2 should be updated
			Expect(updatedWorld.Ships[1].Vel.Length()).To(BeNumerically(">", vel2.Length()))

			// Ships 1 and 3 should be unchanged
			Expect(updatedWorld.Ships[0]).To(Equal(ship1))
			Expect(updatedWorld.Ships[2]).To(Equal(ship3))
		})
	})
})

