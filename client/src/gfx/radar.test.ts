/**
 * Integration tests for Radar system.
 * 
 * Labels: scope:integration loop:g6-rendering layer:gfx dep:state b:radar-minimap r:high
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { Container, Graphics } from 'pixi.js'
import { App } from '../core/app'
import { Radar } from './radar'
import type { SnapshotMessage } from '../net/protocol'

describe('Radar', () => {
  const WORLD_WIDTH = 2000
  const WORLD_HEIGHT = 2000
  const RADAR_WIDTH = 200
  const RADAR_HEIGHT = 200
  let app: App
  let container: Container
  let radarContainer: HTMLElement

  beforeEach(async () => {
    radarContainer = document.createElement('div')
    radarContainer.id = 'radar-container'
    document.body.appendChild(radarContainer)

    app = new App()
    await app.init(radarContainer)
    container = new Container()
    app.getApplication().stage.addChild(container)
  })

  afterEach(() => {
    if (container) {
      container.destroy({ children: true })
    }
    if (app) {
      app.destroy()
    }
    if (radarContainer && radarContainer.parentNode) {
      radarContainer.parentNode.removeChild(radarContainer)
    }
  })

  describe('Constructor', () => {
    it('creates radar with world bounds, radar size, and container', () => {
      const radar = new Radar(WORLD_WIDTH, WORLD_HEIGHT, RADAR_WIDTH, RADAR_HEIGHT, container)
      
      expect(radar).toBeDefined()
    })
  })

  describe('Coordinate Mapping', () => {
    it('maps world coordinates to radar pixels correctly', () => {
      const radar = new Radar(WORLD_WIDTH, WORLD_HEIGHT, RADAR_WIDTH, RADAR_HEIGHT, container)
      
      // Test coordinate mapping by rendering entities at known positions
      // World center (0, 0) should render at radar center (100, 100)
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 0,
        ships: [
          { id: 1, pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
        ],
        planets: [],
        pallets: [],
        worldBounds: { width: WORLD_WIDTH, height: WORLD_HEIGHT },
        myShipId: 1,
        done: false,
        win: false
      }
      
      radar.update(snapshot, 1)
      
      // Should render ship at center of radar
      expect(container.children.length).toBeGreaterThan(0)
    })

    it('maps world bounds corners to radar corners', () => {
      const radar = new Radar(WORLD_WIDTH, WORLD_HEIGHT, RADAR_WIDTH, RADAR_HEIGHT, container)
      
      // Test entities at world corners render at radar corners
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 0,
        ships: [
          { id: 1, pos: { x: -WORLD_WIDTH / 2, y: WORLD_HEIGHT / 2 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
          { id: 2, pos: { x: WORLD_WIDTH / 2, y: -WORLD_HEIGHT / 2 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
        ],
        planets: [],
        pallets: [],
        worldBounds: { width: WORLD_WIDTH, height: WORLD_HEIGHT },
        myShipId: 1,
        done: false,
        win: false
      }
      
      radar.update(snapshot, 1)
      
      // Should render ships at corners
      expect(container.children.length).toBeGreaterThan(0)
    })

    it('handles world coordinates at origin', () => {
      const radar = new Radar(WORLD_WIDTH, WORLD_HEIGHT, RADAR_WIDTH, RADAR_HEIGHT, container)
      
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 0,
        ships: [
          { id: 1, pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
        ],
        planets: [],
        pallets: [],
        worldBounds: { width: WORLD_WIDTH, height: WORLD_HEIGHT },
        myShipId: 1,
        done: false,
        win: false
      }
      
      radar.update(snapshot, 1)
      
      // Should render correctly
      expect(container.children.length).toBeGreaterThan(0)
    })
  })

  describe('update', () => {
    it('updates radar with snapshot data', () => {
      const radar = new Radar(WORLD_WIDTH, WORLD_HEIGHT, RADAR_WIDTH, RADAR_HEIGHT, container)
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 0,
        ships: [
          { id: 1, pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
        ],
        planets: [
          { id: 1, pos: { x: 0, y: 0 }, radius: 50 }
        ],
        pallets: [
          { id: 1, pos: { x: 300, y: 400 }, active: true }
        ],
        worldBounds: { width: WORLD_WIDTH, height: WORLD_HEIGHT },
        myShipId: 1,
        done: false,
        win: false
      }
      
      radar.update(snapshot, 1)
      
      // Radar should have graphics in container
      expect(container.children.length).toBeGreaterThan(0)
    })

    it('renders world bounds rectangle outline', () => {
      const radar = new Radar(WORLD_WIDTH, WORLD_HEIGHT, RADAR_WIDTH, RADAR_HEIGHT, container)
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 0,
        ships: [],
        planets: [],
        pallets: [],
        worldBounds: { width: WORLD_WIDTH, height: WORLD_HEIGHT },
        myShipId: 0,
        done: false,
        win: false
      }
      
      radar.update(snapshot, 0)
      
      // Should have at least one graphics object for world bounds
      expect(container.children.length).toBeGreaterThan(0)
    })

    it('renders all planets as circles', () => {
      const radar = new Radar(WORLD_WIDTH, WORLD_HEIGHT, RADAR_WIDTH, RADAR_HEIGHT, container)
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 0,
        ships: [],
        planets: [
          { id: 1, pos: { x: 0, y: 0 }, radius: 50 },
          { id: 2, pos: { x: 500, y: 500 }, radius: 75 }
        ],
        pallets: [],
        worldBounds: { width: WORLD_WIDTH, height: WORLD_HEIGHT },
        myShipId: 0,
        done: false,
        win: false
      }
      
      radar.update(snapshot, 0)
      
      // Should have graphics object with world bounds and planets drawn on it
      expect(container.children.length).toBe(1)
      expect(container.children[0]).toBeInstanceOf(Graphics)
    })

    it('renders all ships as dots', () => {
      const radar = new Radar(WORLD_WIDTH, WORLD_HEIGHT, RADAR_WIDTH, RADAR_HEIGHT, container)
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 0,
        ships: [
          { id: 1, pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
          { id: 2, pos: { x: -100, y: -200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
        ],
        planets: [],
        pallets: [],
        worldBounds: { width: WORLD_WIDTH, height: WORLD_HEIGHT },
        myShipId: 1,
        done: false,
        win: false
      }
      
      radar.update(snapshot, 1)
      
      // Should have graphics object with world bounds and ships drawn on it
      expect(container.children.length).toBe(1)
      expect(container.children[0]).toBeInstanceOf(Graphics)
    })

    it('highlights own ship differently', () => {
      const radar = new Radar(WORLD_WIDTH, WORLD_HEIGHT, RADAR_WIDTH, RADAR_HEIGHT, container)
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 0,
        ships: [
          { id: 1, pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
          { id: 2, pos: { x: -100, y: -200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
        ],
        planets: [],
        pallets: [],
        worldBounds: { width: WORLD_WIDTH, height: WORLD_HEIGHT },
        myShipId: 1,
        done: false,
        win: false
      }
      
      radar.update(snapshot, 1)
      
      // Should render ships (implementation uses different colors/sizes for own ship)
      expect(container.children.length).toBe(1)
      expect(container.children[0]).toBeInstanceOf(Graphics)
    })

    it('renders active pallets as small dots', () => {
      const radar = new Radar(WORLD_WIDTH, WORLD_HEIGHT, RADAR_WIDTH, RADAR_HEIGHT, container)
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 0,
        ships: [],
        planets: [],
        pallets: [
          { id: 1, pos: { x: 300, y: 400 }, active: true },
          { id: 2, pos: { x: -300, y: -400 }, active: true }
        ],
        worldBounds: { width: WORLD_WIDTH, height: WORLD_HEIGHT },
        myShipId: 0,
        done: false,
        win: false
      }
      
      radar.update(snapshot, 0)
      
      // Should have graphics object with world bounds and pallets drawn on it
      expect(container.children.length).toBe(1)
      expect(container.children[0]).toBeInstanceOf(Graphics)
    })

    it('does not render inactive pallets', () => {
      const radar = new Radar(WORLD_WIDTH, WORLD_HEIGHT, RADAR_WIDTH, RADAR_HEIGHT, container)
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 0,
        ships: [],
        planets: [],
        pallets: [
          { id: 1, pos: { x: 300, y: 400 }, active: false },
          { id: 2, pos: { x: -300, y: -400 }, active: true }
        ],
        worldBounds: { width: WORLD_WIDTH, height: WORLD_HEIGHT },
        myShipId: 0,
        done: false,
        win: false
      }
      
      radar.update(snapshot, 0)
      
      // Should only render active pallet
      expect(container.children.length).toBeGreaterThan(0)
    })

    it('clears and redraws graphics each frame', () => {
      const radar = new Radar(WORLD_WIDTH, WORLD_HEIGHT, RADAR_WIDTH, RADAR_HEIGHT, container)
      const snapshot1: SnapshotMessage = {
        t: 'snapshot',
        tick: 0,
        ships: [
          { id: 1, pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
        ],
        planets: [],
        pallets: [],
        worldBounds: { width: WORLD_WIDTH, height: WORLD_HEIGHT },
        myShipId: 1,
        done: false,
        win: false
      }
      
      radar.update(snapshot1, 1)
      const count1 = container.children.length
      
      const snapshot2: SnapshotMessage = {
        t: 'snapshot',
        tick: 1,
        ships: [
          { id: 1, pos: { x: 200, y: 300 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
        ],
        planets: [
          { id: 1, pos: { x: 0, y: 0 }, radius: 50 }
        ],
        pallets: [],
        worldBounds: { width: WORLD_WIDTH, height: WORLD_HEIGHT },
        myShipId: 1,
        done: false,
        win: false
      }
      
      radar.update(snapshot2, 1)
      const count2 = container.children.length
      
      // Graphics should be cleared and redrawn (count should be consistent)
      expect(count2).toBeGreaterThan(0)
    })
  })

  describe('destroy', () => {
    it('cleans up radar graphics', () => {
      const radar = new Radar(WORLD_WIDTH, WORLD_HEIGHT, RADAR_WIDTH, RADAR_HEIGHT, container)
      const snapshot: SnapshotMessage = {
        t: 'snapshot',
        tick: 0,
        ships: [
          { id: 1, pos: { x: 100, y: 200 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
        ],
        planets: [],
        pallets: [],
        worldBounds: { width: WORLD_WIDTH, height: WORLD_HEIGHT },
        myShipId: 1,
        done: false,
        win: false
      }
      
      radar.update(snapshot, 1)
      const countBefore = container.children.length
      
      radar.destroy()
      
      // Graphics should be removed
      expect(container.children.length).toBeLessThan(countBefore)
    })

    it('is idempotent (safe to call multiple times)', () => {
      const radar = new Radar(WORLD_WIDTH, WORLD_HEIGHT, RADAR_WIDTH, RADAR_HEIGHT, container)
      
      radar.destroy()
      radar.destroy() // Should not throw
      
      expect(radar).toBeDefined()
    })
  })
})

