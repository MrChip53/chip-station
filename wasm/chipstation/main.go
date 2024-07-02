//go:build js && wasm

package main

import (
	_ "embed"
	"fmt"
	"syscall/js"

	webgl "github.com/seqsense/webgl-go"

	"github.com/mrchip53/chip-station/cores/chip8"
	"github.com/mrchip53/chip-station/cores/chip8/webgl"
	"github.com/mrchip53/chip-station/utilities"
)

var done chan struct{}

var (
	gl          *webgl.WebGL
	opcodeSpan  js.Value
	pcSpan      js.Value
	fpsSpan     js.Value
	romSizeSpan js.Value
	width       float64
	height      float64
)

var e *chip8web.Chip8WebEmulator

//go:embed 5-quirks.ch8
var rom4 []byte

func main() {
	var err error

	js.Global().Set("setKeyState", js.FuncOf(setKeyState))
	js.Global().Set("loadRom", js.FuncOf(loadRom))
	js.Global().Set("getRom", js.FuncOf(getRom))
	js.Global().Set("setIpf", js.FuncOf(setIpf))
	opcodeSpan = js.Global().Get("document").Call("getElementById", "opcode")
	pcSpan = js.Global().Get("document").Call("getElementById", "pc")
	fpsSpan = js.Global().Get("document").Call("getElementById", "fps")
	romSizeSpan = js.Global().Get("document").Call("getElementById", "romSize")
	canvas := js.Global().Get("document").Call("getElementById", "screen")

	gl, err = webgl.New(canvas)
	if err != nil {
		panic(err)
	}
	e = chip8web.NewChip8WebEmulator(gl)

	go runGameLoop()
	<-done
}

func setIpf(this js.Value, p []js.Value) interface{} {
	ipf := p[0].Int()
	e.SetIPF(ipf)
	return nil
}

func getRom(this js.Value, p []js.Value) interface{} {
	rom := e.GetRom()
	romBytes := js.Global().Get("Uint8Array").New(len(rom))
	js.CopyBytesToJS(romBytes, rom)
	return romBytes
}

func setKeyState(this js.Value, p []js.Value) interface{} {
	key := p[0].Int()
	state := p[1].Int()
	e.SetKeyState(uint8(key), uint8(state))
	return nil
}

func loadRom(this js.Value, p []js.Value) interface{} {
	romBytes := p[0]
	length := romBytes.Get("length").Int()
	rom := make([]byte, length)
	js.CopyBytesToGo(rom, romBytes)
	e.Stop()
	e.LoadROM([]byte(rom))
	e.Start()
	romSizeSpan.Set("innerText", fmt.Sprintf("%d bytes", length))
	return nil
}

func runGameLoop() {
	e.Initialize(chip8.Hooks{
		Decode: func(pc uint16, opcode uint16, drawCount uint64) bool {
			opcodeSpan.Set("innerText", utilities.Hex(opcode))
			pcSpan.Set("innerText", utilities.Hex(pc))
			return false
		},
		Draw: func(drawCount uint64, fps float64) {
			fpsSpan.Set("innerText", fmt.Sprintf("%.2f", fps))
			e.Draw()
		},
	})
	// e.SetMemory(0x1ff, []byte{1})
	e.LoadROM(rom4)
	romSizeSpan.Set("innerText", fmt.Sprintf("%d bytes", len(rom4)))
	e.Loop()
}
