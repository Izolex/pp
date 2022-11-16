package app

import (
	"log"
	"net"
	"net/http"
	"os/exec"
)

type PP struct {
	config   *Config
	renderer *Renderer
	mux      *http.ServeMux
	listener net.Listener
	address  string
	runners  []*Runner
}

func NewPP(
	config *Config,
	screen *Screen,
	cliArgs [][]string,
	serverNames []string,
) *PP {
	if len(config.Args) != len(cliArgs) {
		log.Fatal("check arguments :(")
	}

	listener, address := NewTcpListener(config.PortRange.Min, config.PortRange.Max)
	renderer := NewRenderer(screen)
	runners := NewRunners(config, cliArgs, serverNames)
	mux := NewServeMux(runners)

	return &PP{
		config,
		renderer,
		mux,
		listener,
		address,
		runners,
	}
}

func (pp *PP) Start() {
	err := http.Serve(pp.listener, pp.mux)
	if err != nil {
		panic(err)
	}
}

func (pp *PP) Run(command string) {
	for _, r := range pp.runners {
		cmd := exec.Command("ssh", r.Addr, command)
		go r.Exec(cmd, pp.config.Codes)
	}

	pp.renderer.Render(pp.address, pp.runners)
}
