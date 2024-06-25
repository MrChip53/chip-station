package utilities

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

func SavePNG(display [64][32]uint8, filename string) {
	img := image.NewGray(image.Rect(0, 0, 64, 32))

	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.Gray{Y: display[x][y] * 255})
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, img); err != nil {
		f.Close()
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
