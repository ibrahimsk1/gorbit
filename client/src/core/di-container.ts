/**
 * Dependency Injection Container for managing application dependencies.
 * 
 * Provides:
 * - Factory pattern registration (all dependencies use factories)
 * - Automatic dependency resolution in correct order
 * - Lifecycle phase management (early/late initialization)
 * - Ownership tracking (container/orchestrator/application)
 * - Singleton and transient instance scoping
 * 
 * Labels: scope:unit loop:g2-app layer:core
 */

import type {
  LifecyclePhase,
  RegistrationOptions,
  Registration,
} from './di-types'

export class DIContainer {
  private registrations = new Map<string, Registration>()
  private instances = new Map<string, unknown>()
  private resolvedKeys = new Set<string>()
  private initializationOrder: string[] = []

  /**
   * Registers a factory function for a dependency.
   * 
   * @param key Unique identifier for the dependency
   * @param factory Function that creates the instance
   * @param options Registration options (lifecycle, scope, dependencies, owner)
   * @throws Error if dependencies are not registered or circular dependency detected
   */
  register<T>(
    key: string,
    factory: () => T,
    options: RegistrationOptions
  ): void {
    // Store registration
    this.registrations.set(key, { factory, options })
    
    // Rebuild initialization order (topological sort of dependency graph)
    // This will detect circular dependencies and missing dependencies
    this.rebuildInitializationOrder()
  }

  /**
   * Resolves a dependency, creating it if necessary.
   * Dependencies are resolved in the correct order automatically.
   * 
   * @param key Dependency key
   * @returns Resolved instance
   * @throws Error if dependency is not registered
   */
  resolve<T>(key: string): T {
    // Check if already resolved (singleton)
    if (this.instances.has(key)) {
      return this.instances.get(key) as T
    }

    // Check if registered
    const registration = this.registrations.get(key)
    if (!registration) {
      throw new Error(`Dependency '${key}' is not registered`)
    }

    // Resolve dependencies first (recursive)
    if (registration.options.dependencies) {
      for (const depKey of registration.options.dependencies) {
        this.resolve(depKey)  // Recursive resolution ensures correct order
      }
    }

    // Create instance
    const instance = registration.factory() as T

    // Cache if singleton
    if (registration.options.scope === 'singleton') {
      this.instances.set(key, instance)
      registration.instance = instance
    }

    this.resolvedKeys.add(key)
    return instance
  }

  /**
   * Initializes all dependencies in a specific lifecycle phase.
   * Dependencies are resolved in the correct order based on their
   * dependency graph. Calls initialization hooks after resolution.
   * 
   * @param phase Lifecycle phase to initialize
   */
  async initializePhase(phase: LifecyclePhase): Promise<void> {
    // Get all keys for this phase in initialization order
    const phaseKeys = this.initializationOrder.filter(key => {
      const reg = this.registrations.get(key)
      return reg?.options.phase === phase
    })

    // Resolve all dependencies in order and call initialization hooks
    for (const key of phaseKeys) {
      const instance = this.resolve(key)
      const reg = this.registrations.get(key)
      
      // Call custom initialization hook if provided
      if (reg?.options.initializeAfter) {
        await reg.options.initializeAfter(instance)
      }
      
      // Call instance.initialize() if it exists and is a function
      if (instance && typeof (instance as { initialize?: () => void | Promise<void> }).initialize === 'function') {
        const result = (instance as { initialize: () => void | Promise<void> }).initialize()
        if (result instanceof Promise) {
          await result
        }
      }
    }
  }

  /**
   * Rebuilds the initialization order using topological sort.
   * Ensures dependencies are resolved before dependents.
   * Detects circular dependencies and missing dependencies.
   * 
   * @throws Error if circular dependency or missing dependency is detected
   */
  private rebuildInitializationOrder(): void {
    const visited = new Set<string>()
    const tempMark = new Set<string>()
    const order: string[] = []

    const visit = (key: string): void => {
      // Check for circular dependency
      if (tempMark.has(key)) {
        throw new Error(`Circular dependency detected involving '${key}'`)
      }
      if (visited.has(key)) {
        return
      }

      // Check if dependency is registered
      const reg = this.registrations.get(key)
      if (!reg) {
        // This shouldn't happen, but handle gracefully
        return
      }

      tempMark.add(key)
      
      // Visit dependencies first
      if (reg.options.dependencies) {
        for (const dep of reg.options.dependencies) {
          // Check if dependency is registered
          if (!this.registrations.has(dep)) {
            throw new Error(
              `Dependency '${dep}' for '${key}' is not registered. ` +
              `Register dependencies before dependents.`
            )
          }
          visit(dep)
        }
      }
      
      tempMark.delete(key)
      visited.add(key)
      order.push(key)
    }

    // Visit all registered keys
    for (const key of this.registrations.keys()) {
      if (!visited.has(key)) {
        visit(key)
      }
    }

    this.initializationOrder = order
  }

  /**
   * Destroys all container-owned instances.
   * Calls destroy() method if it exists on the instance.
   */
  destroy(): void {
    for (const [key, instance] of this.instances.entries()) {
      const reg = this.registrations.get(key)
      if (reg?.options.owner === 'container') {
        const destroyable = instance as { destroy?: () => void }
        if (typeof destroyable.destroy === 'function') {
          try {
            destroyable.destroy()
          } catch (error) {
            console.warn(`Error destroying '${key}':`, error)
          }
        }
      }
    }
    this.instances.clear()
    this.resolvedKeys.clear()
  }

  /**
   * Clears all registrations and instances (for testing).
   */
  clear(): void {
    this.registrations.clear()
    this.instances.clear()
    this.resolvedKeys.clear()
    this.initializationOrder = []
  }

  /**
   * Checks if a dependency is registered.
   * 
   * @param key Dependency key
   * @returns True if registered
   */
  isRegistered(key: string): boolean {
    return this.registrations.has(key)
  }

  /**
   * Gets the initialization order for debugging.
   * 
   * @returns Array of keys in initialization order
   */
  getInitializationOrder(): string[] {
    return [...this.initializationOrder]
  }
}

