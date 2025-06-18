# Makefile for grpc-chat-app with Kafka

.PHONY: help kafka-up kafka-down server consumer client clean build proto test

# 기본 목표
help:
	@echo "Available commands:"
	@echo "  kafka-up    - Start Kafka cluster with Docker Compose"
	@echo "  kafka-down  - Stop Kafka cluster"
	@echo "  server      - Run gRPC chat server"
	@echo "  consumer    - Run Kafka message consumer"
	@echo "  client      - Run interactive chat client (specify USER=<username>)"
	@echo "  build       - Build all binaries"
	@echo "  proto       - Generate protobuf code"
	@echo "  test        - Run tests"
	@echo "  clean       - Clean up binaries and logs"
	@echo ""
	@echo "Example usage:"
	@echo "  make kafka-up"
	@echo "  make server"
	@echo "  make consumer"
	@echo "  make client USER=alice    # Interactive chat client"
	@echo ""
	@echo "How to test chat:"
	@echo "  1. Start Kafka: make kafka-up"
	@echo "  2. Start server: make server (in terminal 1)"
	@echo "  3. Start consumer: make consumer (in terminal 2)"
	@echo "  4. Start clients: make client USER=alice (in terminal 3)"
	@echo "                   make client USER=bob (in terminal 4)"
	@echo "  5. Type messages in client terminals and see real-time chat!"
	@echo "  6. Use '/quit' or Ctrl+C to exit clients"

# Kafka 클러스터 시작
kafka-up:
	@echo "Starting Kafka cluster..."
	docker-compose up -d
	@echo "Waiting for Kafka to be ready..."
	sleep 10
	@echo "Kafka cluster is ready!"
	@echo "Kafka UI available at: http://localhost:8080"

# Kafka 클러스터 종료
kafka-down:
	@echo "Stopping Kafka cluster..."
	docker-compose down

# gRPC 서버 실행
server:
	@echo "Starting gRPC chat server..."
	go run .

# Kafka Consumer 실행 (별도 프로그램으로)
consumer:
	@echo "Starting Kafka consumer..."
	@echo "You can also use: go run . consumer"
	go run . consumer

# 채팅 클라이언트 실행 (별도 프로그램으로)
client:
	@if [ -z "$(USER)" ]; then \
		echo "Please specify USER: make client USER=<username>"; \
		echo "Example: make client USER=alice"; \
		exit 1; \
	fi
	@echo "Starting interactive chat client for user: $(USER)"
	@echo "Type messages and press Enter to send"
	@echo "Use '/quit' or Ctrl+C to exit"
	@echo "You can also use: go run . client $(USER)"
	@echo "========================================"
	go run . client $(USER)

# 빌드
build:
	@echo "Building binaries..."
	go build -o bin/grpc-chat-server main.go
	go build -o bin/kafka-consumer consumer.go
	go build -o bin/chat-client client.go
	@echo "Binaries built in bin/ directory"

# Protobuf 코드 생성
proto:
	@echo "Generating protobuf code..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/chat.proto
	@echo "Protobuf code generated"

# 테스트 실행
test:
	@echo "Running tests..."
	go test ./...

# 정리
clean:
	@echo "Cleaning up..."
	rm -rf bin/
	rm -f *.log
	@echo "Cleanup completed"

# 의존성 설치
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# 전체 환경 설정 (처음 실행 시)
setup: deps proto
	@echo "Environment setup completed!"

# 개발 환경 시작 (Kafka + Server)
dev: kafka-up
	@echo "Waiting for Kafka to be fully ready..."
	sleep 15
	@echo "Starting development environment..."
	@echo "Run 'make consumer' in another terminal"
	@echo "Run 'make client USER=<username>' to test"
	make server

# 모든 서비스 중지
stop:
	@echo "Stopping all services..."
	pkill -f "go run main.go" || true
	pkill -f "go run consumer.go" || true
	pkill -f "go run client.go" || true
	make kafka-down
	@echo "All services stopped" 