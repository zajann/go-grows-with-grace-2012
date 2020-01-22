package main

import (
	"fmt"
	"net/http"

	"golang.org/x/net/websocket"
)

// const listenAddr = "localhost:4000"
const listenAddr = "0.0.0.0:5555"

func main() {
	http.Handle("/", websocket.Handler(handler))
	http.ListenAndServe(listenAddr, nil)
}

func handler(c *websocket.Conn) {
	var s string
	fmt.Fscan(c, &s)
	fmt.Println("Receive: ", s)
	fmt.Fprint(c, "How do you do?")
}
