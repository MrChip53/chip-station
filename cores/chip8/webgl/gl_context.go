//go:build js && wasm

package chip8web

import (
	"fmt"

	"github.com/seqsense/webgl-go"

	"github.com/mrchip53/chip-station/cores/chip8"
	"github.com/mrchip53/chip-station/cores/chip8/webgl/programs"
)

type GlContext struct {
	gl *webgl.WebGL

	colors   []float32
	onColor  Color
	offColor Color

	fullScreen bool

	glPrograms *programs.Programs
}

func NewGlContext(gl *webgl.WebGL, fontSource string) *GlContext {
	context := &GlContext{
		gl:         gl,
		colors:     make([]float32, chip8.SCREEN_WIDTH*chip8.SCREEN_HEIGHT*3*2*3),
		onColor:    NewColor(0xF2CE03),
		offColor:   NewColor(0x8E6903),
		glPrograms: programs.NewPrograms(gl, fontSource),
	}
	return context
}

func (c *GlContext) Draw(e *Chip8WebEmulator) {
	c.calculateColors(e.GetDisplay())

	scale := float32(1)
	x := float32(0)
	y := float32(0)
	if !c.fullScreen {
		x = 0.5
		scale = 0.5
	}
	_, _, _ = scale, x, y
	c.glPrograms.DisplayProgram.Draw(c.gl, c.colors, scale, x, y)
	if !c.fullScreen {
		h := c.gl.Canvas.ClientHeight()
		w := c.gl.Canvas.ClientWidth()
		// textHeight := programs.CHAR_SIZE / float32(h)

		c.DrawWindow("ChipStation CHIP-8 Emulator", 0, 0, float32(w)/4.0, float32(h), []string{
			"Toggle Fullscreen: 'u'",
			fmt.Sprintf("FPS: %.2f", e.GetFps()),
			fmt.Sprintf("PC: 0x%04X", e.GetPc()),
			fmt.Sprintf("Opcode: 0x%04X", e.GetOpCode()),
			fmt.Sprintf("IPF: %d cycles/frame", e.GetIPF()),
			fmt.Sprintf("ROM Size: %d bytes", e.GetRomSize()),
		})

		// c.glPrograms.TextProgram.Draw(c.gl, "ChipStation CHIP-8 Emulator - Press 'u' to toggle the UI", -1, 1)
		// c.glPrograms.TextProgram.Draw(c.gl, fmt.Sprintf("FPS: %.2f", e.GetFps()), -1, 1-textHeight, 1)
		// c.glPrograms.TextProgram.Draw(c.gl, fmt.Sprintf("PC: 0x%04X", e.GetPc()), -1, 1-textHeight*2, 1)
		// c.glPrograms.TextProgram.Draw(c.gl, fmt.Sprintf("Opcode: 0x%04X", e.GetOpCode()), -1, 1-textHeight*3, 1)
		// c.glPrograms.TextProgram.Draw(c.gl, fmt.Sprintf("IPF: %d cycles/frame", e.GetIPF()), -1, 1-textHeight*4, 1)
		// c.glPrograms.TextProgram.Draw(c.gl, fmt.Sprintf("ROM Size: %d bytes", e.GetRomSize()), -1, 1-textHeight*5, 1)
	}
}

func (c *GlContext) DrawWindow(title string, x, y, w, h float32, text []string) {
	ch := float32(c.gl.Canvas.ClientHeight())
	cw := float32(c.gl.Canvas.ClientWidth())

	ix := (x + 10 - cw/2) / (cw / 2)
	iy := (y + 10 - ch/2) / (ch / 2) * -1

	c.glPrograms.WindowProgram.Draw(c.gl, title, x, y, w, h)
	c.glPrograms.TextProgram.Draw(c.gl, title, ix, iy, 0.8)

	textHeight := programs.CHAR_SIZE / float32(h)
	iy -= textHeight

	for _, t := range text {
		c.glPrograms.TextProgram.Draw(c.gl, t, ix, iy, 0.9)
		iy -= textHeight
	}
}

func (c *GlContext) calculateColors(display chip8.Display) {
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			offset := (y*64 + x) * 3 * 2 * 3
			if display[x][y] == 1 {
				c.setGeometryColor(offset, 3, 2, c.onColor)
			} else {
				c.setGeometryColor(offset, 3, 2, c.offColor)
			}
		}
	}
}

func (c *GlContext) setGeometryColor(offset, verticeCount, geometryCount int, color Color) {
	for i := 0; i < geometryCount; i++ {
		for j := 0; j < verticeCount; j++ {
			c.colors[(offset+i*verticeCount*3)+(j*3)] = color.R
			c.colors[(offset+i*verticeCount*3)+(j*3)+1] = color.G
			c.colors[(offset+i*verticeCount*3)+(j*3)+2] = color.B
		}
	}
}
