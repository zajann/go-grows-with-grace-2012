package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"html/template"

	"golang.org/x/net/websocket"
)

// const listenAddr = "localhost:4000"
const listenAddr = "0.0.0.0:5555"

// Prefix is a Markov chain prefix of one or more words.
type Prefix []string

// String returns the Prefix as a string (for use as a map key).
func (p Prefix) String() string {
	return strings.Join(p, " ")
}

// Shift removes the first word from the Prefix and appends the given word.
func (p Prefix) Shift(word string) {
	copy(p, p[1:])
	p[len(p)-1] = word
}

// Chain contains a map ("chain") of prefixes to a list of suffixes.
// A prefix is a string of prefixLen words joined with spaces.
// A suffix is a single word. A prefix can have multiple suffixes.
type Chain struct {
	chain     map[string][]string
	prefixLen int
}

// NewChain returns a new Chain with prefixes of prefixLen words.
func NewChain(prefixLen int) *Chain {
	return &Chain{make(map[string][]string), prefixLen}
}

// Build reads text from the provided Reader and
// parses it into prefixes and suffixes that are stored in Chain.
func (c *Chain) Build(r io.Reader) {
	br := bufio.NewReader(r)
	p := make(Prefix, c.prefixLen)
	for {
		var s string
		if _, err := fmt.Fscan(br, &s); err != nil {
			break
		}
		key := p.String()
		c.chain[key] = append(c.chain[key], s)
		p.Shift(s)
	}
}

// Write parses the bytes into prefixes and suffixes that are stored in Chain.
func (c *Chain) Write(b []byte) (int, error) {
	r := bytes.NewReader(b)
	go c.Build(r)
	return len(b), nil
}

// Generate returns a string of at most n words generated from Chain.
func (c *Chain) Generate(n int) string {
	p := make(Prefix, c.prefixLen)
	var words []string
	for i := 0; i < n; i++ {
		choices := c.chain[p.String()]
		if len(choices) == 0 {
			break
		}
		next := choices[rand.Intn(len(choices))]
		words = append(words, next)
		p.Shift(next)
	}
	return strings.Join(words, " ")
}

type socket struct {
	io.Reader
	io.Writer
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

var chain = NewChain(2) // 2-word prefixes
func socketHandler(ws *websocket.Conn) {
	r, w := io.Pipe()
	go func() {
		_, err := io.Copy(io.MultiWriter(w, chain), ws)
		w.CloseWithError(err)
	}()
	s := socket{r, ws, make(chan bool)}
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
	case <-time.After(5 * time.Second):
		chat(Bot(), c)
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

// Bot returns an io.ReadWriteCloser that responds to
// each incoming write with a generated sentence.
func Bot() io.ReadWriteCloser {
	r, out := io.Pipe() // for outgoing data
	return bot{r, out}
}

type bot struct {
	io.ReadCloser
	out io.Writer
}

func (b bot) Write(buf []byte) (int, error) {
	go b.speak()
	return len(buf), nil
}

func (b bot) speak() {
	time.Sleep(time.Second)
	msg := chain.Generate(10) // at most 10 words
	b.out.Write([]byte(msg))
}

func main() {
	http.HandleFunc("/", httpHandler)
	http.Handle("/socket", websocket.Handler(socketHandler))
	err := http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
