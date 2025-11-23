# Implementation Plan: cu/collision-tests-multiplanet

**CU ID**: `cu/collision-tests-multiplanet`  
**Loop**: `g2-physics`  
**Step**: 6  
**Status**: IN_PROGRESS

## Goal

Update `server/internal/sim/physics/collision_test.go` to ensure all collision tests are properly labeled for the g2-physics loop. The `CheckShipPlanetCollisions` tests are already comprehensive with multiple planets scenarios, but the `ShipPlanetCollision` and `ShipPalletCollision` tests need label updates.

## Files Allowlist

- `server/internal/sim/physics/collision_test.go`

## Proof Labels

- `scope:unit loop:g2-physics layer:physics dep:none b:collision-detection r:medium double:fake`

## Current State

- `CheckShipPlanetCollisions` tests (lines 172-338) already have correct labels: `loop:g2-physics layer:physics`
- `ShipPlanetCollision` tests (lines 17-170) have old labels: `loop:g1-physics layer:sim`
- `ShipPalletCollision` tests (lines 340-498) are inside main "Collision" Describe block with old labels
- `CheckShipPlanetCollisions` already has comprehensive tests for:
  - Multiple planets scenarios
  - Correct planet ID verification
  - Ship between multiple planets
  - Edge cases

## Tasks

### 1. Update Test Labels
- Update main "Collision" Describe block label from `loop:g1-physics layer:sim` to `loop:g2-physics layer:physics`
- Ensure label matches proof requirements: `scope:unit loop:g2-physics layer:physics dep:none b:collision-detection r:medium double:fake`

### 2. Verify Test Coverage
- Confirm `CheckShipPlanetCollisions` tests cover all required scenarios (already done in cu/collision-multiplanet)
- Verify tests check correct planet ID on collision
- Verify scenarios with ship between multiple planets are covered

### 3. Add Additional Tests (if needed)
- Review existing tests to ensure comprehensive coverage
- Add any missing edge cases for multiple planets scenarios

## Test Strategy

### Test-First Approach
1. Update existing test labels
2. Verify all tests still compile and pass
3. Ensure test coverage meets requirements

### Key Test Scenarios (Already Covered)
1. ✅ Multiple Planets: Tests with 2-3 planets
2. ✅ Correct Planet ID: Verified that correct planet ID is returned
3. ✅ Ship Between Planets: Tests for ship positioned between multiple planets
4. ✅ Edge Cases: Empty array, ship at planet center, ship at radius, etc.
5. ✅ Determinism: Tests verify identical results across multiple runs

## Acceptance Criteria

- [ ] All collision test labels updated to `loop:g2-physics layer:physics`
- [ ] All tests compile without errors
- [ ] Tests verify CheckShipPlanetCollisions returns correct planet ID (already done)
- [ ] Tests cover scenarios with ship between multiple planets (already done)
- [ ] Tests pass successfully

## Evidence

### Compilation
- Command: `go test -c ./server/internal/sim/physics -o /dev/null`
- Result: ✅ **SUCCESS** - Tests compile without errors

### Test Execution
- Command: `go test ./server/internal/sim/physics -v -ginkgo.label-filter="loop:g2-physics"`
- Result: _Pending full test run (compilation verified)_

### Notes
- ✅ Main "Collision" test labels updated from `loop:g1-physics layer:sim` to `loop:g2-physics layer:physics`
- ✅ `CheckShipPlanetCollisions` tests already comprehensive and correctly labeled
- ✅ All required test scenarios covered:
  - Multiple planets scenarios ✅
  - Correct planet ID verification ✅
  - Ship between multiple planets ✅
  - Edge cases ✅
  - Determinism tests ✅

