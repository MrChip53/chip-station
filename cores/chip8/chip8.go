package chip8

import (
	"github.com/mrchip53/chip-station/utilities"
)

/**
 * NOTES:
 * Run at ~700 IPS
 */

const (
	ROM_START_ADDRESS = 0x200
	MEMORY_SIZE       = 4096
	SCREEN_WIDTH      = 64
	SCREEN_HEIGHT     = 32
	NUM_KEYS          = 16
	NUM_REGISTERS     = 16
)

type Chip8Emulator struct {
	memory     [MEMORY_SIZE]byte
	display    [SCREEN_WIDTH][SCREEN_HEIGHT]uint8
	stack      *utilities.Stack
	soundTimer uint8
	delayTimer uint8
	keyState   [NUM_KEYS]uint8

	pc uint16
	i  uint16
	v  [NUM_REGISTERS]uint8
}

func (e *Chip8Emulator) Initialize() {
	e.pc = ROM_START_ADDRESS
	// Load fontset
}

func (e *Chip8Emulator) Loop() {
}

func (e *Chip8Emulator) LoadROM(rom []byte) {
	copy(e.memory[ROM_START_ADDRESS:], rom)
}

func (e *Chip8Emulator) cycle() {
}

func (e *Chip8Emulator) fetch() uint16 {
	b1 := uint16(e.memory[e.pc])
	b2 := uint16(e.memory[e.pc+1])
	e.pc += 2
	return b1<<8 | b2
}

func (e *Chip8Emulator) decode(opcode uint16) {
}

func (e *Chip8Emulator) execute(opcode uint16) {
}
