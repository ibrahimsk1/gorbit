/**
 * Contract tests for protocol types, type guards, and validation.
 * 
 * Labels: scope:contract loop:g4-proto layer:contract
 */

import { describe, it, expect } from 'vitest'
import {
  Vec2Snapshot,
  ShipSnapshot,
  PlanetSnapshot,
  PalletSnapshot,
  WorldBounds,
  InputMessage,
  RestartMessage,
  SnapshotMessage,
  Message,
  PlayerInfo,
  CreateRoomMessage,
  JoinRoomMessage,
  LeaveRoomMessage,
  StartMatchMessage,
  RoomCreatedMessage,
  RoomStateMessage,
  PlayerJoinedMessage,
  PlayerLeftMessage,
  MatchStartedMessage,
  MatchEndedMessage,
  isInputMessage,
  isSnapshotMessage,
  isRestartMessage,
  isMessage,
  isPlayerInfo,
  isCreateRoomMessage,
  isJoinRoomMessage,
  isLeaveRoomMessage,
  isStartMatchMessage,
  isRoomCreatedMessage,
  isRoomStateMessage,
  isPlayerJoinedMessage,
  isPlayerLeftMessage,
  isMatchStartedMessage,
  isMatchEndedMessage,
  validateVec2Snapshot,
  validateShipSnapshot,
  validatePlanetSnapshot,
  validatePalletSnapshot,
  createInputMessage,
  createRestartMessage,
  createCreateRoomMessage,
  createJoinRoomMessage,
  createLeaveRoomMessage,
  createStartMatchMessage,
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

  // v0 format (legacy) - will be updated in cu/protocol-tests
  it('should define SnapshotMessage with planets array (v0 format)', () => {
    const snapshot: any = {
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

  // v1 multiplayer format
  // Labels: scope:contract loop:g4-proto layer:contract b:snapshot-format
  it('should define WorldBounds with width and height', () => {
    const bounds: WorldBounds = { width: 2000.0, height: 2000.0 }
    expect(bounds.width).toBe(2000.0)
    expect(bounds.height).toBe(2000.0)
  })

  it('should define ShipSnapshot with id field (v1 format)', () => {
    const ship: ShipSnapshot = {
      id: 1,
      pos: { x: 10, y: 20 },
      vel: { x: 1, y: 2 },
      rot: 1.5,
      energy: 75.0
    }
    expect(ship.id).toBe(1)
    expect(ship.pos.x).toBe(10)
    expect(ship.vel.y).toBe(2)
    expect(ship.rot).toBe(1.5)
    expect(ship.energy).toBe(75.0)
  })

  it('should define PlanetSnapshot with id field (v1 format)', () => {
    const planet: PlanetSnapshot = {
      id: 1,
      pos: { x: 0, y: 0 },
      radius: 5.0
    }
    expect(planet.id).toBe(1)
    expect(planet.pos.x).toBe(0)
    expect(planet.radius).toBe(5.0)
  })

  it('should define SnapshotMessage with ships array, worldBounds, and myShipId (v1 format)', () => {
    const snapshot: SnapshotMessage = {
      t: 'snapshot',
      tick: 42,
      ships: [
        {
          id: 1,
          pos: { x: 10, y: 20 },
          vel: { x: 1, y: 2 },
          rot: 1.5,
          energy: 75.0
        },
        {
          id: 2,
          pos: { x: 30, y: 40 },
          vel: { x: 0, y: 0 },
          rot: 0.0,
          energy: 100.0
        }
      ],
      planets: [
        {
          id: 1,
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
      worldBounds: { width: 2000.0, height: 2000.0 },
      myShipId: 1,
      done: false,
      win: false
    }
    expect(snapshot.t).toBe('snapshot')
    expect(snapshot.tick).toBe(42)
    expect(snapshot.ships).toHaveLength(2)
    expect(snapshot.ships[0].id).toBe(1)
    expect(snapshot.ships[1].id).toBe(2)
    expect(snapshot.worldBounds.width).toBe(2000.0)
    expect(snapshot.worldBounds.height).toBe(2000.0)
    expect(snapshot.myShipId).toBe(1)
    expect(snapshot.planets).toHaveLength(1)
    expect(snapshot.planets[0].id).toBe(1)
    expect(snapshot.pallets).toHaveLength(1)
  })

  // v0 format (legacy) - will be updated in cu/protocol-tests
  it('should support multiple planets in array (v0 format)', () => {
    const snapshot: any = {
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

  it('should support empty planets array (v0 format)', () => {
    const snapshot: any = {
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

  it('should support multiple pallets in array (v0 format)', () => {
    const snapshot: any = {
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
      ships: [
        { id: 1, pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
      ],
      planets: [
        { id: 1, pos: { x: 0, y: 0 }, radius: 5.0 }
      ],
      pallets: [],
      worldBounds: { width: 2000.0, height: 2000.0 },
      myShipId: 1,
      done: false,
      win: false
    }
    expect(input.t).toBe('input')
    expect(restart.t).toBe('restart')
    expect(snapshot.t).toBe('snapshot')
  })

  // Room management message types
  // Labels: scope:contract loop:g4-proto layer:contract b:room-messages
  it('should define PlayerInfo with id and name', () => {
    const player: PlayerInfo = { id: 1, name: 'Player1' }
    expect(player.id).toBe(1)
    expect(player.name).toBe('Player1')
  })

  it('should define CreateRoomMessage with correct structure', () => {
    const msg: CreateRoomMessage = { t: 'createRoom' }
    expect(msg.t).toBe('createRoom')
  })

  it('should define JoinRoomMessage with roomCode', () => {
    const msg: JoinRoomMessage = { t: 'joinRoom', roomCode: 'ABC123' }
    expect(msg.t).toBe('joinRoom')
    expect(msg.roomCode).toBe('ABC123')
  })

  it('should define LeaveRoomMessage with correct structure', () => {
    const msg: LeaveRoomMessage = { t: 'leaveRoom' }
    expect(msg.t).toBe('leaveRoom')
  })

  it('should define StartMatchMessage with correct structure', () => {
    const msg: StartMatchMessage = { t: 'startMatch' }
    expect(msg.t).toBe('startMatch')
  })

  it('should define RoomCreatedMessage with roomCode', () => {
    const msg: RoomCreatedMessage = { t: 'roomCreated', roomCode: 'XYZ789' }
    expect(msg.t).toBe('roomCreated')
    expect(msg.roomCode).toBe('XYZ789')
  })

  it('should define RoomStateMessage with roomCode, players, state, and hostId', () => {
    const msg: RoomStateMessage = {
      t: 'roomState',
      roomCode: 'ABC123',
      players: [{ id: 1, name: 'Player1' }, { id: 2, name: 'Player2' }],
      state: 'lobby',
      hostId: 1
    }
    expect(msg.t).toBe('roomState')
    expect(msg.roomCode).toBe('ABC123')
    expect(msg.players).toHaveLength(2)
    expect(msg.state).toBe('lobby')
    expect(msg.hostId).toBe(1)
  })

  it('should define PlayerJoinedMessage with player', () => {
    const msg: PlayerJoinedMessage = {
      t: 'playerJoined',
      player: { id: 2, name: 'Player2' }
    }
    expect(msg.t).toBe('playerJoined')
    expect(msg.player.id).toBe(2)
    expect(msg.player.name).toBe('Player2')
  })

  it('should define PlayerLeftMessage with playerId', () => {
    const msg: PlayerLeftMessage = { t: 'playerLeft', playerId: 2 }
    expect(msg.t).toBe('playerLeft')
    expect(msg.playerId).toBe(2)
  })

  it('should define MatchStartedMessage with correct structure', () => {
    const msg: MatchStartedMessage = { t: 'matchStarted' }
    expect(msg.t).toBe('matchStarted')
  })

  it('should define MatchEndedMessage with optional winnerId', () => {
    const msg: MatchEndedMessage = { t: 'matchEnded', winnerId: 1 }
    expect(msg.t).toBe('matchEnded')
    expect(msg.winnerId).toBe(1)
  })

  it('should support MatchEndedMessage without winnerId', () => {
    const msg: MatchEndedMessage = { t: 'matchEnded' }
    expect(msg.t).toBe('matchEnded')
    expect(msg.winnerId).toBeUndefined()
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
    it('should return true for valid SnapshotMessage (v1 format)', () => {
      const msg = {
        t: 'snapshot',
        tick: 42,
        ships: [
          { id: 1, pos: { x: 10, y: 20 }, vel: { x: 1, y: 2 }, rot: 1.5, energy: 75 }
        ],
        planets: [{ id: 1, pos: { x: 0, y: 0 }, radius: 5.0 }],
        pallets: [{ id: 1, pos: { x: 5, y: 5 }, active: true }],
        worldBounds: { width: 2000.0, height: 2000.0 },
        myShipId: 1,
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
        ships: [
          { id: 1, pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
        ],
        planets: 'not-an-array',
        pallets: [],
        worldBounds: { width: 2000.0, height: 2000.0 },
        myShipId: 1,
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

    it('should return true for SnapshotMessage (v1 format)', () => {
      const msg = {
        t: 'snapshot',
        tick: 42,
        ships: [
          { id: 1, pos: { x: 0, y: 0 }, vel: { x: 0, y: 0 }, rot: 0, energy: 100 }
        ],
        planets: [],
        pallets: [],
        worldBounds: { width: 2000.0, height: 2000.0 },
        myShipId: 1,
        done: false,
        win: false
      }
      expect(isMessage(msg)).toBe(true)
    })

    // Room management message type guards
    // Labels: scope:contract loop:g4-proto layer:contract b:type-guards
    it('should return true for CreateRoomMessage', () => {
      const msg = { t: 'createRoom' }
      expect(isMessage(msg)).toBe(true)
    })

    it('should return true for JoinRoomMessage', () => {
      const msg = { t: 'joinRoom', roomCode: 'ABC123' }
      expect(isMessage(msg)).toBe(true)
    })

    it('should return true for LeaveRoomMessage', () => {
      const msg = { t: 'leaveRoom' }
      expect(isMessage(msg)).toBe(true)
    })

    it('should return true for StartMatchMessage', () => {
      const msg = { t: 'startMatch' }
      expect(isMessage(msg)).toBe(true)
    })

    it('should return true for RoomCreatedMessage', () => {
      const msg = { t: 'roomCreated', roomCode: 'XYZ789' }
      expect(isMessage(msg)).toBe(true)
    })

    it('should return true for RoomStateMessage', () => {
      const msg = {
        t: 'roomState',
        roomCode: 'ABC123',
        players: [{ id: 1, name: 'Player1' }],
        state: 'lobby',
        hostId: 1
      }
      expect(isMessage(msg)).toBe(true)
    })

    it('should return true for PlayerJoinedMessage', () => {
      const msg = { t: 'playerJoined', player: { id: 2, name: 'Player2' } }
      expect(isMessage(msg)).toBe(true)
    })

    it('should return true for PlayerLeftMessage', () => {
      const msg = { t: 'playerLeft', playerId: 2 }
      expect(isMessage(msg)).toBe(true)
    })

    it('should return true for MatchStartedMessage', () => {
      const msg = { t: 'matchStarted' }
      expect(isMessage(msg)).toBe(true)
    })

    it('should return true for MatchEndedMessage', () => {
      const msg = { t: 'matchEnded', winnerId: 1 }
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

  // Room management message type guards
  // Labels: scope:contract loop:g4-proto layer:contract b:type-guards
  describe('Room Management Type Guards', () => {
    describe('isPlayerInfo', () => {
      it('should return true for valid PlayerInfo', () => {
        const player = { id: 1, name: 'Player1' }
        expect(isPlayerInfo(player)).toBe(true)
      })

      it('should return false for missing id', () => {
        const player = { name: 'Player1' }
        expect(isPlayerInfo(player)).toBe(false)
      })

      it('should return false for missing name', () => {
        const player = { id: 1 }
        expect(isPlayerInfo(player)).toBe(false)
      })

      it('should return false for wrong types', () => {
        expect(isPlayerInfo({ id: '1', name: 'Player1' })).toBe(false)
        expect(isPlayerInfo({ id: 1, name: 123 })).toBe(false)
      })

      it('should return false for null', () => {
        expect(isPlayerInfo(null)).toBe(false)
      })
    })

    describe('isCreateRoomMessage', () => {
      it('should return true for valid CreateRoomMessage', () => {
        const msg = { t: 'createRoom' }
        expect(isCreateRoomMessage(msg)).toBe(true)
      })

      it('should return false for wrong type', () => {
        const msg = { t: 'joinRoom' }
        expect(isCreateRoomMessage(msg)).toBe(false)
      })

      it('should return false for null', () => {
        expect(isCreateRoomMessage(null)).toBe(false)
      })
    })

    describe('isJoinRoomMessage', () => {
      it('should return true for valid JoinRoomMessage', () => {
        const msg = { t: 'joinRoom', roomCode: 'ABC123' }
        expect(isJoinRoomMessage(msg)).toBe(true)
      })

      it('should return false for missing roomCode', () => {
        const msg = { t: 'joinRoom' }
        expect(isJoinRoomMessage(msg)).toBe(false)
      })

      it('should return false for wrong type', () => {
        const msg = { t: 'createRoom', roomCode: 'ABC123' }
        expect(isJoinRoomMessage(msg)).toBe(false)
      })

      it('should return false for null', () => {
        expect(isJoinRoomMessage(null)).toBe(false)
      })
    })

    describe('isLeaveRoomMessage', () => {
      it('should return true for valid LeaveRoomMessage', () => {
        const msg = { t: 'leaveRoom' }
        expect(isLeaveRoomMessage(msg)).toBe(true)
      })

      it('should return false for wrong type', () => {
        const msg = { t: 'createRoom' }
        expect(isLeaveRoomMessage(msg)).toBe(false)
      })

      it('should return false for null', () => {
        expect(isLeaveRoomMessage(null)).toBe(false)
      })
    })

    describe('isStartMatchMessage', () => {
      it('should return true for valid StartMatchMessage', () => {
        const msg = { t: 'startMatch' }
        expect(isStartMatchMessage(msg)).toBe(true)
      })

      it('should return false for wrong type', () => {
        const msg = { t: 'createRoom' }
        expect(isStartMatchMessage(msg)).toBe(false)
      })

      it('should return false for null', () => {
        expect(isStartMatchMessage(null)).toBe(false)
      })
    })

    describe('isRoomCreatedMessage', () => {
      it('should return true for valid RoomCreatedMessage', () => {
        const msg = { t: 'roomCreated', roomCode: 'XYZ789' }
        expect(isRoomCreatedMessage(msg)).toBe(true)
      })

      it('should return false for missing roomCode', () => {
        const msg = { t: 'roomCreated' }
        expect(isRoomCreatedMessage(msg)).toBe(false)
      })

      it('should return false for wrong type', () => {
        const msg = { t: 'createRoom', roomCode: 'XYZ789' }
        expect(isRoomCreatedMessage(msg)).toBe(false)
      })

      it('should return false for null', () => {
        expect(isRoomCreatedMessage(null)).toBe(false)
      })
    })

    describe('isRoomStateMessage', () => {
      it('should return true for valid RoomStateMessage', () => {
        const msg = {
          t: 'roomState',
          roomCode: 'ABC123',
          players: [{ id: 1, name: 'Player1' }],
          state: 'lobby',
          hostId: 1
        }
        expect(isRoomStateMessage(msg)).toBe(true)
      })

      it('should return false for missing roomCode', () => {
        const msg = {
          t: 'roomState',
          players: [{ id: 1, name: 'Player1' }],
          state: 'lobby',
          hostId: 1
        }
        expect(isRoomStateMessage(msg)).toBe(false)
      })

      it('should return false for missing players', () => {
        const msg = {
          t: 'roomState',
          roomCode: 'ABC123',
          state: 'lobby',
          hostId: 1
        }
        expect(isRoomStateMessage(msg)).toBe(false)
      })

      it('should return false for invalid players array', () => {
        const msg = {
          t: 'roomState',
          roomCode: 'ABC123',
          players: 'not-an-array',
          state: 'lobby',
          hostId: 1
        }
        expect(isRoomStateMessage(msg)).toBe(false)
      })

      it('should return false for missing state', () => {
        const msg = {
          t: 'roomState',
          roomCode: 'ABC123',
          players: [{ id: 1, name: 'Player1' }],
          hostId: 1
        }
        expect(isRoomStateMessage(msg)).toBe(false)
      })

      it('should return false for missing hostId', () => {
        const msg = {
          t: 'roomState',
          roomCode: 'ABC123',
          players: [{ id: 1, name: 'Player1' }],
          state: 'lobby'
        }
        expect(isRoomStateMessage(msg)).toBe(false)
      })

      it('should return false for null', () => {
        expect(isRoomStateMessage(null)).toBe(false)
      })
    })

    describe('isPlayerJoinedMessage', () => {
      it('should return true for valid PlayerJoinedMessage', () => {
        const msg = { t: 'playerJoined', player: { id: 2, name: 'Player2' } }
        expect(isPlayerJoinedMessage(msg)).toBe(true)
      })

      it('should return false for missing player', () => {
        const msg = { t: 'playerJoined' }
        expect(isPlayerJoinedMessage(msg)).toBe(false)
      })

      it('should return false for invalid player', () => {
        const msg = { t: 'playerJoined', player: { id: 2 } }
        expect(isPlayerJoinedMessage(msg)).toBe(false)
      })

      it('should return false for null', () => {
        expect(isPlayerJoinedMessage(null)).toBe(false)
      })
    })

    describe('isPlayerLeftMessage', () => {
      it('should return true for valid PlayerLeftMessage', () => {
        const msg = { t: 'playerLeft', playerId: 2 }
        expect(isPlayerLeftMessage(msg)).toBe(true)
      })

      it('should return false for missing playerId', () => {
        const msg = { t: 'playerLeft' }
        expect(isPlayerLeftMessage(msg)).toBe(false)
      })

      it('should return false for wrong type', () => {
        const msg = { t: 'playerJoined', playerId: 2 }
        expect(isPlayerLeftMessage(msg)).toBe(false)
      })

      it('should return false for null', () => {
        expect(isPlayerLeftMessage(null)).toBe(false)
      })
    })

    describe('isMatchStartedMessage', () => {
      it('should return true for valid MatchStartedMessage', () => {
        const msg = { t: 'matchStarted' }
        expect(isMatchStartedMessage(msg)).toBe(true)
      })

      it('should return false for wrong type', () => {
        const msg = { t: 'matchEnded' }
        expect(isMatchStartedMessage(msg)).toBe(false)
      })

      it('should return false for null', () => {
        expect(isMatchStartedMessage(null)).toBe(false)
      })
    })

    describe('isMatchEndedMessage', () => {
      it('should return true for valid MatchEndedMessage with winnerId', () => {
        const msg = { t: 'matchEnded', winnerId: 1 }
        expect(isMatchEndedMessage(msg)).toBe(true)
      })

      it('should return true for valid MatchEndedMessage without winnerId', () => {
        const msg = { t: 'matchEnded' }
        expect(isMatchEndedMessage(msg)).toBe(true)
      })

      it('should return false for wrong type', () => {
        const msg = { t: 'matchStarted', winnerId: 1 }
        expect(isMatchEndedMessage(msg)).toBe(false)
      })

      it('should return false for null', () => {
        expect(isMatchEndedMessage(null)).toBe(false)
      })
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
    it('should return ShipSnapshot for valid data (v1 format with id)', () => {
      const data = {
        id: 1,
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
    it('should return PlanetSnapshot for valid data (v1 format with id)', () => {
      const data = { id: 1, pos: { x: 0, y: 0 }, radius: 5.0 }
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

  // Room management factory functions
  // Labels: scope:contract loop:g4-proto layer:contract b:factory-functions
  describe('Room Management Factory Functions', () => {
    describe('createCreateRoomMessage', () => {
      it('should create valid CreateRoomMessage', () => {
        const msg = createCreateRoomMessage()
        expect(msg).toEqual({ t: 'createRoom' })
      })
    })

    describe('createJoinRoomMessage', () => {
      it('should create valid JoinRoomMessage with roomCode', () => {
        const msg = createJoinRoomMessage('ABC123')
        expect(msg).toEqual({ t: 'joinRoom', roomCode: 'ABC123' })
      })
    })

    describe('createLeaveRoomMessage', () => {
      it('should create valid LeaveRoomMessage', () => {
        const msg = createLeaveRoomMessage()
        expect(msg).toEqual({ t: 'leaveRoom' })
      })
    })

    describe('createStartMatchMessage', () => {
      it('should create valid StartMatchMessage', () => {
        const msg = createStartMatchMessage()
        expect(msg).toEqual({ t: 'startMatch' })
      })
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

