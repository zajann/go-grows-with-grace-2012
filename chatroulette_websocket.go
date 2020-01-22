package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"html/template"

	"golang.org/x/net/websocket"
)

// const listenAddr = "localhost:4000"
const listenAddr = "0.0.0.0:5555"

type socket struct {
	io.ReadWriter
	done chan bool
}

/*
// By embedding the *websocket.Conn as an io.ReadWriter, we cat drop the explicit socket Read and Write methods
func (s socket) Read(b []byte) (int, error) {
	return s.conn.Read(b)
}

func (s socket) Write(b []byte) (int, error) {
	return s.conn.Write(b)
}
*/

func (s socket) Close() error {
	s.done <- true
	return nil
}

var partner = make(chan io.ReadWriteCloser)

var htmlTemplate = template.Must(template.New("root").Parse(`
	<!DOCTYPE html>
	<html>
		<head>
			<title>Websocket chat - Golang</title>
			<meta charset="utf-8" />
			<script>

			var input, output, websocket;

			function showMessage(m) {
			        var p = document.createElement("p");
			        p.innerHTML = m;
			        output.appendChild(p);
			}

			function onMessage(e) {
			        showMessage(e.data);
			}

			function onOpen(e) {
					showMessage("Connection opend.")
			}

			function onClose() {
			        showMessage("Connection closed.");
			}

			function sendMessage() {
			        var m = input.value;
			        input.value = "";
			        websocket.send(m + "\n");
			        showMessage(m);
			}

			function onKey(e) {
			        if (e.keyCode == 13) {
			                sendMessage();
			        }
			}

			function init() {
			        input = document.getElementById("input");
			        input.addEventListener("keyup", onKey, false);

			        output = document.getElementById("output");

					websocket = new WebSocket("ws://61.79.198.2:5555/socket");
			        websocket.onmessage = onMessage;
					websocket.onopen = onOpen;
			        websocket.onclose = onClose;
			}

			window.addEventListener("load", init, false);
			</script>
		</head>
		<body>
			Say: <input id="input" type="text">
			<div id="output"></div>
		</body>
	</html>
`))

func httpHandler(w http.ResponseWriter, r *http.Request) {
	htmlTemplate.Execute(w, listenAddr)
}

func socketHandler(ws *websocket.Conn) {
	s := socket{ws, make(chan bool)}
	go match(s)
	<-s.done
}

func match(c io.ReadWriteCloser) {
	fmt.Fprintln(c, "Waiting for a partner ...")

	select {
	case partner <- c:
		// handled by the other goroutine
	case p := <-partner:
		chat(p, c)
	}
}

func chat(a, b io.ReadWriteCloser) {
	fmt.Fprintln(a, "Found one! Say hi.")
	fmt.Fprintln(b, "Found one! Say hi.")
	errc := make(chan error, 1)
	go cp(a, b, errc)
	go cp(b, a, errc)
	if err := <-errc; err != nil {
		log.Println(err)
	}
	a.Close()
	b.Close()
}

func cp(w io.Writer, r io.Reader, errc chan<- error) {
	_, err := io.Copy(w, r)
	errc <- err
}

func main() {
	http.HandleFunc("/", httpHandler)
	http.Handle("/socket", websocket.Handler(socketHandler))
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
