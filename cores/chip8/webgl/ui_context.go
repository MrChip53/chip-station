//go:build js && wasm

package chip8web

import (
	"errors"
	"fmt"
	"syscall/js"

	"github.com/seqsense/webgl-go"
)

const (
	CHARS_PER_ROW = float32(32)
	CHAR_SIZE     = float32(8)
	SHEET_HEIGHT  = CHAR_SIZE * 3
	SHEET_WIDTH   = CHAR_SIZE * CHARS_PER_ROW
	SCALE         = float32(5)
)

var vTextShader = `
attribute vec2 position;
attribute vec2 texCoord;

varying vec2 vTexCoord;

void main() {
    gl_Position = vec4(position, 0.0, 1.0);
    vTexCoord = texCoord;
}`

var fTextShader = `
precision mediump float;

uniform sampler2D texture;
varying vec2 vTexCoord;
vec4 color;

void main() {
    color = texture2D(texture, vTexCoord);
    if (color.a > 0.0) {
        gl_FragColor = color;
    } else {
        gl_FragColor = vec4(color.rgb, 0.5);
    }
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
	loaded         bool
}

func NewUiContext(gl *webgl.WebGL) *UiContext {
	ui := &UiContext{
		gl:             gl,
		texture:        gl.CreateTexture(),
		texCoordBuffer: gl.CreateBuffer(),
		vertexBuffer:   gl.CreateBuffer(),
	}

	ui.createGlProgram()

	img := js.Global().Get("Image").New()

	var cbFunc js.Func
	cbFunc = js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		defer cbFunc.Release()
		ui.gl.ActiveTexture(ui.gl.TEXTURE0)
		ui.gl.BindTexture(ui.gl.TEXTURE_2D, ui.texture)
		ui.gl.TexImage2D(ui.gl.TEXTURE_2D, 0, ui.gl.RGBA, ui.gl.RGBA, ui.gl.UNSIGNED_BYTE, img)
		bare := ui.gl.JS()
		bare.Call("generateMipmap", bare.Get("TEXTURE_2D"))
		ui.gl.TexParameteri(ui.gl.TEXTURE_2D, ui.gl.TEXTURE_MIN_FILTER, bare.Get("LINEAR_MIPMAP_NEAREST"))
		ui.gl.TexParameteri(ui.gl.TEXTURE_2D, ui.gl.TEXTURE_MAG_FILTER, ui.gl.LINEAR)
		ui.gl.TexParameteri(ui.gl.TEXTURE_2D, ui.gl.TEXTURE_WRAP_S, ui.gl.CLAMP_TO_EDGE)
		ui.gl.TexParameteri(ui.gl.TEXTURE_2D, ui.gl.TEXTURE_WRAP_T, ui.gl.CLAMP_TO_EDGE)
		ui.loaded = true
		return nil
	})

	img.Set("src", "./assets/VT323.png")
	img.Set("onload", cbFunc)

	return ui
}

func (c *UiContext) Draw(e *Chip8WebEmulator) {
	if !c.loaded {
		return
	}
	h := e.glContext.gl.Canvas.ClientHeight()
	textHeight := 8.0 / float32(h) * SCALE

	c.RenderText("ChipStation CHIP-8 Emulator - Press 'u' to toggle the UI", -1, 1)
	c.RenderText(fmt.Sprintf("FPS: %.2f", e.GetFps()), -1, 1-textHeight)
	c.RenderText(fmt.Sprintf("PC: 0x%04X", e.GetPc()), -1, 1-textHeight*2)
	c.RenderText(fmt.Sprintf("Opcode: 0x%04X", e.GetOpCode()), -1, 1-textHeight*3)
}

func (c *UiContext) RenderText(text string, x, y float32) {
	if !c.loaded {
		return
	}

	vertices := []float32{}
	texCoords := []float32{}

	w := (8.0 / float32(c.gl.Canvas.ClientWidth())) * SCALE
	h := (8.0 / float32(c.gl.Canvas.ClientHeight())) * SCALE

	for i, char := range text {
		char := int(char)
		tx := x + float32(i)*w
		ty := y
		coords := c.getTextureCoordinates(char)
		vertices = append(vertices,
			tx, ty,
			tx, ty-h,
			tx+w, ty,
			tx+w, ty-h)
		texCoords = append(texCoords, coords...)
	}

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
	row := float32(i / int(CHARS_PER_ROW))
	col := float32(i % int(CHARS_PER_ROW))
	u := col * CHAR_SIZE / SHEET_WIDTH
	v := row * CHAR_SIZE / SHEET_HEIGHT
	w := CHAR_SIZE / SHEET_WIDTH
	h := CHAR_SIZE / SHEET_HEIGHT
	return []float32{
		u, v,
		u, v + h,
		u + w, v,
		u + w, v + h,
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
	c.positionAttrib = c.gl.GetAttribLocation(program, "position")
	c.texCoordAttrib = c.gl.GetAttribLocation(program, "texCoord")
	c.textureUniform = c.gl.GetUniformLocation(program, "texture")
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
