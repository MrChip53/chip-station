//go:build js && wasm

package chip8web

import (
	"errors"
	"fmt"
	"syscall/js"

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
  gl_FragColor = vec4(vColor, 1.);
}
`

type GlProgram struct {
	gl *webgl.WebGL

	program  webgl.Program
	color    int
	position int

	vertexBuffer webgl.Buffer
	colorBuffer  webgl.Buffer
}

type Chip8WebEmulator struct {
	chip8.Chip8Emulator

	vertices []float32
	colors   []float32

	glProgram *GlProgram
}

func NewChip8WebEmulator(gl *webgl.WebGL) *Chip8WebEmulator {
	e := &Chip8WebEmulator{
		Chip8Emulator: chip8.Chip8Emulator{},
		colors:        []float32{},
		glProgram: &GlProgram{
			gl:           gl,
			vertexBuffer: gl.CreateBuffer(),
			colorBuffer:  gl.CreateBuffer(),
		},
	}
	e.calculateVertices()
	gl.BindBuffer(gl.ARRAY_BUFFER, e.glProgram.vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(e.vertices), gl.STATIC_DRAW)
	e.createGlProgram()
	return e
}

func (e *Chip8WebEmulator) Draw() {
	gl := e.glProgram.gl
	w := gl.Canvas.ClientWidth()
	h := gl.Canvas.ClientHeight()

	e.calculateColors()
	gl.BindBuffer(gl.ARRAY_BUFFER, e.glProgram.colorBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(e.colors), gl.STATIC_DRAW)

	gl.UseProgram(e.glProgram.program)

	gl.BindBuffer(gl.ARRAY_BUFFER, e.glProgram.vertexBuffer)
	gl.VertexAttribPointer(e.glProgram.position, 3, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(e.glProgram.position)

	gl.BindBuffer(gl.ARRAY_BUFFER, e.glProgram.colorBuffer)
	gl.VertexAttribPointer(e.glProgram.color, 3, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(e.glProgram.color)

	gl.ClearColor(0, 0, 0, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.Enable(gl.DEPTH_TEST)
	gl.Viewport(0, 0, w, h)
	gl.DrawArrays(gl.TRIANGLES, 0, len(e.vertices)/3)
}

func (e *Chip8WebEmulator) calculateVertices() {
	display := e.GetDisplay()
	e.vertices = []float32{}
	for y := 0; y < 32; y++ {
		fy := 31 - y
		for x := 0; x < 64; x++ {
			if display[x][y] == 1 {
				e.vertices = append(e.vertices, float32(x)/32-1, float32(fy)/16-1, 0)
				e.vertices = append(e.vertices, float32(x+1)/32-1, float32(fy)/16-1, 0)
				e.vertices = append(e.vertices, float32(x)/32-1, float32(fy+1)/16-1, 0)
				e.vertices = append(e.vertices, float32(x+1)/32-1, float32(fy)/16-1, 0)
				e.vertices = append(e.vertices, float32(x+1)/32-1, float32(fy+1)/16-1, 0)
				e.vertices = append(e.vertices, float32(x)/32-1, float32(fy+1)/16-1, 0)
			} else {
				e.vertices = append(e.vertices, float32(x)/32-1, float32(fy)/16-1, 0)
				e.vertices = append(e.vertices, float32(x+1)/32-1, float32(fy)/16-1, 0)
				e.vertices = append(e.vertices, float32(x)/32-1, float32(fy+1)/16-1, 0)
				e.vertices = append(e.vertices, float32(x+1)/32-1, float32(fy)/16-1, 0)
				e.vertices = append(e.vertices, float32(x+1)/32-1, float32(fy+1)/16-1, 0)
				e.vertices = append(e.vertices, float32(x)/32-1, float32(fy+1)/16-1, 0)
			}
		}
	}
}

func (e *Chip8WebEmulator) calculateColors() {
	display := e.GetDisplay()
	e.colors = []float32{}
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			if display[x][y] == 1 {
				e.colors = append(e.colors, 1, 1, 1)
				e.colors = append(e.colors, 1, 1, 1)
				e.colors = append(e.colors, 1, 1, 1)
				e.colors = append(e.colors, 1, 1, 1)
				e.colors = append(e.colors, 1, 1, 1)
				e.colors = append(e.colors, 1, 1, 1)
			} else {
				e.colors = append(e.colors, 0, 0, 0)
				e.colors = append(e.colors, 0, 0, 0)
				e.colors = append(e.colors, 0, 0, 0)
				e.colors = append(e.colors, 0, 0, 0)
				e.colors = append(e.colors, 0, 0, 0)
				e.colors = append(e.colors, 0, 0, 0)
			}
		}
	}
}

func (e *Chip8WebEmulator) createGlProgram() {
	var err error
	var vs, fs webgl.Shader
	if vs, err = e.initVertexShader(vsSource); err != nil {
		panic(err)
	}

	if fs, err = e.initFragmentShader(fsSource); err != nil {
		panic(err)
	}

	program, err := e.linkShaders(nil, vs, fs)
	if err != nil {
		panic(err)
	}

	e.glProgram.program = program
	e.glProgram.color = e.glProgram.gl.GetAttribLocation(program, "color")
	e.glProgram.position = e.glProgram.gl.GetAttribLocation(program, "position")
}

func (e *Chip8WebEmulator) initVertexShader(src string) (webgl.Shader, error) {
	gl := e.glProgram.gl
	s := gl.CreateShader(gl.VERTEX_SHADER)
	gl.ShaderSource(s, src)
	gl.CompileShader(s)
	if !gl.GetShaderParameter(s, gl.COMPILE_STATUS).(bool) {
		compilationLog := gl.GetShaderInfoLog(s)
		return webgl.Shader(js.Null()), fmt.Errorf("compile failed (VERTEX_SHADER) %v", compilationLog)
	}
	return s, nil
}

func (e *Chip8WebEmulator) initFragmentShader(src string) (webgl.Shader, error) {
	gl := e.glProgram.gl
	s := gl.CreateShader(gl.FRAGMENT_SHADER)
	gl.ShaderSource(s, src)
	gl.CompileShader(s)
	if !gl.GetShaderParameter(s, gl.COMPILE_STATUS).(bool) {
		compilationLog := gl.GetShaderInfoLog(s)
		return webgl.Shader(js.Null()), fmt.Errorf("compile failed (FRAGMENT_SHADER) %v", compilationLog)
	}
	return s, nil
}

func (e *Chip8WebEmulator) linkShaders(fbVarings []string, shaders ...webgl.Shader) (webgl.Program, error) {
	gl := e.glProgram.gl
	program := gl.CreateProgram()
	for _, s := range shaders {
		gl.AttachShader(program, s)
	}
	if len(fbVarings) > 0 {
		gl.TransformFeedbackVaryings(program, fbVarings, gl.SEPARATE_ATTRIBS)
	}
	gl.LinkProgram(program)
	if !gl.GetProgramParameter(program, gl.LINK_STATUS).(bool) {
		return webgl.Program(js.Null()), errors.New("link failed: " + gl.GetProgramInfoLog(program))
	}
	return program, nil
}

func (e *Chip8WebEmulator) SetWebGL(gl *webgl.WebGL) {
	e.glProgram.gl = gl
}
