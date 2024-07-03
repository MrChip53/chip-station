package chip8

type Message interface {
	HandleMessage(e *Chip8Emulator)
	IsCustom() bool
}

type CustomMessage struct{}

func (m CustomMessage) IsCustom() bool {
	return true
}

func (m CustomMessage) HandleMessage(e *Chip8Emulator) {}

type BaseMessage struct {
	custom bool
}

func (m BaseMessage) IsCustom() bool {
	return m.custom
}

type PauseMessage struct {
	BaseMessage
}

func (m PauseMessage) HandleMessage(e *Chip8Emulator) {
	e.hang()
}

type SwapRomMessage struct {
	BaseMessage
	rom []byte
}

func (m SwapRomMessage) HandleMessage(e *Chip8Emulator) {
	e.loadRom(m.rom)
	e.reset()
}

type IpfMessage struct {
	BaseMessage
	ipf int
}

func (m IpfMessage) HandleMessage(e *Chip8Emulator) {
	e.ipf = m.ipf
}

type SetMemoryMessage struct {
	BaseMessage
	address uint16
	data    []byte
}

func (m SetMemoryMessage) HandleMessage(e *Chip8Emulator) {
	copy(e.memory[m.address:], m.data)
}

type KeyStateMessage struct {
	BaseMessage
	key   uint8
	state bool
}

func (m KeyStateMessage) HandleMessage(e *Chip8Emulator) {
	e.keyState.SetKeyState(m.key, m.state)
}
