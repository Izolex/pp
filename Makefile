.DEFAULT_GOAL=no_target

no_target:
	@echo "Grr..."


build.darwin:
	$(call build, darwin)

build.windows:
	$(call build, windows)

build.linux:
	$(call build, linux)

mv:
	mv $(shell pwd)/pp /usr/local/bin


define build
	docker run \
		--rm \
		-e GOOS=$(strip $(1)) \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=0 \
		-v $(shell pwd):/app \
		-w /app \
		golang:alpine \
		go build
endef
