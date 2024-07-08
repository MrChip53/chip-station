//go:build js && wasm

package chip8web

import (
	webgl "github.com/seqsense/webgl-go"

	"github.com/mrchip53/chip-station/cores/chip8"
)

const vsSource = `
attribute vec3 position;
attribute vec3 color;
varying vec3 vColor;

void main(void) {
  gl_Position = vec4(position, 1.0);
  vColor = color;
}
`

const fsSource = `
precision mediump float;
varying vec3 vColor;

void main(void) {
  vec2 uv = gl_FragCoord.xy / vec2(64.0, 32.0);
  float scanline = sin(uv.y * 64.0) * 0.05;
  vec3 color = vColor.rgb + scanline;
  gl_FragColor = vec4(color, 1.);
}
`

type Chip8WebEmulator struct {
	chip8.Chip8Emulator

	glContext *GlContext
	uiContext *UiContext
	beep      *Beep
}

func NewChip8WebEmulator(gl *webgl.WebGL, hooks chip8.Hooks) *Chip8WebEmulator {
	e := &Chip8WebEmulator{
		Chip8Emulator: *chip8.NewChip8Emulator(hooks),
		glContext:     NewGlContext(gl),
		uiContext:     NewUiContext(gl),
	}
	return e
}

func (e *Chip8WebEmulator) Draw() {
	e.glContext.Draw(e.GetDisplay())
	// e.uiContext.RenderText("Hell", 0, 0)
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

func (e *Chip8WebEmulator) SetOffColor(c Color) {
	e.EnqueueMessage(ChangeColorMessage{color: c, off: true})
}

func (e *Chip8WebEmulator) SetOnColor(c Color) {
	e.EnqueueMessage(ChangeColorMessage{color: c, off: false})
}
