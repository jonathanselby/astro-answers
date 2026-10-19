# Astro Answers — Design Notes

Astro Answers was designed by reducing the game until the learning action and the arcade action became the same thing.

> **The answer is the weapon.**

The useful design history is therefore mostly a record of what was kept small.

## Start with one mechanic

The original product question was whether repetitive keyboard and times-table practice could become intrinsically replayable without reward systems, currencies, campaigns or educational dashboards. Correct answers destroy asteroids. That is the product.

Typing and Maths deliberately share the same game. The game state does not know about education; each rock only needs display text and an answer.

## Complexity comes from interaction

The playfield uses a few rules that compound:

- movement;
- screen wrapping;
- fragmentation;
- faster fragments;
- increasing wave pressure;
- time pressure created by the moving field.

This produces apparent complexity without adding conventional game systems.

## Answer selection became simpler through play

The first targeting idea implicitly favoured the nearest asteroid. Play testing showed that this was the wrong rule: the player should solve any threat they can see. Target selection was removed entirely.

A second problem appeared when different numeric answers shared prefixes, for example `12` and `120`. The rule was tightened so different active answers cannot be prefixes of one another. Identical Maths answers may coexist and resolve deterministically.

A final play-test improvement allowed the player to abandon a partial answer simply by typing the first valid character of another visible answer. No target-selection control or extra state was added.

## Visual quality without a visual framework

The game deliberately separates hero artwork from simple geometry.

Raster sprites are used for the station and three asteroid sizes because these are the visual identity of the game. Stars, HUD elements, frames, shield indicators, text backing plates and particles are drawn directly in code.

Vectorising the hero sprites was considered and rejected. The small binary saving was not worth reducing the quality of the parts the player looks at constantly.

The title and result screens use strong type, restrained framing and the same game art rather than introducing a separate UI system. The final treatment keeps cyan lines as structural accents while the decorative outline around styled text is dark grey.

## Sound

Nine short sound cues were created specifically for Astro Answers. They are embedded WAV files and played directly through Ebitengine. A runtime synthesiser or audio subsystem would save little and add machinery, so neither exists.

## Display

V0.1 is Linux desktop only. Gameplay stays at a fixed 960×720 logical 4:3 playfield. Full-screen output is uniformly scaled and centred, with black unused space rather than stretching the geometry.

Mobile was kept out of the code. The PRD records a V0.2 mobile direction—portrait play, falling threats and device-native text/numeric input—but V0.1 contains no abstractions added in anticipation of it.

## Persistence

Only the two top-five score tables persist. They use one local JSON file written through a temporary file and rename. There is no database or repository layer.

## What was deliberately rejected

V0.1 does not contain player movement, aiming, weapons, power-ups, achievements, campaigns, accounts, online leaderboards, adaptive learning, generated content, a reusable game engine, an event bus, a physics engine, a UI framework or runtime AI.

The governing test remained:

> **Does every piece of machinery remove more complexity than it introduces?**
