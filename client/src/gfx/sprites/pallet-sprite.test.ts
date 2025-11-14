/**
 * Integration tests for Pallet sprite factory.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { Graphics, Container } from 'pixi.js'
import type { PalletSnapshot } from '../../net/protocol'
import { PalletSpriteFactory } from './pallet-sprite'

describe('Pallet Sprite Factory', () => {
  let container: Container

  beforeEach(() => {
    container = new Container()
  })

  afterEach(() => {
    container.destroy({ children: true })
  })

  describe('Sprite Creation', () => {
    it('creates sprite from PalletSnapshot', () => {
      const palletState: PalletSnapshot = {
        id: 1,
        pos: { x: 200, y: 150 },
        active: true
      }

      const sprite = PalletSpriteFactory.create(palletState)

      expect(sprite).toBeDefined()
      expect(sprite).toBeInstanceOf(Graphics)
    })

    it('creates sprite with correct initial position', () => {
      const palletState: PalletSnapshot = {
        id: 1,
        pos: { x: 300, y: 400 },
        active: true
      }

      const sprite = PalletSpriteFactory.create(palletState)

      expect(sprite.x).toBe(300)
      expect(sprite.y).toBe(400)
    })

    it('creates sprite with correct active state (visible when active)', () => {
      const palletState: PalletSnapshot = {
        id: 1,
        pos: { x: 0, y: 0 },
        active: true
      }

      const sprite = PalletSpriteFactory.create(palletState)

      expect(sprite.visible).toBe(true)
    })

    it('creates sprite with correct inactive state (hidden when inactive)', () => {
      const palletState: PalletSnapshot = {
        id: 1,
        pos: { x: 0, y: 0 },
        active: false
      }

      const sprite = PalletSpriteFactory.create(palletState)

      expect(sprite.visible).toBe(false)
    })

    it('creates sprite with visual properties', () => {
      const palletState: PalletSnapshot = {
        id: 1,
        pos: { x: 0, y: 0 },
        active: true
      }

      const sprite = PalletSpriteFactory.create(palletState)

      expect(sprite).toBeInstanceOf(Graphics)
    })
  })

  describe('Sprite Updates', () => {
    it('updates sprite position from PalletSnapshot', () => {
      const palletState1: PalletSnapshot = {
        id: 1,
        pos: { x: 100, y: 200 },
        active: true
      }

      const sprite = PalletSpriteFactory.create(palletState1)
      container.addChild(sprite)

      const palletState2: PalletSnapshot = {
        id: 1,
        pos: { x: 500, y: 600 },
        active: true
      }

      PalletSpriteFactory.update(sprite, palletState2)

      expect(sprite.x).toBe(500)
      expect(sprite.y).toBe(600)
    })

    it('updates sprite visibility when active state changes to true', () => {
      const palletState1: PalletSnapshot = {
        id: 1,
        pos: { x: 0, y: 0 },
        active: false
      }

      const sprite = PalletSpriteFactory.create(palletState1)
      container.addChild(sprite)

      expect(sprite.visible).toBe(false)

      const palletState2: PalletSnapshot = {
        id: 1,
        pos: { x: 0, y: 0 },
        active: true
      }

      PalletSpriteFactory.update(sprite, palletState2)

      expect(sprite.visible).toBe(true)
    })

    it('updates sprite visibility when active state changes to false', () => {
      const palletState1: PalletSnapshot = {
        id: 1,
        pos: { x: 0, y: 0 },
        active: true
      }

      const sprite = PalletSpriteFactory.create(palletState1)
      container.addChild(sprite)

      expect(sprite.visible).toBe(true)

      const palletState2: PalletSnapshot = {
        id: 1,
        pos: { x: 0, y: 0 },
        active: false
      }

      PalletSpriteFactory.update(sprite, palletState2)

      expect(sprite.visible).toBe(false)
    })

    it('updates sprite when position and active state change', () => {
      const palletState1: PalletSnapshot = {
        id: 1,
        pos: { x: 50, y: 50 },
        active: true
      }

      const sprite = PalletSpriteFactory.create(palletState1)
      container.addChild(sprite)

      const palletState2: PalletSnapshot = {
        id: 1,
        pos: { x: 200, y: 300 },
        active: false
      }

      PalletSpriteFactory.update(sprite, palletState2)

      expect(sprite.x).toBe(200)
      expect(sprite.y).toBe(300)
      expect(sprite.visible).toBe(false)
    })
  })

  describe('Sprite Destruction', () => {
    it('destroys sprite and removes from parent', () => {
      const palletState: PalletSnapshot = {
        id: 1,
        pos: { x: 0, y: 0 },
        active: true
      }

      const sprite = PalletSpriteFactory.create(palletState)
      container.addChild(sprite)

      expect(container.children).toContain(sprite)

      PalletSpriteFactory.destroy(sprite)

      expect(container.children).not.toContain(sprite)
    })

    it('handles destruction of sprite without parent', () => {
      const palletState: PalletSnapshot = {
        id: 1,
        pos: { x: 0, y: 0 },
        active: true
      }

      const sprite = PalletSpriteFactory.create(palletState)

      // Should not throw
      expect(() => PalletSpriteFactory.destroy(sprite)).not.toThrow()
    })
  })
})

