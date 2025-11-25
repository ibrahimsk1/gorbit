/**
 * Type definitions for Dependency Injection Container.
 * 
 * Labels: scope:unit loop:g2-app layer:core
 */

/**
 * Lifecycle phase for dependency initialization.
 * - 'bootstrap': Core infrastructure (App, NetworkClient, InputHandler) - needed immediately
 * - 'menu': Menu systems (Scene, MainMenu, RoomLobby) - needed for menu UI
 * - 'game': Game systems (Renderer, HUD, StateManager, Simulators) - lazy loaded when match starts
 */
export type LifecyclePhase = 'bootstrap' | 'menu' | 'game'

/**
 * Instance scope for dependency resolution.
 * - 'singleton': Single instance created and reused
 * - 'transient': New instance created on each resolve
 */
export type InstanceScope = 'singleton' | 'transient'

/**
 * Ownership category for dependency lifecycle management.
 * - 'container': Created and destroyed by DI container
 * - 'orchestrator': Created by container, managed by orchestrator
 * - 'application': Created by container, managed by application
 */
export type Ownership = 'container' | 'orchestrator' | 'application'

/**
 * Options for registering a dependency in the DI container.
 */
export interface RegistrationOptions {
  /** Lifecycle phase when this dependency should be initialized */
  phase: LifecyclePhase
  
  /** Instance scope (singleton or transient) */
  scope: InstanceScope
  
  /** Keys of dependencies that must be resolved before this one */
  dependencies?: string[]
  
  /** Who owns the lifecycle of this dependency */
  owner?: Ownership
  
  /** 
   * Post-construction initialization hook.
   * Called after instance is created, before phase initialization completes.
   * Useful for initialization that depends on other phase dependencies.
   * @param instance The created instance
   */
  initializeAfter?: (instance: unknown) => void | Promise<void>
}

/**
 * Internal registration metadata stored in the container.
 */
export interface Registration<T = unknown> {
  /** Factory function that creates the instance */
  factory: () => T
  
  /** Registration options */
  options: RegistrationOptions
  
  /** Cached instance (for singletons) */
  instance?: T
}

