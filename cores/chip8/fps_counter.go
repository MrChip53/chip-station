package chip8

import "time"

type FpsCounter struct {
	frameCount   uint64
	lastTime     time.Time
	savedElapsed float64
}

func NewFpsCounter() *FpsCounter {
	return &FpsCounter{}
}

func (f *FpsCounter) Inc() {
	f.frameCount++
}

func (f *FpsCounter) Reset() {
	f.frameCount = 0
	f.lastTime = time.Now()
	f.savedElapsed = 0
}

func (e *FpsCounter) Pause() {
	e.savedElapsed = e.savedElapsed + time.Since(e.lastTime).Seconds()
}

func (e *FpsCounter) Resume() {
	e.lastTime = time.Now()
}

func (f *FpsCounter) GetFps() float64 {
	elapsed := time.Since(f.lastTime).Seconds()
	return float64(f.frameCount) / (elapsed + f.savedElapsed)
}
