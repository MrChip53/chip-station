//go:build js && wasm

package main

import (
	_ "embed"
	"fmt"
	"syscall/js"
	"time"

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

var csRom = []byte{
	0x63, 0x08, 0x81, 0x30, 0x62, 0x04, 0xA2, 0x4C, 0xD1, 0x2A, 0x71, 0x07, 0xA2, 0x56, 0xD1, 0x2A,
	0x71, 0x07, 0xA2, 0x60, 0xD1, 0x2A, 0x71, 0x03, 0xA2, 0x6A, 0xD1, 0x2D, 0x72, 0x0B, 0x81, 0x30,
	0x71, 0x05, 0xA2, 0x77, 0xD1, 0x2A, 0x71, 0x07, 0xA2, 0x81, 0xD1, 0x2A, 0x71, 0x07, 0xA2, 0x8B,
	0xD1, 0x2A, 0x71, 0x07, 0xA2, 0x81, 0xD1, 0x2A, 0x71, 0x07, 0xA2, 0x60, 0xD1, 0x2A, 0x71, 0x03,
	0xA2, 0x95, 0xD1, 0x2A, 0x71, 0x07, 0xA2, 0x9F, 0xD1, 0x2A, 0x12, 0x4A, 0xFC, 0xFC, 0xC0, 0xC0,
	0xC0, 0xC0, 0xC0, 0xC0, 0xFC, 0xFC, 0xC0, 0xC0, 0xC0, 0xC0, 0xFC, 0xFC, 0xCC, 0xCC, 0xCC, 0xCC,
	0xC0, 0xC0, 0x00, 0x00, 0xC0, 0xC0, 0xC0, 0xC0, 0xC0, 0xC0, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF,
	0xC3, 0xC3, 0xFF, 0xFF, 0xC0, 0xC0, 0xC0, 0xFC, 0xFC, 0xC0, 0xC0, 0xFC, 0xFC, 0x0C, 0x0C, 0xFC,
	0xFC, 0x00, 0x00, 0x30, 0x30, 0xFC, 0xFC, 0x30, 0x30, 0x30, 0x30, 0x00, 0x00, 0xF8, 0xFC, 0x0C,
	0x7C, 0xFC, 0xCC, 0xFC, 0x7C, 0x00, 0x00, 0x00, 0x00, 0xFC, 0xFC, 0xCC, 0xCC, 0xFC, 0xFC, 0x00,
	0x00, 0x00, 0x00, 0xFC, 0xFC, 0xCC, 0xCC, 0xCC, 0xCC,
}

func main() {
	var err error

	canvas := js.Global().Get("document").Call("getElementById", "screen")
	gl, err = webgl.New(canvas)
	if err != nil {
		panic(err)
	}

	e = chip8web.NewChip8WebEmulator(gl)
	go runGameLoop()

	emulatorObj := js.Global().Get("Object").New()
	emulatorObj.Set("setKeyState", js.FuncOf(setKeyState))
	emulatorObj.Set("loadRom", js.FuncOf(loadRom))
	emulatorObj.Set("getRom", js.FuncOf(getRom))
	emulatorObj.Set("setIpf", js.FuncOf(setIpf))
	emulatorObj.Set("pause", js.FuncOf(pause))
	emulatorObj.Set("resume", js.FuncOf(resume))
	emulatorObj.Set("isPaused", js.FuncOf(isPaused))
	emulatorObj.Set("setOnColor", js.FuncOf(setOnColor))
	emulatorObj.Set("setOffColor", js.FuncOf(setOffColor))
	js.Global().Set("emulator", emulatorObj)

	opcodeSpan = js.Global().Get("document").Call("getElementById", "opcode")
	pcSpan = js.Global().Get("document").Call("getElementById", "pc")
	fpsSpan = js.Global().Get("document").Call("getElementById", "fps")
	romSizeSpan = js.Global().Get("document").Call("getElementById", "romSize")

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

func pause(this js.Value, p []js.Value) interface{} {
	e.Pause()
	return nil
}

func resume(this js.Value, p []js.Value) interface{} {
	e.Resume()
	return nil
}

func isPaused(this js.Value, p []js.Value) interface{} {
	return e.IsPaused()
}

func setOnColor(this js.Value, p []js.Value) interface{} {
	rgb := p[0].Int()
	e.SetOnColor(chip8web.NewColor(uint32(rgb)))
	return nil
}

func setOffColor(this js.Value, p []js.Value) interface{} {
	rgb := p[0].Int()
	e.SetOffColor(chip8web.NewColor(uint32(rgb)))
	return nil
}

func loadRom(this js.Value, p []js.Value) interface{} {
	romBytes := p[0]
	length := romBytes.Get("length").Int()
	rom := make([]byte, length)
	js.CopyBytesToGo(rom, romBytes)
	e.SwapROM([]byte(rom))
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
		PlaySound: func() {
			e.PlayBeep()
		},
		StopSound: func() {
			e.StopBeep()
		},
		CustomMessage: func(m chip8.Message) {
			switch m := m.(type) {
			case chip8web.Message:
				m.Handle(e)
			}
		},
	})
	// e.SetMemory(0x1ff, []byte{1})
	romSizeSpan.Set("innerText", fmt.Sprintf("%d bytes", len(csRom)))
	e.SwapROM(csRom)
	go func() {
		time.Sleep(100 * time.Millisecond)
		e.Resume()
	}()
	e.Loop()
}
