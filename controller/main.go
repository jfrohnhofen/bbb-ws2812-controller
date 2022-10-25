package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/aamcrae/pru"
)

//go:generate ../build/pasm -V3 -g -Gmain -CpruCode pru.p

type Frame struct {
	numChannels uint32
	data        []byte
}

type PruBuffer struct {
	addr uint32
	size uint32
	data []byte
	next *PruBuffer
}

const (
	port = 9000

	frameDataOffset = 12
	bytesPerLed     = 3
	bitsPerLed      = 24

	numChannelsPerDataPin = 6
	numDataPins           = 8

	pruEndAddrOffset  = 0
	pruNextAddrOffset = 4
	pruDataOffset     = 8

	perBitTime = 1250 * time.Nanosecond
	resetTime  = 1 * time.Millisecond
)

var (
	prus    *pru.PRU
	ch      = make(chan bool)
	pruPins = []string{"P8_27", "P8_28", "P8_29", "P8_30", "P8_39", "P8_40", "P8_41", "P8_42", "P8_43", "P8_44", "P8_45", "P8_46"}
	showRe  = regexp.MustCompile(`^show ([a-zA-Z0-9_\./-]*) (\d+)$`)
	playRe  = regexp.MustCompile(`^play ([a-zA-Z0-9_\./-]*) (\d+) to (\d+) @ (\d+) fps$`)
	stopRe  = regexp.MustCompile(`^stop$`)
)

func main() {
	setupPru()
	go releaseToken()
	runServer()
}

func setupPru() {
	for _, pin := range pruPins {
		if err := exec.Command("config-pin", pin, "pruout").Run(); err != nil {
			log.Fatalf("Failed to configure pin '%s': %s\n", pin, err)
		}
	}

	var err error
	prus, err = pru.Open(pru.NewConfig().EnableUnit(0).EnableUnit(1))
	if err != nil {
		log.Fatalf("Failed to open PRU device: %s\n", err)
	}

	if err := prus.Unit(1).LoadAt(pruCode, 0x00); err != nil {
		log.Fatalf("Failed to load PRU code: %s\n", err)
	}
}

func runServer() {
	conn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Printf("Listening on port %d\n", port)

	for {
		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Printf("Failed to read from connection: %s\n", err)
			continue
		}
		go handle(conn, addr, string(buf[:n-1]))
	}
}

func handle(conn net.PacketConn, addr net.Addr, cmd string) {
	log.Printf("Command from %s: '%s'\n", addr, cmd)

	if match := showRe.FindStringSubmatch(cmd); match != nil {
		frame, _ := strconv.ParseInt(match[2], 10, 32)
		show(match[1], uint32(frame))
	} else if match := playRe.FindStringSubmatch(cmd); match != nil {
		from, _ := strconv.ParseInt(match[2], 10, 32)
		to, _ := strconv.ParseInt(match[3], 10, 32)
		fps, _ := strconv.ParseInt(match[4], 10, 32)
		play(match[1], uint32(from), uint32(to), uint32(fps))
	} else if stopRe.MatchString(cmd) {
		stop()
	} else {
		log.Printf("Invalid command: '%s'\n", cmd)
	}
}

func show(input string, frame uint32) {
	claimToken()
	for hasToken() {
		if err := outputFrame(input, frame); err != nil {
			log.Printf("Failed to show frame: %s", err)
		} else {
			log.Printf("Done showing frame %d", frame)
			releaseToken()
			return
		}
	}
}

func play(input string, from, to, fps uint32) {
	claimToken()
	start := time.Now()
	for hasToken() {
		frame := from + uint32(time.Since(start).Seconds()*float64(fps))
		if frame > to {
			frame = to
		}

		if err := outputFrame(input, frame); err != nil {
			log.Printf("Failed to show frame: %s", err)
		} else if frame == to {
			log.Printf("Done playing frames %d to %d", from, to)
			releaseToken()
			return
		}

		nextFrameStart := start.Add(time.Duration(float64(frame-from+1) * float64(time.Second) / float64(fps)))
		time.Sleep(time.Until(nextFrameStart))
	}
}

func stop() {
	claimToken()
	releaseToken()
}

