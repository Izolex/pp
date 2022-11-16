package main

import (
	_ "embed"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
)

//go:embed config.yaml
var configFile string

func main() {
	config := NewConfig(configFile)
	args, serverNames := parseCliArgs(os.Args[1:])

	if len(config.Args) != len(args) {
		log.Fatal("check arguments :(")
	}

	runners := NewRunners(config, args, serverNames)
	listener, addr := NewTcpListener(config.PortRange.min, config.PortRange.max)

	go runHttp(listener, NewServeMux(runners))
	go runSsh(config.Command, config.Codes, runners)

	screen := NewScreen(os.Stdout)
	renderer := NewRenderer(screen)
	renderer.Render(addr, runners)

	screen.Row("Waiting for interrupt...")

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	<-signalCh
}

func parseCliArgs(args []string) ([][]string, []string) {
	sanitized := regexp.MustCompile(` *, *`).ReplaceAllString(strings.Join(args, " "), ",")
	groups := strings.Split(sanitized, " ")

	list := make([][]string, 0, len(groups))
	for _, group := range groups {
		list = append(list, strings.Split(group, ","))
	}

	return list[:len(groups)-1], list[len(groups)-1:][0]
}

func runHttp(listener net.Listener, mux *http.ServeMux) {
	err := http.Serve(listener, mux)
	if err != nil {
		panic(err)
	}
}

func runSsh(command string, codes *Codes, runners []*Runner) {
	for _, r := range runners {
		cmd := exec.Command("ssh", r.Addr, command)
		go r.Exec(cmd, codes)
	}
}
