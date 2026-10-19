package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ScoreEntry struct {
	Initials string `json:"initials"`
	Score    int    `json:"score"`
}

type HighScores struct {
	Typing []ScoreEntry `json:"typing"`
	Maths  []ScoreEntry `json:"maths"`
}

func LoadScores(path string) (HighScores, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return HighScores{}, nil
	}
	if err != nil {
		return HighScores{}, err
	}
	var scores HighScores
	if err := json.Unmarshal(data, &scores); err != nil {
		return HighScores{}, fmt.Errorf("decode high scores: %w", err)
	}
	return scores.Normalized(), nil
}

func SaveScores(path string, scores HighScores) error {
	scores = scores.Normalized()
	data, err := json.MarshalIndent(scores, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func (h HighScores) Normalized() HighScores {
	h.Typing = normalizeTable(h.Typing)
	h.Maths = normalizeTable(h.Maths)
	return h
}

func (h HighScores) table(mode Mode) []ScoreEntry {
	if mode == Maths {
		return h.Maths
	}
	return h.Typing
}

func (h *HighScores) Add(mode Mode, entry ScoreEntry) {
	entry.Initials = strings.ToUpper(entry.Initials)
	if mode == Maths {
		h.Maths = normalizeTable(append(h.Maths, entry))
		return
	}
	h.Typing = normalizeTable(append(h.Typing, entry))
}

func (h HighScores) Qualifies(mode Mode, score int) bool {
	table := h.table(mode)
	if len(table) < 5 {
		return true
	}
	return score > table[len(table)-1].Score
}

func normalizeTable(in []ScoreEntry) []ScoreEntry {
	out := make([]ScoreEntry, 0, len(in))
	for _, e := range in {
		e.Initials = strings.ToUpper(strings.TrimSpace(e.Initials))
		if len(e.Initials) != 3 || e.Score < 0 {
			continue
		}
		out = append(out, e)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	if len(out) > 5 {
		out = out[:5]
	}
	return out
}
