package app

import (
	"errors"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[0;32m"
	colorRed    = "\033[0;31m"
	colorYellow = "\033[0;33m"

	stateOk   = colorGreen + "\u2713" + colorReset
	stateErr  = colorRed + "\u2717" + colorReset
	stateWarn = colorYellow + "\u26A0" + colorReset
)

type Renderer struct {
	screen *Screen
}

func NewRenderer(screen *Screen) *Renderer {
	return &Renderer{screen}
}

func (r *Renderer) Render(addr string, runners []*Runner) {
	r.list(addr, runners)
	r.report(runners)
}

func (r *Renderer) list(addr string, runners []*Runner) {
	rendered := false
	loaderPos := 0

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	r.screen.Row("Started at %s\n", time.Now().Format(time.Kitchen))
	for {
		<-ticker.C

		if rendered {
			r.screen.CursorUp(len(runners))
		}

		running := false

		for _, runner := range runners {
			var state string
			switch runner.GetState() {
			case RunnerStateOK:
				state = stateOk
			case RunnerStateWarning:
				state = stateWarn
			case RunnerStateError:
				state = stateErr
			case RunnerStateRunning:
				state = string(`-\|/`[loaderPos])
				running = true
			}
			r.screen.Row("[%s] %s/%s", state, addr, runner.Addr)
		}

		if !running {
			return
		}

		rendered = true
		loaderPos++
		if loaderPos > 3 {
			loaderPos = 0
		}
	}
}

func (r *Renderer) report(runners []*Runner) {
	result := stateErr
	ok, err, warn := 0, 0, 0

	for _, runner := range runners {
		switch runner.GetState() {
		case RunnerStateOK:
			ok++
		case RunnerStateError:
			err++
		case RunnerStateWarning:
			warn++
		default:
			panic(errors.New("something bad happened"))
		}
	}

	if err == 0 {
		if warn == 0 {
			result = stateOk
		} else {
			result = stateWarn
		}
	}

	r.screen.Row(
		"\n[%s] Finished at %s - OK: %s%d%s Warn: %s%d%s Err: %s%d%s",
		result,
		time.Now().Format(time.Kitchen),
		colorGreen,
		ok,
		colorReset,
		colorYellow,
		warn,
		colorReset,
		colorRed,
		err,
		colorReset,
	)
}
