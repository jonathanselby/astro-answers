package game

import (
	"math"
	"reflect"
	"testing"
)

func newTestGame(seed int64, mode Mode) *Game {
	g := New(seed, HighScores{})
	g.startRun(mode)
	g.Rocks = nil
	g.WaveNotice = 0
	return g
}

func TestEmbeddedWordsAreThreeLetters(t *testing.T) {
	if len(words) < 50 {
		t.Fatalf("word list unexpectedly small: %d", len(words))
	}
	for _, word := range words {
		if len(word) != 3 {
			t.Fatalf("word %q is not three letters", word)
		}
	}
}

func TestStraightLineMovementAndWrapping(t *testing.T) {
	g := newTestGame(1, Typing)
	g.Rocks = []Rock{{ID: 1, X: LogicalWidth - 0.25, Y: 10, VX: 1, VY: -11, Answer: "cat"}}
	g.updatePlay(Input{})
	if math.Abs(g.Rocks[0].X-0.75) > 0.0001 {
		t.Fatalf("x wrap: got %v", g.Rocks[0].X)
	}
	if math.Abs(g.Rocks[0].Y-(LogicalHeight-1)) > 0.0001 {
		t.Fatalf("y wrap: got %v", g.Rocks[0].Y)
	}
	if g.Rocks[0].Answer != "cat" || g.Rocks[0].ID != 1 {
		t.Fatal("wrapping changed gameplay state")
	}
}

func TestWrappingAllDirections(t *testing.T) {
	cases := []struct {
		name         string
		x, y         float64
		vx, vy       float64
		wantX, wantY float64
	}{
		{"left", 0.1, 100, -1, 0, LogicalWidth - 0.9, 100},
		{"right", LogicalWidth - 0.1, 100, 1, 0, 0.9, 100},
		{"top", 100, 0.1, 0, -1, 100, LogicalHeight - 0.9},
		{"bottom", 100, LogicalHeight - 0.1, 0, 1, 100, 0.9},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newTestGame(2, Typing)
			g.Rocks = []Rock{{ID: 1, X: tc.x, Y: tc.y, VX: tc.vx, VY: tc.vy, Answer: "cat"}}
			g.updatePlay(Input{})
			if math.Abs(g.Rocks[0].X-tc.wantX) > 0.0001 || math.Abs(g.Rocks[0].Y-tc.wantY) > 0.0001 {
				t.Fatalf("got (%.2f, %.2f), want (%.2f, %.2f)", g.Rocks[0].X, g.Rocks[0].Y, tc.wantX, tc.wantY)
			}
		})
	}
}

func TestFragmentationAndScoring(t *testing.T) {
	g := newTestGame(3, Typing)
	g.Rocks = []Rock{{ID: 7, Size: Large, X: 100, Y: 100, VX: 1, Display: "cat", Answer: "cat"}}
	g.nextID = 8
	g.hitRock(7)
	if len(g.Rocks) != 2 {
		t.Fatalf("large split produced %d children", len(g.Rocks))
	}
	if g.Score != 20 || g.Solved != 1 {
		t.Fatalf("score/solved = %d/%d", g.Score, g.Solved)
	}
	for _, r := range g.Rocks {
		if r.Size != Medium {
			t.Fatalf("child size = %v", r.Size)
		}
		if math.Hypot(r.VX, r.VY) <= 1 {
			t.Fatalf("fragment not faster: %.3f", math.Hypot(r.VX, r.VY))
		}
		if r.Answer == "" {
			t.Fatal("child missing challenge")
		}
	}

	first := g.Rocks[0].ID
	g.hitRock(first)
	mediumChildren := 0
	for _, r := range g.Rocks {
		if r.Size == Small {
			mediumChildren++
		}
	}
	if mediumChildren != 2 {
		t.Fatalf("medium produced %d small children", mediumChildren)
	}

	smallID := -1
	for _, r := range g.Rocks {
		if r.Size == Small {
			smallID = r.ID
			break
		}
	}
	before := len(g.Rocks)
	g.hitRock(smallID)
	if len(g.Rocks) != before-1 {
		t.Fatal("small rock was not removed")
	}
}

func TestStationCollisionCostsOneShieldAndRemovesRock(t *testing.T) {
	g := newTestGame(4, Typing)
	g.Rocks = []Rock{{ID: 1, Size: Small, X: LogicalWidth / 2, Y: LogicalHeight / 2, Answer: "cat"}}
	g.updatePlay(Input{})
	if g.Shield != 2 {
		t.Fatalf("shield = %d", g.Shield)
	}
	if len(g.Rocks) == 0 {
		return
	}
	for _, r := range g.Rocks {
		if r.ID == 1 {
			t.Fatal("colliding rock remained")
		}
	}
}

