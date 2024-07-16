//go:build js && wasm

package programs

import (
	"time"

	"github.com/seqsense/webgl-go"

	"github.com/mrchip53/chip-station/cores/chip8"
)

const vsSource = `
attribute vec3 position;
attribute vec3 color;
uniform float scale;
uniform vec2 scaleOffset;
uniform vec2 offset;
varying vec3 vColor;
varying vec2 vPosition;

void main(void) {
  gl_Position = vec4(position*scale+vec3(scaleOffset, 0.0)+vec3(offset, 0.0), 1.0);
  vColor = color;
  vPosition = position.xy;
}
`

const fsSource = `
precision mediump float;
varying vec3 vColor;
varying vec2 vPosition;

uniform vec2 resolution;
uniform float time;

void main(void) {
  vec2 fragPixelCoord = (vPosition.xy + 1.0) / 2.0 * resolution;
  vec2 fragCoord = abs(fragPixelCoord*2.0-resolution.xy/1.0);

  float line = pow(fragCoord.x/resolution.x, 70.0) + pow((fragCoord.y + (resolution.x - resolution.y))/resolution.x, 70.0);
  float minphase = abs(0.02*sin(time*10.0)+0.2*sin(fragCoord.y));
  float frame = max(min(line+minphase,1.0),0.0);

  gl_FragColor = vec4(vColor-vec3(frame), 1.0);
  //gl_FragColor = vec4(vColor, 1.0);
}
`

type DisplayProgram struct {
	program webgl.Program

	position    int
	color       int
	scale       webgl.Location
	scaleOffset webgl.Location
	offset      webgl.Location
	resolution  webgl.Location
	time        webgl.Location

	vertexBuffer webgl.Buffer
	colorBuffer  webgl.Buffer

	polygonCount int

	start time.Time
}

func NewDisplayProgram(gl *webgl.WebGL) *DisplayProgram {
	c := &DisplayProgram{
		program: gl.CreateProgram(),

		vertexBuffer: gl.CreateBuffer(),
		colorBuffer:  gl.CreateBuffer(),
		start:        time.Now(),
	}
	c.init(gl)
	return c
}

func (c *DisplayProgram) init(gl *webgl.WebGL) {
	var err error
	var vs, fs webgl.Shader
	if vs, err = initVertexShader(gl, vsSource); err != nil {
		panic(err)
	}

	if fs, err = initFragmentShader(gl, fsSource); err != nil {
		panic(err)
	}

	program, err := linkShaders(gl, nil, vs, fs)
	if err != nil {
		panic(err)
	}

	c.program = program
	c.color = gl.GetAttribLocation(program, "color")
	c.position = gl.GetAttribLocation(program, "position")
	c.scale = gl.GetUniformLocation(program, "scale")
	c.offset = gl.GetUniformLocation(program, "offset")
	c.scaleOffset = gl.GetUniformLocation(program, "scaleOffset")
	c.resolution = gl.GetUniformLocation(program, "resolution")
	c.time = gl.GetUniformLocation(program, "time")
	c.generateVertices(gl)
}

func (p *DisplayProgram) generateVertices(gl *webgl.WebGL) {
	vertices := []float32{}

	for y := 0; y < chip8.SCREEN_HEIGHT; y++ {
		fy := (chip8.SCREEN_HEIGHT - 1) - y
		for x := 0; x < chip8.SCREEN_WIDTH; x++ {
			vertices = append(vertices, float32(x)/32-1, float32(fy)/16-1, 0)
			vertices = append(vertices, float32(x+1)/32-1, float32(fy)/16-1, 0)
			vertices = append(vertices, float32(x)/32-1, float32(fy+1)/16-1, 0)
			vertices = append(vertices, float32(x+1)/32-1, float32(fy)/16-1, 0)
			vertices = append(vertices, float32(x+1)/32-1, float32(fy+1)/16-1, 0)
			vertices = append(vertices, float32(x)/32-1, float32(fy+1)/16-1, 0)
		}
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, p.vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(vertices), gl.STATIC_DRAW)
	p.polygonCount = len(vertices) / 3
}

func (c *DisplayProgram) Draw(gl *webgl.WebGL, colors []float32, scale float32, x, y float32) {
	h := float32(gl.Canvas.ClientHeight())
	w := float32(gl.Canvas.ClientWidth())

	gl.UseProgram(c.program)

	gl.BindBuffer(gl.ARRAY_BUFFER, c.vertexBuffer)
	gl.VertexAttribPointer(c.position, 3, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(c.position)

	gl.BindBuffer(gl.ARRAY_BUFFER, c.colorBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(colors), gl.STATIC_DRAW)
	gl.VertexAttribPointer(c.color, 3, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(c.color)

	scaleOffX := float32(0)
	scaleOffY := float32(0)
	if scale != 1.0 {
		scaleOffX = scale * -1
		scaleOffY = scale
	}
	gl.Uniform1f(c.scale, scale)
	gl.Uniform1f(c.time, float32(time.Since(c.start).Seconds()))
	uniform2f(gl, c.scaleOffset, scaleOffX, scaleOffY)
	uniform2f(gl, c.offset, x, y)
	uniform2f(gl, c.resolution, w*scale, h*scale)

	gl.DrawArrays(gl.TRIANGLES, 0, c.polygonCount)
}
