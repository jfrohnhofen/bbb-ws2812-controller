package main

import "log"

//go:generate ../../build/pasm -g -Gmain -CpruCode pru.p

func main() {
	log.Println("Hello World")
}
