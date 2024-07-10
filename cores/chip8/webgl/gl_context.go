//go:build js && wasm

package chip8web

import (
	"errors"
	"fmt"
	"syscall/js"

	"github.com/seqsense/webgl-go"
)

type GlContext struct {
	gl *webgl.WebGL

	program  webgl.Program
	color    int
	position int

	vertexBuffer webgl.Buffer
	colorBuffer  webgl.Buffer

	polygonCount int
	colors       []float32
	onColor      Color
	offColor     Color
}

func NewGlContext(gl *webgl.WebGL) *GlContext {
	context := &GlContext{
		gl:           gl,
		vertexBuffer: gl.CreateBuffer(),
		colorBuffer:  gl.CreateBuffer(),
		colors:       make([]float32, 64*32*3*2*3),
		onColor:      NewColor(0xF2CE03),
		offColor:     NewColor(0x8E6903),
	}
	context.generateVertices(64, 32)
	context.createGlProgram()
	return context
}

func (c *GlContext) Draw(display [64][32]uint8) {
	gl := c.gl

	c.calculateColors(display)
	gl.BindBuffer(gl.ARRAY_BUFFER, c.colorBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(c.colors), gl.STATIC_DRAW)

	gl.UseProgram(c.program)

	gl.BindBuffer(gl.ARRAY_BUFFER, c.vertexBuffer)
	gl.VertexAttribPointer(c.position, 3, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(c.position)

	gl.BindBuffer(gl.ARRAY_BUFFER, c.colorBuffer)
	gl.VertexAttribPointer(c.color, 3, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(c.color)

	gl.DrawArrays(gl.TRIANGLES, 0, c.polygonCount)
}

func (c *GlContext) calculateColors(display [64][32]uint8) {
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

func (c *GlContext) generateVertices(width, height int) {
	vertices := []float32{}

	for y := 0; y < height; y++ {
		fy := (height - 1) - y
		for x := 0; x < width; x++ {
			vertices = append(vertices, float32(x)/32-1, float32(fy)/16-1, 0)
			vertices = append(vertices, float32(x+1)/32-1, float32(fy)/16-1, 0)
			vertices = append(vertices, float32(x)/32-1, float32(fy+1)/16-1, 0)
			vertices = append(vertices, float32(x+1)/32-1, float32(fy)/16-1, 0)
			vertices = append(vertices, float32(x+1)/32-1, float32(fy+1)/16-1, 0)
			vertices = append(vertices, float32(x)/32-1, float32(fy+1)/16-1, 0)
		}
	}

	c.gl.BindBuffer(c.gl.ARRAY_BUFFER, c.vertexBuffer)
	c.gl.BufferData(c.gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(vertices), c.gl.STATIC_DRAW)
	c.polygonCount = len(vertices) / 3
}

func (c *GlContext) createGlProgram() {
	var err error
	var vs, fs webgl.Shader
	if vs, err = c.initVertexShader(vsSource); err != nil {
		panic(err)
	}

	if fs, err = c.initFragmentShader(fsSource); err != nil {
		panic(err)
	}

	program, err := c.linkShaders(nil, vs, fs)
	if err != nil {
		panic(err)
	}

	c.program = program
	c.color = c.gl.GetAttribLocation(program, "color")
	c.position = c.gl.GetAttribLocation(program, "position")
}

func (c *GlContext) initVertexShader(src string) (webgl.Shader, error) {
	gl := c.gl
	s := gl.CreateShader(gl.VERTEX_SHADER)
	gl.ShaderSource(s, src)
	gl.CompileShader(s)
	if !gl.GetShaderParameter(s, gl.COMPILE_STATUS).(bool) {
		compilationLog := gl.GetShaderInfoLog(s)
		return webgl.Shader(js.Null()), fmt.Errorf("compile failed (VERTEX_SHADER) %v", compilationLog)
	}
	return s, nil
}

func (c *GlContext) initFragmentShader(src string) (webgl.Shader, error) {
	gl := c.gl
	s := gl.CreateShader(gl.FRAGMENT_SHADER)
	gl.ShaderSource(s, src)
	gl.CompileShader(s)
	if !gl.GetShaderParameter(s, gl.COMPILE_STATUS).(bool) {
		compilationLog := gl.GetShaderInfoLog(s)
		return webgl.Shader(js.Null()), fmt.Errorf("compile failed (FRAGMENT_SHADER) %v", compilationLog)
	}
	return s, nil
}

func (c *GlContext) linkShaders(fbVarings []string, shaders ...webgl.Shader) (webgl.Program, error) {
	gl := c.gl
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
