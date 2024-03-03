package main

import (
	"bufio"
	"fmt"
	"log"
	"math/bits"
	"os"
	"strconv"
	"time"

	"github.com/appstronomer/sf-hw-26/pipe"
	"github.com/appstronomer/sf-hw-26/ring"
)

var stderr *log.Logger = log.New(os.Stderr, "", 0)

func loopStdin(chOut chan<- int) {
	scanner := bufio.NewScanner(os.Stdin)
	var text string
	for scanner.Scan() {
		text = scanner.Text()
		val, err := strconv.ParseInt(text, 10, bits.UintSize)
		if err != nil {
			stderr.Printf("[ERR] loopStdin: ParseInt error: %v", err)
			continue
		}
		stderr.Printf("[LOG] loopStdin: number parsed: %v", val)
		chOut <- int(val)
	}
	if err := scanner.Err(); err != nil {
		stderr.Printf("[ERR] scanner error: %v", err)
	}
}

func loopStdout(chIn <-chan int) {
	for val := range chIn {
		fmt.Println(val)
	}
}

func loopFilterNegative(chIn <-chan int, chOut chan<- int) {
	for val := range chIn {
		if val < 0 {
			stderr.Printf("[FILTER] negative: %v", val)
			continue
		}
		stderr.Printf("[PASS] not negative: %v", val)
		chOut <- val
	}
}

func loopFilterMul(chIn <-chan int, chOut chan<- int) {
	for val := range chIn {
		if val == 0 || val%3 != 0 {
			stderr.Printf("[FILTER] undividable: %v", val)
			continue
		}
		stderr.Printf("[PASS] divisible by 3: %v", val)
		chOut <- val
	}
}

func loopSabotage(chIn <-chan int, chOut chan<- int) {
	for i := 0; i < 50; i++ {
		val, ok := <-chIn
		if !ok {
			return
		}
		chOut <- val
	}
	for val := range chIn {
		chOut <- val
		time.Sleep(time.Millisecond * 5)
	}
}

func main() {
	const cap = 4
	// data source
	pipe := pipe.NewPipe[int](cap, loopStdin)
	// filters
	pipe.Next(cap, loopFilterNegative).Next(cap, loopFilterMul)
	// ring buffer
	pipe.Next(cap, ring.MakeRingBufferLoop[int](50))
	// sloooow sabotage -> stdout
	pipe.Next(cap, loopSabotage).Last(loopStdout)

	pipe.Wait()
}
