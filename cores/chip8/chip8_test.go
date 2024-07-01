package chip8

import (
	_ "embed"
	"fmt"
	"reflect"
	"testing"

	"github.com/mrchip53/chip-station/utilities"
)

//go:embed 1-chip8-logo.ch8
var rom1 []byte

//go:embed output/1-chip8-logo-test.bin
var rom1Test []byte

func TestChip8Emulator_Chip8Logo(t *testing.T) {
	e := Chip8Emulator{}
	e.Initialize(Hooks{
		Decode: func(opcode uint16, drawCount uint64) bool {
			return opcode&0xF000 == 0x1000
		},
		Draw: func(display [64][32]uint8, drawCount uint64) {
			if drawCount != 11 {
				return
			}

			bytes := make([]byte, 64*32)
			for y := 0; y < 32; y++ {
				for x := 0; x < 64; x++ {
					bytes[y*64+x] = display[x][y]
				}
			}

			if !reflect.DeepEqual(bytes, rom1Test) {
				t.Fatal("Test failed")
			}
		},
	})
	e.LoadROM(rom1)
	e.Loop()
}

//go:embed 2-ibm-logo.ch8
var rom2 []byte

//go:embed output/2-ibm-logo-test.bin
var rom2Test []byte

func TestChip8Emulator_IBMLogo(t *testing.T) {
	e := Chip8Emulator{}
	e.Initialize(Hooks{
		Decode: func(opcode uint16, drawCount uint64) bool {
			return opcode&0xF000 == 0x1000
		},
		Draw: func(display [64][32]uint8, drawCount uint64) {
			if drawCount != 5 {
				return
			}

			bytes := make([]byte, 64*32)
			for y := 0; y < 32; y++ {
				for x := 0; x < 64; x++ {
					bytes[y*64+x] = display[x][y]
				}
			}

			if !reflect.DeepEqual(bytes, rom2Test) {
				t.Fatal("Test failed")
			}
		},
	})
	e.LoadROM(rom2)
	e.Loop()
}

//go:embed 3-corax+.ch8
var rom3 []byte

//go:embed output/3-corax+-test.bin
var rom3Test []byte

func TestChip8Emulator_CoraxPlus(t *testing.T) {
	e := Chip8Emulator{}
	e.Initialize(Hooks{
		Decode: func(opcode uint16, drawCount uint64) bool {
			return drawCount >= 68 && opcode&0xF000 == 0x1000
		},
		Draw: func(display [64][32]uint8, drawCount uint64) {
			if drawCount != 67 {
				return
			}

			bytes := make([]byte, 64*32)
			for y := 0; y < 32; y++ {
				for x := 0; x < 64; x++ {
					bytes[y*64+x] = display[x][y]
				}
			}

			if !reflect.DeepEqual(bytes, rom3Test) {
				t.Fatal("Test failed")
			}
		},
	})
	e.LoadROM(rom3)
	e.Loop()
}

//go:embed 4-flags.ch8
var rom4 []byte

//go:embed output/4-flags-test.bin
var rom4Test []byte

func TestChip8Emulator_Flags(t *testing.T) {
	e := Chip8Emulator{}
	e.Initialize(Hooks{
		Decode: func(opcode uint16, drawCount uint64) bool {
			if drawCount > 78 {
				return opcode&0xF000 == 0x1000
			}
			return false
		},
		Draw: func(display [64][32]uint8, drawCount uint64) {
			if drawCount != 78 {
				return
			}

			bytes := make([]byte, 64*32)
			for y := 0; y < 32; y++ {
				for x := 0; x < 64; x++ {
					bytes[y*64+x] = display[x][y]
				}
			}

			if !reflect.DeepEqual(bytes, rom4Test) {
				t.Fatal("Test failed")
			}
		},
	})
	e.LoadROM(rom4)
	e.Loop()
}

//go:embed 5-quirks.ch8
var rom5 []byte

func TestChip8Emulator_Chip8Quirks(t *testing.T) {
	e := Chip8Emulator{}
	e.Initialize(Hooks{
		Decode: func(opcode uint16, drawCount uint64) bool {
			return false
		},
		Draw: func(display [64][32]uint8, drawCount uint64) {
			bytes := make([]byte, 64*32)
			for y := 0; y < 32; y++ {
				for x := 0; x < 64; x++ {
					bytes[y*64+x] = display[x][y]
				}
			}

			utilities.SavePNG(display, fmt.Sprintf("screenshots/5-quirks-%d.png", drawCount))
		},
	})
	e.SetMemory(0x1FF, []byte{0x01})
	e.LoadROM(rom5)
	e.Loop()
}
