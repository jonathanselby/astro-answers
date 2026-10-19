# Astro Answers --- Engineering Design Document

**Version:** 0.3\
**Status:** Final — V0.1.0, red-teamed against Garage engineering rules\
**Implementation:** Go

## 1. Purpose

Astro Answers is one small real-time arcade game with two challenge sets:
Typing and Maths.

The governing engineering rule is:

> **Keep the game as small as the game actually is.**

The implementation is not a reusable game engine. It is one game whose
richness comes from a few interacting rules: movement, wrapping,
fragmentation, increasing speed and answer matching.

The engine knows nothing about education. A rock has text to display and
text that counts as the correct answer.

## 2. Design constraints

The design follows these constraints:

1.  one authoritative game state;
2.  one explicit update loop;
3.  deterministic rules where practical;
4.  no abstraction before a demonstrated need;
5.  bounded gameplay and cosmetic resources;
6.  no runtime machinery that does not improve the game;
7.  visual polish comes from assets and execution, not architecture;
8.  tests exercise the game rules without requiring the renderer.

If implementation becomes materially more complicated than this EDD,
first ask what can be deleted.

## 3. Runtime dependency

Use Ebitengine for:

-   window and game loop;
-   keyboard input;
-   2D drawing and transforms;
-   text;
-   audio.

Do not add another game, UI or physics framework.

``` text
Ebitengine
    |
    v
  Game
    |
    v
plain Go rules
```

## 4. Authoritative state

The application has one authoritative `Game`.

Conceptually:

``` go
type Game struct {
    Screen Screen
    Mode   Mode

    Wave   int
    Score  int
    Shield int

    Rocks []Rock

    Typed string

    Correct int
    Errors  int
    Solved  int
    Ticks   int

    RNG *rand.Rand

    Particles []Particle
}
```

This is not a required final field layout. It states the design
boundary: game truth lives in one place.

Rendering and sound read or react to this state. They do not own game
truth.

## 5. Rock

A rock is plain data.

``` go
type Rock struct {
    ID   int
    Size Size

    X, Y   float64
    VX, VY float64

    Angle float64
    Spin  float64

    Display string
    Answer  string
}
```

`ID` is deliberately retained.

The previous draft attempted to remove rock identity and use slice
position for removal. Red-teaming showed that this saves one integer but
creates hidden bookkeeping whenever rocks are removed or the slice changes.
A small monotonic integer ID makes removal and deterministic tie handling
simpler and more testable.

This is a useful distinction:

> **Fewer fields is not always simpler. Fewer rules and fewer failure
> paths is simpler.**

There are three sizes:

``` go
const (
    Large Size = iota
    Medium
    Small
)
```

Size determines scale, collision radius, score, speed and fragmentation.

There are no separate asteroid classes.

## 6. Concepts deliberately kept as values

These do not need objects or subsystems:

``` text
station     = fixed centre + radius
shield      = integer
wave        = integer
score       = integer
maths level = wave
typed input = string
```

The station does not need a `Station` type.

Wave does not need a `Wave` type.

Score does not need a scoring service.

Maths does not need an independent level model.

## 7. Challenge generation

There is no challenge-provider interface.

Use ordinary functions.

Conceptually:

``` go
func newChallenge(mode Mode, wave int, rng *rand.Rand) (display, answer string)
```

Typing returns a three-letter word as both values.

Maths generates a multiplication question and its numeric answer.

The two returned strings are copied directly onto a new rock.

There is no persistent `Challenge` object in the runtime because the
rock already contains the only two values the game needs.

### Typing

Use an embedded curated list of common three-letter English words.

Tests validate that every entry is exactly three letters.

Avoid immediate repetition where convenient.

### Maths

Wave is the level:

``` text
Wave 1    ×2
Wave 2    ×3
...
Wave 11   ×12
Wave 12+  mixed ×2–×12
```

For Waves 1--11:

``` go
table := wave + 1
```

Choose the second operand from 1--12.

For Wave 12 onward choose the table from 2--12 and the second operand
from 1--12.

Calculate the answer. Do not store multiplication tables.

## 8. Screens

Only three input contexts are required:

``` text
ATTRACT
PLAY
INITIALS
```

`ATTRACT` covers title, mode selection and high-score display.

`PLAY` runs the game.

`INITIALS` accepts a qualifying three-character high-score entry and returns to `ATTRACT` immediately after the third character.

Game-over and wave announcements are presentation overlays, not architectural states. The game-over overlay explicitly indicates whether initials are next or the game is returning to the title screen, so the transition never appears frozen.

## 9. Update loop

The update sequence remains explicit and single-threaded.

``` text
read input
    |
match / advance typed prefix
    |
move rocks
    |
wrap rocks
    |
resolve station collisions
    |
remove or split completed answer
    |
if no rocks remain:
    wave++
    spawn next wave
    |
update cosmetic particles
```

