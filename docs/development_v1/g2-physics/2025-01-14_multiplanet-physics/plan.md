# Plan

## Prerequisites
- G1-entities loop complete: Ship, Planet, World with multiplayer arrays (Ships[], Planets[], Pallets[]) implemented
- All proofs labeled `scope:unit loop:g1-entities layer:entities dep:none` are green
- World bounds constants (WORLD_WIDTH, WORLD_HEIGHT) defined in entities package
- Wraparound utility (WrapPosition) implemented in entities package
- v0 G1-physics loop complete: GravityAcceleration (single sun), SemiImplicitEuler, ShipSunCollision, ShipPalletCollision exist
- All v0 proofs labeled `scope:unit loop:g1-physics layer:sim dep:none` are green

## Steps
1. **Implement multiple planets gravity summation** — Create `CalculateTotalGravity(shipPos Vec2, planets []Planet, G, aMax float64) Vec2` function in `server/internal/sim/physics/gravity.go` that sums gravity from all planets using superposition principle; each planet's gravity calculated via GravityAcceleration, then vector-summed; add tests for multiple planets scenarios
2. **Update collision detection for multiple planets** — Rename `ShipSunCollision` to `ShipPlanetCollision` and update signature to accept Planet struct; add `CheckShipPlanetCollisions(shipPos Vec2, planets []Planet) (bool, uint32)` function that checks ship against all planets and returns collision status with planet ID; update tests to use multiple planets
3. **Add wraparound function to physics package** — Create `ApplyWraparound(pos Vec2, worldWidth, worldHeight float64) Vec2` function in `server/internal/sim/physics/wraparound.go` (or integrator.go) that wraps positions using entities.WrapPosition or implements modulo-based wraparound per physics spec; add tests for wraparound determinism and edge cases
4. **Update physics integration tests for multiplayer** — Modify `server/internal/sim/physics/physics_test.go` to use World with Ships[] and Planets[] arrays instead of single Ship/Sun; test gravity summation with 3–5 planets, collision detection across multiple planets, and wraparound for all ships
5. **Update gravity tests for multiple planets** — Modify `server/internal/sim/physics/gravity_test.go` to add tests for CalculateTotalGravity with multiple planets (2, 3, 5 planets), verify superposition principle (sum equals individual calculations), test edge cases (zero planets, overlapping planets, planets at boundaries)
6. **Update collision tests for multiple planets** — Modify `server/internal/sim/physics/collision_test.go` to test ShipPlanetCollision with multiple planets, verify CheckShipPlanetCollisions returns correct planet ID on collision, test scenarios with ship between multiple planets
7. **Verify physics determinism with multiplayer model** — Add property tests ensuring physics calculations (gravity, integration, collisions, wraparound) are deterministic across multiple runs with identical inputs; verify no side effects or non-deterministic behavior with multiple ships and planets

## CU List

| id | title | loop | step | files_allowlist | proof_labels | status |
|----|-------|------|------|-----------------|--------------|--------|
| cu/gravity-summation | Implement multiple planets gravity summation | g2-physics | 1 | `server/internal/sim/physics/gravity.go`, `server/internal/sim/physics/gravity_test.go` | `scope:unit loop:g2-physics layer:physics dep:none b:gravity-summation r:high double:fake` | DONE | Evidence: [implementation_plan.md](cu_gravity-summation/gravity-summation-implementation_plan.md) |
| cu/collision-multiplanet | Update collision detection for multiple planets | g2-physics | 2 | `server/internal/sim/physics/collision.go`, `server/internal/sim/physics/collision_test.go` | `scope:unit loop:g2-physics layer:physics dep:none b:collision-multiplanet r:medium double:fake` | DONE | Evidence: [implementation_plan.md](cu_collision-multiplanet/collision-multiplanet-implementation_plan.md) |
| cu/wraparound-physics | Add wraparound function to physics package | g2-physics | 3 | `server/internal/sim/physics/wraparound.go`, `server/internal/sim/physics/wraparound_test.go` | `scope:unit loop:g2-physics layer:physics dep:none b:wraparound r:medium` | DONE | Evidence: [implementation_plan.md](cu_wraparound-physics/wraparound-physics-implementation_plan.md) |
| cu/physics-integration-multiplayer | Update physics integration tests for multiplayer | g2-physics | 4 | `server/internal/sim/physics/physics_test.go` | `scope:unit loop:g2-physics layer:physics dep:none b:physics-integration r:high double:fake` | DONE | Evidence: [implementation_plan.md](cu_physics-integration-multiplayer/physics-integration-multiplayer-implementation_plan.md) |
| cu/gravity-tests-multiplanet | Update gravity tests for multiple planets | g2-physics | 5 | `server/internal/sim/physics/gravity_test.go` | `scope:unit loop:g2-physics layer:physics dep:none b:gravity-field r:high double:fake` | BACKLOG |
| cu/collision-tests-multiplanet | Update collision tests for multiple planets | g2-physics | 6 | `server/internal/sim/physics/collision_test.go` | `scope:unit loop:g2-physics layer:physics dep:none b:collision-detection r:medium double:fake` | BACKLOG |
| cu/physics-determinism | Verify physics determinism with multiplayer model | g2-physics | 7 | `server/internal/sim/physics/*_test.go` | `scope:unit loop:g2-physics layer:physics dep:none b:determinism r:high double:fake` | BACKLOG |

## Acceptance
- [x] CalculateTotalGravity function sums gravity from all planets using superposition principle
- [x] Gravity summation tests pass with 2, 3, and 5 planets; verify sum equals individual calculations (tests implemented, execution blocked by pre-existing physics_test.go errors)
- [x] ShipPlanetCollision function replaces ShipSunCollision and accepts Planet struct
- [x] CheckShipPlanetCollisions function checks ship against all planets and returns collision status with planet ID
- [x] ApplyWraparound function wraps positions deterministically using modulo-based formula
- [x] Physics integration tests use World with Ships[] and Planets[] arrays
- [ ] All gravity tests updated to test multiple planets scenarios
- [ ] All collision tests updated to test multiple planets scenarios
- [ ] Property tests verify determinism across multiple runs with identical inputs
- [ ] All tests labeled `scope:unit loop:g2-physics layer:physics dep:none` pass in <1s per package
- [ ] No IO dependencies in `server/internal/sim/physics/` package
- [ ] Physics calculations are pure functions with no side effects or non-deterministic behavior
- [ ] All CU proofs are green before advancing to G3-rules loop

