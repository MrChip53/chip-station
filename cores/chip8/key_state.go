package chip8

type KeyState struct {
	keys            [NUM_KEYS]bool
	lastKeyReleased uint8
}

func NewKeyState() *KeyState {
	return &KeyState{}
}

func (k *KeyState) IsKeyPressed(key uint8) bool {
	return k.keys[key]
}

func (k *KeyState) SetKeyState(key uint8, state bool) {
	k.keys[key] = state
	if !state {
		k.lastKeyReleased = key
	}
}

func (k *KeyState) GetLastKeyReleased() uint8 {
	return k.lastKeyReleased
}

func (k *KeyState) ResetLastKeyReleased() {
	k.lastKeyReleased = 0xFF
}

func (k *KeyState) Reset() {
	for i := range k.keys {
		k.keys[i] = false
	}
	k.ResetLastKeyReleased()
}
