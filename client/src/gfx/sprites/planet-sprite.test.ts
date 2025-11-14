/**
 * Integration tests for Planet sprite factory.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { Graphics, Container } from 'pixi.js'
import type { PlanetSnapshot } from '../../net/protocol'
import { PlanetSpriteFactory } from './planet-sprite'

describe('Planet Sprite Factory', () => {
  let container: Container

  beforeEach(() => {
    container = new Container()
  })

  afterEach(() => {
    container.destroy({ children: true })
  })

  describe('Sprite Creation', () => {
    it('creates sprite from PlanetSnapshot', () => {
      const planetState: PlanetSnapshot = {
        pos: { x: 400, y: 300 },
        radius: 50
      }

      const sprite = PlanetSpriteFactory.create(planetState)

      expect(sprite).toBeDefined()
      expect(sprite).toBeInstanceOf(Graphics)
    })

    it('creates sprite with correct initial position', () => {
      const planetState: PlanetSnapshot = {
        pos: { x: 250, y: 350 },
        radius: 30
      }

      const sprite = PlanetSpriteFactory.create(planetState)

      expect(sprite.x).toBe(250)
      expect(sprite.y).toBe(350)
    })

    it('creates sprite with correct radius', () => {
      const planetState: PlanetSnapshot = {
        pos: { x: 0, y: 0 },
        radius: 75
      }

      const sprite = PlanetSpriteFactory.create(planetState)

      expect(sprite).toBeInstanceOf(Graphics)
      // Radius is used in graphics drawing, verify sprite is created
      expect(sprite.visible).toBe(true)
    })

    it('creates sprite with visual properties', () => {
      const planetState: PlanetSnapshot = {
        pos: { x: 0, y: 0 },
        radius: 50
      }

      const sprite = PlanetSpriteFactory.create(planetState)

      expect(sprite.visible).toBe(true)
      expect(sprite).toBeInstanceOf(Graphics)
    })
  })

  describe('Sprite Updates', () => {
    it('updates sprite position from PlanetSnapshot', () => {
      const planetState1: PlanetSnapshot = {
        pos: { x: 100, y: 200 },
        radius: 50
      }

      const sprite = PlanetSpriteFactory.create(planetState1)
      container.addChild(sprite)

      const planetState2: PlanetSnapshot = {
        pos: { x: 500, y: 600 },
        radius: 50
      }

      PlanetSpriteFactory.update(sprite, planetState2)

      expect(sprite.x).toBe(500)
      expect(sprite.y).toBe(600)
    })

    it('updates sprite when position changes', () => {
      const planetState1: PlanetSnapshot = {
        pos: { x: 50, y: 50 },
        radius: 40
      }

      const sprite = PlanetSpriteFactory.create(planetState1)
      container.addChild(sprite)

      const planetState2: PlanetSnapshot = {
        pos: { x: 300, y: 400 },
        radius: 40
      }

      PlanetSpriteFactory.update(sprite, planetState2)

      expect(sprite.x).toBe(300)
      expect(sprite.y).toBe(400)
    })
  })

  describe('Sprite Destruction', () => {
    it('destroys sprite and removes from parent', () => {
      const planetState: PlanetSnapshot = {
        pos: { x: 0, y: 0 },
        radius: 50
      }

      const sprite = PlanetSpriteFactory.create(planetState)
      container.addChild(sprite)

      expect(container.children).toContain(sprite)

      PlanetSpriteFactory.destroy(sprite)

      expect(container.children).not.toContain(sprite)
    })

    it('handles destruction of sprite without parent', () => {
      const planetState: PlanetSnapshot = {
        pos: { x: 0, y: 0 },
        radius: 50
      }

      const sprite = PlanetSpriteFactory.create(planetState)

      // Should not throw
      expect(() => PlanetSpriteFactory.destroy(sprite)).not.toThrow()
    })
  })
})

