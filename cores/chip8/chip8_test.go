package chip8

import (
	_ "embed"
	"reflect"
	"testing"
)

//go:embed 1-chip8-logo.ch8
var rom1 []byte

//go:embed output/1-chip8-logo-test.bin
var rom1Test []byte

func TestChip8Emulator_Chip8Logo(t *testing.T) {
	e := Chip8Emulator{}
	e.Initialize(Hooks{
		Decode: func(opcode uint16) bool {
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
		Decode: func(opcode uint16) bool {
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