No gameplay goroutines.

No channels.

No event bus.

No asynchronous actors.

The sequence is intentionally visible because order matters.

## 10. Time and movement

Use Ebitengine's fixed update cycle.

Each tick:

``` go
rock.X += rock.VX
rock.Y += rock.VY
rock.Angle += rock.Spin
```

Asteroids move in straight lines.

There is no:

-   steering;
-   pathfinding;
-   gravity;
-   acceleration model;
-   rock-to-rock collision;
-   orbital simulation.

The combination of straight-line motion, wrapping and fragmentation
supplies the movement complexity.

## 11. Wrapping

The playfield wraps horizontally and vertically.

Conceptually:

``` go
if rock.X < 0 {
    rock.X += width
}
if rock.X >= width {
    rock.X -= width
}

if rock.Y < 0 {
    rock.Y += height
}
if rock.Y >= height {
    rock.Y -= height
}
```

Wrapping changes only position.

It does not reset or otherwise alter the rock.

### Visual continuity

When a sprite overlaps an edge, draw the same rock at the corresponding
opposite edge as well. Near a corner it may be drawn in up to four
places.

There is still one gameplay rock.

This keeps the simulation simple while making wrapping visually
continuous.

## 12. Wave creation and difficulty

A new run starts with:

``` text
Wave   = 1
Shield = 3
Score  = 0
```

A wave spawns a bounded number of large rocks away from the station with
varied trajectories.

Difficulty has only two deliberate inputs:

``` text
starting large-rock count
base rock speed
```

Both are simple functions of wave number.

Fragmentation supplies the rest of the difficulty.

One large rock can become:

``` text
1 large
2 medium
4 small
```

and fragments become faster.

Do not add dynamic difficulty, skill estimation, difficulty profiles or
arbitrary continuous spawning unless play testing demonstrates a real
need.

## 13. Fragmentation

One function owns the gameplay rule:

``` go
hitRock(id)
```

Behaviour:

``` text
Large  -> remove parent, create two Medium
Medium -> remove parent, create two Small
Small  -> remove
```

Children:

-   begin at or near the parent position;
-   receive new display/answer values;
-   inherit the parent's general direction;
-   diverge from one another;
-   move faster;
-   receive their own spin.

Exact speed and divergence values are tuning constants.

Decorative debris is not gameplay state.

## 14. Input matching

The player never aims manually and there is no target lock.

`Typed` is one global input prefix:

1.  append the next character conceptually;
2.  if an active `Answer` begins with that prefix, accept the character and advance `Typed`;
3.  otherwise test the newest character alone as the start of a different visible answer;
4.  if that single-character prefix matches, replace `Typed` with it and continue from the new answer;
5.  otherwise count an error and leave `Typed` unchanged;
6.  if `Typed` exactly matches an active answer, hit one matching rock deterministically by rock ID;
7.  clear `Typed` after a successful hit.

Distance from the station is never used to choose what the player may solve. A
player can spell or solve any visible asteroid simply by entering its answer.

For this to remain unambiguous without another selection control, challenge
generation enforces one small invariant: two *different* active answers may
not be prefixes of one another. For example, `12` and `120` do not coexist.
Identical Maths answers may coexist because the same typed answer is valid for
either; the lower rock ID is removed first. Typing challenges avoid duplicate
active words.

If a station collision removes the last asteroid compatible with the current
`Typed` prefix, clear `Typed`.

## 15. Wrong input

Wrong input already has a natural penalty: time passes while rocks
continue moving.

Therefore an incorrect key:

-   increments `Errors`;
-   produces short visual/audio feedback;
-   changes nothing else.

No score deduction.

No shield damage.

No full-answer reset.

## 16. Station collision

The station is fixed at screen centre.

Use circle collision:

``` text
distance(rock, centre) < station radius + rock radius
```

On collision:

``` text
remove rock
shield--
show impact
play sound
```

At zero shield, end the run.

No pixel-perfect or general collision engine is required.

## 17. Scoring

One function maps size to points.

Initial values:

``` text
Large      20
Medium     50
Small      100
```

On a successful hit:

``` go
Score += points(rock.Size)
```

No combos, multipliers, experience or educational scoring.

## 18. Randomness

Use one seedable gameplay RNG for:

-   spawn position;
-   direction;
-   spin;
-   challenge selection;
-   fragment divergence;
-   sprite selection.

A test can reproduce a run from seed + input + ticks.

Cosmetic randomness that can consume a variable number of random values must
not share the gameplay stream. Particles and other visual-only variance either
use deterministic derivation from existing state or a separate cosmetic RNG.

## 19. Rendering boundary

The renderer reads game state and draws it.

It does not decide:

-   which answer is active;
-   whether a hit is correct;
-   whether a rock splits;
-   whether a wave is complete;
-   whether the player has died.

