/**
 * Local simulation engine matching server physics/rules for client-side prediction.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

import type { GameState } from './state-manager'
import type { Vec2Snapshot, ShipSnapshot, PlanetSnapshot, PalletSnapshot } from '../net/protocol'

// Simulation constants (matching server values)
const DT = 1.0 / 30.0 // 30Hz tick rate
const G_CONST = 1.0 // Gravitational constant
const A_MAX = 100.0 // Maximum acceleration
const PICKUP_RADIUS = 15.0 // Pallet pickup radius (about ship length for better gameplay)

// Input constants (matching server values)
const THRUST_ACCELERATION = 20.0 // Acceleration per unit thrust
const TURN_RATE = 3.0 // Angular velocity per unit turn (rad/s)
const MIN_ENERGY_FOR_THRUST = 0.0 // Minimum energy to thrust

// Energy constants (matching server values)
const MAX_ENERGY = 100.0 // Maximum energy
const THRUST_DRAIN_RATE = 0.5 // Energy drained per tick when thrusting
const PALLET_RESTORE_AMOUNT = 25.0 // Energy restored per pallet pickup

// Default planet mass (since PlanetSnapshot doesn't include mass)
const DEFAULT_PLANET_MASS = 1000.0

/**
 * Vec2 math utilities matching server entities.Vec2 behavior.
 */
function vec2Add(a: Vec2Snapshot, b: Vec2Snapshot): Vec2Snapshot {
  return { x: a.x + b.x, y: a.y + b.y }
}

function vec2Sub(a: Vec2Snapshot, b: Vec2Snapshot): Vec2Snapshot {
  return { x: a.x - b.x, y: a.y - b.y }
}

function vec2Scale(v: Vec2Snapshot, s: number): Vec2Snapshot {
  return { x: v.x * s, y: v.y * s }
}

function vec2LengthSq(v: Vec2Snapshot): number {
  return v.x * v.x + v.y * v.y
}

function vec2Length(v: Vec2Snapshot): number {
  return Math.sqrt(vec2LengthSq(v))
}

function vec2Normalize(v: Vec2Snapshot): Vec2Snapshot {
  const length = vec2Length(v)
  if (length === 0) {
    return { x: 0, y: 0 }
  }
  return vec2Scale(v, 1.0 / length)
}

function vec2Zero(): Vec2Snapshot {
  return { x: 0, y: 0 }
}

/**
 * Clamps input values to valid ranges (matching server rules.ClampInput).
 */
function clampInput(input: { thrust: number, turn: number }): { thrust: number, turn: number } {
  return {
    thrust: Math.max(0.0, Math.min(1.0, input.thrust)),
    turn: Math.max(-1.0, Math.min(1.0, input.turn))
  }
}

/**
 * Normalizes rotation angle to [0, 2π) range (matching server rules.normalizeRotation).
 */
function normalizeRotation(rot: number): number {
  let normalized = rot % (2 * Math.PI)
  if (normalized < 0) {
    normalized += 2 * Math.PI
  }
  return normalized
}

/**
 * Updates rotation based on turn input (matching server rules.UpdateRotation).
 */
function updateRotation(currentRot: number, turnInput: number, dt: number): number {
  const newRot = currentRot + TURN_RATE * turnInput * dt
  return normalizeRotation(newRot)
}

/**
 * Calculates thrust acceleration vector in direction of rotation (matching server rules.CalculateThrustAcceleration).
 * Y component is negated to match screen coordinate system where Y increases downward.
 */
function calculateThrustAcceleration(rotation: number, thrustInput: number): Vec2Snapshot {
  const directionX = Math.cos(rotation)
  const directionY = Math.sin(rotation)
  const magnitude = thrustInput * THRUST_ACCELERATION
  // Negate Y component to match screen coordinate system (Y-down)
  // World uses Y-up, but rendering flips Y, so we flip thrust Y to compensate
  return { x: directionX * magnitude, y: -directionY * magnitude }
}

/**
 * Clamps energy to valid range [0, MAX_ENERGY] (matching server rules.ClampEnergy).
 */
function clampEnergy(energy: number): number {
  return Math.max(0, Math.min(MAX_ENERGY, energy))
}

/**
 * Drains energy when thrusting (matching server rules.DrainEnergyOnThrust).
 */
