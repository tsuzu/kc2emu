package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

func parse(prog string) []uint8 {
	var exe []uint8
	for _, p := range strings.Split(prog, " ") {
		v, err := strconv.ParseInt(p, 16, 16)
		if err != nil {
			panic(err)
		}
		exe = append(exe, uint8(v))
	}

	return exe
}

func stepExecution(r *Runner, writer io.Writer) {
	var halt bool
	var addr uint8
	var prevPhase, nextPhase int
	for {
		if nextPhase == 0 {
			addr = r.PC
		}

		halt, nextPhase = r.RunSinglePhase()

		fmt.Fprintf(writer, "%02x\t%x\t%02x\t%02x\t%02x\t%02x\t%02x\t%02x\n", addr, prevPhase, r.PC, r.FLAG, r.ACC, r.IX, r.MAR, r.IR)
		if halt {
			break
		}
		prevPhase = nextPhase
	}

	fmt.Fprint(writer, "program memory: [")
	for i := 0; i < DataMemoryOffset; i++ {
		fmt.Fprintf(writer, "%02x ", r.Memory[i])
	}
	fmt.Fprint(writer, "]\n")

	fmt.Fprint(writer, "data memory: [")
	for i := DataMemoryOffset; i < len(r.Memory); i++ {
		fmt.Fprintf(writer, "%02x ", r.Memory[i])
	}
	fmt.Fprint(writer, "]\n")

}

func main() {

	sender := NewRunner(parse("67 00 41 41 41 41 b7 01 10 3c 09 ba 02 fa 18 31 00 0f"))
	receiver := NewRunner(parse("34 00 1f 77 01 e2 f0 42 42 42 42 77 00 67 01 e2 0f 77 01 ba 02 fa 18 31 00 0f"))

	for i := 0; i < 24; i++ {
		sender.Memory[i+DataMemoryOffset] = uint8(i % 16)
	}

	NewSyncer(sender, receiver)

	sw, err := os.Create("sender.dump")

	if err != nil {
		panic(err)
	}
	defer sw.Close()

	rw, err := os.Create("receiver.dump")

	if err != nil {
		panic(err)
	}
	defer rw.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		stepExecution(sender, sw)
		wg.Done()
	}()
	go func() {
		stepExecution(receiver, rw)
		wg.Done()
	}()

	wg.Wait()
}
