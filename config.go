package main

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
)

type ServerInfo struct {
	Mask   string
	Params map[string][]string
}

type Codes struct {
	ok      []int
	warning []int
}

type PortRange struct {
	min int
	max int
}

type Config struct {
	Command   string
	Codes     *Codes
	PortRange *PortRange
	Args      []string
	Servers   map[string]*ServerInfo
	Aliases   map[string][]string
}

func NewConfig(configYaml string) *Config {
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBufferString(configYaml))
	if err != nil {
		panic(err)
	}

	codes := viper.Sub("codes")
	portRange := viper.Sub("portRange")
	config := &Config{
		Command: viper.GetString("command"),
		Codes: &Codes{
			ok:      codes.GetIntSlice("ok"),
			warning: codes.GetIntSlice("warning"),
		},
		PortRange: &PortRange{
			min: portRange.GetInt("min"),
			max: portRange.GetInt("max"),
		},
		Args:    viper.GetStringSlice("args"),
		Servers: parseServers(),
		Aliases: viper.GetStringMapStringSlice("aliases"),
	}

	params := viper.GetStringMapStringSlice("params")
	// fill default params to each serverInfo
	for _, server := range config.Servers {
		for name, val := range params {
			if _, ok := server.Params[name]; !ok {
				server.Params[name] = val
			}
		}
	}

	// check aliased server existence
	for _, serverInfos := range config.Aliases {
		for _, server := range serverInfos {
			if _, exists := config.Servers[server]; !exists {
				panic(fmt.Errorf("alias: server \"%s\" not exists", server))
			}
		}
	}

	return config
}

func parseServers() map[string]*ServerInfo {
	servers := make(map[string]*ServerInfo)
	serversConfig := viper.GetStringMap("servers")

	for name, row := range serversConfig {
		var mask string
		params := map[string][]string{}

		switch row.(type) {
		case string:
			mask = viper.Sub("servers").GetString(name)
		case interface{}:
			sub := viper.Sub("servers").Sub(name)
			mask = sub.GetString("mask")
			params = sub.GetStringMapStringSlice("params")
		default:
			panic("servers config is malformed")
		}

		servers[name] = &ServerInfo{
			Mask:   mask,
			Params: params,
		}
	}

	return servers
}