func TestThirdStationHitEndsRun(t *testing.T) {
	g := newTestGame(5, Typing)
	g.Shield = 1
	g.Rocks = []Rock{{ID: 1, Size: Small, X: LogicalWidth / 2, Y: LogicalHeight / 2, Answer: "cat"}}
	g.updatePlay(Input{})
	if g.Shield != 0 {
		t.Fatalf("shield = %d", g.Shield)
	}
	if g.Screen != Play {
		t.Fatal("game-over should be a presentation delay, not a separate screen")
	}
	for i := 0; i < 90; i++ {
		g.updatePlay(Input{})
	}
	if g.Screen != Initials {
		t.Fatalf("qualifying score should enter initials, got screen %v", g.Screen)
	}
}

func TestPlayerCanSpellAnyVisibleAsteroid(t *testing.T) {
	g := newTestGame(6, Typing)
	g.Rocks = []Rock{
		{ID: 1, Size: Small, X: LogicalWidth/2 + 150, Y: LogicalHeight / 2, Display: "cat", Answer: "cat"},
		{ID: 2, Size: Small, X: 80, Y: 80, Display: "car", Answer: "car"},
	}
	g.readInput([]rune{'c', 'a', 'r'})
	if g.Errors != 0 {
		t.Fatalf("valid far-away word counted as error: %d", g.Errors)
	}
	if g.pendingHitID != 2 {
		t.Fatalf("pending hit = %d, want far-away asteroid 2", g.pendingHitID)
	}
	g.updatePlay(Input{})
	for _, r := range g.Rocks {
		if r.ID == 2 {
			t.Fatal("far-away asteroid remained after spelling its word")
		}
	}
}

func TestPlayerCanSolveAnyVisibleMathsAsteroid(t *testing.T) {
	g := newTestGame(611, Maths)
	g.Wave = 11
	g.Rocks = []Rock{
		{ID: 1, Size: Small, X: LogicalWidth/2 + 150, Y: LogicalHeight / 2, Display: "12 x 10", Answer: "120"},
		{ID: 2, Size: Small, X: 80, Y: 80, Display: "12 x 11", Answer: "132"},
	}
	g.readInput([]rune{'1', '3', '2'})
	if g.Errors != 0 {
		t.Fatalf("valid far-away answer counted as error: %d", g.Errors)
	}
	if g.pendingHitID != 2 {
		t.Fatalf("pending hit = %d, want far-away asteroid 2", g.pendingHitID)
	}
}

func TestTypingCanSwitchMidAnswer(t *testing.T) {
	g := newTestGame(613, Typing)
	g.Rocks = []Rock{{ID: 1, Size: Small, X: 120, Y: 120, Display: "cow", Answer: "cow"}, {ID: 2, Size: Small, X: 700, Y: 520, Display: "ant", Answer: "ant"}}
	g.readInput([]rune{'c', 'o', 'a', 'n', 't'})
	if g.Errors != 0 { t.Fatalf("switching to another visible word counted as error: %d", g.Errors) }
	if g.pendingHitID != 2 { t.Fatalf("pending hit = %d, want asteroid 2", g.pendingHitID) }
	if g.Typed != "ant" { t.Fatalf("typed prefix = %q, want ant", g.Typed) }
}

func TestMathsCanSwitchMidAnswer(t *testing.T) {
	g := newTestGame(614, Maths)
	g.Rocks = []Rock{{ID: 1, Size: Small, X: 120, Y: 120, Display: "7 x 6", Answer: "42"}, {ID: 2, Size: Small, X: 700, Y: 520, Display: "8 x 8", Answer: "64"}}
	g.readInput([]rune{'4', '6', '4'})
	if g.Errors != 0 { t.Fatalf("switching to another visible maths answer counted as error: %d", g.Errors) }
	if g.pendingHitID != 2 { t.Fatalf("pending hit = %d, want asteroid 2", g.pendingHitID) }
	if g.Typed != "64" { t.Fatalf("typed prefix = %q, want 64", g.Typed) }
}

func TestNonQualifyingGameOverReturnsToAttract(t *testing.T) {
	scores := HighScores{Typing: []ScoreEntry{{Initials: "AAA", Score: 500}, {Initials: "BBB", Score: 400}, {Initials: "CCC", Score: 300}, {Initials: "DDD", Score: 200}, {Initials: "EEE", Score: 100}}}
	g := New(615, scores)
	g.startRun(Typing)
	g.Score = 20
	g.Shield = 0
	for i := 0; i < 90; i++ { g.updatePlay(Input{}) }
	if g.Screen != Attract { t.Fatalf("non-qualifying game over should return to attract, got %v", g.Screen) }
}