func outputFrame(input string, frameIdx uint32) error {
	frame, err := loadFrame(input, frameIdx)
	if err != nil {
		return fmt.Errorf("failed to load frame '%s:%d': %s", input, frameIdx, err)
	}

	if frame.numChannels != numChannelsPerDataPin*numDataPins {
		return fmt.Errorf("frame has invalid number of cahnnels")
	}

	sharedRam := PruBuffer{addr: 0x0001_0000, size: 12 * 1024, data: prus.SharedRam}
	pru0Ram := PruBuffer{addr: 0x0000_2000, size: 8 * 1024, data: prus.Unit(0).Ram, next: &sharedRam}
	pru1Ram := PruBuffer{addr: 0x0000_0000, size: 8 * 1024, data: prus.Unit(1).Ram, next: &pru0Ram}
	sharedRam.next = &pru1Ram
	buffers := []PruBuffer{pru1Ram, pru0Ram, sharedRam}

	numBits := len(frame.data) / numChannelsPerDataPin
	pruTimeout := time.Duration(numBits)*perBitTime + resetTime

	for _, buf := range buffers {
		binary.LittleEndian.PutUint32(buf.data[pruEndAddrOffset:], 0x0000_0000)
		binary.LittleEndian.PutUint32(buf.data[pruNextAddrOffset:], buf.next.addr)
		if len(frame.data) > 0 {
			fillPruBuffer(&frame, buf)
		}
	}

	prus.Unit(1).RunAt(0)
	start := time.Now()

	buf := &buffers[0]
	for prus.Unit(1).IsRunning() && time.Since(start) < pruTimeout {
		if len(frame.data) > 0 && binary.LittleEndian.Uint32(buf.data[pruEndAddrOffset:]) == 0x0000_0000 {
			fillPruBuffer(&frame, *buf)
			buf = buf.next
		}
		time.Sleep(10 * time.Microsecond)
	}

	if len(frame.data) > 0 {
		return fmt.Errorf("controller has not written all data")
	}

	if prus.Unit(1).IsRunning() {
		return fmt.Errorf("PRU timed out")
	}
	for _, buf := range buffers {
		if binary.LittleEndian.Uint32(buf.data[pruEndAddrOffset:]) != 0x0000_0000 {
			return fmt.Errorf("PRU has not written all data")
		}
	}

	return nil
}

func loadFrame(input string, frameIdx uint32) (Frame, error) {
	frame := Frame{}

	file, err := os.Open(input)
	if err != nil {
		return frame, err
	}
	defer file.Close()

	var numFrames, numLeds uint32
	if err := binary.Read(file, binary.LittleEndian, &frame.numChannels); err != nil {
		return frame, err
	}
	if err := binary.Read(file, binary.LittleEndian, &numLeds); err != nil {
		return frame, err
	}
	if err := binary.Read(file, binary.LittleEndian, &numFrames); err != nil {
		return frame, err
	}
	if frameIdx >= numFrames {
		return frame, fmt.Errorf("frame index out of range")
	}

	size := frame.numChannels * numLeds * bytesPerLed
	offset := int64(frameDataOffset + frameIdx*size)
	frame.data = make([]byte, size)
	_, err = file.ReadAt(frame.data, offset)
	return frame, err
}

func fillPruBuffer(frame *Frame, buf PruBuffer) {
	numBits := (buf.size - pruDataOffset) / numChannelsPerDataPin
	if bitsLeft := uint32(len(frame.data)) / numChannelsPerDataPin; bitsLeft <= numBits {
		numBits = bitsLeft
	}
	numBits = numBits / bitsPerLed * bitsPerLed
	size := numBits * numChannelsPerDataPin
	endAddr := buf.addr + pruDataOffset + size - numChannelsPerDataPin
	binary.LittleEndian.PutUint32(buf.data[pruEndAddrOffset:], endAddr)
	copy(buf.data[pruDataOffset:], frame.data[:size])
	frame.data = frame.data[size:]
}

func claimToken() {
	<-ch
}

func hasToken() bool {
	select {
	case ch <- true:
		return false
	default:
		return true
	}
}

func releaseToken() {
	ch <- true
}
