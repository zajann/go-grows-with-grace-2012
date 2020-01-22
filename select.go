package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	boom := time.After(10 * time.Second)
	tick := time.NewTicker(499 * time.Millisecond)

	for {
		select {
		case <-tick.C:
			fmt.Println("tick", time.Since(start))
		case <-boom:
			fmt.Println("BOOM", time.Since(start))
			return
		}
	}
}