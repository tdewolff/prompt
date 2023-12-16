package main

import (
	"fmt"
	"time"

	"github.com/tdewolff/prompt"
)

func main() {
	p := prompt.NewPercentProgress("Progres bar: ", 1.0, prompt.DefaultProgressStyle)

	p.Start()
	for i := 0; i <= 100; i++ {
		f := float64(i) / 100.0
		p.Set(f)
		time.Sleep(10 * time.Millisecond)
	}
	p.Stop()
	fmt.Println("abc")
}
