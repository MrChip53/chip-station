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
	showUi    bool
}

func NewChip8WebEmulator(gl *webgl.WebGL, hooks chip8.Hooks) *Chip8WebEmulator {
	e := &Chip8WebEmulator{
		Chip8Emulator: *chip8.NewChip8Emulator(hooks),
		glContext:     NewGlContext(gl),
		uiContext:     NewUiContext(gl),
		showUi:        true,
	}
	return e
}

func (e *Chip8WebEmulator) Draw() {
	w := e.glContext.gl.Canvas.ClientWidth()
	h := e.glContext.gl.Canvas.ClientHeight()

	e.glContext.gl.ClearColor(e.glContext.offColor.R, e.glContext.offColor.G, e.glContext.offColor.B, 1)
	e.glContext.gl.Clear(e.glContext.gl.COLOR_BUFFER_BIT)
	e.glContext.gl.Enable(e.glContext.gl.DEPTH_TEST)
	e.glContext.gl.DepthFunc(e.glContext.gl.LEQUAL)
	e.glContext.gl.Enable(e.glContext.gl.BLEND)
	e.glContext.gl.BlendFunc(e.glContext.gl.SRC_ALPHA, e.glContext.gl.ONE_MINUS_SRC_ALPHA)
	e.glContext.gl.Viewport(0, 0, w, h)

	e.glContext.Draw(e.GetDisplay())

	if e.showUi {
		e.uiContext.Draw(e)
	}
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
