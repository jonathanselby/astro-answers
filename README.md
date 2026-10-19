# Astro Answers

**The answer is the weapon.**

Astro Answers is a small full-screen arcade game written in Go. A stationary space station is surrounded by moving asteroids. Each asteroid carries either a three-letter spelling challenge or a multiplication problem. Type the answer to hit it.

Large asteroids split into two faster medium asteroids. Medium asteroids split into two faster small asteroids. Small asteroids disappear. The same game supports both learning modes; only the challenge source changes.

The repository is also a Jonno's Garage engineering example: polished output from a deliberately small runtime design.

## Play

```text
M          start Maths
T          start Spelling
letters    answer any visible spelling challenge
0-9        answer any visible maths challenge
Esc        quit
```

There is no aiming control. Input is global. If a partially entered answer can no longer continue but the newest key begins another visible answer, Astro Answers switches immediately to that answer.

## V0.1

V0.1 targets Linux desktop. It launches full-screen using a fixed 960×720 logical playfield and preserves 4:3 geometry on wider or taller displays with centred black pillarboxing or letterboxing.

The game includes:

- Spelling using a curated embedded three-letter word list;
- Maths from ×2 through ×12, then mixed ×2–×12 play;
- three shield segments;
- large → medium → small fragmentation;
- wrapping movement;
- separate persistent top-five Maths and Spelling score tables;
- three-character high-score initials;
- end-of-run learning statistics;
- embedded original artwork and nine original sound cues;
- no network, telemetry, accounts, advertising or runtime AI.

V0.2 mobile direction is recorded in [`PRD.md`](PRD.md): portrait play, falling asteroids, station at the bottom, and device keyboard/number-pad input. No mobile machinery is present in V0.1.

## Build

Astro Answers uses Go 1.25+ and Ebitengine 2.10.5.

```bash
go build -trimpath -ldflags="-s -w" -o ./bin/astro-answers ./cmd/astroanswers
./bin/astro-answers
```

The executable embeds the required PNG and WAV assets. High scores are stored locally in `scores.json` beside the working directory.

## Engineering gate

```bash
gofmt -w ./cmd ./internal
go vet ./...
go test -race -count=1 ./...
go build ./cmd/astroanswers
GOOS=linux GOARCH=arm64 go build ./cmd/astroanswers
```

The gameplay package is deterministic for a fixed seed and can be tested without opening a graphical window. CI enforces formatting, vet, race tests, Linux builds and core acceptance tests.

## Architecture

```text
keyboard
   |
   v
Game.Update()  <---- one authoritative Game state
   |
   +-- input matching
   +-- move / wrap
   +-- station collision
   +-- hit / split
   +-- wave progression
   |
   v
Game state ----> thin Ebitengine adapter ----> screen + sound

local JSON <---- top-five scores only
```

There is no reusable game engine, event bus, service layer, repository abstraction, physics framework or UI framework. See [`EDD.md`](EDD.md) and [`architecture.txt`](architecture.txt).

## Design and evidence

- [`PRD.md`](PRD.md) — product contract and acceptance criteria.
- [`EDD.md`](EDD.md) — hand-written engineering design.
- [`DESIGN-NOTES.md`](DESIGN-NOTES.md) — material design decisions and rejected complexity.
- [`TESTING.md`](TESTING.md) — automated and human testing, including failures found during play testing and the packaging mistake that was corrected.
- [`RELEASE.md`](RELEASE.md) — release gate and evidence status.
- [`generated/EDD.md`](generated/EDD.md) — code-derived engineering view pinned to the release source snapshot.
- [`generated/COMPARISON.md`](generated/COMPARISON.md) — design versus implementation comparison.

## Licence

MIT. See [`LICENSE`](LICENSE).
