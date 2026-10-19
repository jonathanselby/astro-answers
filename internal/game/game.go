package game

import (
	_ "embed"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"unicode"
)

const (
	LogicalWidth   = 960
	LogicalHeight  = 720
	TicksPerSecond = 60

	InitialShield = 3
	MaxRocks      = 40
	MaxParticles  = 256
	StationRadius = 56.0
)

type Screen uint8

const (
	Attract Screen = iota
	Play
	Initials
)

type Mode uint8

const (
	Typing Mode = iota
	Maths
)

type Size uint8

const (
	Large Size = iota
	Medium
	Small
)

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

func (r Rock) Radius() float64 {
	switch r.Size {
	case Large:
		return 50
	case Medium:
		return 34
	default:
		return 22
	}
}

type Particle struct {
	X, Y   float64
	VX, VY float64
	Life   int
	Kind   uint8
}

type Input struct {
	Chars       []rune
	Backspace   bool
	StartTyping bool
	StartMaths  bool
}

type RunStats struct {
	Accuracy            float64
	WordsPerMinute      float64
	AverageResponseTime float64
	HighestTable        int
}

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

	RNG         *rand.Rand
	CosmeticRNG *rand.Rand

	Particles []Particle
	Scores    HighScores
	Initials  string

	WaveNotice     int
	GameOverNotice int

	nextID        int
	pendingHitID  int
	lastChallenge string
}

//go:embed words.txt
var wordData string

var words = strings.Fields(wordData)

func New(seed int64, scores HighScores) *Game {
	return &Game{
		Screen:      Attract,
		Shield:      InitialShield,
		RNG:         rand.New(rand.NewSource(seed)),
		CosmeticRNG: rand.New(rand.NewSource(seed ^ 0x5deece66d)),
		Scores:      scores.Normalized(),
	}
}

func (g *Game) Update(in Input) {
	switch g.Screen {
	case Attract:
		if in.StartTyping {
			g.startRun(Typing)
			return
		}
		if in.StartMaths {
			g.startRun(Maths)
			return
		}
	case Initials:
		g.updateInitials(in)
	case Play:
		g.updatePlay(in)
	}
}

func (g *Game) startRun(mode Mode) {
	g.Screen = Play
	g.Mode = mode
	g.Wave = 1
	g.Score = 0
	g.Shield = InitialShield
	g.Rocks = g.Rocks[:0]
	g.Typed = ""
	g.Correct = 0
	g.Errors = 0
	g.Solved = 0
	g.Ticks = 0
	g.Particles = g.Particles[:0]
	g.Initials = ""
	g.pendingHitID = -1
	g.GameOverNotice = 0
	g.WaveNotice = 0
	g.spawnWave()
}

func (g *Game) updatePlay(in Input) {
	if g.Shield <= 0 {
		g.GameOverNotice++
		g.updateParticles()
		if g.GameOverNotice >= 90 {
			g.finishRun()
		}
		return
	}

	g.Ticks++
	if g.WaveNotice > 0 {
		g.WaveNotice--
	}

	g.readInput(in.Chars)
	g.moveRocks()
	g.wrapRocks()
	g.resolveStationCollisions()

	if g.Shield <= 0 {
		g.pendingHitID = -1
		g.Typed = ""
		g.updateParticles()
		return
	}

	if g.pendingHitID >= 0 {
		g.hitRock(g.pendingHitID)
		g.pendingHitID = -1
	}

	if len(g.Rocks) == 0 {
		g.Wave++
		g.spawnWave()
	}

	g.updateParticles()
}

func (g *Game) finishRun() {
	g.GameOverNotice = 0
	if g.Scores.Qualifies(g.Mode, g.Score) {
		g.Screen = Initials
		g.Initials = ""
		return
	}
	g.Screen = Attract
}

