/**
 * Integration tests for protocol types, type guards, and validation.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

import { describe, it, expect } from 'vitest'
import {
  Vec2Snapshot,
  ShipSnapshot,
  PlanetSnapshot,
  PalletSnapshot,
  InputMessage,
  RestartMessage,
  SnapshotMessage,
  Message,
  isInputMessage,
  isSnapshotMessage,
  isRestartMessage,
  isMessage,
  validateVec2Snapshot,
  validateShipSnapshot,
  validatePlanetSnapshot,
  validatePalletSnapshot,
  createInputMessage,
  createRestartMessage,
  PROTOCOL_VERSION
} from './protocol'

describe('Protocol Type Definitions', () => {
  it('should define Vec2Snapshot with x and y', () => {
    const vec: Vec2Snapshot = { x: 10.5, y: 20.3 }
    expect(vec.x).toBe(10.5)
    expect(vec.y).toBe(20.3)
  })

  it('should define ShipSnapshot with all required fields', () => {
    const ship: ShipSnapshot = {
      pos: { x: 10, y: 20 },
      vel: { x: 1, y: 2 },
      rot: 1.5,
      energy: 75.0
    }
    expect(ship.pos.x).toBe(10)
    expect(ship.vel.y).toBe(2)
    expect(ship.rot).toBe(1.5)
    expect(ship.energy).toBe(75.0)
  })

  it('should define PlanetSnapshot with pos and radius', () => {
    const planet: PlanetSnapshot = {
      pos: { x: 0, y: 0 },
      radius: 5.0
    }
    expect(planet.pos.x).toBe(0)
    expect(planet.radius).toBe(5.0)
  })

  it('should define PalletSnapshot with id, pos, and active', () => {
    const pallet: PalletSnapshot = {
      id: 1,
      pos: { x: 5, y: 5 },
      active: true
    }
    expect(pallet.id).toBe(1)
    expect(pallet.active).toBe(true)
  })

  it('should define InputMessage with correct structure', () => {
    const input: InputMessage = {
      t: 'input',
      seq: 1,
      thrust: 0.5,
      turn: 0.3
    }
    expect(input.t).toBe('input')
    expect(input.seq).toBe(1)
    expect(input.thrust).toBe(0.5)
    expect(input.turn).toBe(0.3)
  })

  it('should define RestartMessage with correct structure', () => {
    const restart: RestartMessage = {
      t: 'restart'
    }
    expect(restart.t).toBe('restart')
  })

  it('should define SnapshotMessage with planets array', () => {
    const snapshot: SnapshotMessage = {
      t: 'snapshot',
      tick: 42,
      ship: {
        pos: { x: 10, y: 20 },
        vel: { x: 1, y: 2 },
        rot: 1.5,
        energy: 75.0
      },
      planets: [
        {
          pos: { x: 0, y: 0 },
          radius: 5.0
        }
      ],
      pallets: [
        {
          id: 1,
          pos: { x: 5, y: 5 },
          active: true
        }
      ],
      done: false,
      win: false
    }
    expect(snapshot.t).toBe('snapshot')
    expect(snapshot.tick).toBe(42)
    expect(snapshot.planets).toHaveLength(1)
    expect(snapshot.pallets).toHaveLength(1)
  })

  it('should support multiple planets in array', () => {
    const snapshot: SnapshotMessage = {
      t: 'snapshot',
      tick: 1,
      ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
      planets: [
        { pos: { x: 0, y: 0 }, radius: 5.0 },
        { pos: { x: 100, y: 100 }, radius: 3.0 }
      ],
      pallets: [],
      done: false,
      win: false
    }
    expect(snapshot.planets).toHaveLength(2)
    expect(snapshot.planets[0].radius).toBe(5.0)
    expect(snapshot.planets[1].radius).toBe(3.0)
  })

  it('should support empty planets array', () => {
    const snapshot: SnapshotMessage = {
      t: 'snapshot',
      tick: 1,
      ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
      planets: [],
      pallets: [],
      done: false,
      win: false
    }
    expect(snapshot.planets).toHaveLength(0)
  })

  it('should support multiple pallets in array', () => {
    const snapshot: SnapshotMessage = {
      t: 'snapshot',
      tick: 1,
      ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
      planets: [{ pos: { x: 0, y: 0 }, radius: 5.0 }],
      pallets: [
        { id: 1, pos: { x: 5, y: 5 }, active: true },
        { id: 2, pos: { x: 10, y: 10 }, active: false },
        { id: 3, pos: { x: 15, y: 15 }, active: true }
      ],
      done: false,
      win: false
    }
    expect(snapshot.pallets).toHaveLength(3)
    expect(snapshot.pallets[0].active).toBe(true)
    expect(snapshot.pallets[1].active).toBe(false)
  })

  it('should support Message union type', () => {
    const input: Message = { t: 'input', seq: 1, thrust: 0.5, turn: 0.3 }
    const restart: Message = { t: 'restart' }
    const snapshot: Message = {
      t: 'snapshot',
      tick: 1,
      ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
      planets: [],
      pallets: [],
      done: false,
      win: false
    }
    expect(input.t).toBe('input')
    expect(restart.t).toBe('restart')
    expect(snapshot.t).toBe('snapshot')
  })
})

describe('Type Guards', () => {
  describe('isInputMessage', () => {
    it('should return true for valid InputMessage', () => {
      const msg = { t: 'input', seq: 1, thrust: 0.5, turn: 0.3 }
      expect(isInputMessage(msg)).toBe(true)
    })

    it('should return false for invalid type discriminator', () => {
      const msg = { t: 'snapshot', seq: 1, thrust: 0.5, turn: 0.3 }
      expect(isInputMessage(msg)).toBe(false)
    })

    it('should return false for missing fields', () => {
      const msg = { t: 'input', seq: 1 }
      expect(isInputMessage(msg)).toBe(false)
    })

    it('should return false for wrong types', () => {
      const msg = { t: 'input', seq: '1', thrust: 0.5, turn: 0.3 }
      expect(isInputMessage(msg)).toBe(false)
    })

    it('should return false for null', () => {
      expect(isInputMessage(null)).toBe(false)
    })

    it('should return false for undefined', () => {
      expect(isInputMessage(undefined)).toBe(false)
    })
  })

  describe('isSnapshotMessage', () => {
    it('should return true for valid SnapshotMessage', () => {
      const msg = {
        t: 'snapshot',
        tick: 42,
        ship: { pos: { x: 10, y: 20 }, vel: { x: 1, y: 2 }, rot: 1.5, energy: 75 },
        planets: [{ pos: { x: 0, y: 0 }, radius: 5.0 }],
        pallets: [{ id: 1, pos: { x: 5, y: 5 }, active: true }],
        done: false,
        win: false
      }
      expect(isSnapshotMessage(msg)).toBe(true)
    })

    it('should return false for invalid type discriminator', () => {
      const msg = { t: 'input', tick: 42 }
      expect(isSnapshotMessage(msg)).toBe(false)
    })

    it('should return false for missing required fields', () => {
      const msg = { t: 'snapshot', tick: 42 }
      expect(isSnapshotMessage(msg)).toBe(false)
    })

    it('should return false for invalid planets array', () => {
      const msg = {
        t: 'snapshot',
        tick: 42,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: 'not-an-array',
        pallets: [],
        done: false,
        win: false
      }
      expect(isSnapshotMessage(msg)).toBe(false)
    })

    it('should return false for null', () => {
      expect(isSnapshotMessage(null)).toBe(false)
    })
  })

  describe('isRestartMessage', () => {
    it('should return true for valid RestartMessage', () => {
      const msg = { t: 'restart' }
      expect(isRestartMessage(msg)).toBe(true)
    })

    it('should return false for invalid type discriminator', () => {
      const msg = { t: 'input' }
      expect(isRestartMessage(msg)).toBe(false)
    })

    it('should return false for null', () => {
      expect(isRestartMessage(null)).toBe(false)
    })
  })

  describe('isMessage', () => {
    it('should return true for InputMessage', () => {
      const msg = { t: 'input', seq: 1, thrust: 0.5, turn: 0.3 }
      expect(isMessage(msg)).toBe(true)
    })

    it('should return true for RestartMessage', () => {
      const msg = { t: 'restart' }
      expect(isMessage(msg)).toBe(true)
    })

    it('should return true for SnapshotMessage', () => {
      const msg = {
        t: 'snapshot',
        tick: 42,
        ship: { pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 },
        planets: [],
        pallets: [],
        done: false,
        win: false
      }
      expect(isMessage(msg)).toBe(true)
    })

    it('should return false for unknown message type', () => {
      const msg = { t: 'unknown' }
      expect(isMessage(msg)).toBe(false)
    })

    it('should return false for null', () => {
      expect(isMessage(null)).toBe(false)
    })
  })
})

describe('Validation Functions', () => {
  describe('validateVec2Snapshot', () => {
    it('should return Vec2Snapshot for valid data', () => {
      const result = validateVec2Snapshot({ x: 10.5, y: 20.3 })
      expect(result).toEqual({ x: 10.5, y: 20.3 })
    })

    it('should return null for missing x', () => {
      const result = validateVec2Snapshot({ y: 20.3 })
      expect(result).toBeNull()
    })

    it('should return null for missing y', () => {
      const result = validateVec2Snapshot({ x: 10.5 })
      expect(result).toBeNull()
    })

    it('should return null for wrong types', () => {
      expect(validateVec2Snapshot({ x: '10', y: 20 })).toBeNull()
      expect(validateVec2Snapshot({ x: 10, y: '20' })).toBeNull()
    })

    it('should return null for null', () => {
      expect(validateVec2Snapshot(null)).toBeNull()
    })

    it('should return null for undefined', () => {
      expect(validateVec2Snapshot(undefined)).toBeNull()
    })
  })

  describe('validateShipSnapshot', () => {
    it('should return ShipSnapshot for valid data', () => {
      const data = {
        pos: { x: 10, y: 20 },
        vel: { x: 1, y: 2 },
        rot: 1.5,
        energy: 75.0
      }
      const result = validateShipSnapshot(data)
      expect(result).toEqual(data)
    })

    it('should return null for missing fields', () => {
      expect(validateShipSnapshot({ pos: { x: 10, y: 20 } })).toBeNull()
    })

    it('should return null for invalid nested types', () => {
      const data = {
        pos: 'invalid',
        vel: { x: 1, y: 2 },
        rot: 1.5,
        energy: 75.0
      }
      expect(validateShipSnapshot(data)).toBeNull()
    })

    it('should return null for null', () => {
      expect(validateShipSnapshot(null)).toBeNull()
    })
  })

  describe('validatePlanetSnapshot', () => {
    it('should return PlanetSnapshot for valid data', () => {
      const data = { pos: { x: 0, y: 0 }, radius: 5.0 }
      const result = validatePlanetSnapshot(data)
      expect(result).toEqual(data)
    })

    it('should return null for missing fields', () => {
      expect(validatePlanetSnapshot({ pos: { x: 0, y: 0 } })).toBeNull()
      expect(validatePlanetSnapshot({ radius: 5.0 })).toBeNull()
    })

    it('should return null for invalid types', () => {
      expect(validatePlanetSnapshot({ pos: { x: 0, y: 0 }, radius: '5' })).toBeNull()
    })

    it('should return null for null', () => {
      expect(validatePlanetSnapshot(null)).toBeNull()
    })
  })

  describe('validatePalletSnapshot', () => {
    it('should return PalletSnapshot for valid data', () => {
      const data = { id: 1, pos: { x: 5, y: 5 }, active: true }
      const result = validatePalletSnapshot(data)
      expect(result).toEqual(data)
    })

    it('should return null for missing fields', () => {
      expect(validatePalletSnapshot({ id: 1, pos: { x: 5, y: 5 } })).toBeNull()
    })

    it('should return null for invalid types', () => {
      expect(validatePalletSnapshot({ id: '1', pos: { x: 5, y: 5 }, active: true })).toBeNull()
    })

    it('should return null for null', () => {
      expect(validatePalletSnapshot(null)).toBeNull()
    })
  })
})

describe('Helper Functions', () => {
  describe('createInputMessage', () => {
    it('should create valid InputMessage', () => {
      const msg = createInputMessage(1, 0.5, 0.3)
      expect(msg).toEqual({
        t: 'input',
        seq: 1,
        thrust: 0.5,
        turn: 0.3
      })
    })
  })

  describe('createRestartMessage', () => {
    it('should create valid RestartMessage', () => {
      const msg = createRestartMessage()
      expect(msg).toEqual({ t: 'restart' })
    })
  })
})

describe('Protocol Versioning', () => {
  it('should define PROTOCOL_VERSION constant', () => {
    expect(PROTOCOL_VERSION).toBeDefined()
    expect(typeof PROTOCOL_VERSION).toBe('number')
  })

  it('should support optional version field in messages', () => {
    const input: InputMessage = {
      t: 'input',
      seq: 1,
      thrust: 0.5,
      turn: 0.3,
      version: 1
    }
    expect(input.version).toBe(1)
  })
})

