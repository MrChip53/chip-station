//go:build js && wasm

package chip8web

import (
	webgl "github.com/seqsense/webgl-go"

	"github.com/mrchip53/chip-station/cores/chip8"
)

type Chip8WebEmulator struct {
	chip8.Chip8Emulator

	gl        *webgl.WebGL
	glContext *GlContext
	beep      *Beep
}

func NewChip8WebEmulator(gl *webgl.WebGL, hooks chip8.Hooks) *Chip8WebEmulator {
	e := &Chip8WebEmulator{
		Chip8Emulator: *chip8.NewChip8Emulator(hooks),
		gl:            gl,
		glContext:     NewGlContext(gl),
	}
	return e
}

func (e *Chip8WebEmulator) Draw() {
	w := e.gl.Canvas.ClientWidth()
	h := e.gl.Canvas.ClientHeight()

	e.gl.ClearColor(0, 0, 0, 1)
	e.gl.Clear(e.gl.COLOR_BUFFER_BIT)
	e.gl.Enable(e.gl.DEPTH_TEST)
	e.gl.DepthFunc(e.gl.LEQUAL)
	e.gl.Enable(e.gl.BLEND)
	e.gl.BlendFunc(e.gl.SRC_ALPHA, e.gl.ONE_MINUS_SRC_ALPHA)
	e.gl.Viewport(0, 0, w, h)

	e.glContext.Draw(e)
}

func (e *Chip8WebEmulator) PlayBeep() {
	if e.beep == nil {
		e.beep = NewBeep()
	}
	e.beep.Play()
}

func (e *Chip8WebEmulator) StopBeep() {
	if e.beep != nil {
		e.beep.Stop()
	}
}

func (e *Chip8WebEmulator) ToggleUi() {
	e.EnqueueMessage(ToggleUiMessage{})
}

func (e *Chip8WebEmulator) SetOffColor(c Color) {
	e.EnqueueMessage(ChangeColorMessage{color: c, off: true})
}

func (e *Chip8WebEmulator) SetOnColor(c Color) {
	e.EnqueueMessage(ChangeColorMessage{color: c, off: false})
}
