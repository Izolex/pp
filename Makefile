.DEFAULT_GOAL=no_target

no_target:
	@echo "Grr..."


build.go:
	go build -o "$(shell pwd)"

build.silicon:
	$(call build, arm64, darwin)

build.darwin:
	$(call build, amd64, darwin)

build.windows:
	$(call build, amd64, windows)

build.linux:
	$(call build, amd64, linux)

mv:
	mv "$(shell pwd)/pp" /usr/local/bin


define build
	docker run \
		--rm \
		-e GOOS=$(strip $(2)) \
		-e GOARCH=$(strip $(1)) \
		-e CGO_ENABLED=0 \
		-v "$(shell pwd):/app" \
		-w /app \
		golang:alpine \
		go build -o /app/pp -buildvcs=false
endef
