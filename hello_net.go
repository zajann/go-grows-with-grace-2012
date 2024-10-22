package main

import (
	"fmt"
	"log"
	"net"
)

const listenAddr = "localhost:5555"

func main() {
	l, err := net.Listen("tcp", listenAddr)

	if err != nil {
		log.Fatal(err)
	}
	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintln(c, "Hello!")
		c.Close()
	}
}
