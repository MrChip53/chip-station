package utilities

import (
	"log"
	"os"
)

func SaveBytesToFile(data []byte, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := f.Write(data); err != nil {
		f.Close()
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
