# Astro Answers — Design / Implementation Comparison

**Implementation evidence:** `5acdc382a4bc51a6e617a525f3ad4ac108cbde10`  
**Generated:** 2026-10-19

| Design area | Intended contract | Recovered implementation | Assessment |
|---|---|---|---|
| one authoritative state | one `Game`; renderer owns no truth | one `Game` owns gameplay, counters and score state | MATCH |
| modes | one game, two challenge sets | Typing/Maths share the same update and rock model | MATCH |
| screens | Attract / Play / Initials only | exactly those three `Screen` values | MATCH |
| update model | explicit single-threaded loop | input → move → wrap → collision → hit → wave → particles | MATCH |
| targeting | any visible answer; no proximity lock | one global prefix, no distance-based target selection | MATCH |
| mid-answer switching | newest valid starting key may switch answer | implemented in `readInput` | MATCH |
| prefix ambiguity | distinct active answers must not prefix-conflict | `answerCompatible` enforces the invariant | MATCH |
| duplicate Maths answers | allowed; deterministic removal | lowest matching rock ID wins | MATCH |
| movement | straight-line + wrapping | velocity addition plus per-axis wrapping | MATCH |
| fragmentation | large → 2 medium → 2 small → removed | implemented in `hitRock` / `split` | MATCH |
| difficulty | bounded count and base speed | eight-large-rock cap; speed cap 1.75 | MATCH |
| resource bounds | bounded gameplay and particles | 40 rocks; 256 particles | MATCH |
| RNG | deterministic gameplay; cosmetics isolated | separate gameplay and cosmetic RNGs | MATCH |
| persistence | local JSON top-five tables only | `scores.go`; temp-file + rename | MATCH |
| renderer | thin Ebitengine adapter | one `cmd/astroanswers/main.go` adapter | MATCH |
| display | fixed 960×720 4:3, no stretch | logical layout + black final-screen fill | MATCH |
| hero assets | raster sprites retained | four embedded PNGs | MATCH |
| simple effects | code-drawn stars/HUD/particles | direct Ebitengine drawing | MATCH |
| audio | nine original short cues; no audio subsystem | nine embedded WAVs, direct players | MATCH |
| title/end treatment | text fill unchanged; styled outline dark grey; cyan lines retained | renderer uses dark-grey glow/outline and cyan structural accents | MATCH |
| network/privacy | no network, telemetry, accounts or AI | no network/runtime AI path in source | MATCH |
| CI | format, vet, race, amd64, arm64, acceptance | encoded in `.github/workflows/test.yml` | MATCH |

## Material corrections made during development

Play testing changed targeting from nearest-compatible asteroid selection to global visible-answer selection. This removed an unnecessary spatial rule and made the input mechanic match the product idea: the answer itself is the weapon.

Maths testing exposed answer-prefix ambiguity such as `12` versus `120`. The final design prevents different active answers from being prefixes of one another while still permitting identical Maths answers with deterministic ID tie-breaking.

Play testing also showed that a player needed to be able to abandon a partially-entered answer without a separate targeting control. The final input rule therefore switches to a new visible answer whenever the newest key cannot continue the current prefix but can start another one.

Late pre-release testing exposed an engineering-process failure rather than a game failure: a proven Linux launcher/packaging path was changed at the same time as visual and gameplay work. The correction was to return to the last known-good harness and change only the product source. This is recorded in `TESTING.md` rather than hidden.

The final visual pass changed only the decorative outline/glow around styled text from blue to dark grey. Text fills and cyan structural lines remain unchanged.

## Open evidence

The standard Garage constrained-hardware benchmark has not yet been recorded on the selected physical target. The repository states that explicitly rather than claiming evidence that does not exist.

## Result

**MATCH after correction.** The V0.1.0 implementation aligns with the final hand-written PRD and EDD. The material divergences found during design and play testing were corrected before the evidence snapshot; no unexplained architecture mismatch remains.