Rendering is an adapter over the authoritative game state.

## 20. Visual presentation

The logical playfield is fixed at 960×720.

The Linux V0.1 application launches full-screen by default and preserves the
4:3 logical aspect ratio on wider or taller physical displays. The final image
is scaled uniformly and centred; unused display area remains black.

Do not create a bespoke viewport system. Use Ebitengine's normal logical
layout and final screen scaling.

The game uses four embedded production sprites:

-   large asteroid;
-   medium asteroid;
-   small asteroid;
-   station.

Stars, shield arcs, HUD geometry, flashes and particles are drawn in code.

Hero sprites remain raster assets. Converting them to procedural/vector art
solely to reduce bytes would reduce visual quality while increasing rendering
code.

Asteroid challenge text remains upright while sprites rotate beneath it. It is
larger than secondary HUD copy, uses high-contrast light text and may be given
a small dark backing plate so asteroid texture can never reduce contrast.
A tiny bitmap face is sufficient; do not add a font asset pipeline.

Readability wins over decoration.

## 21. Cosmetic particles

Particles are the one small cosmetic data structure justified by the
visual requirement.

Conceptually:

``` go
type Particle struct {
    X, Y   float64
    VX, VY float64
    Life   int
    Kind   uint8
}
```

Each tick:

``` text
move
decrement life
remove when expired
```

Particle count is explicitly bounded.

When full, drop additional cosmetic particles.

Never drop gameplay state to preserve an effect.

There is no general effects engine.

## 22. Sound

Required cues:

-   accepted input;
-   incorrect input;
-   large split;
-   medium split;
-   small destruction;
-   station impact;
-   new wave;
-   game over;
-   high score.

Use Ebitengine's audio facilities directly.

No separate audio architecture is required.

Music is out of V0.1.

## 23. HUD

Typing:

``` text
SCORE          WAVE          SHIELD
```

Maths Waves 1--11:

``` text
SCORE           ×7           SHIELD
```

Maths Wave 12 onward:

``` text
SCORE          MIXED         SHIELD
```

Do not display redundant information.

## 24. Statistics

Do not create a `Stats` subsystem unless the code becomes clearer with a
tiny struct.

The authoritative values required during a run are only counters:

``` text
correct input
errors
challenges solved
elapsed ticks
```

Typing derives accuracy and WPM.

Maths derives accuracy, average response time and highest table reached.

V0.1 does not persist learning history.

## 25. High scores

High scores are the only required persistent state.

Maintain:

``` text
Typing top 5
Maths top 5
```

Each entry contains:

``` text
three-character initials
score
```

Use one small local JSON file.

Persistence remains ordinary functions:

``` go
loadScores()
saveScores()
addScore()
```

No repository pattern, adapter or database.

Write safely using a temporary file followed by rename.

If the file is absent, start empty.

If it is malformed, report the problem and start empty.

High-score corruption must not prevent the game from running.

## 26. Assets

Embed required assets in the executable.

Expected asset categories are small:

``` text
asteroids
station
effects
audio
font
word list
```

A starting asteroid set may be only a handful of excellent images.

Do not create asset quantity as a proxy for quality.

Missing required visual assets fail startup visibly.

Missing non-critical sound may degrade gracefully.

## 27. Resource budgets

V0.1 has two explicit runtime ceilings:

``` text
MaxRocks = 40
MaxParticles = 256
```

The maximum starting wave is eight large asteroids, which can produce at
most 32 simultaneous rocks after fragmentation, leaving headroom below the
40-rock ceiling. Additional particles are dropped when the cosmetic ceiling
is full; gameplay state is never dropped for an effect.

Sound uses nine short embedded cues and Ebitengine audio directly. V0.1 does
not add a sound-pool or player manager because measured gameplay did not
justify another subsystem. Asset memory is dominated by four small PNGs and
nine short WAV files and is recorded as build evidence rather than guarded by
an arbitrary runtime limit.

Do not introduce object pooling unless measurement shows garbage collection
materially affects play.

> **Bound gameplay state. Measure the rest. Optimise only when evidence says to.**

## 28. Repository shape

The V0.1 implementation remains deliberately small:

``` text
cmd/
    astroanswers/
        main.go
        assets/
            4 PNG sprites
            9 WAV cues

internal/
    game/
        game.go
        game_test.go
        scores.go
        scores_test.go
        test_helpers_test.go
        words.txt
```

Assets live beside the only adapter that embeds and renders them. The word
list lives beside the game rules that consume it. Scores earned their own
file because persistence is a separate concern; input did not earn a package
or subsystem.

Split files only when the code becomes materially clearer.

Do not create package hierarchies from architectural vocabulary.

No default:

``` text
engine/
domain/
services/
providers/
repositories/
controllers/
managers/
systems/
```