func (g *Game) updateInitials(in Input) {
	if in.Backspace && len(g.Initials) > 0 {
		g.Initials = g.Initials[:len(g.Initials)-1]
	}
	for _, r := range in.Chars {
		if !unicode.IsLetter(r) || len(g.Initials) >= 3 {
			continue
		}
		g.Initials += strings.ToUpper(string(r))
		if len(g.Initials) == 3 {
			g.Scores.Add(g.Mode, ScoreEntry{Initials: g.Initials, Score: g.Score})
			g.Screen = Attract
			return
		}
	}
}

func (g *Game) readInput(chars []rune) {
	for _, r := range chars {
		if g.pendingHitID >= 0 {
			return
		}
		ch := normaliseInput(r)
		if ch == 0 {
			continue
		}

		// Keep extending the current answer while it still matches a visible
		// asteroid. If it does not, treat the newest key as the beginning of a
		// different visible answer. This lets the player change their mind at any
		// time without adding a target-selection control.
		prefix := g.Typed + string(ch)
		if !g.hasAnswerPrefix(prefix) {
			prefix = string(ch)
			if !g.hasAnswerPrefix(prefix) {
				g.Errors++
				g.addFeedbackParticle(2)
				continue
			}
		}

		g.Typed = prefix
		g.Correct++
		g.addFeedbackParticle(1)

		if id, ok := g.exactAnswerID(prefix); ok {
			g.pendingHitID = id
			return
		}
	}
}

func normaliseInput(r rune) rune {
	r = unicode.ToLower(r)
	if r >= 'a' && r <= 'z' {
		return r
	}
	if r >= '0' && r <= '9' {
		return r
	}
	return 0
}

func (g *Game) hasAnswerPrefix(prefix string) bool {
	for _, r := range g.Rocks {
		if strings.HasPrefix(r.Answer, prefix) {
			return true
		}
	}
	return false
}

func (g *Game) exactAnswerID(answer string) (int, bool) {
	bestID := 0
	found := false
	for _, r := range g.Rocks {
		if r.Answer != answer {
			continue
		}
		if !found || r.ID < bestID {
			bestID = r.ID
			found = true
		}
	}
	return bestID, found
}

func (g *Game) moveRocks() {
	for i := range g.Rocks {
		g.Rocks[i].X += g.Rocks[i].VX
		g.Rocks[i].Y += g.Rocks[i].VY
		g.Rocks[i].Angle += g.Rocks[i].Spin
	}
}

func (g *Game) wrapRocks() {
	for i := range g.Rocks {
		if g.Rocks[i].X < 0 {
			g.Rocks[i].X += LogicalWidth
		}
		if g.Rocks[i].X >= LogicalWidth {
			g.Rocks[i].X -= LogicalWidth
		}
		if g.Rocks[i].Y < 0 {
			g.Rocks[i].Y += LogicalHeight
		}
		if g.Rocks[i].Y >= LogicalHeight {
			g.Rocks[i].Y -= LogicalHeight
		}
	}
}

func (g *Game) resolveStationCollisions() {
	cx, cy := LogicalWidth/2.0, LogicalHeight/2.0
	for i := len(g.Rocks) - 1; i >= 0; i-- {
		r := g.Rocks[i]
		dx, dy := r.X-cx, r.Y-cy
		limit := StationRadius + r.Radius()
		if dx*dx+dy*dy >= limit*limit {
			continue
		}

		g.addBurst(r.X, r.Y, 3, 8)
		if r.ID == g.pendingHitID {
			g.pendingHitID = -1
		}
		g.removeRockAt(i)
		g.Shield--
		if g.Shield <= 0 {
			return
		}
	}
	if g.Typed != "" && !g.hasAnswerPrefix(g.Typed) {
		g.Typed = ""
	}
}

