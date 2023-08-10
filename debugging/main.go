package main

import (
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
	pru1 := prus.Unit(0)

	if err := pru0.LoadAt(pru1Code, 0x00); err != nil {
		log.Fatalf("Failed to load PRU0 code: %s\n", err)
	}

	if err := pru1.LoadAt(pru1Code, 0x00); err != nil {
		log.Fatalf("Failed to load PRU1 code: %s\n", err)
	}

	pru0.RunAt(0)
	pru1.RunAt(0)

	start := time.Now()
	for (pru0.IsRunning() || pru1.IsRunning()) && time.Since(start) < time.Second {
		time.Sleep(10 * time.Microsecond)
	}

	fmt.Println("Done")
}
