package main

import (
	"fmt"
	"image/color"
	"log"
	"strings"
	"sync"

	"midi_listener/notes"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	screenWidth  = 800
	screenHeight = 400
	buttonWidth  = 120
	buttonHeight = 40
)

var (
	mplusNormalFont font.Face
)

type Button struct {
	x, y          int
	width, height int
	text          string
	clicked       bool
}

func (b *Button) Contains(x, y int) bool {
	return x >= b.x && x < b.x+b.width && y >= b.y && y < b.y+b.height
}

func (b *Button) Draw(screen *ebiten.Image) {
	buttonColor := color.RGBA{200, 200, 200, 255}
	if b.clicked {
		buttonColor = color.RGBA{150, 150, 150, 255}
	}

	// Draw button background
	buttonRect := ebiten.NewImage(b.width, b.height)
	buttonRect.Fill(buttonColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(b.x), float64(b.y))
	screen.DrawImage(buttonRect, op)

	// Draw button text
	textBounds := text.BoundString(mplusNormalFont, b.text)
	textX := b.x + (b.width-textBounds.Dx())/2
	textY := b.y + (b.height+textBounds.Dy())/2
	text.Draw(screen, b.text, mplusNormalFont, textX, textY, color.Black)
}

type Game struct {
	ports        []drivers.In
	selectedPort int
	deviceName   string
	currentNotes []string
	stop         func()
	mu           sync.Mutex
	buttons      []Button
	started      bool
	currentNote  *Note
}

type Note struct {
	midiNote uint8
	velocity uint8
}

// Convert MIDI note number to staff position (0 = middle C, positive = above, negative = below)
func getStaffPosition(midiNote uint8) int {
	return notes.GetStaffPosition(midiNote)
}

func GetStaffPosition(midiNote uint8) int {
	return int(midiNote) - 60
}

func (g *Game) drawStaff(screen *ebiten.Image) {
	if !g.started || len(g.currentNotes) == 0 {
		return
	}

	staffY := 200
	staffWidth := 200
	lineSpacing := 10

	// Draw five staff lines
	for i := 0; i < 5; i++ {
		ebitenutil.DrawLine(screen, 50, float64(staffY+i*lineSpacing), float64(50+staffWidth), float64(staffY+i*lineSpacing), color.White)
	}

	// Draw treble clef (simplified as a vertical line for now)
	ebitenutil.DrawLine(screen, 60, float64(staffY-lineSpacing), 60, float64(staffY+5*lineSpacing), color.White)

	// Draw the current note if exists
	if g.currentNote != nil {
		pos := getStaffPosition(g.currentNote.midiNote)
		// Position notes relative to the staff
		// Add 6 to shift the position so that A4 (pos=0) is on the second space
		noteY := staffY + 4*lineSpacing - (pos+6)*(lineSpacing/2)

		// Draw note head (circle)
		radius := float64(lineSpacing) * 0.4
		for y := -radius; y <= radius; y++ {
			for x := -radius; x <= radius; x++ {
				if x*x+y*y <= radius*radius {
					screen.Set(int(100+x), int(float64(noteY)+y), color.White)
				}
			}
		}

		// Draw ledger lines if needed
		if pos <= -5 { // Below staff (including middle C)
			// Start from the first ledger line below staff (C4)
			for y := staffY + 4*lineSpacing + lineSpacing; y <= noteY+lineSpacing/2; y += lineSpacing {
				ebitenutil.DrawLine(screen, 90, float64(y), 110, float64(y), color.White)
			}
		} else if pos >= 5 { // Above staff
			// Start from the first ledger line above staff
			for y := staffY; y >= noteY-lineSpacing/2; y -= lineSpacing {
				ebitenutil.DrawLine(screen, 90, float64(y), 110, float64(y), color.White)
			}
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{40, 40, 40, 255})

	if !g.started {
		// Draw device selection buttons
		for _, btn := range g.buttons {
			btn.Draw(screen)
		}
	} else {
		// Draw device name
		text.Draw(screen, g.deviceName, mplusNormalFont, 20, 30, color.White)

		// Draw current notes
		for i, note := range g.currentNotes {
			text.Draw(screen, note, mplusNormalFont, 20, 60+i*20, color.White)
		}

		// Draw staff and current note
		g.drawStaff(screen)
	}

	// Draw close button
	closeBtn := Button{
		x:      screenWidth - buttonWidth - 20,
		y:      20,
		width:  buttonWidth,
		height: buttonHeight,
		text:   "Close",
	}
	closeBtn.Draw(screen)
}

func (g *Game) Update() error {
	if g.buttons == nil {
		g.buttons = make([]Button, len(g.ports))
		for i, port := range g.ports {
			g.buttons[i] = Button{
				x:      20,
				y:      60 + i*50,
				width:  400,
				height: 40,
				text:   port.String(),
			}
		}
	}

	// Handle mouse input
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()

		// Check close button
		closeBtn := Button{
			x:      screenWidth - buttonWidth - 20,
			y:      20,
			width:  buttonWidth,
			height: buttonHeight,
			text:   "Close",
		}
		if closeBtn.Contains(x, y) {
			closeBtn.clicked = true
			return ebiten.Termination
		}

		// Check port buttons
		for i, btn := range g.buttons {
			if btn.Contains(x, y) && !g.started {
				g.selectedPort = i
				// Start MIDI listening for selected port
				g.startMIDIListening()
				g.started = true
				break
			}
		}
	} else {
		closeBtn := Button{
			x:      screenWidth - buttonWidth - 20,
			y:      20,
			width:  buttonWidth,
			height: buttonHeight,
			text:   "Close",
		}
		closeBtn.clicked = false
		for _, btn := range g.buttons {
			btn.clicked = false
		}
	}

	return nil
}

