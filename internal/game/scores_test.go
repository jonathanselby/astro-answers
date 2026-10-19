package game

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestScoresIndependentSortedAndFiveOnly(t *testing.T) {
	h := HighScores{}
	for _, s := range []int{20, 100, 40, 80, 60, 120} {
		h.Add(Typing, ScoreEntry{Initials: "abc", Score: s})
	}
	h.Add(Maths, ScoreEntry{Initials: "xyz", Score: 77})
	if len(h.Typing) != 5 {
		t.Fatalf("typing entries = %d", len(h.Typing))
	}
	want := []int{120, 100, 80, 60, 40}
	for i, score := range want {
		if h.Typing[i].Score != score {
			t.Fatalf("typing[%d] = %d", i, h.Typing[i].Score)
		}
	}
	if len(h.Maths) != 1 || h.Maths[0].Score != 77 {
		t.Fatal("maths table was not independent")
	}
}

func TestScoreQualification(t *testing.T) {
	h := HighScores{Typing: []ScoreEntry{
		{"AAA", 100}, {"BBB", 90}, {"CCC", 80}, {"DDD", 70}, {"EEE", 60},
	}}
	if !h.Qualifies(Typing, 61) {
		t.Fatal("61 should qualify")
	}
	if h.Qualifies(Typing, 60) {
		t.Fatal("tie with fifth should not qualify")
	}
	if !h.Qualifies(Maths, 0) {
		t.Fatal("empty table should accept a score")
	}
}

func TestScoreSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scores.json")
	want := HighScores{
		Typing: []ScoreEntry{{"AAA", 100}, {"BBB", 50}},
		Maths:  []ScoreEntry{{"CCC", 75}},
	}
	if err := SaveScores(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadScores(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestMissingScoresStartEmpty(t *testing.T) {
	got, err := LoadScores(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Typing) != 0 || len(got.Maths) != 0 {
		t.Fatal("missing score file did not start empty")
	}
}

func TestMalformedScoresReportAndStartEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scores.json")
	if err := os.WriteFile(path, []byte("not-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadScores(path)
	if err == nil {
		t.Fatal("malformed score file should report an error")
	}
	if len(got.Typing) != 0 || len(got.Maths) != 0 {
		t.Fatal("malformed score file should return empty scores")
	}
}

func TestInitialEntryAddsHighScore(t *testing.T) {
	g := New(1, HighScores{})
	g.Screen = Initials
	g.Mode = Typing
	g.Score = 321
	g.Update(Input{Chars: []rune{'j', 'o', 'n'}})
	if g.Screen != Attract {
		t.Fatalf("screen = %v", g.Screen)
	}
	if len(g.Scores.Typing) != 1 || g.Scores.Typing[0].Initials != "JON" || g.Scores.Typing[0].Score != 321 {
		t.Fatalf("scores = %#v", g.Scores.Typing)
	}
}