function drainEnergyOnThrust(currentEnergy: number, isThrusting: boolean): number {
  if (!isThrusting) {
    return currentEnergy
  }
  return clampEnergy(currentEnergy - THRUST_DRAIN_RATE)
}

/**
 * Restores energy on pallet pickup (matching server rules.RestoreEnergyOnPickup).
 */
function restoreEnergyOnPickup(currentEnergy: number): number {
  return clampEnergy(currentEnergy + PALLET_RESTORE_AMOUNT)
}

/**
 * Applies input commands to ship (matching server rules.ApplyInput).
 */
function applyInput(ship: ShipSnapshot, input: { thrust: number, turn: number }, dt: number): ShipSnapshot {
  const clampedInput = clampInput(input)
  
  // Update rotation (always works, regardless of energy)
  const newRot = updateRotation(ship.rot, clampedInput.turn, dt)
  
  // Calculate thrust acceleration
  const thrustAcc = calculateThrustAcceleration(newRot, clampedInput.thrust)
  
  // Determine if thrust should be applied (only when energy > 0)
  const shouldThrust = ship.energy > MIN_ENERGY_FOR_THRUST && clampedInput.thrust > 0.0
  
  // Update velocity (only if thrusting and energy available)
  let newVel = { ...ship.vel }
  if (shouldThrust) {
    // Apply thrust acceleration: v_new = v_old + a * dt
    newVel = vec2Add(newVel, vec2Scale(thrustAcc, dt))
  }
  
  // Update energy (drain if thrusting)
  const newEnergy = drainEnergyOnThrust(ship.energy, shouldThrust)
  
  // Return updated ship (position unchanged by input, handled by physics step)
  return {
    pos: { ...ship.pos },
    vel: newVel,
    rot: newRot,
    energy: newEnergy
  }
}

/**
 * Calculates gravity acceleration (matching server physics.GravityAcceleration).
 */
function gravityAcceleration(
  shipPos: Vec2Snapshot,
  planetPos: Vec2Snapshot,
  planetMass: number,
  G: number,
  aMax: number
): Vec2Snapshot {
  // Handle zero mass (early return)
  if (planetMass === 0) {
    return vec2Zero()
  }
  
  // Calculate direction vector from ship to planet
  const direction = vec2Sub(planetPos, shipPos)
  const distanceSq = vec2LengthSq(direction)
  
  // Handle zero distance (ship exactly at planet position)
  if (distanceSq === 0) {
    return vec2Zero()
  }
  
  // Calculate acceleration magnitude using inverse-square law: |a| = G * M / r²
  let accMagnitude = G * planetMass / distanceSq
  
  // Clamp acceleration magnitude to aMax
  if (accMagnitude > aMax) {
    accMagnitude = aMax
  }
  
  // Normalize direction and scale by acceleration magnitude
  const directionNormalized = vec2Normalize(direction)
  return vec2Scale(directionNormalized, accMagnitude)
}

/**
 * Performs semi-implicit Euler integration (matching server physics.SemiImplicitEuler).
 */
function semiImplicitEuler(
  pos: Vec2Snapshot,
  vel: Vec2Snapshot,
  acc: Vec2Snapshot,
  dt: number
): { pos: Vec2Snapshot, vel: Vec2Snapshot } {
  // Step 1: Update velocity: v_new = v_old + a * dt
  const newVel = vec2Add(vel, vec2Scale(acc, dt))
  
  // Step 2: Update position using new velocity: p_new = p_old + v_new * dt
  const newPos = vec2Add(pos, vec2Scale(newVel, dt))
  
  return { pos: newPos, vel: newVel }
}

/**
 * Checks if ship collides with planet (matching server physics.ShipSunCollision).
 */
function shipPlanetCollision(shipPos: Vec2Snapshot, planetPos: Vec2Snapshot, planetRadius: number): boolean {
  const direction = vec2Sub(planetPos, shipPos)
  const distanceSq = vec2LengthSq(direction)
  const radiusSq = planetRadius * planetRadius
  
  // Check if distance squared <= radius squared (avoiding square root)
  return distanceSq <= radiusSq
}

/**
 * Checks if ship collides with pallet (matching server physics.ShipPalletCollision).
 */
function shipPalletCollision(shipPos: Vec2Snapshot, palletPos: Vec2Snapshot, pickupRadius: number): boolean {
  const direction = vec2Sub(palletPos, shipPos)
  const distanceSq = vec2LengthSq(direction)
  const radiusSq = pickupRadius * pickupRadius
  
  // Check if distance squared <= radius squared (avoiding square root)
  return distanceSq <= radiusSq
}