func TestQualifyingGameOverInitialsReturnToAttract(t *testing.T) {
	g := New(616, HighScores{})
	g.startRun(Maths)
	g.Score = 120
	g.Shield = 0
	for i := 0; i < 90; i++ { g.updatePlay(Input{}) }
	if g.Screen != Initials { t.Fatalf("qualifying game over should enter initials, got %v", g.Screen) }
	g.Update(Input{Chars: []rune{'j', 's', 'l'}})
	if g.Screen != Attract { t.Fatalf("three initials should return to attract, got %v", g.Screen) }
	if len(g.Scores.Maths) != 1 || g.Scores.Maths[0].Initials != "JSL" || g.Scores.Maths[0].Score != 120 { t.Fatalf("score not recorded correctly: %#v", g.Scores.Maths) }
}

func TestMathsChallengeAvoidsActivePrefixAmbiguity(t *testing.T) {
	g := newTestGame(612, Maths)
	g.Wave = 11
	g.Rocks = []Rock{{ID: 1, Display: "12 x 1", Answer: "12"}}
	for i := 0; i < 100; i++ {
		_, answer := g.newChallenge()
		if answer != "12" && (len(answer) >= 2 && answer[:2] == "12") { t.Fatalf("generated prefix-conflicting answer %q while 12 was active", answer) }
	}
}

func TestWrongInputKeepsProgress(t *testing.T) {
	g := newTestGame(7, Typing)
	g.Rocks = []Rock{{ID: 1, X: 100, Y: 100, Answer: "cat"}}
	g.readInput([]rune{'c', 'x'})
	if g.Typed != "c" { t.Fatalf("wrong input reset progress: %q", g.Typed) }
	if g.Errors != 1 || g.Score != 0 || g.Shield != InitialShield { t.Fatalf("wrong-input effects errors=%d score=%d shield=%d", g.Errors, g.Score, g.Shield) }
}

func TestCorrectSequenceHitsTarget(t *testing.T) {
	g := newTestGame(8, Typing)
	g.Rocks = []Rock{{ID: 1, Size: Small, X: 100, Y: 100, Answer: "cat", Display: "cat"}}
	g.updatePlay(Input{Chars: []rune{'c', 'a', 't'}})
	for _, r := range g.Rocks { if r.ID == 1 { t.Fatal("completed target remained") } }
	if g.Score != 100 || g.Typed != "" { t.Fatalf("score/typed = %d/%q", g.Score, g.Typed) }
}

func TestCollisionClearsUnmatchedTypedPrefix(t *testing.T) {
	g := newTestGame(9, Typing)
	g.Rocks = []Rock{{ID: 1, Size: Small, X: LogicalWidth / 2, Y: LogicalHeight / 2, Answer: "cat"}}
	g.Typed = "c"
	g.updatePlay(Input{})
	if g.Typed != "" { t.Fatalf("typed prefix not cleared after its only matching asteroid collided: %q", g.Typed) }
}

func TestWaveAdvancesWhenFieldCleared(t *testing.T) {
	g := newTestGame(10, Typing)
	g.Rocks = []Rock{{ID: 1, Size: Small, X: 100, Y: 100, Answer: "a"}}
	g.updatePlay(Input{Chars: []rune{'a'}})
	if g.Wave != 2 { t.Fatalf("wave = %d", g.Wave) }
	if len(g.Rocks) != startingLargeCount(2) { t.Fatalf("next wave rocks = %d", len(g.Rocks)) }
}

func TestMathsWaveProgression(t *testing.T) {
	for wave := 1; wave <= 11; wave++ {
		g := newTestGame(int64(100+wave), Maths)
		g.Wave = wave
		for i := 0; i < 100; i++ {
			display, answer := g.newChallenge()
			table := wave + 1
			var a, b int
			if _, err := fmtSscanf(display, &a, &b); err != nil { t.Fatalf("parse %q: %v", display, err) }
			if a != table || b < 1 || b > 12 { t.Fatalf("wave %d generated %q", wave, display) }
			if answer != intString(a*b) { t.Fatalf("wrong answer for %q: %q", display, answer) }
		}
	}
}

func TestMathsMixedWaveRange(t *testing.T) {
	g := newTestGame(222, Maths)
	g.Wave = 12
	for i := 0; i < 500; i++ {
		display, answer := g.newChallenge()
		var a, b int
		if _, err := fmtSscanf(display, &a, &b); err != nil { t.Fatalf("parse %q: %v", display, err) }
		if a < 2 || a > 12 || b < 1 || b > 12 { t.Fatalf("out-of-range mixed question: %q", display) }
		if answer != intString(a*b) { t.Fatalf("wrong answer for %q", display) }
	}
}

