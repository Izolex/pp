package app

import (
	"bytes"
	"fmt"
	"html/template"
)

func NewRunners(config *Config, cliArgs [][]string, serverNames []string) []*Runner {
	list := make([]*Runner, 0)
	used := map[string]struct{}{}
	argsCom := combine(newArgsMap(config.Args, cliArgs))

	for _, info := range newServerInfosList(config, serverNames) {
		paramsCom := combine(info.Params)

		for _, args := range argsCom {
			for _, params := range paramsCom {
				data := newTemplateData(args, params)
				addr := parseAddr(info.Mask, data)

				if _, exists := used[addr]; !exists {
					used[addr] = struct{}{}
					list = append(list, NewRunner(addr))
				}
			}
		}
	}

	return list
}

func newServerInfosList(config *Config, serverNames []string) []*ServerInfo {
	servers := make([]*ServerInfo, 0, len(serverNames))
	for _, serverName := range serverNames {
		if aliases, isAlias := config.Aliases[serverName]; isAlias {
			for _, server := range aliases {
				servers = append(servers, config.Servers[server])
			}
			continue
		}

		if _, ok := config.Servers[serverName]; !ok {
			panic(fmt.Errorf("server \"%s\" does not exists", serverName))
		}

		servers = append(servers, config.Servers[serverName])
	}
	return servers
}

func newArgsMap(configArgs []string, args [][]string) map[string][]string {
	argsMap := make(map[string][]string)
	for i, argsList := range args {
		name := configArgs[i]
		argsMap[name] = argsList
	}
	return argsMap
}

func newTemplateData(args, params map[string]string) map[string]string {
	data := make(map[string]string)

	for name, value := range args {
		data[name] = value
	}
	for name, value := range params {
		data[name] = value
	}

	return data
}

func parseAddr(mask string, data map[string]string) string {
	t := template.Must(template.New(mask).Parse(mask))

	buf := bytes.NewBuffer(nil)
	err := t.Execute(buf, data)

	if err != nil {
		panic(err)
	}

	return buf.String()
}

func combine(data map[string][]string) []map[string]string {
	result := make([]map[string]string, 0)

	first := true
	for name, dataVal := range data {
		// init fill res
		if first { // for over map is unstable
			for _, val := range data[name] {
				result = append(result, map[string]string{
					name: val,
				})
			}

			first = false
			continue
		}

		filler := make([]map[string]string, 0)

		// combine each item of res with current item of arr
		for _, resultVal := range result {
			for _, val := range dataVal {
				resultVal[name] = val
				filler = append(filler, resultVal)
			}
		}

		result = filler
	}

	return result
}
