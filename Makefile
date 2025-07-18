# docker image



# build binary
.PHONY: adapter
adapter:
	@echo === building alertmanager-hook-adapter binary
	@CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -ldflags="-s -w" -o alertmanager-hook-adapter  ./cmd/main.go


.PHONY: image
image: adapter
	@echo === running docker build
	@docker build --no-cache -t alertmanager-hook-adapter:v2 -f deploy/image/Dockerfile .

.PHONY: clean
clean:
	rm -f ./alertmanager-hook-adapter
	go clean -testcache