## 29. Failure behaviour

  Failure                             Behaviour
  ----------------------------------- ----------------------
  high-score file absent              start empty
  high-score file malformed           report; start empty
  optional sound unavailable          continue without it
  required visual asset unavailable   fail startup visibly
  invalid embedded word data          fail test/startup
  programming/runtime failure         fail visibly

Failure behaviour remains explicit but proportionate to a local arcade
game.

## 30. Testing

The game rules must run without opening a graphical window.

### Movement

Test:

-   straight-line movement;
-   all four wrapping directions;
-   gameplay state unchanged by wrapping.

### Fragmentation

Test:

-   Large -\> two Medium;
-   Medium -\> two Small;
-   Small -\> removed;
-   fragments are faster;
-   children receive new challenges.

### Station

Test:

-   collision removes rock;
-   one collision removes one shield;
-   third hit ends the run.

### Input matching

Test:

-   a player can complete a farther spelling asteroid even when a nearer asteroid shares its prefix;
-   a player can complete any visible Maths answer without proximity selection;
-   different active answers are not allowed to create prefix ambiguity;
-   identical Maths answers resolve deterministically by ID;
-   incorrect input does not erase progress;
-   a collision clears `Typed` when no remaining answer matches the prefix.

### Waves

Test:

-   empty field increments wave;
-   next wave spawns;
-   count and speed functions behave as intended.

### Typing

Test every embedded word is exactly three letters.

### Maths

Test:

``` text
Wave 1   -> ×2 only
Wave 11  -> ×12 only
Wave 12+ -> ×2..×12 only
```

Verify generated answers mathematically.

### Scores

Test:

-   Typing and Maths remain independent;
-   descending order;
-   only five retained;
-   non-qualifying score rejected;
-   save/load round trip;
-   malformed file behaviour.

### Determinism

Given the same:

``` text
seed
input sequence
tick count
```

the gameplay state should reproduce.

### Build evidence

CI runs:

``` text
go vet ./...
go test -race ./...
linux/amd64 build
linux/arm64 build
```

## 31. Constrained-hardware benchmark

After the game is correct and polished, run the standard Garage
constrained-hardware benchmark.

Record:

-   hardware;
-   OS;
-   logical and displayed resolution;
-   frame rate;
-   CPU;
-   memory;
-   executable size;
-   maximum observed active rocks;
-   maximum observed particles.

The benchmark proves the design; it does not drive premature
optimisation.

## 32. Privacy and network boundary

V0.1:

-   requires no network;
-   sends no telemetry;
-   has no account;
-   has no advertising;
-   has no online leaderboard;
-   stores no learning profile;
-   requires no secrets.

Only local initials and scores persist.

## 33. Deliberately absent architecture

EDD v0.3 deliberately does not contain:

``` text
Challenge type in runtime state
ChallengeSource interface
content plugin framework
Station object
Wave object
Maths level object
score service
score-store abstraction
difficulty subsystem
event bus
threat prediction
gameplay goroutines
channels
persistent WaveClear state
persistent GameOver state
persistent learning history
general effects engine
general collision engine
large package hierarchy
object pooling without evidence
```

One earlier deletion has been reversed:

``` text
Rock.ID
```

It earns its keep because it removes slice-index bookkeeping and makes
answer/removal behaviour simpler and more robust.

## 34. Architecture summary

``` text
                  KEYBOARD
                     |
                     v
               +-----------+
               |   Game    |
               +-----------+
                |    |    |
                |    |    +---- wave / score / shield
                |    |
                |    +--------- typed prefix
                |
                +-------------- []Rock
                                    |
                         position / velocity
                         size / spin
                         display / answer

                  Game.Update()
                       |
        +--------------+--------------+
        |              |              |
      input           move          collide
        |              |              |
      input           wrap          shield
        |
       hit
        |
      split
        |
    next wave


                  Game.Draw()
                       |
                  Ebitengine
```

Persistent state:

``` text
two top-five score tables
```

Content:

``` text
one curated three-letter word list
+
multiplication generated from integers
```

Cosmetics:

``` text
embedded art and sound
+
one bounded particle slice
```

That is the application.

## 35. Design conclusion

The game should look substantially more sophisticated than its
implementation.

Wrapping creates movement complexity.

Fragmentation creates population complexity.

Fragment speed creates pressure.

Challenge text creates the learning mechanic.

Good artwork, rotation, particles and sound create visual richness.

A top-five score table creates replay motivation.

The red-team test is therefore not whether the design has the fewest
possible structs or fields.

It is:

> **Does every piece of machinery remove more complexity than it
> introduces?**

For EDD v0.3, the answer leaves us with one authoritative game state, plain
rock data, ordinary functions, one update loop, one local score file and
a thin 2D runtime dependency.

Anything substantially beyond that must earn its place.
