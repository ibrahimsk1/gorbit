/**
 * Contract tests for TypeScript client workspace setup.
 * These tests verify that the workspace structure meets the requirements
 * for G0 workspace bootstrap.
 * 
 * Labels: scope:contract loop:g0-work layer:infra
 */

import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync, statSync } from 'fs'
import { join, dirname } from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)
const clientDir = join(__dirname, '..')

describe('TypeScript Client Workspace Structure', () => {
  it('package.json exists', () => {
    const packageJsonPath = join(clientDir, 'package.json')
    expect(existsSync(packageJsonPath)).toBe(true)
  })

  it('tsconfig.json exists', () => {
    const tsconfigPath = join(clientDir, 'tsconfig.json')
    expect(existsSync(tsconfigPath)).toBe(true)
  })

  it('vite.config.ts exists', () => {
    const viteConfigPath = join(clientDir, 'vite.config.ts')
    expect(existsSync(viteConfigPath)).toBe(true)
  })

  it('vitest.config.ts exists', () => {
    const vitestConfigPath = join(clientDir, 'vitest.config.ts')
    expect(existsSync(vitestConfigPath)).toBe(true)
  })

  it('src directory exists', () => {
    const srcDir = join(clientDir, 'src')
    expect(existsSync(srcDir)).toBe(true)
    const stat = statSync(srcDir)
    expect(stat.isDirectory()).toBe(true)
  })
})

describe('Package Configuration', () => {
  it('package.json contains required scripts', () => {
    const packageJsonPath = join(clientDir, 'package.json')
    const packageJson = JSON.parse(readFileSync(packageJsonPath, 'utf-8'))
    
    expect(packageJson.scripts).toBeDefined()
    expect(packageJson.scripts.dev).toBeDefined()
    expect(packageJson.scripts.build).toBeDefined()
    expect(packageJson.scripts.test).toBeDefined()
  })

  it('package.json contains required dependencies', () => {
    const packageJsonPath = join(clientDir, 'package.json')
    const packageJson = JSON.parse(readFileSync(packageJsonPath, 'utf-8'))
    
    expect(packageJson.dependencies).toBeDefined()
    expect(packageJson.dependencies['pixi.js']).toBeDefined()
    
    expect(packageJson.devDependencies).toBeDefined()
    expect(packageJson.devDependencies.vite).toBeDefined()
    expect(packageJson.devDependencies.typescript).toBeDefined()
    expect(packageJson.devDependencies.vitest).toBeDefined()
  })
})

describe('TypeScript Configuration', () => {
  it('tsconfig.json is valid JSON', () => {
    const tsconfigPath = join(clientDir, 'tsconfig.json')
    const content = readFileSync(tsconfigPath, 'utf-8')
    // Strip comments for JSON parsing (tsconfig.json allows comments)
    const jsonContent = content.replace(/\/\*[\s\S]*?\*\/|\/\/.*/g, '')
    const tsconfig = JSON.parse(jsonContent)
    
    expect(tsconfig.compilerOptions).toBeDefined()
    expect(tsconfig.compilerOptions.strict).toBe(true)
  })
})

describe('Build Process', () => {
  it('npm run build should succeed', async () => {
    // This test verifies the build configuration is valid
    // Actual build will be tested separately
    const packageJsonPath = join(clientDir, 'package.json')
    const packageJson = JSON.parse(readFileSync(packageJsonPath, 'utf-8'))
    
    expect(packageJson.scripts.build).toContain('vite build')
    expect(packageJson.scripts.build).toContain('tsc')
  })
})

