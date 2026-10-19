package main

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"io"
	"log"
	"math"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/jonathanselby/astro-answers/internal/game"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
)

//go:embed assets/*.png assets/*.wav
var assetFS embed.FS

const (
	scoreFile       = "scores.json"
	audioSampleRate = 44100
)

type sounds struct {
	accepted      []byte
	incorrect     []byte
	largeSplit    []byte
	mediumSplit   []byte
	smallDestroy  []byte
	stationImpact []byte
	newWave       []byte
	gameOver      []byte
	highScore     []byte
}

type app struct {
	game    *game.Game
	large   *ebiten.Image
	medium  *ebiten.Image
	small   *ebiten.Image
	station *ebiten.Image

	audio  *audio.Context
	sounds sounds

	titleFace   font.Face
	headingFace font.Face
	labelFace   font.Face
}

func main() {
	scores, err := game.LoadScores(scoreFile)
	if err != nil {
		log.Printf("high scores: %v; starting empty", err)
	}

	a := &app{
		game:  game.New(time.Now().UnixNano(), scores),
		audio: audio.NewContext(audioSampleRate),
	}
	if err := a.loadAssets(); err != nil {
		log.Fatal(err)
	}
	if err := a.loadFonts(); err != nil {
		log.Fatal(err)
	}
	a.loadSounds()

	ebiten.SetWindowSize(game.LogicalWidth, game.LogicalHeight)
	ebiten.SetWindowResizable(false)
	ebiten.SetWindowTitle("Astro Answers")
	ebiten.SetFullscreen(true)
	if err := ebiten.RunGame(a); err != nil {
		log.Fatal(err)
	}
}

func (a *app) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	in := game.Input{
		StartTyping: inpututil.IsKeyJustPressed(ebiten.KeyT),
		StartMaths:  inpututil.IsKeyJustPressed(ebiten.KeyM),
		Backspace:   inpututil.IsKeyJustPressed(ebiten.KeyBackspace),
	}

	for key := ebiten.KeyA; key <= ebiten.KeyZ; key++ {
		if inpututil.IsKeyJustPressed(key) {
			in.Chars = append(in.Chars, rune('a'+(key-ebiten.KeyA)))
		}
	}
	for key := ebiten.Key0; key <= ebiten.Key9; key++ {
		if inpututil.IsKeyJustPressed(key) {
			in.Chars = append(in.Chars, rune('0'+(key-ebiten.Key0)))
		}
	}

	beforeScreen := a.game.Screen
	beforeCorrect := a.game.Correct
	beforeErrors := a.game.Errors
	beforeScore := a.game.Score
	beforeShield := a.game.Shield
	beforeWave := a.game.Wave

	a.game.Update(in)

	if a.game.Correct > beforeCorrect {
		a.play(a.sounds.accepted)
	}
	if a.game.Errors > beforeErrors {
		a.play(a.sounds.incorrect)
	}
	switch a.game.Score - beforeScore {
	case 20:
		a.play(a.sounds.largeSplit)
	case 50:
		a.play(a.sounds.mediumSplit)
	case 100:
		a.play(a.sounds.smallDestroy)
	}
	if a.game.Shield < beforeShield {
		a.play(a.sounds.stationImpact)
	}
	if (beforeScreen != game.Play && a.game.Screen == game.Play) || a.game.Wave > beforeWave {
		a.play(a.sounds.newWave)
	}
	if beforeShield > 0 && a.game.Shield <= 0 {
		a.play(a.sounds.gameOver)
	}
	if beforeScreen == game.Play && a.game.Screen == game.Initials {
		a.play(a.sounds.highScore)
	}

	if beforeScreen == game.Initials && a.game.Screen == game.Attract {
		if err := game.SaveScores(scoreFile, a.game.Scores); err != nil {
			log.Printf("save high scores: %v", err)
		}
	}
	return nil
}

func (a *app) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{6, 8, 14, 255})
	a.drawStars(screen)

	switch a.game.Screen {
	case game.Attract:
		a.drawAttract(screen)
	case game.Play:
		a.drawPlay(screen)
	case game.Initials:
		a.drawInitials(screen)
	}
}

