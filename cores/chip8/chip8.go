package chip8

import (
	_ "embed"
	"log"
	"time"

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
	IPS               = 700
)

//go:embed font.bin
var defaultFont []byte

type (
	DecodeHook func(opcode uint16, drawCount uint64) bool
	DrawHook   func(display [SCREEN_WIDTH][SCREEN_HEIGHT]uint8, drawCount uint64)
)

type Hooks struct {
	Decode DecodeHook
	Draw   DrawHook
}

type Chip8Emulator struct {
	memory     [MEMORY_SIZE]byte
	display    [SCREEN_WIDTH][SCREEN_HEIGHT]uint8
	stack      *utilities.Stack
	soundTimer uint8
	delayTimer uint8
	keyState   [NUM_KEYS]uint8
	hooks      Hooks

	cycleCount uint64
	drawCount  uint64

	pc uint16
	i  uint16
	v  [NUM_REGISTERS]uint8

	// Testing stuff
	abort bool
}

func (e *Chip8Emulator) Initialize(hooks Hooks) {
	e.pc = ROM_START_ADDRESS
	copy(e.memory[:], defaultFont)
	e.stack = utilities.NewStack(16)
	e.hooks = hooks
}

func (e *Chip8Emulator) Loop() {
	for {
		if e.abort {
			break
		}
		e.cycle()
		time.Sleep(time.Second / IPS)
	}
}

func (e *Chip8Emulator) LoadROM(rom []byte) {
	copy(e.memory[ROM_START_ADDRESS:], rom)
}

func (e *Chip8Emulator) cycle() {
	opcode := e.fetch()
	e.decode(opcode)
	e.cycleCount++
}

func (e *Chip8Emulator) fetch() uint16 {
	b1 := uint16(e.memory[e.pc])
	b2 := uint16(e.memory[e.pc+1])
	e.pc += 2
	return b1<<8 | b2
}

func (e *Chip8Emulator) decode(opcode uint16) {
	op := opcode & 0xF000 >> 12
	x := (opcode & 0x0F00) >> 8
	y := (opcode & 0x00F0) >> 4
	n := opcode & 0x000F
	nn := opcode & 0x00FF
	nnn := opcode & 0x0FFF

	e.abort = e.hooks.Decode != nil && e.hooks.Decode(opcode, e.drawCount)

	if op == 0x0 {
		if opcode == 0x00E0 {
			e.clearDisplay()
		} else if opcode == 0x00EE {
			e.pc = e.stack.Pop()
		} else {
			log.Printf("op: %x, x: %x, y: %x, n: %x, nn: %x, nnn: %x, cycle: %d\n", op, x, y, n, nn, nnn, e.cycleCount)
			panic("not implemented")
		}
		return
	}

	switch op {
	case 0x1:
		e.pc = nnn
	case 0x2:
		e.stack.Push(e.pc)
		e.pc = nnn
	case 0x3:
		if e.v[x] == uint8(nn) {
			e.pc += 2
		}
	case 0x4:
		if e.v[x] != uint8(nn) {
			e.pc += 2
		}
	case 0x5:
		if e.v[x] == e.v[y] {
			e.pc += 2
		}
	case 0x6:
		e.v[x] = uint8(nn)
	case 0x7:
		e.v[x] += uint8(nn)
	case 0x8:
		switch n {
		case 0x0:
			e.v[x] = e.v[y]
		case 0x1:
			e.v[x] |= e.v[y]
		case 0x2:
			e.v[x] &= e.v[y]
		case 0x3:
			e.v[x] ^= e.v[y]
		case 0x4:
			e.v[x] += e.v[y]
		case 0x5:
			e.v[x] -= e.v[y]
		case 0x6:
			e.v[x] = e.v[y] >> 1
		case 0x7:
			e.v[x] = e.v[y] - e.v[x]
		case 0xE:
			e.v[x] = e.v[y] << 1
		}
	case 0x9:
		if e.v[x] != e.v[y] {
			e.pc += 2
		}
	case 0xA:
		e.i = nnn
	case 0xD:
		x := uint16(e.v[x] % SCREEN_WIDTH)
		y := uint16(e.v[y] % SCREEN_HEIGHT)
		e.v[0xF] = 0
		for i := uint16(0); i < n; i++ {
			sprite := e.memory[e.i+i]
			for j := uint16(0); j < 8; j++ {
				if (sprite & (0x80 >> j)) != 0 {
					if e.display[x+j][y+i] == 1 {
						e.v[0xF] = 1
					}
					e.display[x+j][y+i] ^= 1
				}
			}
		}
		if e.hooks.Draw != nil {
			e.hooks.Draw(e.display, e.drawCount)
		}
		e.drawCount++
	case 0xF:
		switch nn {
		case 0x1E:
			e.i += uint16(e.v[x])
		case 0x33:
			e.memory[e.i] = e.v[x] / 100
			e.memory[e.i+1] = (e.v[x] / 10) % 10
			e.memory[e.i+2] = e.v[x] % 10
		case 0x55:
			for i := 0; i <= int(x); i++ {
				e.memory[e.i+uint16(i)] = e.v[i]
			}
		case 0x65:
			for i := 0; i <= int(x); i++ {
				e.v[i] = e.memory[e.i+uint16(i)]
			}
		}
	default:
		log.Printf("op: %x, x: %x, y: %x, n: %x, nn: %x, nnn: %x, cycle: %d\n", op, x, y, n, nn, nnn, e.cycleCount)
		panic("opcode not implemented")
	}
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

func (e *Chip8Emulator) GetDisplay() [SCREEN_WIDTH][SCREEN_HEIGHT]uint8 {
	return e.display
}

func (e *Chip8Emulator) ScreenshotPNG(filename string) {
	utilities.SavePNG(e.display, filename)
}
