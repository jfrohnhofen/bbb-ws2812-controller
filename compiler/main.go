package main

import (
	"encoding/binary"
	"flag"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/schollz/progressbar/v3"
)

var (
	configFile   = flag.String("config", "", "config file")
	outputFile   = flag.String("output", "", "output file")
	inputPattern = flag.String("pattern", "", "input file pattern")
)

func main() {
	flag.Parse()

	config, err := ReadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to read config: %s", err)
	}

	out, err := os.Create(*outputFile)
	if err != nil {
		log.Fatalf("Failed to crreate output: %s", err)
	}
	defer out.Close()

	inputs, err := filepath.Glob(*inputPattern)
	if err != nil {
		log.Fatalf("Failed to list frames: %s", err)
	}
	sort.Strings(inputs)
	numFrames := len(inputs)

	numChannels := uint((NumChannels+7)/8) * 8
	numLeds := uint(0)
	for _, panel := range config.Panels {
		if panel.NumLeds() > numLeds {
			numLeds = panel.NumLeds()
		}
	}

	log.Printf("Number of channels: %d\n", numChannels)
	log.Printf("Number of LEDs: %d\n", numLeds)
	log.Printf("Number of frames: %d\n", numFrames)

	binary.Write(out, binary.LittleEndian, uint32(numChannels))
	binary.Write(out, binary.LittleEndian, uint32(numLeds))
	binary.Write(out, binary.LittleEndian, uint32(numFrames))

	bar := progressbar.Default(int64(numFrames))

	for _, inputFile := range inputs {
		img, err := LoadImage(inputFile)
		if err != nil {
			log.Fatalf("Failed to load image: %s", err)
		}

		frame := make([]uint8, numChannels*numLeds*3)

		for _, panel := range config.Panels {
			for ledIdx, led := range panel.Leds {
				color := GetPixel(img, led.X()*config.ImageScale, led.Y()*config.ImageScale)
				const bitsPerLed = 24
				for i := uint(0); i < bitsPerLed; i++ {
					bitValue := uint8(((uint32(color.G)<<16)|(uint32(color.R)<<8)|uint32(color.B))>>(23-i)) & 0b1
					globalBitIdx := (uint(ledIdx)+panel.NumBufferLeds)*bitsPerLed*numChannels +
						i*numChannels +
						panel.Channel
					byteIdx := globalBitIdx / 8
					bitIdx := globalBitIdx % 8
					frame[byteIdx] = frame[byteIdx] | (bitValue << bitIdx)
				}
			}
		}

		if _, err := out.Write(frame); err != nil {
			log.Fatalf("Failed to write frame data to file: %s", err)
		}
		bar.Add(1)
	}
}
