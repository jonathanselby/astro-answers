# Astro Answers — Testing and Debugging Record

Testing was split deliberately between deterministic game-rule tests and real play. Both were necessary.

## Automated game-rule tests

`internal/game` runs without Ebitengine or a graphical window. The test suite covers:

- straight-line movement and wrapping in all four directions;
- large → two medium → four small fragmentation;
- fragment speed and scoring;
- station collision and shield loss;
- third-hit game over;
- wave progression;
- three-letter embedded words;
- Maths table progression from ×2 through ×12 and mixed play after Wave 11;
- answer selection independent of distance;
- spelling and Maths mid-answer switching;
- prefix-ambiguity prevention;
- deterministic resolution of identical Maths answers;
- wrong-input behaviour;
- game-over return paths;
- high-score initials;
- separate sorted top-five score tables;
- score-file save/load and malformed-file behaviour;
- deterministic replay for a fixed seed;
- isolation of cosmetic randomness from gameplay;
- bounded difficulty functions and derived run statistics.

Release commands are:

```bash
gofmt -w ./cmd ./internal
go vet ./...
go test -race -count=1 ./...
go build -trimpath -ldflags="-s -w" -o ./bin/astro-answers ./cmd/astroanswers
GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o ./bin/astro-answers-linux-arm64 ./cmd/astroanswers
```

## Human play testing

Automated tests did not find every important problem. Full-screen Linux play testing exposed several product-level failures:

1. **Numeric prefix ambiguity.** In later Maths play, a correct answer could appear to be rejected when another visible answer shared its prefix. Generation was changed so different active answers cannot be prefixes of one another.
2. **Unwanted proximity targeting.** Early behaviour effectively preferred a nearer compatible asteroid. This was removed. Any visible answer can be completed regardless of distance.
3. **Partial-answer trapping.** Once a player had started one answer, changing their mind was awkward. Input now continues the current prefix when possible; otherwise the newest key immediately starts another visible answer when valid.
4. **End-of-game presentation.** Maths and Spelling both appeared to freeze at the end of a run. Explicit game-over timing and transition tests were added for both the normal return path and the high-score initials path.
5. **Challenge readability.** Asteroid text was enlarged, made heavier and given a dark backing plate.
6. **Title/result presentation.** Functional diagnostic-looking screens were replaced by a deliberate graphic hierarchy while keeping the layout simple.

The result of this loop was not more architecture. In each case the preferred fix was a smaller rule or clearer presentation.

## When the test process itself went wrong

Late in the pre-release cycle, a proven Linux test launcher used for Builds 1–4 was changed at the same time as gameplay and presentation code. The new launcher embedded its source archive differently and failed before the game could start. A follow-up package fixed that mistake but then exposed an unrelated renderer type error during `go vet`.

The correction was to return to the last known-good Build 4 delivery mechanism and apply only the intended game changes. The launcher format, dependency path and build sequence were frozen again.

The lesson is simple:

> **Do not change a proven test harness while changing the system under test unless the harness itself needs to change.**

That failure is recorded because it is useful engineering evidence, not because the packaging experiment is part of the product. None of the failed launchers belongs in the public repository.

## Play-test evidence for V0.1

The complete V0.1 game was run on Linux in full-screen mode with:

- both learning modes;
- fragmentation and wave progression;
- unrestricted answer selection;
- mid-answer switching;
- shield/game-over flow;
- persistent high scores and initials;
- embedded artwork;
- all nine sound cues;
- 4:3 aspect-preserving full-screen output.

The final pre-publication feedback was that gameplay and features were complete. The only final presentation correction was changing the blue outline/glow around styled text to dark grey while leaving the text fill and cyan structural lines unchanged.

## Constrained-hardware benchmark

The Garage constrained-hardware benchmark remains a separate physical evidence step. It must record hardware, OS, displayed resolution, frame rate, CPU, memory, executable size, maximum observed rocks and maximum observed particles.

Do not substitute a desktop or synthetic benchmark and call it the constrained-hardware result. Until that physical run is recorded, the repository can be published and tested, but the benchmark item in `RELEASE.md` remains open.