func (g *Game) startMIDIListening() {
	if g.stop != nil {
		g.stop()
	}

	if g.selectedPort >= 0 && g.selectedPort < len(g.ports) {
		selectedPort := g.ports[g.selectedPort]
		g.deviceName = selectedPort.String()
		fmt.Printf("Starting to listen to MIDI device: %s\n", g.deviceName)

		stop, err := midi.ListenTo(selectedPort, func(msg midi.Message, timestampms int32) {
			msgStr := msg.String()
			fmt.Printf("Raw MIDI message: %s\n", msgStr)

			// Check for NoteOn messages
			if strings.HasPrefix(msgStr, "NoteOn") {
				// Parse the message which is in format: "NoteOn channel: X key: Y velocity: Z"
				var channel, key, velocity int
				_, err := fmt.Sscanf(msgStr, "NoteOn channel: %d key: %d velocity: %d", &channel, &key, &velocity)
				if err == nil && velocity > 0 {
					noteStr := getNoteString(uint8(key))
					g.mu.Lock()
					g.currentNotes = append([]string{noteStr}, g.currentNotes...)
					if len(g.currentNotes) > 5 {
						g.currentNotes = g.currentNotes[:5]
					}
					g.currentNote = &Note{
						midiNote: uint8(key),
						velocity: uint8(velocity),
					}
					g.mu.Unlock()
					fmt.Printf("Note On: %s\n", noteStr)
				}
			} else if strings.HasPrefix(msgStr, "NoteOff") {
				g.mu.Lock()
				if g.currentNote != nil {
					var channel, key int
					_, err := fmt.Sscanf(msgStr, "NoteOff channel: %d key: %d", &channel, &key)
					if err == nil && uint8(key) == g.currentNote.midiNote {
						g.currentNote = nil
					}
				}
				g.mu.Unlock()
			}
		})

		if err != nil {
			log.Printf("Error starting MIDI listening: %v", err)
			return
		}
		g.stop = stop
		fmt.Println("MIDI listening started successfully")
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func getNoteString(note uint8) string {
	noteNames := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
	noteName := noteNames[note%12]
	octave := int(note/12) - 1
	return fmt.Sprintf("%s%d", noteName, octave)
}

func init() {
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	mplusNormalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	defer midi.CloseDriver()

	// List all MIDI ports
	inPorts := midi.GetInPorts()
	if len(inPorts) == 0 {
		log.Fatal("no MIDI input ports found")
	}

	game := &Game{
		ports:        inPorts,
		currentNotes: make([]string, 0),
		selectedPort: -1,
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("MIDI Note Display")

	if err := ebiten.RunGame(game); err != nil {
		if err != ebiten.Termination {
			log.Fatal(err)
		}
	}
}
