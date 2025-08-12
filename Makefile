# docker image
GIT_COMMIT := $(shell git rev-parse --short HEAD)


# build binary
.PHONY: adapter
adapter:
	@echo === building alertmanager-hook-adapter binary
	@CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -ldflags="-s -w" -o alertmanager-hook-adapter  ./cmd/main.go


.PHONY: image
image: adapter
	@echo === running docker build
	@docker build --no-cache -t alertmanager-hook-adapter:${GIT_COMMIT} -f deploy/image/Dockerfile .

.PHONY: clean
clean:
	rm -f ./alertmanager-hook-adapter
	go clean -testcache

.PHONY: lint
lint:
	golangci-lint run --timeout=5m

.PHONY: test
test:
	go test ./... -v
