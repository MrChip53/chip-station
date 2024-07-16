//go:build js && wasm

package programs

import (
	"errors"
	"fmt"
	"syscall/js"

	"github.com/seqsense/webgl-go"
)

func createGlProgram(gl *webgl.WebGL) webgl.Program {
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

	return program
}

func initVertexShader(gl *webgl.WebGL, src string) (webgl.Shader, error) {
	s := gl.CreateShader(gl.VERTEX_SHADER)
	gl.ShaderSource(s, src)
	gl.CompileShader(s)
	if !gl.GetShaderParameter(s, gl.COMPILE_STATUS).(bool) {
		compilationLog := gl.GetShaderInfoLog(s)
		return webgl.Shader(js.Null()), fmt.Errorf("compile failed (VERTEX_SHADER) %v", compilationLog)
	}
	return s, nil
}

func initFragmentShader(gl *webgl.WebGL, src string) (webgl.Shader, error) {
	s := gl.CreateShader(gl.FRAGMENT_SHADER)
	gl.ShaderSource(s, src)
	gl.CompileShader(s)
	if !gl.GetShaderParameter(s, gl.COMPILE_STATUS).(bool) {
		compilationLog := gl.GetShaderInfoLog(s)
		return webgl.Shader(js.Null()), fmt.Errorf("compile failed (FRAGMENT_SHADER) %v", compilationLog)
	}
	return s, nil
}

func linkShaders(gl *webgl.WebGL, fbVarings []string, shaders ...webgl.Shader) (webgl.Program, error) {
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

func uniform2f(gl *webgl.WebGL, loc webgl.Location, x, y float32) {
	gl.JS().Call("uniform2f", js.Value(loc), x, y)
}
