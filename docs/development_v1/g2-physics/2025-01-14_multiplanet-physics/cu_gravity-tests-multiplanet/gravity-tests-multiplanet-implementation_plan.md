# Implementation Plan: cu/gravity-tests-multiplanet

**CU ID**: `cu/gravity-tests-multiplanet`  
**Loop**: `g2-physics`  
**Step**: 5  
**Status**: IN_PROGRESS

## Goal

Update `server/internal/sim/physics/gravity_test.go` to ensure all gravity tests are properly labeled for the g2-physics loop and test multiple planets scenarios where appropriate. The `CalculateTotalGravity` tests are already comprehensive, but the `GravityAcceleration` tests need label updates.

## Files Allowlist

- `server/internal/sim/physics/gravity_test.go`

## Proof Labels

- `scope:unit loop:g2-physics layer:physics dep:none b:gravity-field r:high double:fake`

## Current State

- `CalculateTotalGravity` tests (lines 332-516) already have correct labels: `loop:g2-physics layer:physics`
- `GravityAcceleration` tests (lines 17-330) have old labels: `loop:g1-physics layer:sim`
- `CalculateTotalGravity` already has comprehensive tests for:
  - 2, 3, and 5 planets
  - Superposition principle verification
  - Edge cases (empty array, single planet, zero mass, overlapping planets, planets at boundaries)
  - Determinism

## Tasks

### 1. Update Test Labels
- Update `GravityAcceleration` test suite label from `loop:g1-physics layer:sim` to `loop:g2-physics layer:physics`
- Ensure label matches proof requirements: `scope:unit loop:g2-physics layer:physics dep:none b:gravity-field r:high double:fake`

### 2. Verify Test Coverage
- Confirm `CalculateTotalGravity` tests cover all required scenarios (already done in cu/gravity-summation)
- Ensure tests verify superposition principle (sum equals individual calculations)
- Verify edge cases are covered (zero planets, overlapping planets, planets at boundaries)

### 3. Add Context Tests (Optional)
- Consider adding tests that demonstrate `GravityAcceleration` is used correctly by `CalculateTotalGravity`
- Ensure tests show the relationship between single-planet and multi-planet gravity calculations

## Test Strategy

### Test-First Approach
1. Update existing test labels
2. Verify all tests still compile and pass
3. Ensure test coverage meets requirements

### Key Test Scenarios (Already Covered)
1. ✅ Multiple Planets Gravity: 2, 3, and 5 planets tested
2. ✅ Superposition Principle: Verified that sum equals individual calculations
3. ✅ Edge Cases: Empty array, single planet, zero mass, overlapping planets, planets at boundaries
4. ✅ Determinism: Tests verify identical results across multiple runs

## Acceptance Criteria

- [ ] All gravity test labels updated to `loop:g2-physics layer:physics`
- [ ] All tests compile without errors
- [ ] Tests verify superposition principle (already done)
- [ ] Tests cover multiple planets scenarios (2, 3, 5 planets) (already done)
- [ ] Edge cases covered (zero planets, overlapping planets, planets at boundaries) (already done)
- [ ] Tests pass successfully

## Evidence

### Compilation
- Command: `go test -c ./server/internal/sim/physics -o /dev/null`
- Result: ✅ **SUCCESS** - Tests compile without errors

### Test Execution
- Command: `go test ./server/internal/sim/physics -v -ginkgo.label-filter="loop:g2-physics"`
- Result: _Pending full test run (compilation verified)_

### Notes
- ✅ `GravityAcceleration` test labels updated from `loop:g1-physics layer:sim` to `loop:g2-physics layer:physics`
- ✅ `CalculateTotalGravity` tests already comprehensive and correctly labeled
- ✅ All required test scenarios covered:
  - Multiple planets (2, 3, 5 planets) ✅
  - Superposition principle verification ✅
  - Edge cases (zero planets, overlapping planets, planets at boundaries) ✅
  - Determinism tests ✅

