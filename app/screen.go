package app

import (
	"fmt"
	"io"
)

type Screen struct {
	output io.Writer
}

func NewScreen(output io.Writer) *Screen {
	return &Screen{output}
}

func (s *Screen) Row(format string, a ...interface{}) {
	s.write(format+"\n", a...)
}

func (s *Screen) CursorUp(lines int) {
	s.write("\033[%dA", lines)
}

func (s *Screen) write(format string, a ...interface{}) {
	_, err := fmt.Fprintf(s.output, format, a...)
	if err != nil {
		panic(err)
	}
}
