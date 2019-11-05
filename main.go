package main

import (
	"fmt"
	"strconv"
	"strings"
)

func main() {
	prog := "62 48 10 3c 03 62 65 10 3c 08 62 6c 10 3c 0d 62 6c 10 3c 12 62 6f 10 3c 17 62 2c 10 3c 1c 62 20 10 3c 21 62 6b 10 3c 26 62 75 10 3c 2b 62 65 10 3c 30 62 63 10 3c 35 62 68 10 3c 3a 62 69 10 3c 3f 62 70 10 3c 44 0f"

	var exe []uint8
	for _, p := range strings.Split(prog, " ") {
		v, err := strconv.ParseInt(p, 16, 16)
		if err != nil {
			panic(err)
		}
		exe = append(exe, uint8(v))
	}

	runner := NewRunner(exe)

	NewSyncer(runner, &SyncableDstStdout{})

	prevPhase := 0
	for {
		addr := runner.PC
		halt, nextPhase := runner.RunSinglePhase()

		fmt.Printf("%02x\t%x\t%02x\t%02x\t%02x\t%02x\t%02x\t%02x\n", addr, prevPhase, runner.PC, runner.FLAG, runner.ACC, runner.IX, runner.MAR, runner.IR)
		if halt {
			break
		}
		prevPhase = nextPhase
	}

}
