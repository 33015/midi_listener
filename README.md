# MIDI Listener

A Go application that visualizes MIDI input on a musical staff in real-time using the Ebitengine game framework.

## Features

- Real-time MIDI input visualization
- Musical staff rendering
- Note position mapping from MIDI input to staff notation
- Smooth graphics using Ebitengine

## Prerequisites

- Go 1.16 or higher
- RtMidi library installed on your system
- A MIDI input device (keyboard, controller, etc.)

## Installation

1. Install the RtMidi library:
   - **Linux**: `sudo apt-get install librtmidi-dev`
   - **macOS**: `brew install rtmidi`
   - **Windows**: Download and install from the RtMidi website

2. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/midi_listener.git
   cd midi_listener
   ```

3. Install Go dependencies:
   ```bash
   go mod download
   ```

## Running the Application

```bash
go run main.go
```

Connect your MIDI device and start playing - notes will be displayed on the staff in real-time.

## Dependencies

- [Ebitengine](https://github.com/hajimehoshi/ebiten) - 2D game engine
- [RtMidi](https://github.com/thestk/rtmidi) - MIDI input/output
- [gomidi](https://gitlab.com/gomidi/midi/) - Go MIDI library

