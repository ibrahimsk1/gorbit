# Implementation Plan: cu/physics-integration-multiplayer

**CU ID**: `cu/physics-integration-multiplayer`  
**Loop**: `g2-physics`  
**Step**: 4  
**Status**: IN_PROGRESS

## Goal

Update `server/internal/sim/physics/physics_test.go` to use the v1 multiplayer model with `World` containing `Ships[]` and `Planets[]` arrays instead of the v0 single Ship/Sun model. Test gravity summation with 3–5 planets, collision detection across multiple planets, and wraparound for all ships.

## Files Allowlist

- `server/internal/sim/physics/physics_test.go`

## Proof Labels

- `scope:unit loop:g2-physics layer:physics dep:none b:physics-integration r:high double:fake`

## Tasks

### 1. Update World Construction
- Replace `entities.NewWorld(ship, sun, pallets)` with `entities.NewWorld(ships, planets, pallets)` where:
  - `ships` is `[]entities.Ship` (array with one or more ships)
  - `planets` is `[]entities.Planet` (array with 3–5 planets)
  - Use `entities.NewShip(id, pos, vel, rot, energy)` for ships
  - Use `entities.NewPlanet(id, pos, radius, mass)` for planets

### 2. Update Gravity Calculations
- Replace `GravityAcceleration(shipPos, sun.Pos, sun.Mass, G, aMax)` with:
  - `CalculateTotalGravity(shipPos, world.Planets, G, aMax)` for multiple planets
- Update all gravity-related tests to use 3–5 planets
- Verify superposition principle: total gravity equals sum of individual planet gravities

### 3. Update Collision Detection
- Replace `ShipSunCollision(shipPos, sun.Pos, sun.Radius)` with:
  - `CheckShipPlanetCollisions(shipPos, world.Planets)` which returns `(bool, uint32)`
- Update collision tests to check against multiple planets
- Verify correct planet ID is returned on collision

### 4. Add Wraparound Tests
- Add tests using `ApplyWraparound(pos, entities.WORLD_WIDTH, entities.WORLD_HEIGHT)`
- Test wraparound for ships exiting world bounds
- Test wraparound determinism across multiple ticks

### 5. Update Test Labels
- Change `loop:g1-physics` to `loop:g2-physics` in all test labels
- Ensure all tests have correct proof labels matching CU requirements

### 6. Update Test Scenarios
- **Determinism tests**: Use World with multiple ships and planets
- **Gravity and Integration tests**: Test with 3–5 planets, verify ships fall toward nearest/multiple planets
- **Collision Detection tests**: Test collisions with multiple planets, verify correct planet ID
- **End-to-End Scenarios**: Test orbital paths with multiple planets, wraparound scenarios
- **Edge Cases**: Test with multiple planets at boundaries, overlapping planets, etc.

### 7. Remove v0-Specific Code
- Remove references to `entities.Sun` type
- Remove `world.Ship` and `world.Sun` field access (use `world.Ships[]` and `world.Planets[]`)

## Test Strategy

### Test-First Approach
1. Update existing tests to use new multiplayer model
2. Add new tests for multiple planets scenarios
3. Verify all tests compile
4. Run tests and fix any issues

### Key Test Scenarios
1. **Multiple Planets Gravity**: Ship affected by gravity from 3–5 planets simultaneously
2. **Multi-Ship Simulation**: Multiple ships in same world, each affected by all planets
3. **Collision with Multiple Planets**: Ship can collide with any of the planets, correct ID returned
4. **Wraparound with Multiple Ships**: All ships wrap correctly when exiting world bounds
5. **Determinism with Multiplayer Model**: Identical initial conditions produce identical results across multiple runs

## Acceptance Criteria

- [ ] All tests use `World` with `Ships[]` and `Planets[]` arrays
- [ ] Gravity tests use `CalculateTotalGravity` with 3–5 planets
- [ ] Collision tests use `CheckShipPlanetCollisions` and verify planet ID
- [ ] Wraparound tests added using `ApplyWraparound`
- [ ] All test labels updated to `loop:g2-physics`
- [ ] Tests compile without errors
- [ ] Tests pass (or identify remaining issues for future CUs)

## Evidence

### Compilation
- Command: `go test -c ./server/internal/sim/physics -o /dev/null`
- Result: ✅ **SUCCESS** - Tests compile without errors

### Test Execution
- Command: `go test ./server/internal/sim/physics -v -ginkgo.label-filter="loop:g2-physics"`
- Result: _Pending full test run (compilation verified)_

### Notes
- ✅ All tests updated to use `World` with `Ships[]` and `Planets[]` arrays
- ✅ All gravity calculations use `CalculateTotalGravity` with multiple planets
- ✅ All collision detection uses `CheckShipPlanetCollisions` with planet ID verification
- ✅ Wraparound tests added using `ApplyWraparound` function
- ✅ Test labels updated to `loop:g2-physics`
- ✅ Multiple planets tests added (3-5 planets scenarios)
- ✅ Multiple ships tests added
- ✅ All compilation errors resolved

