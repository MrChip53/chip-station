package chip8

type FpsCounter struct {
	fps  float64
	last float64
}

func NewFpsCounter() *FpsCounter {
	return &FpsCounter{}
}

func (f *FpsCounter) UpdateFps(now float64) {
	now *= 0.001
	d := now - f.last
	f.last = now
	f.fps = 1 / d
}

func (f *FpsCounter) Reset() {
	f.fps = 0
	f.last = 0
}

func (f *FpsCounter) GetFps() float64 {
	return f.fps
}
