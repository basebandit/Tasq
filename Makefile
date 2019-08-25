SHELL := /bin/bash
TARGET_SERVER=todo-server
TARGET_CLIENT_REST=client-rest
TARGET_CLIENT_GRPC=client-grpc


test:
	@go test -v ./pkg/protocol/service/v1/

proto: third_party/protoc-gen.sh
	@chmod u+x ./third_party/protoc-gen.sh
	./third_party/protoc-gen.sh

build:
	@go build -o ./$(TARGET_SERVER) ./cmd/server/main.go
	@go build -o ./$(TARGET_CLIENT_REST) ./cmd/client-rest/main.go
	@go build -o ./$(TARGET_CLIENT_GRPC) ./cmd/client-grpc/main.go


server: todo-server
	./todo-server -grpc-port=9090 -http-port=8080 -host=localhost -user=mars -password=mars -db=ToDo -migrations=. -log-level=-1 -log-time-format=2006-01-02T15:04:05.999999999Z07:00

rest: client-rest
	./client-rest -server=http://localhost:8080

grpc: client-grpc
	./client-grpc -server=localhost:9090

clean: todo-server client-rest client-grpc
	@rm todo-server
	@rm client-rest
	@rm client-grpc
