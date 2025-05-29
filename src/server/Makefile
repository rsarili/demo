VERSION:=0.1

.PHONY: build-server
build:
	go build

.PHONY: run-server
run-server:
	go run main.go

.PHONY: format
format:
	go fmt
	
.PHONY: build-image
build-image:
	docker build --tag go-server:v${VERSION} .