func (a *app) Layout(_, _ int) (int, int) {
	// Gameplay always uses the same 4:3 logical playfield. Ebitengine's
	// final-screen geometry fits it to the physical display without
	// changing the aspect ratio.
	return game.LogicalWidth, game.LogicalHeight
}

func (a *app) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, geoM ebiten.GeoM) {
	// Fullscreen displays can be wider or taller than the 4:3 playfield.
	// Keep the unused area black instead of stretching the game.
	screen.Fill(color.RGBA{0, 0, 0, 255})
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM = geoM
	screen.DrawImage(offscreen, op)
}

func (a *app) loadAssets() error {
	var err error
	if a.large, err = loadImage("assets/asteroid_large.png"); err != nil {
		return err
	}
	if a.medium, err = loadImage("assets/asteroid_medium.png"); err != nil {
		return err
	}
	if a.small, err = loadImage("assets/asteroid_small.png"); err != nil {
		return err
	}
	if a.station, err = loadImage("assets/station.png"); err != nil {
		return err
	}
	return nil
}

func (a *app) loadFonts() error {
	parsed, err := opentype.Parse(gobold.TTF)
	if err != nil {
		return fmt.Errorf("parse built-in title font: %w", err)
	}
	newFace := func(size float64) (font.Face, error) {
		face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
			Size:    size,
			DPI:     72,
			Hinting: font.HintingFull,
		})
		if err != nil {
			return nil, err
		}
		return face, nil
	}
	if a.titleFace, err = newFace(54); err != nil {
		return fmt.Errorf("title font: %w", err)
	}
	if a.headingFace, err = newFace(24); err != nil {
		return fmt.Errorf("heading font: %w", err)
	}
	if a.labelFace, err = newFace(18); err != nil {
		return fmt.Errorf("label font: %w", err)
	}
	return nil
}

func (a *app) loadSounds() {
	var err error
	if a.sounds.accepted, err = loadSound("assets/accepted.wav"); err != nil {
		log.Printf("sound accepted: %v", err)
	}
	if a.sounds.incorrect, err = loadSound("assets/incorrect.wav"); err != nil {
		log.Printf("sound incorrect: %v", err)
	}
	if a.sounds.largeSplit, err = loadSound("assets/large_split.wav"); err != nil {
		log.Printf("sound large split: %v", err)
	}
	if a.sounds.mediumSplit, err = loadSound("assets/medium_split.wav"); err != nil {
		log.Printf("sound medium split: %v", err)
	}
	if a.sounds.smallDestroy, err = loadSound("assets/small_destroy.wav"); err != nil {
		log.Printf("sound small destroy: %v", err)
	}
	if a.sounds.stationImpact, err = loadSound("assets/station_impact.wav"); err != nil {
		log.Printf("sound station impact: %v", err)
	}
	if a.sounds.newWave, err = loadSound("assets/new_wave.wav"); err != nil {
		log.Printf("sound new wave: %v", err)
	}
	if a.sounds.gameOver, err = loadSound("assets/game_over.wav"); err != nil {
		log.Printf("sound game over: %v", err)
	}
	if a.sounds.highScore, err = loadSound("assets/high_score.wav"); err != nil {
		log.Printf("sound high score: %v", err)
	}
}

