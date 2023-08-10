package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/aamcrae/pru"
)

//go:generate ../build/pasm -V3 -g -Gmain -Cpru0Code pru0.p
//go:generate ../build/pasm -V3 -g -Gmain -Cpru1Code pru1.p

var (
	pruInputs  = []string{"P9_27", "P9_28", "P9_29", "P9_30", "P9_31"}
	pruOutputs = []string{"P8_27", "P8_28", "P8_29", "P8_30", "P8_39", "P8_40", "P8_41", "P8_42", "P8_43", "P8_44", "P8_45", "P8_46"}
)

func main() {
	for _, pin := range pruOutputs {
		if err := exec.Command("config-pin", pin, "pruout").Run(); err != nil {
			log.Fatalf("Failed to configure pin '%s': %s\n", pin, err)
		}
	}
	for _, pin := range pruInputs {
		if err := exec.Command("config-pin", pin, "pruin").Run(); err != nil {
			log.Fatalf("Failed to configure pin '%s': %s\n", pin, err)
		}
	}

	prus, err := pru.Open(pru.NewConfig().EnableUnit(0).EnableUnit(1))
	if err != nil {
		log.Fatalf("Failed to open PRU device: %s\n", err)
	}

	pru0 := prus.Unit(0)
	if err := pru0.LoadAt(pru0Code, 0x00); err != nil {
		log.Fatalf("Failed to load PRU0 code: %s\n", err)
	}

	pru1 := prus.Unit(1)
	if err := pru1.LoadAt(pru1Code, 0x00); err != nil {
		log.Fatalf("Failed to load PRU1 code: %s\n", err)
	}

	binary.LittleEndian.PutUint32(pru0.Ram, 0)
	pru0.RunAt(0)
	pru1.RunAt(0)

	start := time.Now()
	for (pru0.IsRunning() || pru1.IsRunning()) && time.Since(start) < time.Second {
		time.Sleep(10 * time.Microsecond)
	}

	if pru0.IsRunning() {
		log.Println("PRU0 is still busy")
		return
	}
	count0 := binary.LittleEndian.Uint32(pru0.Ram[0:]) - 4
	fmt.Println("PRU0 cycles:", count0)

	samples := pru0.Ram[4 : 4*4*28+4]

	for burst := 0; burst < 4; burst++ {
		for ch := 0; ch < 6; ch++ {
			fmt.Printf("CH%d ", ch)
			for sample := 0; sample < 4*28; sample++ {
				if samples[burst*28*4+sample]&(1<<ch) == 0 {
					fmt.Print("_")
				} else {
					fmt.Print("\u25AE")
				}
			}
			fmt.Println()
		}
		fmt.Println()
	}

	if pru1.IsRunning() {
		log.Println("PRU1 is still busy")
	}
	count1 := binary.LittleEndian.Uint32(pru1.Ram[0:])
	sum1 := binary.LittleEndian.Uint32(pru1.Ram[4:]) - 4*count1
	fmt.Println("PRU1 time: count", count1, "sum", sum1, "avg", float64(sum1)/float64(count1))

	fmt.Println("Done")
}
