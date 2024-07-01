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
	IPF               = 12
)

//go:embed font.bin
var defaultFont []byte

type (
	DecodeHook func(pc uint16, opcode uint16, drawCount uint64) bool
	DrawHook   func(drawCount uint64, fps float64)
)

type Hooks struct {
	Decode DecodeHook
	Draw   DrawHook
}

type Chip8Emulator struct {
	memory          [MEMORY_SIZE]byte
	display         [SCREEN_WIDTH][SCREEN_HEIGHT]uint8
	stack           *utilities.Stack
	soundTimer      uint8
	delayTimer      uint8
	keyState        [NUM_KEYS]uint8
	lastKeyReleased uint8
	hooks           Hooks
	draw            bool

	cycleCount uint64
	drawCount  uint64
	frameCount uint64
	start      time.Time

	pc uint16
	i  uint16
	v  [NUM_REGISTERS]uint8

	lastRomSize int

	// Testing stuff
	abort   bool
	stopped bool
}

func (e *Chip8Emulator) Initialize(hooks Hooks) {
	e.pc = ROM_START_ADDRESS
	copy(e.memory[:], defaultFont)
	e.stack = utilities.NewStack(16)
	e.hooks = hooks
	e.lastKeyReleased = 0xFF
}

func (e *Chip8Emulator) Start() {
	e.display = [SCREEN_WIDTH][SCREEN_HEIGHT]uint8{}
	e.stack = utilities.NewStack(16)
	e.pc = ROM_START_ADDRESS
	e.stopped = false
}

func (e *Chip8Emulator) Stop() {
	e.stopped = true
}

func (e *Chip8Emulator) Loop() {
	e.start = time.Now()
	for {
		if e.abort {
			break
		}

		if e.stopped {
			time.Sleep(time.Second / 60)
			continue
		}

		start := time.Now()

		if start.Sub(e.start) > time.Second*15 {
			e.start = time.Now()
			e.frameCount = 0
		}

		if e.hooks.Draw != nil {
			fps := float64(e.frameCount) / time.Since(e.start).Seconds()
			e.hooks.Draw(e.drawCount, fps)
		}
		e.drawCount++
		for i := 0; i < IPF; i++ {
			if e.abort || e.draw {
				break
			}
			e.cycle()
		}
		e.draw = false
		if e.delayTimer > 0 {
			e.delayTimer--
		}
		if e.soundTimer > 0 {
			e.soundTimer--
		}
		e.lastKeyReleased = 0xFF
		e.frameCount++
		elapsed := time.Since(start)
		time.Sleep(time.Second/60 - elapsed)
	}
}

func (e *Chip8Emulator) LoadROM(rom []byte) {
	copy(e.memory[ROM_START_ADDRESS:], rom)
	e.lastRomSize = len(rom)
}

func (e *Chip8Emulator) GetRomSize() int {
	return e.lastRomSize
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

	e.abort = e.hooks.Decode != nil && e.hooks.Decode(e.pc-2, opcode, e.drawCount)

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
			e.v[0xF] = 0
		case 0x2:
			e.v[x] &= e.v[y]
			e.v[0xF] = 0
		case 0x3:
			e.v[x] ^= e.v[y]
			e.v[0xF] = 0
		case 0x4:
			c := 0
			if int(e.v[x])+int(e.v[y]) > 255 {
				c = 1
			}
			e.v[x] += e.v[y]
			e.v[0xF] = uint8(c)
		case 0x5:
			c := 1
			if e.v[y] > e.v[x] {
				c = 0
			}
			e.v[x] -= e.v[y]
			e.v[0xF] = uint8(c)
		case 0x6:
			c := e.v[x] & 0x01
			e.v[x] = e.v[y] >> 1
			e.v[0xF] = uint8(c)
		case 0x7:
			c := 0x01
			if e.v[x] > e.v[y] {
				c = 0x00
			}
			e.v[x] = e.v[y] - e.v[x]
			e.v[0xF] = uint8(c)
		case 0xE:
			c := e.v[y] & 0x80
			e.v[x] = e.v[y] << 1
			e.v[0xF] = uint8(c) >> 7
		}
	case 0x9:
		if e.v[x] != e.v[y] {
			e.pc += 2
		}
	case 0xA:
		e.i = nnn
	case 0xB:
		e.pc = nnn + uint16(e.v[0])
	case 0xC:
		e.v[x] = uint8(uint16(utilities.RandInt(0, 255)) & nn)
	case 0xD:
		x := uint16(e.v[x] % SCREEN_WIDTH)
		y := uint16(e.v[y] % SCREEN_HEIGHT)
		e.v[0xF] = 0
		for i := uint16(0); i < n; i++ {
			sprite := e.memory[e.i+i]
			for j := uint16(0); j < 8; j++ {
				if (sprite & (0x80 >> j)) != 0 {
					if x+j >= SCREEN_WIDTH || y+i >= SCREEN_HEIGHT {
						continue
					}
					if e.display[x+j][y+i] == 1 {
						e.v[0xF] = 1
					}
					e.display[x+j][y+i] ^= 1
				}
			}
		}
		e.draw = true
	case 0xE:
		switch nn {
		case 0x9E:
			if e.keyState[e.v[x]] == 1 {
				e.pc += 2
			}
		case 0xA1:
			if e.keyState[e.v[x]] == 0 {
				e.pc += 2
			}
		}
	case 0xF:
		switch nn {
		case 0x07:
			e.v[x] = e.delayTimer
		case 0x15:
			e.delayTimer = e.v[x]
		case 0x18:
			e.soundTimer = e.v[x]
		case 0x1E:
			e.i += uint16(e.v[x])
		case 0x29:
			e.i = uint16(e.v[x]) * 5
		case 0x33:
			e.memory[e.i] = e.v[x] / 100
			e.memory[e.i+1] = (e.v[x] / 10) % 10
			e.memory[e.i+2] = e.v[x] % 10
		case 0x55:
			for i := 0; i <= int(x); i++ {
				e.memory[e.i] = e.v[i]
				e.i++
			}
		case 0x65:
			for i := 0; i <= int(x); i++ {
				e.v[i] = e.memory[e.i]
				e.i++
			}
		case 0x0A:
			if e.lastKeyReleased < 0xFF {
				e.v[x] = e.lastKeyReleased
			} else {
				e.pc -= 2
			}
		default:
			log.Printf("op: %x, x: %x, y: %x, n: %x, nn: %x, nnn: %x, cycle: %d\n", op, x, y, n, nn, nnn, e.cycleCount)
			panic("opcode not implemented")
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

func (e *Chip8Emulator) SetMemory(address uint16, data []byte) {
	copy(e.memory[address:], data)
}

func (e *Chip8Emulator) SetKeyState(key, state uint8) {
	e.keyState[key] = state
	if state == 0 {
		e.lastKeyReleased = key
	}
}
