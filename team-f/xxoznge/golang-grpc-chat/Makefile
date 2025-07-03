init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

generate:
	protoc --go_out=. --go-grpc_out=. --proto_path=./protos protos/chat.proto

run:
	go run cmd/main.go
