package main

import (
	_ "embed"
	"os"
	"os/signal"
	"pp/app"
	"regexp"
	"strings"
)

//go:embed config.yaml
var configFile string

func main() {
	screen := app.NewScreen(os.Stdout)
	config := app.NewConfig(configFile)
	args, serverNames := parseCliArgs(os.Args[1:])

	pipi := app.NewPP(config, screen, args, serverNames)
	go pipi.Start()
	pipi.Run(config.Command)

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