func loadSound(path string) ([]byte, error) {
	data, err := assetFS.ReadFile(path)
	if err != nil {
		return nil, err
	}
	stream, err := wav.DecodeWithoutResampling(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if stream.SampleRate() != audioSampleRate {
		return nil, fmt.Errorf("%s sample rate is %d, want %d", path, stream.SampleRate(), audioSampleRate)
	}
	pcm, err := io.ReadAll(stream)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return pcm, nil
}

func (a *app) play(pcm []byte) {
	if len(pcm) == 0 {
		return
	}
	a.audio.NewPlayerFromBytes(pcm).Play()
}

func loadImage(path string) (*ebiten.Image, error) {
	f, err := assetFS.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return ebiten.NewImageFromImage(img), nil
}

func (a *app) drawPlay(screen *ebiten.Image) {
	a.drawStation(screen)
	for _, rock := range a.game.Rocks {
		a.drawRockWrapped(screen, rock)
	}
	for _, p := range a.game.Particles {
		clr := color.RGBA{220, 220, 220, 220}
		if p.Kind == 1 {
			clr = color.RGBA{130, 220, 160, 220}
		} else if p.Kind == 2 {
			clr = color.RGBA{230, 120, 120, 220}
		} else if p.Kind == 3 {
			clr = color.RGBA{240, 180, 80, 230}
		}
		ebitenutil.DrawRect(screen, p.X-1, p.Y-1, 3, 3, clr)
	}

	a.drawHUD(screen)

	if a.game.WaveNotice > 0 {
		drawCentered(screen, fmt.Sprintf("WAVE %d", a.game.Wave), game.LogicalHeight/2-110)
	}
	if a.game.Shield <= 0 {
		a.drawGameOver(screen)
	}
}

func (a *app) drawGameOver(screen *ebiten.Image) {
	cyan := color.RGBA{92, 216, 238, 255}
	ice := color.RGBA{236, 248, 255, 255}
	textGlow := color.RGBA{64, 68, 74, 255}
	textOutline := color.RGBA{24, 27, 31, 230}

	// Darken the action but keep the finished run visible behind the result.
	ebitenutil.DrawRect(screen, 170, 165, 620, 390, color.RGBA{3, 6, 12, 235})
	drawPanelFrame(screen, 170, 165, 620, 390, cyan)
	drawStyledCentered(screen, a.titleFace, "GAME OVER", game.LogicalWidth/2, 240, ice, textGlow, textOutline)
	drawStyledCentered(screen, a.headingFace, fmt.Sprintf("SCORE %06d", a.game.Score), game.LogicalWidth/2, 290, ice, textGlow, textOutline)
	a.drawStatsStyled(screen, 360)

	if a.game.Scores.Qualifies(a.game.Mode, a.game.Score) {
		drawStyledCentered(screen, a.labelFace, "NEW HIGH SCORE", game.LogicalWidth/2, 475, ice, textGlow, textOutline)
		drawStyledCentered(screen, a.labelFace, "INITIALS NEXT", game.LogicalWidth/2, 510, color.RGBA{180, 202, 214, 255}, textGlow, textOutline)
	} else {
		drawStyledCentered(screen, a.labelFace, "RETURNING TO TITLE", game.LogicalWidth/2, 495, color.RGBA{180, 202, 214, 255}, textGlow, textOutline)
	}
}

func (a *app) drawStation(screen *ebiten.Image) {
	w, h := a.station.Bounds().Dx(), a.station.Bounds().Dy()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(-w/2), float64(-h/2))
	op.GeoM.Translate(game.LogicalWidth/2, game.LogicalHeight/2)
	screen.DrawImage(a.station, op)

	for i := 0; i < a.game.Shield; i++ {
		a := float64(i)*2*math.Pi/3 - math.Pi/2
		x := game.LogicalWidth/2 + math.Cos(a)*76
		y := game.LogicalHeight/2 + math.Sin(a)*76
		ebitenutil.DrawRect(screen, x-7, y-3, 14, 6, color.RGBA{130, 210, 240, 220})
	}
}

func (a *app) drawRockWrapped(screen *ebiten.Image, rock game.Rock) {
	r := rock.Radius()
	xs := []float64{rock.X}
	ys := []float64{rock.Y}
	if rock.X-r < 0 {
		xs = append(xs, rock.X+game.LogicalWidth)
	}
	if rock.X+r >= game.LogicalWidth {
		xs = append(xs, rock.X-game.LogicalWidth)
	}
	if rock.Y-r < 0 {
		ys = append(ys, rock.Y+game.LogicalHeight)
	}
	if rock.Y+r >= game.LogicalHeight {
		ys = append(ys, rock.Y-game.LogicalHeight)
	}
	for _, x := range xs {
		for _, y := range ys {
			a.drawRock(screen, rock, x, y)
		}
	}
}

func (a *app) drawRock(screen *ebiten.Image, rock game.Rock, x, y float64) {
	img := a.small
	if rock.Size == game.Large {
		img = a.large
	} else if rock.Size == game.Medium {
		img = a.medium
	}
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(-w/2), float64(-h/2))
	op.GeoM.Rotate(rock.Angle)
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)

	// Challenge text stays upright and deliberately stronger than the artwork.
	// A typed prefix highlights every compatible asteroid: there is no target
	// lock and therefore no implied "closest asteroid" requirement.
	compatible := a.game.Typed != "" && strings.HasPrefix(rock.Answer, a.game.Typed)
	drawChallengeText(screen, rock.Display, x, y, compatible)
}

