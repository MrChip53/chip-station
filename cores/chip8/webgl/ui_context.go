//go:build js && wasm

package chip8web

import (
	"errors"
	"fmt"
	"syscall/js"

	"github.com/seqsense/webgl-go"
)

const (
	CHARS_PER_ROW = 8
	CHAR_SIZE     = 8
	SHEET_SIZE    = 64
)

var vTextShader = `attribute vec2 a_position;
attribute vec2 a_texCoord;

varying vec2 v_texCoord;

void main() {
    gl_Position = vec4(a_position / vec2(400.0, 300.0) * 2.0 - 1.0, 0.0, 1.0);
    v_texCoord = a_texCoord;
}`

var fTextShader = `precision mediump float;

uniform sampler2D u_texture;
varying vec2 v_texCoord;

void main() {
    gl_FragColor = texture2D(u_texture, v_texCoord);
}`

type UiContext struct {
	gl *webgl.WebGL

	program        webgl.Program
	texture        webgl.Texture
	vertexBuffer   webgl.Buffer
	texCoordBuffer webgl.Buffer
	texCoordAttrib int
	positionAttrib int
	textureUniform webgl.Location
}

func NewUiContext(gl *webgl.WebGL) *UiContext {
	ui := &UiContext{
		gl:             gl,
		texture:        gl.CreateTexture(),
		texCoordBuffer: gl.CreateBuffer(),
		vertexBuffer:   gl.CreateBuffer(),
	}

	img := js.Global().Get("Image").New()

	var cbFunc js.Func
	cbFunc = js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		defer cbFunc.Release()
		ui.gl.BindTexture(ui.gl.TEXTURE_2D, ui.texture)
		ui.gl.TexImage2D(ui.gl.TEXTURE_2D, 0, ui.gl.RGBA, ui.gl.RGBA, ui.gl.UNSIGNED_BYTE, img)
		return nil
	})

	img.Set("src", "http://localhost:8080/app/assets/charmap_black.png")
	img.Set("onload", cbFunc)

	ui.createGlProgram()

	return ui
}

func (c *UiContext) Draw() {
	c.RenderText("Hello, World!", 0, 0)
}

func (c *UiContext) RenderText(text string, x, y float32) {
	vertices := []float32{}
	texCoords := []float32{}

	for i, char := range text {
		char := int(char)
		tx := x + float32(i)*CHAR_SIZE
		ty := y
		coords := c.getTextureCoordinates(char)
		vertices = append(vertices, tx, ty, tx+CHAR_SIZE, ty, tx+CHAR_SIZE, ty+CHAR_SIZE, tx, ty+CHAR_SIZE)
		texCoords = append(texCoords, coords...)
	}

	fmt.Println(vertices)

	c.gl.BindBuffer(c.gl.ARRAY_BUFFER, c.vertexBuffer)
	c.gl.BufferData(c.gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(vertices), c.gl.STATIC_DRAW)

	c.gl.BindBuffer(c.gl.ARRAY_BUFFER, c.texCoordBuffer)
	c.gl.BufferData(c.gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(texCoords), c.gl.STATIC_DRAW)

	c.gl.UseProgram(c.program)
	c.gl.BindBuffer(c.gl.ARRAY_BUFFER, c.vertexBuffer)
	c.gl.VertexAttribPointer(c.positionAttrib, 2, c.gl.FLOAT, false, 0, 0)
	c.gl.EnableVertexAttribArray(c.positionAttrib)

	c.gl.BindBuffer(c.gl.ARRAY_BUFFER, c.texCoordBuffer)
	c.gl.VertexAttribPointer(c.texCoordAttrib, 2, c.gl.FLOAT, false, 0, 0)
	c.gl.EnableVertexAttribArray(c.texCoordAttrib)

	c.gl.ActiveTexture(c.gl.TEXTURE0)
	c.gl.BindTexture(c.gl.TEXTURE_2D, c.texture)
	c.gl.Uniform1i(c.textureUniform, 0)

	c.gl.DrawArrays(c.gl.TRIANGLE_STRIP, 0, len(vertices)/2)
}

func (c *UiContext) getTextureCoordinates(charCode int) []float32 {
	i := charCode - 32
	row := charCode / CHARS_PER_ROW
	col := i % CHARS_PER_ROW
	x := float32(col * CHAR_SIZE / SHEET_SIZE)
	y := float32(row * CHAR_SIZE / SHEET_SIZE)
	size := float32(CHAR_SIZE / SHEET_SIZE)
	return []float32{
		x, y,
		x + size, y,
		x + size, y + size,
		x, y + size,
	}
}

func (c *UiContext) createGlProgram() {
	var err error
	var vs, fs webgl.Shader
	if vs, err = c.initVertexShader(vTextShader); err != nil {
		panic(err)
	}

	if fs, err = c.initFragmentShader(fTextShader); err != nil {
		panic(err)
	}

	program, err := c.linkShaders(nil, vs, fs)
	if err != nil {
		panic(err)
	}

	c.program = program
	c.positionAttrib = c.gl.GetAttribLocation(program, "a_position")
	c.texCoordAttrib = c.gl.GetAttribLocation(program, "a_texCoord")
	c.textureUniform = c.gl.GetUniformLocation(program, "u_texture")
}

func (c *UiContext) initVertexShader(src string) (webgl.Shader, error) {
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

func (c *UiContext) initFragmentShader(src string) (webgl.Shader, error) {
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

func (c *UiContext) linkShaders(fbVarings []string, shaders ...webgl.Shader) (webgl.Program, error) {
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
