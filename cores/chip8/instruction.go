package chip8

import (
	"math/rand"
)

var instructions = map[uint16]Instruction{
	0x00E0: &ClearScreen{},
	0x00EE: &Return{},
	0x1000: &Jump{},
	0x2000: &Call{},
	0x3000: &SkipIfEqualImmediate{},
	0x4000: &SkipIfNotEqualImmediate{},
	0x5000: &SkipIfEqualRegister{},
	0x6000: &SetRegisterImmediate{},
	0x7000: &AddRegisterImmediate{},
	0x8000: &SetRegister{},
	0x8001: &BinaryOrRegister{},
	0x8002: &BinaryAndRegister{},
	0x8003: &BinaryXorRegister{},
	0x8004: &AddRegisterXY{},
	0x8005: &SubtractRegisterXY{},
	0x8006: &ShiftRight{},
	0x8007: &SubtractRegisterYX{},
	0x800E: &ShiftLeft{},
	0x9000: &SkipIfNotEqualRegister{},
	0xA000: &SetIndex{},
	0xB000: &JumpPlusOffset{},
	0xC000: &Random{},
	0xD000: &Draw{},
	0xE09E: &SkipIfKeyPressed{},
	0xE0A1: &SkipIfKeyNotPressed{},
	0xF007: &SetRegisterWithDelayTimer{},
	0xF00A: &WaitForKey{},
	0xF015: &SetDelayTimer{},
	0xF018: &SetSoundTimer{},
	0xF01E: &AddRegisterToIndex{},
	0xF029: &SetIndexToSprite{},
	0xF033: &StoreBCD{},
	0xF055: &StoreRegisters{},
	0xF065: &LoadRegisters{},
}

func getOpKey(opcode uint16) uint16 {
	op := opcode & 0xF000
	if op == 0x8000 {
		return opcode & 0xF00F
	} else if op == 0xE000 || op == 0xF000 {
		return opcode & 0xF0FF
	}
	return op
}

type Instruction interface {
	Execute(*Chip8Emulator)
	Fill(uint16)
}

type BaseInstruction struct {
	opcode uint16
	op     uint8
	x      uint8
	y      uint8
	n      uint8
	nn     uint8
	nnn    uint16
}

func (b *BaseInstruction) Fill(opcode uint16) {
	b.opcode = opcode
	b.op = uint8(opcode & 0xF000 >> 12)
	b.x = uint8((opcode & 0x0F00) >> 8)
	b.y = uint8((opcode & 0x00F0) >> 4)
	b.n = uint8(opcode & 0x000F)
	b.nn = uint8(opcode & 0x00FF)
	b.nnn = opcode & 0x0FFF
}

type ClearScreen struct {
	BaseInstruction
}

func (c ClearScreen) Execute(e *Chip8Emulator) {
	e.clearDisplay()
}

type Return struct {
	BaseInstruction
}

func (r Return) Execute(e *Chip8Emulator) {
	e.pc = e.stack.Pop()
}

type Jump struct {
	BaseInstruction
}

func (j Jump) Execute(e *Chip8Emulator) {
	e.pc = j.nnn
}

type Call struct {
	BaseInstruction
}

func (c Call) Execute(e *Chip8Emulator) {
	e.stack.Push(e.pc)
	e.pc = c.nnn
}

type SkipIfNotEqualImmediate struct {
	BaseInstruction
}

func (s SkipIfNotEqualImmediate) Execute(e *Chip8Emulator) {
	if e.v[s.x] != s.nn {
		e.pc += 2
	}
}

type SkipIfEqualImmediate struct {
	BaseInstruction
}

func (s SkipIfEqualImmediate) Execute(e *Chip8Emulator) {
	if e.v[s.x] == s.nn {
		e.pc += 2
	}
}

type SkipIfEqualRegister struct {
	BaseInstruction
}

func (s SkipIfEqualRegister) Execute(e *Chip8Emulator) {
	if e.v[s.x] == e.v[s.y] {
		e.pc += 2
	}
}

type SetRegisterImmediate struct {
	BaseInstruction
}

func (s SetRegisterImmediate) Execute(e *Chip8Emulator) {
	e.v[s.x] = s.nn
}

type AddRegisterImmediate struct {
	BaseInstruction
}

func (a AddRegisterImmediate) Execute(e *Chip8Emulator) {
	e.v[a.x] += a.nn
}

type SkipIfNotEqualRegister struct {
	BaseInstruction
}

func (s SkipIfNotEqualRegister) Execute(e *Chip8Emulator) {
	if e.v[s.x] != e.v[s.y] {
		e.pc += 2
	}
}

type SetIndex struct {
	BaseInstruction
}

func (s SetIndex) Execute(e *Chip8Emulator) {
	e.i = s.nnn
}

type JumpPlusOffset struct {
	BaseInstruction
}

func (j JumpPlusOffset) Execute(e *Chip8Emulator) {
	e.pc = j.nnn + uint16(e.v[0])
}

type Random struct {
	BaseInstruction
}

func (r Random) Execute(e *Chip8Emulator) {
	e.v[r.x] = uint8(rand.Intn(256)) & r.nn
}

type Draw struct {
	BaseInstruction
}

func (d Draw) Execute(e *Chip8Emulator) {
	x := uint16(e.v[d.x] % 64)
	y := uint16(e.v[d.y] % 32)
	e.v[0xF] = 0
	for i := uint16(0); i < uint16(d.n); i++ {
		sprite := e.memory[e.i+i]
		for j := uint16(0); j < 8; j++ {
			if (sprite & (0x80 >> j)) != 0 {
				if x+j >= 64 || y+i >= 32 {
					continue
				}
				if e.display[x+j][y+i] == 1 {
					e.v[0xF] = 1
				}
				e.display[x+j][y+i] ^= 1
			}
		}
	}
}

