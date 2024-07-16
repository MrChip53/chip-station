package programs

type Color struct {
	R, G, B float32
	RGB     uint32
}

func NewColor(rgb uint32) Color {
	return Color{
		R:   float32((rgb>>16)&0xFF) / 255,
		G:   float32((rgb>>8)&0xFF) / 255,
		B:   float32(rgb&0xFF) / 255,
		RGB: rgb,
	}
}
