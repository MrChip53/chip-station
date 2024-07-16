//go:build js && wasm

package programs

import (
	"github.com/seqsense/webgl-go"
)

const vWindowShader = `
attribute vec3 position;
attribute vec3 color;
varying vec3 vColor;
varying vec2 vPosition;

void main(void) {
  gl_Position = vec4(position, 1.0);
  vPosition = position.xy;
  vColor = color;
}
`

const fWindowShader = `
precision mediump float;
varying vec3 vColor;
varying vec2 vPosition;

uniform vec2 resolution;
uniform vec2 size;
uniform float radius;

void main(void) {
  vec2 pos = vPosition * resolution * 0.5 + resolution * 0.5;
  vec2 rectSize = vec2(size.x - radius * 2.0, size.y - radius * 2.0);
  vec2 d = abs(pos - size * 0.5) - rectSize * 0.5 + vec2(radius);

  float dist = length(max(d, vec2(0.0))) - radius;
  float alpha = smoothstep(0.0, 1.0, -dist);
  gl_FragColor = vec4(vColor.rgb, alpha);
}
`

type WindowProgram struct {
	program        webgl.Program
	position       int
	color          int
	resolution     webgl.Location
	radius         webgl.Location
	size           webgl.Location
	vertexBuffer   webgl.Buffer
	colorBuffer    webgl.Buffer
	windowColor    Color
	secondaryColor Color
}

func NewWindowProgram(gl *webgl.WebGL) *WindowProgram {
	p := &WindowProgram{
		vertexBuffer:   gl.CreateBuffer(),
		colorBuffer:    gl.CreateBuffer(),
		windowColor:    NewColor(0x379634),
		secondaryColor: NewColor(0x7CFFCB),
	}
	p.init(gl)
	return p
}

func (p *WindowProgram) Draw(gl *webgl.WebGL, title string, x, y, w, h float32) {
	ch := float32(gl.Canvas.ClientHeight())
	cw := float32(gl.Canvas.ClientWidth())

	nx := (x - cw/2) / (cw / 2)
	ny := (y - ch/2) / (ch / 2) * -1
	nw := (1 / cw) * w * 2
	nh := (1 / ch) * h * 2 * -1

	vertices := []float32{
		nx, ny, 0,
		nx, ny + nh, 0,
		nx + nw, ny, 0,

		nx + nw, ny, 0,
		nx, ny + nh, 0,
		nx + nw, ny + nh, 0,
	}

	colors := []float32{
		p.windowColor.R, p.windowColor.G, p.windowColor.B,
		p.windowColor.R, p.windowColor.G, p.windowColor.B,
		p.windowColor.R, p.windowColor.G, p.windowColor.B,

		p.windowColor.R, p.windowColor.G, p.windowColor.B,
		p.windowColor.R, p.windowColor.G, p.windowColor.B,
		p.windowColor.R, p.windowColor.G, p.windowColor.B,
	}

	gl.UseProgram(p.program)

	gl.BindBuffer(gl.ARRAY_BUFFER, p.vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(vertices), gl.STATIC_DRAW)
	gl.VertexAttribPointer(p.position, 3, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(p.position)

	gl.BindBuffer(gl.ARRAY_BUFFER, p.colorBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(colors), gl.STATIC_DRAW)
	gl.VertexAttribPointer(p.color, 3, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(p.color)

	uniform2f(gl, p.resolution, cw, ch)
	uniform2f(gl, p.size, w, h)
	gl.Uniform1f(p.radius, 10.0)

	gl.DrawArrays(gl.TRIANGLES, 0, len(vertices)/3)
}

func (p *WindowProgram) init(gl *webgl.WebGL) {
	var err error
	var vs, fs webgl.Shader
	if vs, err = initVertexShader(gl, vWindowShader); err != nil {
		panic(err)
	}

	if fs, err = initFragmentShader(gl, fWindowShader); err != nil {
		panic(err)
	}

	program, err := linkShaders(gl, nil, vs, fs)
	if err != nil {
		panic(err)
	}

	p.program = program
	p.color = gl.GetAttribLocation(program, "color")
	p.position = gl.GetAttribLocation(program, "position")
	p.resolution = gl.GetUniformLocation(program, "resolution")
	p.radius = gl.GetUniformLocation(program, "radius")
	p.size = gl.GetUniformLocation(program, "size")
}
