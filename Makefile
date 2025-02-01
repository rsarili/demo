SERVER_SRC=src/server
CLIENT_SRC=src/client

.PHONY: build-server
build:
	cd $(SERVER_SRC) && go build

.PHONY: run-server
run:
	cd $(SERVER_SRC) && go run main.go


.PHONY: build-client
build-client:
	cd $(CLIENT_SRC) && cargo build

.PHONY: run-client
run-client:
	cd $(CLIENT_SRC) && cargo run