func (a *app) drawHUD(screen *ebiten.Image) {
	left := fmt.Sprintf("SCORE %06d", a.game.Score)
	centre := fmt.Sprintf("WAVE %d", a.game.Wave)
	if a.game.Mode == game.Maths {
		if a.game.Wave <= 11 {
			centre = fmt.Sprintf("x%d", a.game.Wave+1)
		} else {
			centre = "MIXED"
		}
	}
	right := fmt.Sprintf("SHIELD %d", a.game.Shield)
	ebitenutil.DebugPrintAt(screen, left, 24, 20)
	ebitenutil.DebugPrintAt(screen, centre, game.LogicalWidth/2-len(centre)*3, 20)
	ebitenutil.DebugPrintAt(screen, right, game.LogicalWidth-24-len(right)*6, 20)
}

func (a *app) drawAttract(screen *ebiten.Image) {
	ice := color.RGBA{236, 248, 255, 255}
	cyan := color.RGBA{92, 216, 238, 255}
	textGlow := color.RGBA{64, 68, 74, 255}
	textOutline := color.RGBA{24, 27, 31, 230}

	// The layout remains deliberately simple; polish comes from typography,
	// framing and the existing game art rather than another UI system.
	drawStyledCentered(screen, a.titleFace, "ASTRO ANSWERS", game.LogicalWidth/2, 82, ice, textGlow, textOutline)
	drawSectionHeading(screen, a.headingFace, "CHOOSE YOUR GAME", 124, cyan)

	a.drawScorePanel(screen, 60, 165, 370, 350, "MATHS", "PRESS M TO PLAY", a.game.Scores.Maths)
	a.drawScorePanel(screen, 530, 165, 370, 350, "SPELLING", "PRESS T TO PLAY", a.game.Scores.Typing)

	// Existing game art anchors the title screen so it immediately reads as
	// the same world as the game itself.
	drawImageCentered(screen, a.station, game.LogicalWidth/2, 600, 0)
	drawImageCentered(screen, a.small, game.LogicalWidth/2-88, 598, -0.35)
	drawImageCentered(screen, a.medium, game.LogicalWidth/2+100, 596, 0.28)

	drawSectionHeading(screen, a.labelFace, "THE ANSWER IS THE WEAPON", 681, cyan)
}

func (a *app) drawInitials(screen *ebiten.Image) {
	ice := color.RGBA{236, 248, 255, 255}
	cyan := color.RGBA{92, 216, 238, 255}
	textGlow := color.RGBA{64, 68, 74, 255}
	textOutline := color.RGBA{24, 27, 31, 230}

	drawStyledCentered(screen, a.titleFace, "NEW HIGH SCORE", game.LogicalWidth/2, 150, ice, textGlow, textOutline)
	drawPanelFrame(screen, 245, 205, 470, 355, cyan)
	drawStyledCentered(screen, a.headingFace, fmt.Sprintf("%06d", a.game.Score), game.LogicalWidth/2, 260, ice, textGlow, textOutline)
	drawStyledCentered(screen, a.labelFace, "TYPE THREE INITIALS", game.LogicalWidth/2, 315, ice, textGlow, textOutline)
	drawInitialBoxes(screen, a.game.Initials, 360, cyan)
	a.drawStatsStyled(screen, 445)
	drawStyledCentered(screen, a.labelFace, "SAVE + RETURN TO TITLE", game.LogicalWidth/2, 535, color.RGBA{190, 205, 215, 255}, textGlow, textOutline)
}

func (a *app) drawStatsStyled(screen *ebiten.Image, y int) {
	s := a.game.Stats()
	textGlow := color.RGBA{64, 68, 74, 255}
	textOutline := color.RGBA{24, 27, 31, 230}
	muted := color.RGBA{200, 214, 222, 255}
	drawStyledCentered(screen, a.labelFace, fmt.Sprintf("ACCURACY %3.0f%%", s.Accuracy*100), game.LogicalWidth/2, y, muted, textGlow, textOutline)
	if a.game.Mode == game.Typing {
		drawStyledCentered(screen, a.labelFace, fmt.Sprintf("WPM %.1f", s.WordsPerMinute), game.LogicalWidth/2, y+34, muted, textGlow, textOutline)
	} else {
		drawStyledCentered(screen, a.labelFace, fmt.Sprintf("AVG %.2fs", s.AverageResponseTime), game.LogicalWidth/2, y+34, muted, textGlow, textOutline)
		drawStyledCentered(screen, a.labelFace, fmt.Sprintf("HIGHEST x%d", s.HighestTable), game.LogicalWidth/2, y+68, muted, textGlow, textOutline)
	}
}

