# Implementation Plan: cu/physics-determinism

**CU ID**: `cu/physics-determinism`  
**Loop**: `g2-physics`  
**Step**: 7  
**Status**: IN_PROGRESS

## Goal

Add comprehensive property tests ensuring physics calculations (gravity, integration, collisions, wraparound) are deterministic across multiple runs with identical inputs, and verify no side effects or non-deterministic behavior with multiple ships and planets.

## Files Allowlist

- `server/internal/sim/physics/*_test.go`

## Proof Labels

- `scope:unit loop:g2-physics layer:physics dep:none b:determinism r:high double:fake`

## Current State

- Individual functions have determinism tests:
  - `GravityAcceleration` - has determinism tests ✅
  - `CalculateTotalGravity` - has determinism tests ✅
  - `SemiImplicitEuler` - has determinism tests (but labeled `loop:g1-physics`) ⚠️
  - `ShipPlanetCollision` - has determinism tests ✅
  - `CheckShipPlanetCollisions` - has determinism tests ✅
  - `ApplyWraparound` - has determinism tests ✅
- Integration tests in `physics_test.go` have determinism tests but test with single ship/planet
- Missing: Comprehensive property test combining all physics functions with multiple ships and planets

## Tasks

### 1. Update Integrator Test Labels
- Update `integrator_test.go` label from `loop:g1-physics layer:sim` to `loop:g2-physics layer:physics`

### 2. Add Comprehensive Property Test
- Add a new test suite in `physics_test.go` (or create new file) that:
  - Tests complete physics simulation with multiple ships (2-4 ships) and multiple planets (3-5 planets)
  - Runs simulation for multiple ticks (50-100 ticks)
  - Verifies identical results across multiple runs with identical inputs
  - Tests all physics functions together: gravity, integration, collisions, wraparound
  - Verifies no side effects (pure functions)
  - Label: `scope:unit loop:g2-physics layer:physics dep:none b:determinism r:high double:fake`

### 3. Test Scenarios
- Multiple ships with multiple planets
- Ships at different positions and velocities
- Ships wrapping around world boundaries
- Ships colliding with planets
- Long simulation runs (100+ ticks)

## Test Strategy

### Test-First Approach
1. Update integrator test labels
2. Add comprehensive property test
3. Verify all tests compile and pass

### Key Test Scenarios
1. **Complete Physics Simulation**: Multiple ships, multiple planets, all functions combined
2. **Multiple Runs**: Same inputs produce identical results across 10+ runs
3. **No Side Effects**: Functions don't modify input parameters
4. **Long Simulation**: Determinism maintained over 100+ ticks
5. **Edge Cases**: Wraparound, collisions, boundaries

## Acceptance Criteria

- [ ] Integrator test labels updated to `loop:g2-physics layer:physics`
- [ ] Comprehensive property test added for multiple ships and planets
- [ ] Test verifies determinism across multiple runs
- [ ] Test verifies no side effects (pure functions)
- [ ] All tests compile without errors
- [ ] Tests pass successfully

## Evidence

### Compilation
- Command: `go test -c ./server/internal/sim/physics -o /dev/null`
- Result: ✅ **SUCCESS** - Tests compile without errors

### Test Execution
- Command: `go test ./server/internal/sim/physics -v -ginkgo.label-filter="b:determinism"`
- Result: _Pending full test run (compilation verified)_

### Notes
- ✅ Integrator test labels updated from `loop:g1-physics layer:sim` to `loop:g2-physics layer:physics`
- ✅ Comprehensive property test added for multiple ships and planets
- ✅ Test verifies determinism across multiple runs with identical inputs
- ✅ Test verifies no side effects (pure functions)
- ✅ Tests cover:
  - Complete physics simulation with 3-4 ships and 4-5 planets
  - Extended simulation runs (100-200 ticks)
  - All physics functions combined: gravity, integration, collisions, wraparound
  - Pure function verification (no input modification)

