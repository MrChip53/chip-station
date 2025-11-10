//go:build js && wasm

package programs

import "github.com/seqsense/webgl-go"

type Programs struct {
	DisplayProgram *DisplayProgram
	WindowProgram  *WindowProgram
	TextProgram    *TextProgram
}

func NewPrograms(gl *webgl.WebGL, fontSource string) *Programs {
	return &Programs{
		DisplayProgram: NewDisplayProgram(gl),
		WindowProgram:  NewWindowProgram(gl),
		TextProgram:    NewTextProgramWithFontSource(gl, fontSource),
	}
}
