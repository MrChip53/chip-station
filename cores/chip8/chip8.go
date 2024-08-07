package chip8

import (
	_ "embed"
	"log"
	"time"

	"github.com/mrchip53/chip-station/utilities"
)

const (
	ROM_START_ADDRESS  = 0x200
	MEMORY_SIZE        = 4096
	SCREEN_WIDTH       = 64
	SCREEN_HEIGHT      = 32
	NUM_KEYS           = 16
	NUM_REGISTERS      = 16
	IPF                = 20
	MESSAGES_PER_FRAME = 20
)

//go:embed font.bin
var defaultFont []byte

type (
	DecodeHook        func(pc uint16, opcode uint16, drawCount uint64) bool
	DrawHook          func()
	SoundHook         func()
	CustomMessageHook func(m Message)
	Display           [SCREEN_WIDTH][SCREEN_HEIGHT]uint8
)

type Hooks struct {
	Decode        DecodeHook
	Draw          DrawHook
	PlaySound     SoundHook
	StopSound     SoundHook
	CustomMessage CustomMessageHook
}

type Chip8Emulator struct {
	memory [MEMORY_SIZE]byte

	display Display
	draw    bool

	stack      *utilities.Stack
	soundTimer *SoundTimer
	delayTimer *DelayTimer
	hooks      Hooks
	fps        *FpsCounter
	keyState   *KeyState

	messageChan chan Message

	ipf int

	pc uint16
	i  uint16
	v  [NUM_REGISTERS]uint8

	lastRomSize int

	cycleCount uint64
	drawCount  uint64
	paused     bool
}

func NewChip8Emulator(hooks Hooks) *Chip8Emulator {
	e := &Chip8Emulator{
		delayTimer:  NewDelayTimer(),
		fps:         NewFpsCounter(),
		keyState:    NewKeyState(),
		soundTimer:  NewSoundTimer(),
		messageChan: make(chan Message, 20),
		pc:          ROM_START_ADDRESS,
		stack:       utilities.NewStack(16),
		ipf:         IPF,
		hooks:       hooks,
	}
	copy(e.memory[:], defaultFont)
	e.messageChan <- PauseMessage{}
	return e
}

func (e *Chip8Emulator) Start() {
	e.reset()
}

func (e *Chip8Emulator) reset() {
	e.hooks.StopSound()
	e.display = Display{}
	e.stack = utilities.NewStack(16)
	e.soundTimer.Reset()
	e.delayTimer.Reset()
	e.keyState.Reset()
	e.pc = ROM_START_ADDRESS
	e.i = 0
	e.v = [NUM_REGISTERS]uint8{}
	e.paused = false
	e.fps.Reset()
}

func (e *Chip8Emulator) IsPaused() bool {
	return e.paused
}

func (e *Chip8Emulator) Cycle(now float64) bool {
	for i := 0; i < MESSAGES_PER_FRAME; i++ {
		select {
		case m := <-e.messageChan:
			if !m.IsCustom() {
				m.HandleMessage(e)
			} else {
				if e.hooks.CustomMessage != nil {
					e.hooks.CustomMessage(m)
				}
			}
		default:
			break
		}
	}

	if e.hooks.Draw != nil {
		e.hooks.Draw()
	}
	e.drawCount++

	if e.paused {
		return true
	}

	for i := 0; i < e.ipf; i++ {
		opcode, ok := e.cycle()
		if !ok {
			return false
		}
		if opcode&0xF000 == 0xD000 {
			break
		}
	}
	e.draw = false

	e.delayTimer.Decrement()
	e.soundTimer.Decrement(e.hooks.StopSound)
	e.keyState.ResetLastKeyReleased()
	e.fps.UpdateFps(now)

	return true
}

