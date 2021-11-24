package main

import (
	"bytes"
	_ "embed"
	"github.com/buildkite/terminal-to-html"
	"net"
	"net/http"
	"strconv"
	"text/template"
)

//go:embed head.html
var headHtml string

//go:embed terminal.css
var terminalCss string

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

		buf := bytes.NewBuffer(nil)
		t := template.Must(template.New(title).Parse(headHtml))
		err := t.Execute(buf, map[string]string{
			"title": title,
			"css":   terminalCss,
		})
		if err != nil {
			panic(err)
		}

		_, err = writer.Write(buf.Bytes())
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
