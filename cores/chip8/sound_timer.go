package chip8

type SoundTimer struct {
	timer uint8
}

func NewSoundTimer() *SoundTimer {
	return &SoundTimer{}
}

func (s *SoundTimer) SetTimer(timer uint8, hook SoundHook) {
	s.timer = timer
	if s.timer > 0 && hook != nil {
		hook()
	}
}

func (s *SoundTimer) GetTimer() uint8 {
	return s.timer
}

func (s *SoundTimer) Decrement(hook SoundHook) {
	if s.timer > 0 {
		s.timer--
		if s.timer == 0 && hook != nil {
			hook()
		}
	}
}

func (s *SoundTimer) Resume(hook SoundHook) {
	if s.timer > 0 && hook != nil {
		hook()
	}
}

func (s *SoundTimer) Reset() {
	s.timer = 0
}