type SetRegister struct {
	BaseInstruction
}

func (s SetRegister) Execute(e *Chip8Emulator) {
	e.v[s.x] = e.v[s.y]
}

type BinaryOrRegister struct {
	BaseInstruction
}

func (b BinaryOrRegister) Execute(e *Chip8Emulator) {
	e.v[b.x] |= e.v[b.y]
	e.v[0xF] = 0
}

type BinaryAndRegister struct {
	BaseInstruction
}

func (b BinaryAndRegister) Execute(e *Chip8Emulator) {
	e.v[b.x] &= e.v[b.y]
	e.v[0xF] = 0
}

type BinaryXorRegister struct {
	BaseInstruction
}

func (b BinaryXorRegister) Execute(e *Chip8Emulator) {
	e.v[b.x] ^= e.v[b.y]
	e.v[0xF] = 0
}

type AddRegister struct {
	BaseInstruction
}

func (a AddRegister) Execute(e *Chip8Emulator) {
	c := 0
	if int(e.v[a.x])+int(e.v[a.y]) > 255 {
		c = 1
	}
	e.v[a.x] += e.v[a.y]
	e.v[0xF] = uint8(c)
}

type SubtractRegisterXY struct {
	BaseInstruction
}

func (s SubtractRegisterXY) Execute(e *Chip8Emulator) {
	c := 1
	if e.v[s.y] > e.v[s.x] {
		c = 0
	}
	e.v[s.x] -= e.v[s.y]
	e.v[0xF] = uint8(c)
}

type SubtractRegisterYX struct {
	BaseInstruction
}

func (s SubtractRegisterYX) Execute(e *Chip8Emulator) {
	c := 1
	if e.v[s.x] > e.v[s.y] {
		c = 0
	}
	e.v[s.x] = e.v[s.y] - e.v[s.x]
	e.v[0xF] = uint8(c)
}

type AddRegisterXY struct {
	BaseInstruction
}

func (a AddRegisterXY) Execute(e *Chip8Emulator) {
	c := 0
	if int(e.v[a.x])+int(e.v[a.y]) > 255 {
		c = 1
	}
	e.v[a.x] += e.v[a.y]
	e.v[0xF] = uint8(c)
}

type ShiftRight struct {
	BaseInstruction
}

func (s ShiftRight) Execute(e *Chip8Emulator) {
	c := e.v[s.x] & 0x01
	e.v[s.x] = e.v[s.y] >> 1
	e.v[0xF] = uint8(c)
}

type ShiftLeft struct {
	BaseInstruction
}

func (s ShiftLeft) Execute(e *Chip8Emulator) {
	c := e.v[s.x] >> 7
	e.v[s.x] = e.v[s.y] << 1
	e.v[0xF] = uint8(c)
}

type SkipIfKeyPressed struct {
	BaseInstruction
}

func (s SkipIfKeyPressed) Execute(e *Chip8Emulator) {
	if e.keyState.IsKeyPressed(e.v[s.x]) {
		e.pc += 2
	}
}

type SkipIfKeyNotPressed struct {
	BaseInstruction
}

func (s SkipIfKeyNotPressed) Execute(e *Chip8Emulator) {
	if !e.keyState.IsKeyPressed(e.v[s.x]) {
		e.pc += 2
	}
}

type SetRegisterWithDelayTimer struct {
	BaseInstruction
}

func (s SetRegisterWithDelayTimer) Execute(e *Chip8Emulator) {
	e.v[s.x] = e.delayTimer.GetTimer()
}

type SetDelayTimer struct {
	BaseInstruction
}

func (s SetDelayTimer) Execute(e *Chip8Emulator) {
	e.delayTimer.SetTimer(e.v[s.x])
}

type SetSoundTimer struct {
	BaseInstruction
}

func (s SetSoundTimer) Execute(e *Chip8Emulator) {
	e.soundTimer.SetTimer(e.v[s.x], e.hooks.PlaySound)
}

type AddRegisterToIndex struct {
	BaseInstruction
}

func (a AddRegisterToIndex) Execute(e *Chip8Emulator) {
	e.i += uint16(e.v[a.x])
}

type SetIndexToSprite struct {
	BaseInstruction
}

func (s SetIndexToSprite) Execute(e *Chip8Emulator) {
	e.i = uint16(e.v[s.x]) * 5
}

type StoreBCD struct {
	BaseInstruction
}

func (s StoreBCD) Execute(e *Chip8Emulator) {
	e.memory[e.i] = e.v[s.x] / 100
	e.memory[e.i+1] = (e.v[s.x] / 10) % 10
	e.memory[e.i+2] = e.v[s.x] % 10
}

type StoreRegisters struct {
	BaseInstruction
}

func (s StoreRegisters) Execute(e *Chip8Emulator) {
	for i := 0; i <= int(s.x); i++ {
		e.memory[e.i] = e.v[i]
		e.i++
	}
}

type LoadRegisters struct {
	BaseInstruction
}

func (l LoadRegisters) Execute(e *Chip8Emulator) {
	for i := 0; i <= int(l.x); i++ {
		e.v[i] = e.memory[e.i]
		e.i++
	}
}

type WaitForKey struct {
	BaseInstruction
}

func (w WaitForKey) Execute(e *Chip8Emulator) {
	lastKey := e.keyState.GetLastKeyReleased()
	if lastKey < 0xFF {
		e.v[w.x] = lastKey
	} else {
		e.pc -= 2
	}
}
