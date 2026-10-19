# Astro Answers — Generated Engineering Design Document

**Evidence commit:** `fba80594deed4d8083152f15bea5ce8b5be94ff3`  
**Method:** code-derived review using the Jonno's Garage `codebase-to-EDD` evidence discipline  
**Generated:** 2026-10-19

This document describes the implementation at the evidence commit. The hand-written `PRD.md` and `EDD.md` are comparison targets, not runtime evidence.

## 1. Executive description

**KNOWN** — Astro Answers is a small Linux desktop arcade game written in Go. The deterministic gameplay rules live in `internal/game`; `cmd/astroanswers` is a thin Ebitengine adapter for keyboard input, rendering, embedded assets, audio and local score persistence.

**KNOWN** — the same game supports Typing and Maths. Mode changes challenge generation and statistics; it does not select a second game implementation.

## 2. Evidence inspected

- `internal/game/game.go` — authoritative game state, update loop, movement, wrapping, collisions, fragmentation, waves, challenge generation and statistics;
- `internal/game/game_test.go` — gameplay and determinism evidence;
- `internal/game/scores.go` and `scores_test.go` — local top-five score persistence;
- `internal/game/words.txt` — embedded spelling content;
- `cmd/astroanswers/main.go` — Ebitengine adapter, renderer, fullscreen scaling, embedded assets and sound;
- `cmd/astroanswers/assets/` — four PNG sprites and nine WAV cues;
- `.github/workflows/test.yml` — engineering gate;
- `go.mod`, `go.sum`, `VERSION`, `CHANGELOG.md` — dependency and release boundary.

## 3. Runtime structure

```text
keyboard
   |
   v
cmd/astroanswers
   |
   v
Game.Update(Input)
   |
   +-- global answer prefix
   +-- move / wrap rocks
   +-- station collisions
   +-- hit / split
   +-- wave progression
   +-- bounded particles
   |
   v
Game state
   |
   +--> renderer + embedded PNGs
   +--> audio + embedded WAVs
   +--> local JSON top-five scores
```

There is no event bus, service layer, repository abstraction, general game engine, physics framework, network path or runtime model/AI dependency.

## 4. Authoritative state

**KNOWN** — one `Game` owns screen, mode, wave, score, shield, rocks, typed prefix, counters, gameplay RNG, cosmetic RNG, particles, scores, initials and short presentation counters.

**KNOWN** — there are exactly three input contexts: `Attract`, `Play`, and `Initials`. Game-over and wave notices are counters/overlays rather than additional state-machine types.

## 5. Rock model

**KNOWN** — `Rock` is plain data: monotonic ID, size, position, velocity, angle, spin, display text and answer text.

**KNOWN** — size determines collision radius, score and fragmentation behaviour. There are no separate asteroid classes.

## 6. Update order

**KNOWN** — play is single-threaded and explicit:

```text
read input
move rocks
wrap rocks
resolve station collisions
resolve pending answer hit
advance wave when field is empty
update particles
```

**KNOWN** — when shield reaches zero, gameplay freezes into a 90-tick presentation interval and then routes either to initials entry or back to attract mode.

## 7. Answer selection

**KNOWN** — there is no target lock and distance is not used to select an asteroid.

**KNOWN** — `Typed` is one global prefix. If the next character cannot continue the current prefix but can start another visible answer, the prefix switches immediately to that new answer.

**KNOWN** — different active answers are prevented from being prefixes of one another. Identical Maths answers may coexist and are resolved by the lowest rock ID.

## 8. Challenge generation

**KNOWN** — Typing uses one embedded list of three-letter words.

**KNOWN** — Maths uses wave number directly: Wave 1 is ×2 through Wave 11 ×12; Wave 12 onward mixes ×2–×12. The second operand is 1–12 and answers are calculated rather than stored in a multiplication table.

## 9. Movement, collision and fragmentation

**KNOWN** — rocks move by adding velocity once per tick and wrap independently on both axes.

**KNOWN** — station collision is a circle-distance test against the fixed centre and costs one shield.

**KNOWN** — a large rock splits into two medium rocks, a medium into two small rocks, and a small rock is removed. Children are faster than their parent and receive new challenges.

## 10. Difficulty and resource bounds

**KNOWN** — starting large-rock count grows slowly and caps at eight. Base speed grows slowly and caps at 1.75.

**KNOWN** — gameplay is bounded by `MaxRocks = 40`; cosmetic particles are bounded by `MaxParticles = 256`. Cosmetic overflow is dropped, gameplay state is not.

## 11. Determinism

**KNOWN** — gameplay and cosmetic randomness use separate seeded RNGs.

**KNOWN** — tests assert that the same seed, inputs and tick count reproduce gameplay state and that cosmetic random activity does not perturb gameplay state.

## 12. Persistence

**KNOWN** — the only persisted application state is two independent top-five score tables in local JSON: Typing and Maths.

**KNOWN** — writes use a temporary file followed by rename. Missing files start empty; malformed data returns an error to the adapter, which reports it and continues with empty scores.

## 13. Adapter and presentation

**KNOWN** — Ebitengine provides the window/game loop, keyboard input, rendering, transforms and audio.

**KNOWN** — the logical playfield is fixed at 960×720. Final-screen drawing preserves 4:3 geometry and fills unused physical display space black rather than stretching the game.

**KNOWN** — four PNGs provide the station and three asteroid sizes. Stars, panel geometry, text treatment and particles are code-drawn. Nine short WAV cues are embedded and decoded at startup.

**KNOWN** — styled title/end-screen text uses unchanged light fills, dark-grey glow/outline treatment and cyan structural lines/panel accents.

## 14. Network and privacy boundary

**KNOWN** — source inspection finds no network client, HTTP dependency, telemetry path, account system, advertising, secrets or runtime AI/model integration.

## 15. Engineering evidence

The release workflow requires:

```text
go mod tidy with no module-file diff
gofmt cleanliness
go vet ./...
go test -race -count=1 ./...
Linux/amd64 build
Linux/arm64 build
core gameplay acceptance tests
```

The constrained-hardware benchmark is deliberately separate physical evidence and remains open until run on the selected Garage target.
