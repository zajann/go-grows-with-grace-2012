package main

import (
	"io"
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
		io.Copy(c, c)
	}
}
