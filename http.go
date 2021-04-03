package main

import (
	"bytes"
	"fmt"
	"github.com/buildkite/terminal-to-html"
	"net"
	"net/http"
	"strconv"
)

func NewTcpListener(port, maxPort int) (net.Listener, string) {
	for {
		addr := ":" + strconv.Itoa(port)
		conn, err := net.Listen("tcp", addr)
		if err != nil {
			port++

			if port > maxPort {
				panic("can not open tcp listener")
			}

			continue
		}

		if port == 80 {
			addr = ""
		}

		return conn, "http://localhost" + addr
	}
}

func NewMux(runners []*Runner) *http.ServeMux {
	mux := http.NewServeMux()
	for _, r := range runners {
		mux.Handle("/"+r.Addr, NewHandler(r.Addr, r.Broadcast))
	}
	return mux
}

func NewHandler(title string, broadcast *Broadcast) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html")

		flusher := writer.(http.Flusher)

		_, err := writer.Write([]byte(fmt.Sprintf(`
<html>
<head>
	<title>%s</title>
	<!-- https://raw.githubusercontent.com/buildkite/terminal/v3.2.0/assets/terminal.css -->
	<link rel="stylesheet" type="text/css" href="http://gitcdn.link/repo/buildkite/terminal/v3.2.0/assets/terminal.css">
	<style>
		* { padding: 0; margin: 0; border: 0; outline: 0; }
	</style>
</head>
<body class="term-container">
`, title)))
		if err != nil {
			return // client disconnected?
		}
		flusher.Flush()

		ch := broadcast.Subscribe()

		for {
			msg, ok := <-ch
			if !ok {
				writer.Write([]byte("</body></html>"))
				return
			}

			_, err := writer.Write(renderAnsi(msg))
			if err != nil {
				broadcast.Unsubscribe(ch)
			}

			flusher.Flush()
		}
	}
}

func renderAnsi(data []byte) []byte {
	return bytes.ReplaceAll(terminal.Render(data), []byte("\n"), []byte("<br>"))
}
