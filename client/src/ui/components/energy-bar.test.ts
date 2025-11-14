/**
 * Integration tests for Energy Bar component.
 * 
 * Labels: scope:integration loop:g6-client layer:client dep:pixi
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { Container } from 'pixi.js'
import { EnergyBar } from './energy-bar'

describe('Energy Bar', () => {
  let container: Container

  beforeEach(() => {
    container = new Container()
  })

  afterEach(() => {
    container.destroy({ children: true })
  })

  describe('Creation', () => {
    it('creates energy bar with default configuration', () => {
      const energyBar = new EnergyBar(container)

      expect(energyBar).toBeDefined()
      expect(container.children.length).toBeGreaterThan(0)
    })

    it('creates energy bar with custom configuration', () => {
      const energyBar = new EnergyBar(container, {
        x: 50,
        y: 50,
        width: 200,
        height: 20
      })

      expect(energyBar).toBeDefined()
      expect(container.children.length).toBeGreaterThan(0)
    })
  })

  describe('Energy Updates', () => {
    it('updates bar width based on energy ratio at 0%', () => {
      const energyBar = new EnergyBar(container)
      const maxEnergy = 100.0

      energyBar.update(0, maxEnergy)

      // Bar should be at minimum width (0% of max)
      const energyBarContainer = container.children.find(child => 
        child.label === 'energy-bar'
      )
      expect(energyBarContainer).toBeDefined()
      const foregroundBar = energyBarContainer?.children.find(child => 
        child.label === 'energy-foreground'
      )
      expect(foregroundBar).toBeDefined()
      if (foregroundBar) {
        expect(foregroundBar.width).toBe(0)
      }
    })

    it('updates bar width based on energy ratio at 50%', () => {
      const energyBar = new EnergyBar(container, { width: 200 })
      const maxEnergy = 100.0

      energyBar.update(50, maxEnergy)

      // Bar should be at 50% width
      const energyBarContainer = container.children.find(child => 
        child.label === 'energy-bar'
      )
      const foregroundBar = energyBarContainer?.children.find(child => 
        child.label === 'energy-foreground'
      )
      expect(foregroundBar).toBeDefined()
      if (foregroundBar) {
        expect(foregroundBar.width).toBeCloseTo(100, 1) // 50% of 200
      }
    })

    it('updates bar width based on energy ratio at 100%', () => {
      const energyBar = new EnergyBar(container, { width: 200 })
      const maxEnergy = 100.0

      energyBar.update(100, maxEnergy)

      // Bar should be at full width
      const energyBarContainer = container.children.find(child => 
        child.label === 'energy-bar'
      )
      const foregroundBar = energyBarContainer?.children.find(child => 
        child.label === 'energy-foreground'
      )
      expect(foregroundBar).toBeDefined()
      if (foregroundBar) {
        expect(foregroundBar.width).toBeCloseTo(200, 1) // 100% of 200
      }
    })

    it('clamps energy to valid range [0, maxEnergy]', () => {
      const energyBar = new EnergyBar(container, { width: 200 })
      const maxEnergy = 100.0

      // Test energy > maxEnergy
      energyBar.update(150, maxEnergy)
      const energyBarContainer1 = container.children.find(child => 
        child.label === 'energy-bar'
      )
      const foregroundBar1 = energyBarContainer1?.children.find(child => 
        child.label === 'energy-foreground'
      )
      expect(foregroundBar1).toBeDefined()
      if (foregroundBar1) {
        expect(foregroundBar1.width).toBeLessThanOrEqual(200)
      }

      // Test energy < 0
      energyBar.update(-10, maxEnergy)
      const energyBarContainer2 = container.children.find(child => 
        child.label === 'energy-bar'
      )
      const foregroundBar2 = energyBarContainer2?.children.find(child => 
        child.label === 'energy-foreground'
      )
      expect(foregroundBar2).toBeDefined()
      if (foregroundBar2) {
        expect(foregroundBar2.width).toBeGreaterThanOrEqual(0)
      }
    })

    it('handles edge case: energy = 0', () => {
      const energyBar = new EnergyBar(container)
      const maxEnergy = 100.0

      energyBar.update(0, maxEnergy)

      const energyBarContainer = container.children.find(child => 
        child.label === 'energy-bar'
      )
      const foregroundBar = energyBarContainer?.children.find(child => 
        child.label === 'energy-foreground'
      )
      expect(foregroundBar).toBeDefined()
      if (foregroundBar) {
        expect(foregroundBar.width).toBe(0)
      }
    })

    it('handles edge case: energy = maxEnergy', () => {
      const energyBar = new EnergyBar(container, { width: 200 })
      const maxEnergy = 100.0

      energyBar.update(maxEnergy, maxEnergy)

      const energyBarContainer = container.children.find(child => 
        child.label === 'energy-bar'
      )
      const foregroundBar = energyBarContainer?.children.find(child => 
        child.label === 'energy-foreground'
      )
      expect(foregroundBar).toBeDefined()
      if (foregroundBar) {
        expect(foregroundBar.width).toBeCloseTo(200, 1)
      }
    })
  })

  describe('Destruction', () => {
    it('destroys bar and cleans up resources', () => {
      const energyBar = new EnergyBar(container)
      const initialChildCount = container.children.length

      energyBar.destroy()

      expect(container.children.length).toBeLessThan(initialChildCount)
    })
  })
})

