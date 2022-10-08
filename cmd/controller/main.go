package main

import (
	"encoding/binary"
	"log"
	"math"
	"os"
	"os/exec"
	"time"

	"github.com/aamcrae/pru"
)

//go:generate ../../build/pasm -V3 -g -Gmain -CpruCode pru.p

const (
	inputDataOffset       = 8
	bitsPerLed            = 24
	numChannelsPerDataPin = 6
	numDataPins           = 8

	pruEndAddrOffset  = 0
	pruNextAddrOffset = 4
	pruDataOffset     = 8
	pruTimeout        = 15 * time.Millisecond
)

var pruPins = []string{"P8_27", "P8_28", "P8_29", "P8_30", "P8_39", "P8_40", "P8_41", "P8_42", "P8_43", "P8_44", "P8_45", "P8_46"}

type buffer struct {
	addr uint32
	size uint32
	data []byte
	next *buffer
}

func main() {
	for _, pin := range pruPins {
		if err := exec.Command("config-pin", pin, "pruout").Run(); err != nil {
			log.Fatalf("Failed to configure pin '%s': %s\n", pin, err)
		}
	}

	config := pru.NewConfig().EnableUnit(0).EnableUnit(1)
	prus, err := pru.Open(config)
	if err != nil {
		log.Fatalf("Failed to open PRU device: %s\n", err)
	}
	defer prus.Close()

	pru1 := prus.Unit(1)
	if err := pru1.LoadAt(pruCode, 0x00); err != nil {
		log.Fatalf("Failed to load PRU code: %s\n", err)
	}

	sharedRam := buffer{addr: 0x0001_0000, size: 12 * 1024, data: prus.SharedRam}
	pru0Ram := buffer{addr: 0x0000_2000, size: 8 * 1024, data: prus.Unit(0).Ram, next: &sharedRam}
	pru1Ram := buffer{addr: 0x0000_0000, size: 8 * 1024, data: prus.Unit(1).Ram, next: &pru0Ram}
	sharedRam.next = &pru1Ram
	buffers := []buffer{pru1Ram, pru0Ram, sharedRam}

	frames, err := os.Open("output.dat")
	if err != nil {
		log.Fatalf("Failed to open input: %s", err)
	}
	defer frames.Close()

	var numChannels, numLeds uint32
	binary.Read(frames, binary.LittleEndian, &numChannels)
	binary.Read(frames, binary.LittleEndian, &numLeds)
	numBits := numLeds * bitsPerLed

	if numChannels != numChannelsPerDataPin*numDataPins {
		log.Fatalf("Input has unexpected number of channels")
	}

	for frameIdx := 0; frameIdx < 1077; frameIdx++ {
		frame := make([]byte, numChannels*numBits/8)
		frameOffset := int64(inputDataOffset + len(frame)*frameIdx)
		if _, err := frames.ReadAt(frame, frameOffset); err != nil {
			log.Fatalf("Failed to read frame %d: %s", frameIdx, err)
		}

		for _, buf := range buffers {
			frame = fillBuffer(frame, buf)
		}

		pru1.RunAt(0)
		start := time.Now()

		buf := &buffers[0]
		for pru1.IsRunning() && time.Since(start) < pruTimeout {
			if len(frame) > 0 && binary.LittleEndian.Uint32(buf.data[pruEndAddrOffset:]) == 0 {
				frame = fillBuffer(frame, *buf)
				buf = buf.next
			}
			time.Sleep(10 * time.Microsecond)
		}

		log.Printf("Frame %d took %dus\n", frameIdx, time.Since(start).Microseconds())

		time.Sleep(20 * time.Millisecond)
	}

	if pru1.IsRunning() {
		log.Fatalf("PRU timed out")
	}

	for _, buf := range buffers {
		if binary.LittleEndian.Uint32(buf.data[pruEndAddrOffset:]) != 0 {
			log.Fatalf("PRU failed to write all data")
		}
	}
}

func fillBuffer(frame []byte, buf buffer) []byte {
	numBits := (buf.size - pruDataOffset) / numChannelsPerDataPin
	nextAddr := buf.next.addr
	if bitsLeft := uint32(len(frame)) / numChannelsPerDataPin; bitsLeft <= numBits {
		numBits = bitsLeft
		nextAddr = math.MaxUint32
	}

	size := numBits * numChannelsPerDataPin
	binary.LittleEndian.PutUint32(buf.data[pruEndAddrOffset:], buf.addr+pruDataOffset+size-numChannelsPerDataPin)
	binary.LittleEndian.PutUint32(buf.data[pruNextAddrOffset:], nextAddr)
	copy(buf.data[pruDataOffset:], frame[:size])
	return frame[size:]
}
