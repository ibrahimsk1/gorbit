/**
 * Integration tests for Ship sprite factory.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { Graphics, Container } from 'pixi.js'
import type { ShipSnapshot } from '../../net/protocol'
import { ShipSpriteFactory } from './ship-sprite'

describe('Ship Sprite Factory', () => {
  let container: Container

  beforeEach(() => {
    container = new Container()
  })

  afterEach(() => {
    container.destroy({ children: true })
  })

  describe('Sprite Creation', () => {
    it('creates sprite from ShipSnapshot', () => {
      const shipState: ShipSnapshot = {
        pos: { x: 100, y: 200 },
        vel: { x: 5, y: -3 },
        rot: Math.PI / 4,
        energy: 100
      }

      const sprite = ShipSpriteFactory.create(shipState)

      expect(sprite).toBeDefined()
      expect(sprite).toBeInstanceOf(Graphics)
    })

    it('creates sprite with correct initial position', () => {
      const shipState: ShipSnapshot = {
        pos: { x: 150, y: 250 },
        vel: { x: 0, y: 0 },
        rot: 0,
        energy: 100
      }

      const sprite = ShipSpriteFactory.create(shipState)

      expect(sprite.x).toBe(150)
      expect(sprite.y).toBe(250)
    })

    it('creates sprite with correct initial rotation', () => {
      const shipState: ShipSnapshot = {
        pos: { x: 0, y: 0 },
        vel: { x: 0, y: 0 },
        rot: Math.PI / 2,
        energy: 100
      }

      const sprite = ShipSpriteFactory.create(shipState)

      expect(sprite.rotation).toBeCloseTo(Math.PI / 2)
    })

    it('creates sprite with visual properties', () => {
      const shipState: ShipSnapshot = {
        pos: { x: 0, y: 0 },
        vel: { x: 0, y: 0 },
        rot: 0,
        energy: 100
      }

      const sprite = ShipSpriteFactory.create(shipState)

      expect(sprite.visible).toBe(true)
      expect(sprite).toBeInstanceOf(Graphics)
    })
  })

  describe('Sprite Updates', () => {
    it('updates sprite position from ShipSnapshot', () => {
      const shipState1: ShipSnapshot = {
        pos: { x: 100, y: 200 },
        vel: { x: 0, y: 0 },
        rot: 0,
        energy: 100
      }

      const sprite = ShipSpriteFactory.create(shipState1)
      container.addChild(sprite)

      const shipState2: ShipSnapshot = {
        pos: { x: 300, y: 400 },
        vel: { x: 0, y: 0 },
        rot: 0,
        energy: 100
      }

      ShipSpriteFactory.update(sprite, shipState2)

      expect(sprite.x).toBe(300)
      expect(sprite.y).toBe(400)
    })

    it('updates sprite rotation from ShipSnapshot', () => {
      const shipState1: ShipSnapshot = {
        pos: { x: 0, y: 0 },
        vel: { x: 0, y: 0 },
        rot: 0,
        energy: 100
      }

      const sprite = ShipSpriteFactory.create(shipState1)
      container.addChild(sprite)

      const shipState2: ShipSnapshot = {
        pos: { x: 0, y: 0 },
        vel: { x: 0, y: 0 },
        rot: Math.PI,
        energy: 100
      }

      ShipSpriteFactory.update(sprite, shipState2)

      expect(sprite.rotation).toBeCloseTo(Math.PI)
    })

    it('updates sprite when position and rotation change', () => {
      const shipState1: ShipSnapshot = {
        pos: { x: 50, y: 50 },
        vel: { x: 0, y: 0 },
        rot: 0,
        energy: 100
      }

      const sprite = ShipSpriteFactory.create(shipState1)
      container.addChild(sprite)

      const shipState2: ShipSnapshot = {
        pos: { x: 200, y: 300 },
        vel: { x: 0, y: 0 },
        rot: Math.PI / 3,
        energy: 100
      }

      ShipSpriteFactory.update(sprite, shipState2)

      expect(sprite.x).toBe(200)
      expect(sprite.y).toBe(300)
      expect(sprite.rotation).toBeCloseTo(Math.PI / 3)
    })
  })

  describe('Sprite Destruction', () => {
    it('destroys sprite and removes from parent', () => {
      const shipState: ShipSnapshot = {
        pos: { x: 0, y: 0 },
        vel: { x: 0, y: 0 },
        rot: 0,
        energy: 100
      }

      const sprite = ShipSpriteFactory.create(shipState)
      container.addChild(sprite)

      expect(container.children).toContain(sprite)

      ShipSpriteFactory.destroy(sprite)

      expect(container.children).not.toContain(sprite)
    })

    it('handles destruction of sprite without parent', () => {
      const shipState: ShipSnapshot = {
        pos: { x: 0, y: 0 },
        vel: { x: 0, y: 0 },
        rot: 0,
        energy: 100
      }

      const sprite = ShipSpriteFactory.create(shipState)

      // Should not throw
      expect(() => ShipSpriteFactory.destroy(sprite)).not.toThrow()
    })
  })
})