func (g *Game) hitRock(id int) {
	idx := g.indexByID(id)
	if idx < 0 {
		return
	}
	parent := g.Rocks[idx]
	g.removeRockAt(idx)
	g.Score += points(parent.Size)
	g.Solved++
	g.Typed = ""
	g.addBurst(parent.X, parent.Y, 0, 6)

	switch parent.Size {
	case Large:
		g.split(parent, Medium)
	case Medium:
		g.split(parent, Small)
	}
}

func (g *Game) split(parent Rock, childSize Size) {
	if len(g.Rocks)+2 > MaxRocks {
		panic("gameplay rock ceiling exceeded")
	}
	base := math.Atan2(parent.VY, parent.VX)
	parentSpeed := math.Hypot(parent.VX, parent.VY)
	childSpeed := parentSpeed * 1.20
	if childSpeed == 0 {
		childSpeed = baseSpeed(g.Wave) * 1.20
	}
	for _, sign := range []float64{-1, 1} {
		diverge := (0.24 + g.RNG.Float64()*0.12) * sign
		angle := base + diverge
		display, answer := g.newChallenge()
		child := Rock{
			ID:      g.takeID(),
			Size:    childSize,
			X:       parent.X,
			Y:       parent.Y,
			VX:      math.Cos(angle) * childSpeed,
			VY:      math.Sin(angle) * childSpeed,
			Angle:   parent.Angle,
			Spin:    randomSpin(g.RNG),
			Display: display,
			Answer:  answer,
		}
		g.Rocks = append(g.Rocks, child)
	}
}

func (g *Game) spawnWave() {
	count := startingLargeCount(g.Wave)
	if count*4 > MaxRocks {
		panic("wave cannot fit gameplay rock ceiling")
	}
	for i := 0; i < count; i++ {
		x, y := g.spawnPosition()
		angle := g.RNG.Float64() * 2 * math.Pi
		speed := baseSpeed(g.Wave) * (0.90 + g.RNG.Float64()*0.20)
		display, answer := g.newChallenge()
		g.Rocks = append(g.Rocks, Rock{
			ID:      g.takeID(),
			Size:    Large,
			X:       x,
			Y:       y,
			VX:      math.Cos(angle) * speed,
			VY:      math.Sin(angle) * speed,
			Angle:   g.RNG.Float64() * 2 * math.Pi,
			Spin:    randomSpin(g.RNG),
			Display: display,
			Answer:  answer,
		})
	}
	g.WaveNotice = 60
}

func (g *Game) spawnPosition() (float64, float64) {
	cx, cy := LogicalWidth/2.0, LogicalHeight/2.0
	const safeRadius = 230.0
	for attempt := 0; attempt < 32; attempt++ {
		x := g.RNG.Float64() * LogicalWidth
		y := g.RNG.Float64() * LogicalHeight
		dx, dy := x-cx, y-cy
		if dx*dx+dy*dy >= safeRadius*safeRadius {
			return x, y
		}
	}
	return 80, 80
}

func (g *Game) newChallenge() (string, string) {
	// A player chooses a rock only by typing its visible answer. Distinct active
	// answers therefore must not be prefixes of one another: e.g. "12" and
	// "120" cannot coexist, because no input-only control could distinguish
	// which rock the player intended when "12" was completed. Identical maths
	// answers are allowed; completing that answer removes one deterministically.
	for attempt := 0; attempt < 64; attempt++ {
		display, answer := g.rawChallenge()
		if g.answerCompatible(answer) {
			return display, answer
		}
	}

	// With an already-populated field there is always a safe fallback: reusing
	// an existing maths answer cannot create a new prefix ambiguity. Typing mode
	// has enough content that the bounded search above should never reach here.
	if g.Mode == Maths && len(g.Rocks) > 0 {
		r := g.Rocks[g.RNG.Intn(len(g.Rocks))]
		return r.Display, r.Answer
	}
	return g.rawChallenge()
}