func (e *Chip8Emulator) Loop() {
DrawLoop:
	for {
		start := time.Now()

	MessageLoop:
		for i := 0; i < MESSAGES_PER_FRAME; i++ {
			select {
			case m := <-e.messageChan:
				if !m.IsCustom() {
					m.HandleMessage(e)
				} else {
					if e.hooks.CustomMessage != nil {
						e.hooks.CustomMessage(m)
					}
				}
			default:
				break MessageLoop
			}
		}

		if e.hooks.Draw != nil {
			e.hooks.Draw()
		}
		e.drawCount++
		for i := 0; i < e.ipf; i++ {
			opcode, ok := e.cycle()
			if !ok {
				break DrawLoop
			}
			if opcode&0xF000 == 0xD000 {
				break
			}
		}
		e.draw = false

		e.delayTimer.Decrement()
		e.soundTimer.Decrement(e.hooks.StopSound)
		e.keyState.ResetLastKeyReleased()

		elapsed := time.Since(start)
		time.Sleep(time.Second/60 - elapsed)
	}
}

func (e *Chip8Emulator) wipeRom() {
	for i := ROM_START_ADDRESS; i < MEMORY_SIZE; i++ {
		e.memory[i] = 0
	}
}

func (e *Chip8Emulator) loadRom(rom []byte) {
	e.wipeRom()
	copy(e.memory[ROM_START_ADDRESS:], rom)
	e.lastRomSize = len(rom)
}

func (e *Chip8Emulator) GetRomSize() int {
	return e.lastRomSize
}

func (e *Chip8Emulator) EnqueueMessage(m Message) {
	e.messageChan <- m
}

func (e *Chip8Emulator) Pause() {
	e.EnqueueMessage(PauseMessage{})
}

func (e *Chip8Emulator) Resume() {
	e.EnqueueMessage(ResumeMessage{})
}

func (e *Chip8Emulator) SwapROM(rom []byte) {
	e.EnqueueMessage(SwapRomMessage{rom: rom})
}

func (e *Chip8Emulator) pause() {
	e.paused = true
	if e.hooks.StopSound != nil {
		e.hooks.StopSound()
	}
}

func (e *Chip8Emulator) resume() {
	e.paused = false
	e.soundTimer.Resume(e.hooks.PlaySound)
}

func (e *Chip8Emulator) cycle() (uint16, bool) {
	opcode := e.fetch()
	abort := e.decode(opcode)
	e.cycleCount++
	return opcode, !abort
}

func (e *Chip8Emulator) fetch() uint16 {
	b1 := uint16(e.memory[e.pc])
	b2 := uint16(e.memory[e.pc+1])
	e.pc += 2
	return b1<<8 | b2
}

func (e *Chip8Emulator) decode(opcode uint16) bool {
	if e.hooks.Decode != nil && e.hooks.Decode(e.pc-2, opcode, e.drawCount) {
		return true
	}

	opKey := getOpKey(opcode)

	instruction, ok := instructions[opKey]
	if !ok {
		log.Printf("op: %x, cycle: %d\n", opcode, e.cycleCount)
		panic("opcode not implemented")
	}
	instruction.Fill(opcode)
	instruction.Execute(e)

	return false
}

func (e *Chip8Emulator) execute(opcode uint16) {
}

func (e *Chip8Emulator) clearDisplay() {
	for i := 0; i < SCREEN_WIDTH; i++ {
		for j := 0; j < SCREEN_HEIGHT; j++ {
			e.display[i][j] = 0
		}
	}
}

func (e *Chip8Emulator) GetDisplay() Display {
	return e.display
}

func (e *Chip8Emulator) GetRom() []byte {
	return e.memory[ROM_START_ADDRESS : ROM_START_ADDRESS+uint16(e.lastRomSize)]
}

func (e *Chip8Emulator) SetMemory(address uint16, data []byte) {
	e.EnqueueMessage(SetMemoryMessage{address: address, data: data})
}

func (e *Chip8Emulator) SetKeyState(key, state uint8) {
	e.EnqueueMessage(KeyStateMessage{key: key, state: state == 1})
}

func (e *Chip8Emulator) SetIPF(ipf int) {
	e.EnqueueMessage(IpfMessage{ipf: ipf})
}

func (e *Chip8Emulator) GetIPF() int {
	return e.ipf
}

func (e *Chip8Emulator) GetFps() float64 {
	return e.fps.GetFps()
}

func (e *Chip8Emulator) GetPc() uint16 {
	return e.pc
}

func (e *Chip8Emulator) GetOpCode() uint16 {
	return uint16(e.memory[e.pc])<<8 | uint16(e.memory[e.pc+1])
}

func (e *Chip8Emulator) ResetFps() {
	e.fps.Reset()
}
