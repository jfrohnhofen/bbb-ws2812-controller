package main

import (
	"crypto/rand"
	"encoding/binary"
	"log"
	"math"
	"os/exec"
	"time"

	"github.com/aamcrae/pru"
)

//go:generate ../../build/pasm -V3 -g -Gmain -CpruCode pru.p

const (
	endAddrOffset  = 0
	nextAddrOffset = 4
	dataOffset     = 8

	numChannels = 6
	timeout     = 15 * time.Millisecond
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

	numBits := 24
	frame := make([]byte, 6*numBits)
	rand.Read(frame)

	for _, buf := range buffers {
		frame = fillBuffer(frame, buf)
	}

	pru1.RunAt(0)
	start := time.Now()

	buf := &buffers[0]
	for pru1.IsRunning() && time.Since(start) < timeout {
		if len(frame) > 0 && binary.LittleEndian.Uint32(buf.data[endAddrOffset:]) == 0 {
			frame = fillBuffer(frame, *buf)
			buf = buf.next
		}
		time.Sleep(10 * time.Microsecond)
	}

	log.Printf("PRU took %dus\n", time.Since(start).Microseconds())

	if pru1.IsRunning() {
		log.Fatalf("PRU timed out\n")
	}

	for _, buf := range buffers {
		if binary.LittleEndian.Uint32(buf.data[endAddrOffset:]) != 0 {
			log.Fatalf("PRU failed to write all data\n")
		}
	}
}

func fillBuffer(frame []byte, buf buffer) []byte {
	numBits := (buf.size - dataOffset) / numChannels
	nextAddr := buf.next.addr
	if bitsLeft := uint32(len(frame)) / numChannels; bitsLeft <= numBits {
		numBits = bitsLeft
		nextAddr = math.MaxUint32
	}

	size := numBits * numChannels
	binary.LittleEndian.PutUint32(buf.data[endAddrOffset:], buf.addr+dataOffset+size-numChannels)
	binary.LittleEndian.PutUint32(buf.data[nextAddrOffset:], nextAddr)
	copy(buf.data[dataOffset:], frame[:size])
	return frame[size:]
}