func (g *Game) answerCompatible(answer string) bool {
	for _, r := range g.Rocks {
		if r.Answer == answer {
			if g.Mode == Typing {
				return false
			}
			continue
		}
		if strings.HasPrefix(r.Answer, answer) || strings.HasPrefix(answer, r.Answer) {
			return false
		}
	}
	return true
}

func (g *Game) rawChallenge() (string, string) {
	if g.Mode == Typing {
		idx := g.RNG.Intn(len(words))
		word := words[idx]
		if len(words) > 1 && word == g.lastChallenge {
			word = words[(idx+1+g.RNG.Intn(len(words)-1))%len(words)]
		}
		g.lastChallenge = word
		return word, word
	}

	table := g.Wave + 1
	if table > 12 {
		table = 2 + g.RNG.Intn(11)
	}
	other := 1 + g.RNG.Intn(12)
	answer := fmt.Sprintf("%d", table*other)
	display := fmt.Sprintf("%d x %d", table, other)
	g.lastChallenge = display
	return display, answer
}

func (g *Game) indexByID(id int) int {
	for i := range g.Rocks {
		if g.Rocks[i].ID == id {
			return i
		}
	}
	return -1
}

func (g *Game) removeRockAt(i int) {
	copy(g.Rocks[i:], g.Rocks[i+1:])
	g.Rocks = g.Rocks[:len(g.Rocks)-1]
}

func (g *Game) takeID() int {
	id := g.nextID
	g.nextID++
	return id
}

func startingLargeCount(wave int) int {
	count := 2 + (wave-1)/3
	if count > 8 {
		return 8
	}
	return count
}

func baseSpeed(wave int) float64 {
	speed := 0.85 + float64(wave-1)*0.05
	if speed > 1.75 {
		return 1.75
	}
	return speed
}

func randomSpin(rng *rand.Rand) float64 {
	spin := 0.004 + rng.Float64()*0.010
	if rng.Intn(2) == 0 {
		spin = -spin
	}
	return spin
}

func points(size Size) int {
	switch size {
	case Large:
		return 20
	case Medium:
		return 50
	default:
		return 100
	}
}

func (g *Game) addFeedbackParticle(kind uint8) {
	g.addBurst(LogicalWidth/2, LogicalHeight/2, kind, 2)
}

func (g *Game) addBurst(x, y float64, kind uint8, count int) {
	for i := 0; i < count && len(g.Particles) < MaxParticles; i++ {
		angle := g.CosmeticRNG.Float64() * 2 * math.Pi
		speed := 0.4 + g.CosmeticRNG.Float64()*1.6
		g.Particles = append(g.Particles, Particle{
			X:    x,
			Y:    y,
			VX:   math.Cos(angle) * speed,
			VY:   math.Sin(angle) * speed,
			Life: 12 + g.CosmeticRNG.Intn(24),
			Kind: kind,
		})
	}
}

func (g *Game) updateParticles() {
	out := g.Particles[:0]
	for _, p := range g.Particles {
		p.X += p.VX
		p.Y += p.VY
		p.Life--
		if p.Life > 0 {
			out = append(out, p)
		}
	}
	g.Particles = out
}

func (g *Game) Stats() RunStats {
	attempts := g.Correct + g.Errors
	accuracy := 0.0
	if attempts > 0 {
		accuracy = float64(g.Correct) / float64(attempts)
	}
	minutes := float64(g.Ticks) / TicksPerSecond / 60
	wpm := 0.0
	if minutes > 0 {
		wpm = (float64(g.Correct) / 5.0) / minutes
	}
	avg := 0.0
	if g.Solved > 0 {
		avg = (float64(g.Ticks) / TicksPerSecond) / float64(g.Solved)
	}
	highest := 0
	if g.Mode == Maths {
		highest = g.Wave + 1
		if highest > 12 {
			highest = 12
		}
	}
	return RunStats{
		Accuracy:            accuracy,
		WordsPerMinute:      wpm,
		AverageResponseTime: avg,
		HighestTable:        highest,
	}
}