func drawChallengeText(screen *ebiten.Image, label string, x, y float64, compatible bool) {
	const scale = 1.45
	bounds := text.BoundString(basicfont.Face7x13, label)
	w := float64(bounds.Dx()) * scale
	h := float64(bounds.Dy()) * scale

	bg := color.RGBA{4, 6, 10, 215}
	fg := color.RGBA{245, 245, 242, 255}
	if compatible {
		bg = color.RGBA{10, 44, 54, 230}
		fg = color.RGBA{205, 250, 255, 255}
	}
	ebitenutil.DrawRect(screen, x-w/2-5, y-h/2-4, w+10, h+8, bg)

	// Draw twice one logical pixel apart for a restrained heavier weight. The
	// source font is deliberately tiny; scaling and the dark backing plate make
	// the challenge readable without adding a font asset or text subsystem.
	drawScaledText(screen, label, x, y, scale, fg)
	drawScaledText(screen, label, x+1, y, scale, fg)
}

func drawScaledText(screen *ebiten.Image, value string, x, y, scale float64, clr color.Color) {
	bounds := text.BoundString(basicfont.Face7x13, value)
	w := float64(bounds.Dx()) * scale
	h := float64(bounds.Dy()) * scale
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(-bounds.Min.X), float64(-bounds.Min.Y))
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x-w/2, y-h/2)
	op.ColorScale.ScaleWithColor(clr)
	text.DrawWithOptions(screen, value, basicfont.Face7x13, op)
}

func (a *app) drawScorePanel(screen *ebiten.Image, x, y, w, h int, title, action string, table []game.ScoreEntry) {
	cyan := color.RGBA{82, 198, 225, 255}
	ice := color.RGBA{232, 244, 250, 255}
	textGlow := color.RGBA{64, 68, 74, 255}
	textOutline := color.RGBA{24, 27, 31, 230}
	drawPanelFrame(screen, x, y, w, h, cyan)

	drawStyledCentered(screen, a.headingFace, title, x+w/2, y+42, ice, textGlow, textOutline)
	drawStyledCentered(screen, a.labelFace, "HIGH SCORES", x+w/2, y+72, color.RGBA{170, 195, 210, 255}, textGlow, textOutline)
	ebitenutil.DrawRect(screen, float64(x+42), float64(y+86), float64(w-84), 1, color.RGBA{55, 95, 120, 180})

	for i := 0; i < 5; i++ {
		initials := "---"
		score := 0
		if i < len(table) {
			initials = table[i].Initials
			score = table[i].Score
		}
		line := fmt.Sprintf("%d   %s   %06d", i+1, initials, score)
		drawTextCenteredAt(screen, line, x+w/2, y+113+i*38)
	}

	ebitenutil.DrawRect(screen, float64(x+42), float64(y+h-70), float64(w-84), 1, color.RGBA{55, 95, 120, 180})
	drawStyledCentered(screen, a.labelFace, action, x+w/2, y+h-33, ice, textGlow, textOutline)
}

func drawStyledCentered(screen *ebiten.Image, face font.Face, value string, centerX, baselineY int, fill, glow, outline color.RGBA) {
	width := font.MeasureString(face, value).Round()
	x := centerX - width/2

	// A restrained layered treatment gives the title-screen type a proper
	// arcade identity without adding image assets or a UI framework.
	for _, off := range [][2]int{{-3, 0}, {3, 0}, {0, -3}, {0, 3}} {
		text.Draw(screen, value, face, x+off[0], baselineY+off[1], color.RGBA{glow.R, glow.G, glow.B, 70})
	}
	for _, off := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {-1, -1}, {1, -1}, {-1, 1}, {1, 1}} {
		text.Draw(screen, value, face, x+off[0], baselineY+off[1], outline)
	}
	text.Draw(screen, value, face, x, baselineY, fill)
}

