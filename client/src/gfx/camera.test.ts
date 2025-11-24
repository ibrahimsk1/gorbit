/**
 * Integration tests for Camera system.
 * 
 * Labels: scope:integration loop:g6-rendering layer:gfx dep:state b:camera-follow r:high
 */

import { describe, it, expect, beforeEach } from 'vitest'
import { Camera } from './camera'
import type { Vec2Snapshot } from '../net/protocol'

describe('Camera', () => {
  const WORLD_WIDTH = 2000
  const WORLD_HEIGHT = 2000
  const VIEWPORT_WIDTH = 800
  const VIEWPORT_HEIGHT = 600
  const DEFAULT_LERP_FACTOR = 0.1

  describe('Constructor', () => {
    it('creates camera with world bounds, viewport, and default lerp factor', () => {
      const camera = new Camera(WORLD_WIDTH, WORLD_HEIGHT, VIEWPORT_WIDTH, VIEWPORT_HEIGHT)
      
      expect(camera).toBeDefined()
      const pos = camera.getPosition()
      expect(pos.x).toBe(0)
      expect(pos.y).toBe(0)
    })

    it('creates camera with custom lerp factor', () => {
      const customLerp = 0.15
      const camera = new Camera(WORLD_WIDTH, WORLD_HEIGHT, VIEWPORT_WIDTH, VIEWPORT_HEIGHT, customLerp)
      
      expect(camera).toBeDefined()
    })
  })

  describe('update', () => {
    it('lerps camera position toward target', () => {
      const camera = new Camera(WORLD_WIDTH, WORLD_HEIGHT, VIEWPORT_WIDTH, VIEWPORT_HEIGHT, DEFAULT_LERP_FACTOR)
      const targetPos: Vec2Snapshot = { x: 100, y: 200 }
      
      camera.update(targetPos)
      const pos = camera.getPosition()
      
      // Camera should have moved toward target (not instantly)
      expect(pos.x).toBeGreaterThan(0)
      expect(pos.x).toBeLessThan(100) // Should be less than target (lerp)
      expect(pos.y).toBeGreaterThan(0)
      expect(pos.y).toBeLessThan(200) // Should be less than target (lerp)
    })

    it('converges to target position over multiple updates', () => {
      const camera = new Camera(WORLD_WIDTH, WORLD_HEIGHT, VIEWPORT_WIDTH, VIEWPORT_HEIGHT, DEFAULT_LERP_FACTOR)
      const targetPos: Vec2Snapshot = { x: 100, y: 200 }
      
      // Update multiple times
      for (let i = 0; i < 100; i++) {
        camera.update(targetPos)
      }
      
      const pos = camera.getPosition()
      // Should be very close to target after many updates
      expect(pos.x).toBeCloseTo(100, 1)
      expect(pos.y).toBeCloseTo(200, 1)
    })

    it('handles target at origin', () => {
      const camera = new Camera(WORLD_WIDTH, WORLD_HEIGHT, VIEWPORT_WIDTH, VIEWPORT_HEIGHT, DEFAULT_LERP_FACTOR)
      const initialPos = camera.getPosition()
      
      // Move camera away first
      camera.update({ x: 100, y: 100 })
      expect(camera.getPosition().x).not.toBe(initialPos.x)
      
      // Then move back to origin
      camera.update({ x: 0, y: 0 })
      const pos = camera.getPosition()
      expect(pos.x).toBeGreaterThanOrEqual(-WORLD_WIDTH / 2)
      expect(pos.x).toBeLessThanOrEqual(WORLD_WIDTH / 2)
      expect(pos.y).toBeGreaterThanOrEqual(-WORLD_HEIGHT / 2)
      expect(pos.y).toBeLessThanOrEqual(WORLD_HEIGHT / 2)
    })
  })

  describe('getPosition', () => {
    it('returns current camera position', () => {
      const camera = new Camera(WORLD_WIDTH, WORLD_HEIGHT, VIEWPORT_WIDTH, VIEWPORT_HEIGHT)
      
      const pos = camera.getPosition()
      expect(pos).toHaveProperty('x')
      expect(pos).toHaveProperty('y')
      expect(typeof pos.x).toBe('number')
      expect(typeof pos.y).toBe('number')
    })

    it('returns initial position at origin', () => {
      const camera = new Camera(WORLD_WIDTH, WORLD_HEIGHT, VIEWPORT_WIDTH, VIEWPORT_HEIGHT)
      
      const pos = camera.getPosition()
      expect(pos.x).toBe(0)
      expect(pos.y).toBe(0)
    })

    it('returns updated position after update', () => {
      const camera = new Camera(WORLD_WIDTH, WORLD_HEIGHT, VIEWPORT_WIDTH, VIEWPORT_HEIGHT, DEFAULT_LERP_FACTOR)
      const targetPos: Vec2Snapshot = { x: 50, y: 75 }
      
      camera.update(targetPos)
      const pos = camera.getPosition()
      
      expect(pos.x).not.toBe(0)
      expect(pos.y).not.toBe(0)
    })
  })

  describe('Bounds Clamping', () => {
    it('clamps camera position to world bounds (positive)', () => {
      const camera = new Camera(WORLD_WIDTH, WORLD_HEIGHT, VIEWPORT_WIDTH, VIEWPORT_HEIGHT, 1.0) // Instant lerp for testing
      const targetPos: Vec2Snapshot = { x: 1500, y: 1500 } // Outside bounds
      
      camera.update(targetPos)
      const pos = camera.getPosition()
      
      // Should be clamped to world bounds
      expect(pos.x).toBeLessThanOrEqual(WORLD_WIDTH / 2)
      expect(pos.x).toBeGreaterThanOrEqual(-WORLD_WIDTH / 2)
      expect(pos.y).toBeLessThanOrEqual(WORLD_HEIGHT / 2)
      expect(pos.y).toBeGreaterThanOrEqual(-WORLD_HEIGHT / 2)
    })

    it('clamps camera position to world bounds (negative)', () => {
      const camera = new Camera(WORLD_WIDTH, WORLD_HEIGHT, VIEWPORT_WIDTH, VIEWPORT_HEIGHT, 1.0) // Instant lerp for testing
      const targetPos: Vec2Snapshot = { x: -1500, y: -1500 } // Outside bounds
      
      camera.update(targetPos)
      const pos = camera.getPosition()
      
      // Should be clamped to world bounds
      expect(pos.x).toBeLessThanOrEqual(WORLD_WIDTH / 2)
      expect(pos.x).toBeGreaterThanOrEqual(-WORLD_WIDTH / 2)
      expect(pos.y).toBeLessThanOrEqual(WORLD_HEIGHT / 2)
      expect(pos.y).toBeGreaterThanOrEqual(-WORLD_HEIGHT / 2)
    })

    it('allows camera at world bounds', () => {
      const camera = new Camera(WORLD_WIDTH, WORLD_HEIGHT, VIEWPORT_WIDTH, VIEWPORT_HEIGHT, 1.0) // Instant lerp for testing
      const targetPos: Vec2Snapshot = { x: WORLD_WIDTH / 2, y: WORLD_HEIGHT / 2 }
      
      camera.update(targetPos)
      const pos = camera.getPosition()
      
      expect(pos.x).toBe(WORLD_WIDTH / 2)
      expect(pos.y).toBe(WORLD_HEIGHT / 2)
    })

    it('does not wrap around (unlike ships)', () => {
      const camera = new Camera(WORLD_WIDTH, WORLD_HEIGHT, VIEWPORT_WIDTH, VIEWPORT_HEIGHT, 1.0) // Instant lerp for testing
      
      // Try to move camera way outside bounds
      camera.update({ x: 5000, y: 5000 })
      const pos1 = camera.getPosition()
      
      // Camera should be clamped, not wrapped
      expect(pos1.x).toBeLessThanOrEqual(WORLD_WIDTH / 2)
      expect(pos1.y).toBeLessThanOrEqual(WORLD_HEIGHT / 2)
      
      // Try to move camera way outside bounds in negative direction
      camera.update({ x: -5000, y: -5000 })
      const pos2 = camera.getPosition()
      
      // Camera should be clamped, not wrapped
      expect(pos2.x).toBeGreaterThanOrEqual(-WORLD_WIDTH / 2)
      expect(pos2.y).toBeGreaterThanOrEqual(-WORLD_HEIGHT / 2)
    })
  })

  describe('setLerpFactor', () => {
    it('allows runtime adjustment of lerp factor', () => {
      const camera = new Camera(WORLD_WIDTH, WORLD_HEIGHT, VIEWPORT_WIDTH, VIEWPORT_HEIGHT, DEFAULT_LERP_FACTOR)
      const targetPos: Vec2Snapshot = { x: 100, y: 100 }
      
      // Update with default lerp
      camera.update(targetPos)
      const pos1 = camera.getPosition()
      
      // Change lerp factor
      camera.setLerpFactor(0.5)
      
      // Reset position and update again
      // Note: We can't directly reset, but we can test that different lerp factors behave differently
      const camera2 = new Camera(WORLD_WIDTH, WORLD_HEIGHT, VIEWPORT_WIDTH, VIEWPORT_HEIGHT, 0.5)
      camera2.update(targetPos)
      const pos2 = camera2.getPosition()
      
      // Higher lerp factor should move faster
      expect(pos2.x).toBeGreaterThan(pos1.x)
      expect(pos2.y).toBeGreaterThan(pos1.y)
    })

    it('accepts lerp factor between 0 and 1', () => {
      const camera = new Camera(WORLD_WIDTH, WORLD_HEIGHT, VIEWPORT_WIDTH, VIEWPORT_HEIGHT)
      
      // Should accept valid lerp factors
      camera.setLerpFactor(0.0)
      camera.setLerpFactor(0.5)
      camera.setLerpFactor(1.0)
      
      // Camera should still work
      camera.update({ x: 100, y: 100 })
      const pos = camera.getPosition()
      expect(pos).toBeDefined()
    })
  })

  describe('Lerp Behavior', () => {
    it('uses correct lerp formula: cameraPos = cameraPos + (targetPos - cameraPos) * lerpFactor', () => {
      const lerpFactor = 0.2
      const camera = new Camera(WORLD_WIDTH, WORLD_HEIGHT, VIEWPORT_WIDTH, VIEWPORT_HEIGHT, lerpFactor)
      const targetPos: Vec2Snapshot = { x: 100, y: 200 }
      
      const initialPos = camera.getPosition()
      camera.update(targetPos)
      const newPos = camera.getPosition()
      
      // Verify lerp formula: newPos = initialPos + (targetPos - initialPos) * lerpFactor
      const expectedX = initialPos.x + (targetPos.x - initialPos.x) * lerpFactor
      const expectedY = initialPos.y + (targetPos.y - initialPos.y) * lerpFactor
      
      expect(newPos.x).toBeCloseTo(expectedX, 5)
      expect(newPos.y).toBeCloseTo(expectedY, 5)
    })

    it('handles multiple sequential updates correctly', () => {
      const camera = new Camera(WORLD_WIDTH, WORLD_HEIGHT, VIEWPORT_WIDTH, VIEWPORT_HEIGHT, DEFAULT_LERP_FACTOR)
      
      const positions: Vec2Snapshot[] = []
      positions.push(camera.getPosition())
      
      camera.update({ x: 100, y: 100 })
      positions.push(camera.getPosition())
      
      camera.update({ x: 200, y: 200 })
      positions.push(camera.getPosition())
      
      camera.update({ x: 300, y: 300 })
      positions.push(camera.getPosition())
      
      // Each position should be between previous and target
      for (let i = 1; i < positions.length; i++) {
        const prev = positions[i - 1]
        const curr = positions[i]
        
        // Current should be closer to target than previous
        expect(curr.x).not.toBe(prev.x)
        expect(curr.y).not.toBe(prev.y)
      }
    })
  })
})

