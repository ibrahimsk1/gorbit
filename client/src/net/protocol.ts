/**
 * Protocol types matching G4 contracts.
 * 
 * Labels: scope:integration loop:g6-client layer:client
 */

export interface Vec2Snapshot {
  x: number
  y: number
}

export interface ShipSnapshot {
  pos: Vec2Snapshot
  vel: Vec2Snapshot
  rot: number
  energy: number
}

export interface SunSnapshot {
  pos: Vec2Snapshot
  radius: number
}

export interface PalletSnapshot {
  id: number
  pos: Vec2Snapshot
  active: boolean
}

export interface InputMessage {
  t: 'input'
  seq: number
  thrust: number
  turn: number
}

export interface RestartMessage {
  t: 'restart'
}

export interface SnapshotMessage {
  t: 'snapshot'
  tick: number
  ship: ShipSnapshot
  sun: SunSnapshot
  pallets: PalletSnapshot[]
  done: boolean
  win: boolean
}

export type Message = InputMessage | RestartMessage | SnapshotMessage