func drawSectionHeading(screen *ebiten.Image, face font.Face, value string, baselineY int, accent color.RGBA) {
	width := font.MeasureString(face, value).Round()
	center := game.LogicalWidth / 2
	x := center - width/2
	lineY := float64(baselineY - 8)
	gap := 22
	leftEnd := x - gap
	rightStart := x + width + gap
	if leftEnd > 90 {
		ebitenutil.DrawRect(screen, 90, lineY, float64(leftEnd-90), 1, color.RGBA{accent.R, accent.G, accent.B, 135})
	}
	if rightStart < game.LogicalWidth-90 {
		ebitenutil.DrawRect(screen, float64(rightStart), lineY, float64(game.LogicalWidth-90-rightStart), 1, color.RGBA{accent.R, accent.G, accent.B, 135})
	}
	drawStyledCentered(
		screen, face, value, center, baselineY,
		color.RGBA{225, 239, 246, 255},
		color.RGBA{64, 68, 74, 255},
		color.RGBA{24, 27, 31, 230},
	)
}

func drawPanelFrame(screen *ebiten.Image, x, y, w, h int, accent color.RGBA) {
	fill := color.RGBA{7, 13, 23, 220}
	outer := color.RGBA{accent.R, accent.G, accent.B, 205}
	inner := color.RGBA{40, 75, 98, 190}
	ebitenutil.DrawRect(screen, float64(x), float64(y), float64(w), float64(h), fill)

	// Double frame plus brighter corner brackets: enough visual structure to
	// feel designed, still only a handful of rectangles.
	for _, r := range []struct{ x, y, w, h int }{
		{x, y, w, 2}, {x, y + h - 2, w, 2}, {x, y, 2, h}, {x + w - 2, y, 2, h},
		{x + 6, y + 6, w - 12, 1}, {x + 6, y + h - 7, w - 12, 1}, {x + 6, y + 6, 1, h - 12}, {x + w - 7, y + 6, 1, h - 12},
	} {
		clr := inner
		if r.w == 2 || r.h == 2 {
			clr = outer
		}
		ebitenutil.DrawRect(screen, float64(r.x), float64(r.y), float64(r.w), float64(r.h), clr)
	}
	const c = 28
	const t = 3
	for _, r := range []struct{ x, y, w, h int }{
		{x, y, c, t}, {x, y, t, c}, {x + w - c, y, c, t}, {x + w - t, y, t, c},
		{x, y + h - t, c, t}, {x, y + h - c, t, c}, {x + w - c, y + h - t, c, t}, {x + w - t, y + h - c, t, c},
	} {
		ebitenutil.DrawRect(screen, float64(r.x), float64(r.y), float64(r.w), float64(r.h), outer)
	}
}

func drawInitialBoxes(screen *ebiten.Image, initials string, y int, accent color.RGBA) {
	const box = 58
	const gap = 18
	total := box*3 + gap*2
	start := game.LogicalWidth/2 - total/2
	for i := 0; i < 3; i++ {
		x := start + i*(box+gap)
		drawPanelFrame(screen, x, y, box, box, accent)
		value := "_"
		if i < len(initials) {
			value = strings.ToUpper(initials[i : i+1])
		}
		drawTextCenteredAt(screen, value, x+box/2, y+23)
	}
}

func drawImageCentered(screen *ebiten.Image, img *ebiten.Image, x, y float64, angle float64) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(-w)/2, float64(-h)/2)
	op.GeoM.Rotate(angle)
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

func drawTextCenteredAt(screen *ebiten.Image, text string, x, y int) {
	ebitenutil.DebugPrintAt(screen, text, x-len(text)*3, y)
}

func drawCentered(screen *ebiten.Image, text string, y int) {
	ebitenutil.DebugPrintAt(screen, text, game.LogicalWidth/2-len(text)*3, y)
}

func (a *app) drawStars(screen *ebiten.Image) {
	for i := 0; i < 70; i++ {
		x := float64((i*137 + 31) % game.LogicalWidth)
		y := float64((i*83 + 47) % game.LogicalHeight)
		ebitenutil.DrawRect(screen, x, y, 1, 1, color.RGBA{110, 120, 140, 170})
	}
}
