package physics

import "github.com/gorbit/orbitalrush/internal/sim/entities"

// SemiImplicitEuler performs a semi-implicit Euler (symplectic Euler) integration step.
// This method updates velocity first, then uses the new velocity to update position.
//
// Algorithm:
//  1. v_new = v_old + a * dt
//  2. p_new = p_old + v_new * dt
//
// This method is symplectic, meaning it better conserves energy compared to
// explicit Euler, making it suitable for physics simulations.
//
// Parameters:
//   - pos: Current position
//   - vel: Current velocity
//   - acc: Acceleration (constant for this step)
//   - dt: Time step
//
// Returns:
//   - newPos: Updated position
//   - newVel: Updated velocity
func SemiImplicitEuler(pos, vel, acc entities.Vec2, dt float64) (newPos, newVel entities.Vec2) {
	// Step 1: Update velocity: v_new = v_old + a * dt
	newVel = vel.Add(acc.Scale(dt))

	// Step 2: Update position using new velocity: p_new = p_old + v_new * dt
	newPos = pos.Add(newVel.Scale(dt))

	return newPos, newVel
}
