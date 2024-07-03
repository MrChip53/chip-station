package chip8

type DelayTimer struct {
	timer uint8
}

func NewDelayTimer() *DelayTimer {
	return &DelayTimer{}
}

func (d *DelayTimer) Decrement() {
	if d.timer > 0 {
		d.timer--
	}
}

func (d *DelayTimer) SetTimer(t uint8) {
	d.timer = t
}

func (d *DelayTimer) GetTimer() uint8 {
	return d.timer
}

func (d *DelayTimer) Reset() {
	d.timer = 0
}