/**
 * Checks win condition (matching server rules.CheckWinCondition).
 */
function checkWinCondition(state: GameState): boolean {
  // No pallets means there is no collection objective yet
  if (state.pallets.length === 0) {
    return false
  }
  
  // Check if all pallets are collected (Active=false)
  for (const pallet of state.pallets) {
    if (pallet.active) {
      return false
    }
  }
  
  return true
}

/**
 * Checks lose condition (matching server rules.CheckLoseCondition).
 */
function checkLoseCondition(state: GameState): boolean {
  // Check collision with all planets
  for (const planet of state.planets) {
    if (shipPlanetCollision(state.ship.pos, planet.pos, planet.radius)) {
      return true
    }
  }
  return false
}

/**
 * Evaluates game state (matching server rules.EvaluateGameState).
 */
function evaluateGameState(state: GameState): GameState {
  // If game is already done, return unchanged (idempotent)
  if (state.done) {
    return state
  }
  
  // Check win condition first (takes precedence)
  if (checkWinCondition(state)) {
    return {
      ...state,
      done: true,
      win: true
    }
  }
  
  // Check lose condition
  if (checkLoseCondition(state)) {
    return {
      ...state,
      done: true,
      win: false
    }
  }
  
  // Neither condition met, leave Done and Win unchanged
  return state
}

/**
 * Performs one complete game loop step (matching server rules.Step).
 */
function step(
  state: GameState,
  input: { thrust: number, turn: number },
  dt: number,
  G: number,
  aMax: number,
  pickupRadius: number
): GameState {
  // If game is already done, skip processing and only increment tick
  if (state.done) {
    return {
      ...state,
      tick: state.tick + 1
    }
  }
  
  // Step 1: Apply Input
  // Process player input (thrust, turn) - updates rotation, velocity, and energy
  let ship = applyInput(state.ship, input, dt)
  
  // Step 2: Update Physics
  // Calculate gravity acceleration from all planets (sum accelerations)
  let totalAcc = vec2Zero()
  for (const planet of state.planets) {
    const planetAcc = gravityAcceleration(ship.pos, planet.pos, DEFAULT_PLANET_MASS, G, aMax)
    totalAcc = vec2Add(totalAcc, planetAcc)
  }
  
  // Clamp total acceleration magnitude to aMax (if needed)
  const totalAccMag = vec2Length(totalAcc)
  if (totalAccMag > aMax) {
    totalAcc = vec2Scale(vec2Normalize(totalAcc), aMax)
  }
  
  // Integrate position and velocity
  const { pos: newPos, vel: newVel } = semiImplicitEuler(ship.pos, ship.vel, totalAcc, dt)
  ship = {
    ...ship,
    pos: newPos,
    vel: newVel
  }
  
  // Step 3: Process Collisions
  // Check for pallet pickups and process them (deactivate pallet, restore energy)
  const pallets = state.pallets.map(pallet => {
    if (pallet.active && shipPalletCollision(ship.pos, pallet.pos, pickupRadius)) {
      // Deactivate pallet
      // Restore energy
      ship = {
        ...ship,
        energy: restoreEnergyOnPickup(ship.energy)
      }
      return {
        ...pallet,
        active: false
      }
    }
    return pallet
  })
  
  // Step 4: Evaluate Rules
  // Check win/lose conditions and update Done/Win flags
  let newState: GameState = {
    ...state,
    ship,
    pallets,
    tick: state.tick + 1
  }
  newState = evaluateGameState(newState)
  
  return newState
}

/**
 * Local simulator that runs deterministic physics/rules matching server behavior.
 */
export class LocalSimulator {
  private dt: number
  private G: number
  private aMax: number
  private pickupRadius: number

  constructor(
    dt: number = DT,
    G: number = G_CONST,
    aMax: number = A_MAX,
    pickupRadius: number = PICKUP_RADIUS
  ) {
    this.dt = dt
    this.G = G
    this.aMax = aMax
    this.pickupRadius = pickupRadius
  }

  /**
   * Performs one simulation step with the given input.
   */
  step(state: GameState, input: { thrust: number, turn: number }): GameState {
    return step(state, input, this.dt, this.G, this.aMax, this.pickupRadius)
  }
}

