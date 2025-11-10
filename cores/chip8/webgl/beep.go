//go:build js && wasm

package chip8web

import "syscall/js"

type Beep struct {
	audio              js.Value
	timeUpdateListener js.Func
}

func NewBeep() *Beep {
	return NewBeepWithSource("beep.ogg")
}

func NewBeepWithSource(beepSource string) *Beep {
	audio := js.Global().Get("document").Call("createElement", "audio")
	audio.Set("src", beepSource)
	audio.Set("preload", "auto")
	audio.Set("volume", 0.5)

	timeUpdateListener := js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		buffer := 0.5
		currentTime := this.Get("currentTime").Float()
		duration := this.Get("duration").Float()
		if currentTime > duration-buffer {
			this.Set("currentTime", 0)
			this.Call("play")
		}
		return nil
	})

	audio.Call("addEventListener", "timeupdate", timeUpdateListener)

	return &Beep{
		audio:              audio,
		timeUpdateListener: timeUpdateListener,
	}
}

func (b *Beep) Play() {
	b.audio.Call("play")
}

func (b *Beep) Stop() {
	b.audio.Call("pause")
	b.audio.Set("currentTime", 0)
}
