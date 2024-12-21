package notes

// Note represents a musical note with its MIDI number and staff position
type Note struct {
	Name     string
	MIDINote uint8
	Position int // Position on the staff (0 = middle C, positive = above, negative = below)
}

// Musical notes in the treble clef, centered around middle C (C4)
var (
	C3 = Note{"C3", 48, -8}
	D3 = Note{"D3", 50, -7}
	E3 = Note{"E3", 52, -6}
	F3 = Note{"F3", 53, -5}
	G3 = Note{"G3", 55, -4}
	A3 = Note{"A3", 57, -3}
	B3 = Note{"B3", 59, -2}
	C4 = Note{"C4", 60, -1}
	D4 = Note{"D4", 62, 0}
	E4 = Note{"E4", 64, 1}
	F4 = Note{"F4", 65, 2}
	G4 = Note{"G4", 67, 3}
	A4 = Note{"A4", 69, 4}
	B4 = Note{"B4", 71, 5}
	C5 = Note{"C5", 72, 2}
	D5 = Note{"D5", 74, 3}
	E5 = Note{"E5", 76, 4}
	F5 = Note{"F5", 77, 5}
	G5 = Note{"G5", 79, 6}
	A5 = Note{"A5", 81, 7}
	B5 = Note{"B5", 83, 8}
	C6 = Note{"C6", 84, 9}
)

// AllNotes contains all defined notes in order
var AllNotes = []Note{
	C3, D3, E3, F3, G3, A3, B3,
	C4, D4, E4, F4, G4, A4, B4,
	C5, D5, E5, F5, G5, A5, B5, C6,
}

// GetNoteByMIDI returns the Note struct for a given MIDI note number
func GetNoteByMIDI(midiNote uint8) Note {
	for _, note := range AllNotes {
		if note.MIDINote == midiNote {
			return note
		}
	}
	// If not found, calculate position based on middle C
	position := (int(midiNote) - 60) / 2
	return Note{
		Name:     "",
		MIDINote: midiNote,
		Position: position,
	}
}

// GetStaffPosition returns the staff position for a given MIDI note
func GetStaffPosition(midiNote uint8) int {
	return GetNoteByMIDI(midiNote).Position
}
