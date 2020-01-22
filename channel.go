package main

import (
	"fmt"
)

func fibs(c chan int) {
	a, b := 0, 1

	for {
		c <- a
		a, b = b, a+b
	}
}

func main() {
	c := make(chan int)

	go fibs(c)

	for i := 0; i < 20; i++ {
		fmt.Println(<-c)
	}
}
