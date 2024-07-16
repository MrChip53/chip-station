//go:build js && wasm

package chip8web

import "github.com/mrchip53/chip-station/cores/chip8"

type Message interface {
	Handle(*Chip8WebEmulator)
}

type ChangeColorMessage struct {
	chip8.CustomMessage
	color Color
	off   bool
}

func (m ChangeColorMessage) Handle(e *Chip8WebEmulator) {
	if m.off {
		e.glContext.offColor = m.color
	} else {
		e.glContext.onColor = m.color
	}
}

type ToggleUiMessage struct {
	chip8.CustomMessage
}

func (m ToggleUiMessage) Handle(e *Chip8WebEmulator) {
	e.ResetFps()
	e.glContext.fullScreen = !e.glContext.fullScreen
}
