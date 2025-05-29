VERSION:=0.1

.PHONY: build
build:
	go build

.PHONY: run
run:
	go run main.go

.PHONY: format
format:
	go fmt
	
.PHONY: image
image:
	docker build --tag go-server:v${VERSION} .