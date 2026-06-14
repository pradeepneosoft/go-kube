.PHONY: tidy build run-user run-order run-notification proto deploy-local

tidy:
	go mod tidy

build:
	go build -o bin/user-service ./services/user-service/cmd
	go build -o bin/order-service ./services/order-service/cmd
	go build -o bin/notification-service ./services/notification-service/cmd

run-user:
	go run ./services/user-service/cmd

run-order:
	go run ./services/order-service/cmd

run-notification:
	go run ./services/notification-service/cmd

proto:
	protoc \
		--go_out=gen --go_opt=paths=source_relative \
		--go-grpc_out=gen --go-grpc_opt=paths=source_relative \
		proto/user/v1/user.proto proto/order/v1/order.proto

deploy-local:
	./scripts/deploy-local-kind.sh
