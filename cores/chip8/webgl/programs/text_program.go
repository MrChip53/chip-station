//go:build js && wasm

package programs

import (
	"syscall/js"

	"github.com/seqsense/webgl-go"
)

const (
	CHARS_PER_ROW = float32(32)
	CHAR_SIZE     = float32(64)
	SHEET_HEIGHT  = CHAR_SIZE * 3
	SHEET_WIDTH   = CHAR_SIZE * CHARS_PER_ROW
	PAD_SIZE      = float32(0)
	H_PAD         = PAD_SIZE / SHEET_WIDTH
	V_PAD         = PAD_SIZE / SHEET_HEIGHT
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
        discard;
  }
}`

type TextProgram struct {
	program        webgl.Program
	texture        webgl.Texture
	vertexBuffer   webgl.Buffer
	texCoordBuffer webgl.Buffer
	texCoordAttrib int
	positionAttrib int
	textureUniform webgl.Location
	loaded         bool
}

func NewTextProgram(gl *webgl.WebGL) *TextProgram {
	c := &TextProgram{
		program:        gl.CreateProgram(),
		texture:        gl.CreateTexture(),
		vertexBuffer:   gl.CreateBuffer(),
		texCoordBuffer: gl.CreateBuffer(),
	}
	c.init(gl)
	return c
}

func (c *TextProgram) Draw(gl *webgl.WebGL, text string, x, y, size float32) {
	if !c.loaded {
		return
	}

	vertices := []float32{}
	texCoords := []float32{}

	w := (CHAR_SIZE / float32(gl.Canvas.ClientWidth())) * size
	h := (CHAR_SIZE / float32(gl.Canvas.ClientHeight())) * size

	for i, char := range text {
		char := int(char)
		tx := (x + float32(i)*w) - (w/2.5)*float32(i)
		ty := y
		coords := c.getTextureCoordinates(char)
		vertices = append(vertices,
			tx, ty,
			tx, ty-h,
			tx+w, ty,

			tx+w, ty,
			tx, ty-h,
			tx+w, ty-h)
		texCoords = append(texCoords, coords...)
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, c.vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ARRAY_BUFFER, c.texCoordBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, webgl.Float32ArrayBuffer(texCoords), gl.STATIC_DRAW)

	gl.UseProgram(c.program)

	gl.BindBuffer(gl.ARRAY_BUFFER, c.vertexBuffer)
	gl.VertexAttribPointer(c.positionAttrib, 2, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(c.positionAttrib)

	gl.BindBuffer(gl.ARRAY_BUFFER, c.texCoordBuffer)
	gl.VertexAttribPointer(c.texCoordAttrib, 2, gl.FLOAT, false, 0, 0)
	gl.EnableVertexAttribArray(c.texCoordAttrib)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, c.texture)
	gl.Uniform1i(c.textureUniform, 0)

	gl.DrawArrays(gl.TRIANGLES, 0, len(vertices)/2)
}

func (c *TextProgram) init(gl *webgl.WebGL) {
	var err error
	var vs, fs webgl.Shader
	if vs, err = initVertexShader(gl, vTextShader); err != nil {
		panic(err)
	}

	if fs, err = initFragmentShader(gl, fTextShader); err != nil {
		panic(err)
	}

	program, err := linkShaders(gl, nil, vs, fs)
	if err != nil {
		panic(err)
	}

	c.program = program
	c.positionAttrib = gl.GetAttribLocation(c.program, "position")
	c.texCoordAttrib = gl.GetAttribLocation(c.program, "texCoord")
	c.textureUniform = gl.GetUniformLocation(c.program, "texture")

	img := js.Global().Get("Image").New()

	var cbFunc js.Func
	cbFunc = js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		defer cbFunc.Release()
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, c.texture)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, img)
		bare := gl.JS()
		bare.Call("generateMipmap", bare.Get("TEXTURE_2D"))
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, bare.Get("LINEAR_MIPMAP_LINEAR"))
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		c.loaded = true
		return nil
	})

	img.Set("src", "./assets/font.png")
	img.Set("onload", cbFunc)
}

func (c *TextProgram) getTextureCoordinates(charCode int) []float32 {
	i := charCode - 32
	row := float32(i / int(CHARS_PER_ROW))
	col := float32(i % int(CHARS_PER_ROW))
	u := col * CHAR_SIZE / SHEET_WIDTH
	v := row * CHAR_SIZE / SHEET_HEIGHT
	w := CHAR_SIZE / SHEET_WIDTH
	h := CHAR_SIZE / SHEET_HEIGHT
	hPad := float32(0.25 / SHEET_WIDTH)
	vPad := float32(0.25 / SHEET_HEIGHT)
	return []float32{
		u + hPad, v + vPad,
		u + hPad, v - vPad + h,
		u - hPad + w, v + vPad,

		u - hPad + w, v + vPad,
		u + hPad, v - vPad + h,
		u - hPad + w, v - vPad + h,
	}
}
