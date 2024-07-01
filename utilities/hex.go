package utilities

import "fmt"

func Hex(opcode uint16) string {
	return "0x" + fmt.Sprintf("%04X", opcode)
}
