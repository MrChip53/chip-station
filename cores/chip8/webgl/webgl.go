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

vec3 rotateX(vec3 v) {
    mat3 rotationMatrix = mat3(
        1.0, 0.0,  0.0,
        0.0, -1.0, 0.0,
        0.0, 0.0, -1.0
    );
    return rotationMatrix * v;
}

void main(void) {
  gl_Position = vec4(rotateX(position), 1.0);
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

type Chip8WebEmulator struct {
	chip8.Chip8Emulator

	numVertices int
	vertices    []float32
	colors      []float32

	width        int
	height       int
	gl           *webgl.WebGL
	vertexBuffer webgl.Buffer
	colorBuffer  webgl.Buffer
	glProgram    webgl.Program
}

func NewChip8WebEmulator(gl *webgl.WebGL) *Chip8WebEmulator {
	e := &Chip8WebEmulator{
		Chip8Emulator: chip8.Chip8Emulator{},
		colors:        []float32{},
		gl:            gl,
		width:         -1,
		height:        -1,
		vertexBuffer:  gl.CreateBuffer(),
		colorBuffer:   gl.CreateBuffer(),
	}
	e.calculateVertices()
	e.createGlProgram()
	return e
}

func (e *Chip8WebEmulator) Draw() {
	w := e.gl.Canvas.ClientWidth()
	h := e.gl.Canvas.ClientHeight()

	if e.width != w || e.height != h {
		fmt.Println("Re-calculating vertices")
		e.width = w
		e.height = h

		e.calculateVertices()
		e.gl.BindBuffer(e.gl.ARRAY_BUFFER, e.vertexBuffer)
		e.gl.BufferData(e.gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(e.vertices), e.gl.STATIC_DRAW)
	}

	e.calculateColors()
	e.gl.BindBuffer(e.gl.ARRAY_BUFFER, e.colorBuffer)
	e.gl.BufferData(e.gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(e.colors), e.gl.STATIC_DRAW)

	e.gl.UseProgram(e.glProgram)

	e.gl.BindBuffer(e.gl.ARRAY_BUFFER, e.vertexBuffer)
	position := e.gl.GetAttribLocation(e.glProgram, "position")
	e.gl.VertexAttribPointer(position, 3, e.gl.FLOAT, false, 0, 0)
	e.gl.EnableVertexAttribArray(position)

	e.gl.BindBuffer(e.gl.ARRAY_BUFFER, e.colorBuffer)
	color := e.gl.GetAttribLocation(e.glProgram, "color")
	e.gl.VertexAttribPointer(color, 3, e.gl.FLOAT, false, 0, 0)
	e.gl.EnableVertexAttribArray(color)

	e.gl.ClearColor(0, 0, 0, 1)
	e.gl.Clear(e.gl.COLOR_BUFFER_BIT)
	e.gl.Enable(e.gl.DEPTH_TEST)
	e.gl.Viewport(0, 0, w, h)
	e.gl.DrawArrays(e.gl.TRIANGLES, 0, len(e.vertices)/3)
}

func (e *Chip8WebEmulator) calculateVertices() {
	display := e.GetDisplay()
	e.vertices = []float32{}
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			if display[x][y] == 1 {
				e.vertices = append(e.vertices, float32(x)/32-1, float32(y)/16-1, 0)
				e.vertices = append(e.vertices, float32(x+1)/32-1, float32(y)/16-1, 0)
				e.vertices = append(e.vertices, float32(x)/32-1, float32(y+1)/16-1, 0)
				e.vertices = append(e.vertices, float32(x+1)/32-1, float32(y)/16-1, 0)
				e.vertices = append(e.vertices, float32(x+1)/32-1, float32(y+1)/16-1, 0)
				e.vertices = append(e.vertices, float32(x)/32-1, float32(y+1)/16-1, 0)
			} else {
				e.vertices = append(e.vertices, float32(x)/32-1, float32(y)/16-1, 0)
				e.vertices = append(e.vertices, float32(x+1)/32-1, float32(y)/16-1, 0)
				e.vertices = append(e.vertices, float32(x)/32-1, float32(y+1)/16-1, 0)
				e.vertices = append(e.vertices, float32(x+1)/32-1, float32(y)/16-1, 0)
				e.vertices = append(e.vertices, float32(x+1)/32-1, float32(y+1)/16-1, 0)
				e.vertices = append(e.vertices, float32(x)/32-1, float32(y+1)/16-1, 0)
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

	e.glProgram = program
}

func (e *Chip8WebEmulator) initVertexShader(src string) (webgl.Shader, error) {
	s := e.gl.CreateShader(e.gl.VERTEX_SHADER)
	e.gl.ShaderSource(s, src)
	e.gl.CompileShader(s)
	if !e.gl.GetShaderParameter(s, e.gl.COMPILE_STATUS).(bool) {
		compilationLog := e.gl.GetShaderInfoLog(s)
		return webgl.Shader(js.Null()), fmt.Errorf("compile failed (VERTEX_SHADER) %v", compilationLog)
	}
	return s, nil
}

func (e *Chip8WebEmulator) initFragmentShader(src string) (webgl.Shader, error) {
	s := e.gl.CreateShader(e.gl.FRAGMENT_SHADER)
	e.gl.ShaderSource(s, src)
	e.gl.CompileShader(s)
	if !e.gl.GetShaderParameter(s, e.gl.COMPILE_STATUS).(bool) {
		compilationLog := e.gl.GetShaderInfoLog(s)
		return webgl.Shader(js.Null()), fmt.Errorf("compile failed (FRAGMENT_SHADER) %v", compilationLog)
	}
	return s, nil
}

func (e *Chip8WebEmulator) linkShaders(fbVarings []string, shaders ...webgl.Shader) (webgl.Program, error) {
	program := e.gl.CreateProgram()
	for _, s := range shaders {
		e.gl.AttachShader(program, s)
	}
	if len(fbVarings) > 0 {
		e.gl.TransformFeedbackVaryings(program, fbVarings, e.gl.SEPARATE_ATTRIBS)
	}
	e.gl.LinkProgram(program)
	if !e.gl.GetProgramParameter(program, e.gl.LINK_STATUS).(bool) {
		return webgl.Program(js.Null()), errors.New("link failed: " + e.gl.GetProgramInfoLog(program))
	}
	return program, nil
}

func (e *Chip8WebEmulator) SetWebGL(gl *webgl.WebGL) {
	e.gl = gl
}
