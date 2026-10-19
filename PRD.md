# Astro Answers — Product Requirements Document

**Version:** 0.1  
**Status:** Final — V0.1.0  
**Product name:** Astro Answers

## Problem

Typing practice and multiplication-table practice are repetitive by nature. Most learning software exposes that repetition directly and then attempts to make it engaging by adding rewards, progression systems or educational UI around it.

The repetition itself can instead be the game mechanic.

Astro Answers uses a small arcade game in which answering correctly is how the player destroys threats. The game should be enjoyable without needing a separate reward system wrapped around the learning.

## Product objective

Build a polished, lightweight arcade game in Go in which a stationary space station is defended from asteroids by correctly answering the challenge displayed on each asteroid.

> **The answer is the weapon.**

Large asteroids split into two faster medium asteroids. Medium asteroids split into two faster small asteroids. Small asteroids are destroyed.

Asteroids move freely across a wrapping playfield: leaving one edge returns them from the opposite edge.

The product has two modes:

- **Typing** — common three-letter words;
- **Maths** — multiplication tables from ×2 through ×12, followed by mixed endless play.

Both modes use the same arcade game.

The V0.1 product question is:

> **Can a very small set of arcade rules make repetitive typing and times-table practice sufficiently enjoyable that the player voluntarily starts another game?**

## Users

Primary users are children practising keyboard skills or multiplication tables.

Secondary users are anyone who enjoys a simple score-driven arcade game.

The repository also serves as a Garage engineering example: a visually polished application produced from a deliberately small runtime design.

## In scope

- stationary central space station;
- three shield segments;
- large, medium and small asteroids;
- free asteroid movement;
- continuous screen wrapping on all four edges;
- asteroid rotation;
- large asteroid splitting into two medium asteroids;
- medium asteroid splitting into two small asteroids;
- increasing fragment speed;
- wave progression;
- keyboard-only gameplay;
- free answer selection from player input;
- Typing mode using common three-letter English words;
- Maths mode using multiplication tables;
- Maths Wave 1 = ×2 through Wave 11 = ×12;
- Maths Wave 12 onward = mixed ×2–×12;
- increasing arcade difficulty after the educational progression is exhausted;
- scoring;
- separate persistent top-five high-score tables for Typing and Maths;
- three-character high-score initials;
- arcade-style attract screen;
- immediate restart;
- high-quality 2D asteroid, station, fragmentation and explosion graphics;
- concise sound effects;
- end-of-game learning statistics;
- deterministic core gameplay suitable for automated testing;
- Linux desktop V0.1 build;
- full-screen 4:3 presentation with aspect-ratio-preserving scaling;
- Linux/ARM64 build proof.

## Out of scope

- player movement;
- aiming;
- conventional weapons;
- physics simulation;
- 3D graphics;
- power-ups;
- upgrades;
- inventories;
- currencies;
- campaigns;
- achievements;
- player accounts;
- online leaderboards;
- multiplayer;
- cloud services;
- adaptive learning;
- generated educational content;
- AI or model runtime;
- longer words or progressively harder vocabulary;
- mathematics beyond multiplication tables ×2–×12;
- Windows and macOS packaging in V0.1;
- phone and tablet support in V0.1;
- touch input or responsive mobile layouts in V0.1.

## V0.2 mobile direction

V0.2 extends Astro Answers to iPhone and iPad with a mobile-specific playfield rather than shrinking the Linux 4:3 layout.

On iPhone and iPad:

- the playfield is portrait-oriented;
- asteroids enter from the top of the screen and move downward;
- the space station is positioned along the bottom of the screen;
- Typing mode uses the device's on-screen keyboard;
- Maths mode uses the device's on-screen numeric keypad.

The mobile version preserves the learning mechanic, challenge generation, fragmentation, scoring, waves and shields from V0.1. Spawn, movement and collision geometry are adapted to the mobile layout; the Linux wrapping playfield is not simply scaled onto the device.

This is forward scope only. It does not change the V0.1 Linux implementation.

These exclusions are deliberate. Difficulty should come primarily from the interaction of movement, fragmentation, speed and the player's response time rather than additional game systems.

## Functional requirements

### FR1 — Continuous playfield

Asteroids move freely across the playfield. An asteroid crossing one boundary re-enters from the corresponding opposite boundary without changing its state or trajectory.

### FR2 — Fragmentation

A correctly answered large asteroid becomes two medium asteroids.

A correctly answered medium asteroid becomes two small asteroids.

A correctly answered small asteroid is removed.

Fragments move faster than their parent.