func TestDeterministicGameplay(t *testing.T) {
	a := New(123456, HighScores{})
	b := New(123456, HighScores{})
	inputs := []Input{{StartTyping: true}, {}, {}, {}, {Chars: []rune{'c'}}, {}, {Chars: []rune{'a'}}, {}, {Chars: []rune{'t'}}}
	for i := 0; i < 300; i++ {
		in := Input{}
		if i < len(inputs) { in = inputs[i] }
		a.Update(in); b.Update(in)
	}
	if !reflect.DeepEqual(gameplaySnapshot(a), gameplaySnapshot(b)) { t.Fatalf("same seed + input + ticks diverged\nA=%#v\nB=%#v", gameplaySnapshot(a), gameplaySnapshot(b)) }
}

func TestCosmeticRandomnessDoesNotChangeGameplay(t *testing.T) {
	a := New(333, HighScores{}); b := New(333, HighScores{})
	a.startRun(Typing); b.startRun(Typing)
	for i := 0; i < 100; i++ { a.addBurst(10, 10, 0, 8); b.addBurst(10, 10, 0, 1) }
	a.Rocks = nil; b.Rocks = nil; a.Wave = 4; b.Wave = 4; a.spawnWave(); b.spawnWave()
	if !reflect.DeepEqual(gameplaySnapshot(a), gameplaySnapshot(b)) { t.Fatal("cosmetic RNG altered gameplay") }
}

func TestDifficultyFunctionsAreBounded(t *testing.T) {
	lastCount := 0; lastSpeed := 0.0
	for wave := 1; wave <= 1000; wave++ {
		count := startingLargeCount(wave); speed := baseSpeed(wave)
		if count < lastCount || speed < lastSpeed { t.Fatalf("difficulty decreased at wave %d", wave) }
		if count*4 > MaxRocks { t.Fatalf("wave %d can exceed rock ceiling", wave) }
		lastCount, lastSpeed = count, speed
	}
}

func gameplaySnapshot(g *Game) any {
	return struct { Screen Screen; Mode Mode; Wave int; Score int; Shield int; Rocks []Rock; Typed string; Correct int; Errors int; Solved int; Ticks int }{g.Screen, g.Mode, g.Wave, g.Score, g.Shield, append([]Rock(nil), g.Rocks...), g.Typed, g.Correct, g.Errors, g.Solved, g.Ticks}
}

func TestAttractStartsEitherMode(t *testing.T) {
	typing := New(41, HighScores{}); typing.Update(Input{StartTyping: true})
	if typing.Screen != Play || typing.Mode != Typing || typing.Wave != 1 || typing.Shield != InitialShield || len(typing.Rocks) == 0 { t.Fatalf("typing start state incorrect: screen=%v mode=%v wave=%d shield=%d rocks=%d", typing.Screen, typing.Mode, typing.Wave, typing.Shield, len(typing.Rocks)) }
	maths := New(42, HighScores{}); maths.Update(Input{StartMaths: true})
	if maths.Screen != Play || maths.Mode != Maths || maths.Wave != 1 || maths.Shield != InitialShield || len(maths.Rocks) == 0 { t.Fatalf("maths start state incorrect: screen=%v mode=%v wave=%d shield=%d rocks=%d", maths.Screen, maths.Mode, maths.Wave, maths.Shield, len(maths.Rocks)) }
}

func TestInitialsBackspaceAndLetterFilter(t *testing.T) {
	g := New(43, HighScores{}); g.Screen = Initials; g.Score = 100
	g.Update(Input{Chars: []rune{'a', '1', '!'}})
	if g.Initials != "A" { t.Fatalf("initials = %q", g.Initials) }
	g.Update(Input{Backspace: true})
	if g.Initials != "" { t.Fatalf("backspace left %q", g.Initials) }
}

func TestStatsDeriveFromRunCounters(t *testing.T) {
	g := New(44, HighScores{}); g.Mode = Typing; g.Correct = 45; g.Errors = 5; g.Solved = 10; g.Ticks = 60 * 60
	s := g.Stats()
	if math.Abs(s.Accuracy-0.9) > 0.0001 { t.Fatalf("accuracy = %f", s.Accuracy) }
	if math.Abs(s.WordsPerMinute-9) > 0.0001 { t.Fatalf("wpm = %f", s.WordsPerMinute) }
	if math.Abs(s.AverageResponseTime-6) > 0.0001 { t.Fatalf("average response = %f", s.AverageResponseTime) }
	g.Mode = Maths; g.Wave = 12; s = g.Stats()
	if s.HighestTable != 12 { t.Fatalf("highest table = %d", s.HighestTable) }
}
