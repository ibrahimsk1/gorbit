/**
 * Unit tests for DI Container.
 * 
 * Labels: scope:unit loop:g2-app layer:core
 */

import { describe, it, expect, beforeEach } from 'vitest'
import { DIContainer } from './di-container'
import { DI_KEYS } from './di-keys'
import type { RegistrationOptions } from './di-types'

// Test doubles
class TestService {
  constructor(public name: string) {}
  destroy(): void {}
}

class TestDependency {
  constructor(public value: number) {}
}

describe('DIContainer', () => {
  let container: DIContainer

  beforeEach(() => {
    container = new DIContainer()
  })

  describe('Registration', () => {
    it('should register a dependency', () => {
      container.register(
        'test',
        () => new TestService('test'),
        { phase: 'early', scope: 'singleton' }
      )

      expect(container.isRegistered('test')).toBe(true)
    })

    it('should throw error if dependency not found', () => {
      expect(() => {
        container.register(
          'dependent',
          () => new TestService('test'),
          {
            phase: 'early',
            scope: 'singleton',
            dependencies: ['missing']
          }
        )
      }).toThrow("Dependency 'missing' for 'dependent' is not registered")
    })

    it('should allow registering without dependencies', () => {
      container.register(
        'independent',
        () => new TestService('test'),
        { phase: 'early', scope: 'singleton' }
      )

      expect(container.isRegistered('independent')).toBe(true)
    })
  })

  describe('Resolution', () => {
    it('should resolve a registered dependency', () => {
      container.register(
        'test',
        () => new TestService('test'),
        { phase: 'early', scope: 'singleton' }
      )

      const instance = container.resolve<TestService>('test')
      expect(instance).toBeInstanceOf(TestService)
      expect(instance.name).toBe('test')
    })

    it('should throw error if dependency not registered', () => {
      expect(() => {
        container.resolve('missing')
      }).toThrow("Dependency 'missing' is not registered")
    })

    it('should resolve dependencies in correct order', () => {
      container.register(
        'dependency',
        () => new TestDependency(42),
        { phase: 'early', scope: 'singleton' }
      )

      container.register(
        'dependent',
        () => {
          const dep = container.resolve<TestDependency>('dependency')
          return new TestService(`value-${dep.value}`)
        },
        {
          phase: 'early',
          scope: 'singleton',
          dependencies: ['dependency']
        }
      )

      const instance = container.resolve<TestService>('dependent')
      expect(instance.name).toBe('value-42')
    })

    it('should cache singleton instances', () => {
      container.register(
        'singleton',
        () => new TestService('test'),
        { phase: 'early', scope: 'singleton' }
      )

      const instance1 = container.resolve<TestService>('singleton')
      const instance2 = container.resolve<TestService>('singleton')

      expect(instance1).toBe(instance2)
    })

    it('should create new instances for transient scope', () => {
      container.register(
        'transient',
        () => new TestService('test'),
        { phase: 'early', scope: 'transient' }
      )

      const instance1 = container.resolve<TestService>('transient')
      const instance2 = container.resolve<TestService>('transient')

      expect(instance1).not.toBe(instance2)
      expect(instance1.name).toBe(instance2.name)
    })
  })

  describe('Circular Dependencies', () => {
    it('should detect circular dependencies', () => {
      // Register both first without dependencies
      container.register(
        'a',
        () => new TestService('a'),
        { phase: 'early', scope: 'singleton' }
      )

      container.register(
        'b',
        () => new TestService('b'),
        { phase: 'early', scope: 'singleton' }
      )

      // Now update 'a' to depend on 'b'
      container.register(
        'a',
        () => new TestService('a'),
        {
          phase: 'early',
          scope: 'singleton',
          dependencies: ['b']
        }
      )

      // Now update 'b' to depend on 'a' - creates cycle
      // This should be detected when rebuildInitializationOrder is called
      expect(() => {
        container.register(
          'b',
          () => new TestService('b'),
          {
            phase: 'early',
            scope: 'singleton',
            dependencies: ['a']  // Creates cycle: a -> b -> a
          }
        )
      }).toThrow('Circular dependency detected')
    })
  })

  describe('Lifecycle Phases', () => {
    it('should initialize early phase dependencies', async () => {
      let initialized = false

      container.register(
        'early',
        () => {
          initialized = true
          return new TestService('early')
        },
        { phase: 'early', scope: 'singleton' }
      )

      await container.initializePhase('early')

      expect(initialized).toBe(true)
      expect(container.isRegistered('early')).toBe(true)
    })

    it('should initialize late phase dependencies', async () => {
      let initialized = false

      container.register(
        'late',
        () => {
          initialized = true
          return new TestService('late')
        },
        { phase: 'late', scope: 'singleton' }
      )

      await container.initializePhase('late')

      expect(initialized).toBe(true)
    })

    it('should not initialize late phase when initializing early phase', async () => {
      let earlyInitialized = false
      let lateInitialized = false

      container.register(
        'early',
        () => {
          earlyInitialized = true
          return new TestService('early')
        },
        { phase: 'early', scope: 'singleton' }
      )

      container.register(
        'late',
        () => {
          lateInitialized = true
          return new TestService('late')
        },
        { phase: 'late', scope: 'singleton' }
      )

      await container.initializePhase('early')

      expect(earlyInitialized).toBe(true)
      expect(lateInitialized).toBe(false)
    })

    it('should resolve dependencies in correct order across phases', async () => {
      const initOrder: string[] = []

      container.register(
        'early1',
        () => {
          initOrder.push('early1')
          return new TestService('early1')
        },
        { phase: 'early', scope: 'singleton' }
      )

      container.register(
        'early2',
        () => {
          initOrder.push('early2')
          return new TestService('early2')
        },
        {
          phase: 'early',
          scope: 'singleton',
          dependencies: ['early1']
        }
      )

      container.register(
        'late1',
        () => {
          initOrder.push('late1')
          return new TestService('late1')
        },
        {
          phase: 'late',
          scope: 'singleton',
          dependencies: ['early2']
        }
      )

      await container.initializePhase('early')
      await container.initializePhase('late')

      expect(initOrder).toEqual(['early1', 'early2', 'late1'])
    })
  })

  describe('Ownership and Cleanup', () => {
    it('should destroy container-owned instances', () => {
      let destroyed = false

      class DestroyableService {
        destroy(): void {
          destroyed = true
        }
      }

      container.register(
        'container-owned',
        () => new DestroyableService(),
        {
          phase: 'early',
          scope: 'singleton',
          owner: 'container'
        }
      )

      container.resolve('container-owned')
      container.destroy()

      expect(destroyed).toBe(true)
    })

    it('should not destroy orchestrator-owned instances', () => {
      let destroyed = false

      class DestroyableService {
        destroy(): void {
          destroyed = true
        }
      }

      container.register(
        'orchestrator-owned',
        () => new DestroyableService(),
        {
          phase: 'early',
          scope: 'singleton',
          owner: 'orchestrator'
        }
      )

      container.resolve('orchestrator-owned')
      container.destroy()

      expect(destroyed).toBe(false)
    })

    it('should handle missing destroy method gracefully', () => {
      container.register(
        'no-destroy',
        () => new TestService('test'),
        {
          phase: 'early',
          scope: 'singleton',
          owner: 'container'
        }
      )

      container.resolve('no-destroy')

      expect(() => {
        container.destroy()
      }).not.toThrow()
    })
  })

  describe('Clear', () => {
    it('should clear all registrations and instances', () => {
      container.register(
        'test',
        () => new TestService('test'),
        { phase: 'early', scope: 'singleton' }
      )

      container.resolve('test')
      container.clear()

      expect(container.isRegistered('test')).toBe(false)
      expect(() => {
        container.resolve('test')
      }).toThrow()
    })
  })

  describe('Initialization Order', () => {
    it('should return correct initialization order', () => {
      container.register(
        'a',
        () => new TestService('a'),
        { phase: 'early', scope: 'singleton' }
      )

      container.register(
        'b',
        () => new TestService('b'),
        {
          phase: 'early',
          scope: 'singleton',
          dependencies: ['a']
        }
      )

      container.register(
        'c',
        () => new TestService('c'),
        {
          phase: 'early',
          scope: 'singleton',
          dependencies: ['b']
        }
      )

      const order = container.getInitializationOrder()
      expect(order).toEqual(['a', 'b', 'c'])
    })
  })
})

