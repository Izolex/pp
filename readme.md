# pp

Run SSH command on multiple servers at once.

- clone this repo
- create your own `config.yaml`
- run `make build.darwin` || `make build.windows` || `make build.linux`
- move `pp` binary file to your local bin dir (`make mv`)
- run (with src/config.example.yaml):
  - `pp dev fe`
  - `pp dev,stage api,web`
