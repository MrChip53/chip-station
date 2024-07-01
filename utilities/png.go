package utilities

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

func ScaleImage(img image.Image, width, height int) image.Image {
	return resizeImage(img, width, height)
}

func resizeImage(img image.Image, width, height int) image.Image {
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			newImg.Set(x, y, img.At(x*img.Bounds().Dx()/width, y*img.Bounds().Dy()/height))
		}
	}
	return newImg
}

func GetPNG(display [64][32]uint8) image.Image {
	img := image.NewGray(image.Rect(0, 0, 64, 32))

	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.Gray{Y: display[x][y] * 255})
		}
	}

	return img
}

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