### FR3 — Station damage

An asteroid intersecting the station is removed and costs one shield segment.

The game ends after the third shield segment is lost.

### FR4 — Waves

A wave is complete when no asteroids remain.

Later waves increase pressure through more starting asteroids and/or increased speed.

### FR5 — Typing mode

Each asteroid displays a common three-letter English word.

Correctly entering the word hits the asteroid.

Word length remains three characters throughout the game.

### FR6 — Maths mode

Each asteroid displays a multiplication question and is hit by entering the numeric answer.

Maths progression is:

```text
Wave 1     ×2
Wave 2     ×3
Wave 3     ×4
Wave 4     ×5
Wave 5     ×6
Wave 6     ×7
Wave 7     ×8
Wave 8     ×9
Wave 9     ×10
Wave 10    ×11
Wave 11    ×12
Wave 12+   mixed ×2–×12
```

Each table uses multipliers from 1 to 12.

### FR7 — Answer selection

The player does not manually aim and is never required to solve the asteroid closest to the station.

Keyboard input forms one global answer prefix. A character is accepted whenever the resulting prefix still matches at least one visible asteroid. If the current prefix cannot be continued but the newest character can begin another visible answer, the current prefix is abandoned and the new answer starts immediately. The player may therefore change their mind mid-answer without a targeting control.

Completing an answer destroys a matching asteroid. Challenge generation prevents two different active answers from being prefixes of one another, so a shorter valid answer cannot block entry of a longer visible answer. Identical Maths answers may coexist; completing that answer removes one deterministically.

### FR8 — High scores

Typing and Maths maintain separate persistent top-five score tables.

A qualifying score prompts the player for three initials.

Only the five highest scores for each mode are retained.

### FR9 — Attract mode

When idle, the game shows one arcade-style attract screen containing the title, game selection and both persistent top-five high-score tables at the same time. Maths scores appear on the left and Typing mode is presented to the player as **Spelling** on the right.

The title screen uses a deliberate graphic hierarchy: a strong **Astro Answers** logotype treatment, a smaller matching **Choose Your Game** treatment, visually framed Maths and Spelling score panels, and **The Answer Is The Weapon** in the same visual family.

A new game can be started directly from that screen using `M` for Maths or `T` for Spelling/Typing, without navigating a conventional application menu.

### FR10 — Visual quality

The game uses a restrained retro arcade visual language executed to a high modern standard.

The small visual system should concentrate quality on:

- asteroid artwork and rotation;
- fragmentation;
- explosions and debris;
- the central station and shield damage;
- crisp challenge text.

Effects must never obscure the challenge.

Challenge text remains readable and does not rotate with the asteroid beneath it.

## Acceptance criteria

V0.1 is accepted when automated and playable evidence shows that:

1. Typing and Maths run through the same arcade game without duplicated game implementations;
2. asteroids wrap correctly across every playfield boundary;
3. large → two medium → four small fragmentation behaves deterministically;
4. fragments are faster than their parent;
5. station collisions remove one shield and three collisions end the game;
6. a cleared field starts the next wave;
7. Typing challenges contain three-letter words and are completed from keyboard input;
8. Maths waves progress from ×2 through ×12 and then remain in mixed ×2–×12 play;
9. the player can complete any visible challenge without proximity-based target selection and can switch to another visible answer mid-entry;
10. Typing and Maths retain independent top-five high-score tables between application runs;
11. game over reliably transitions either to three-character high-score entry or back to the attract screen;
12. qualifying scores support three-character initial entry and return to the attract screen after entry;
13. the game remains immediately playable using only the keyboard;
14. challenge text is prominent, high-contrast and legible during asteroid movement and rotation;
15. gameplay logic can be tested without running the graphical renderer;
16. Linux V0.1 launches directly into full-screen play;
17. the 4:3 logical playfield preserves its aspect ratio on wider or taller monitors, using centred black pillarboxing or letterboxing rather than stretching;
18. `go vet ./...`, `go test -race ./...`, Linux/amd64 build and Linux/ARM64 build pass;
19. the finished application is run through the Garage constrained-hardware benchmark.

## Definition of done

V0.1 is done when all acceptance criteria have executable or observable evidence, the game can be understood and played without instructions, both learning modes are complete, the top-five tables persist correctly, and the visual execution is good enough that simplicity of implementation is not visible as lack of polish.

The hand-written EDD is then recovered independently from the implementation using the Garage codebase-to-EDD process and compared with the intended design. Material divergence is either corrected or documented.

The final product test remains:

> **The player dies and immediately wants another go.**
