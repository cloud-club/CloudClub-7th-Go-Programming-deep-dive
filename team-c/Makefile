# CloudClub Chat - 간단한 Makefile

.PHONY: help start stop server server-stop

# Variables
SERVER_DIR=server
CLIENT_DIR=client
INFRA_DIR=infra
BINARY_DIR=bin
CHAT_SERVER_BINARY=$(BINARY_DIR)/chat-server

help: ## 사용 가능한 명령어 보기
	@echo "CloudClub Chat - 실시간 채팅 시스템"
	@echo ""
	@echo "사용 가능한 명령어:"
	@echo "  start        - 전체 프로젝트 실행 (Kafka + Go서버 + Next.js)"
	@echo "  stop         - 전체 프로젝트 중단"
	@echo "  server       - Go 서버만 시작"
	@echo "  server-stop  - Go 서버만 중단"

# 프로젝트 전체 실행
start: ## 전체 프로젝트 실행
	@echo "🚀 CloudClub Chat 전체 시스템 시작..."
	@echo "1️⃣ Kafka 인프라 시작..."
	cd $(INFRA_DIR) && docker-compose up -d
	@echo "⏳ Kafka 시작 대기 중... (20초)"
	@sleep 20
	@echo "✅ Kafka 준비 완료!"
	@echo "2️⃣ Protobuf 컴파일..."
	cd $(SERVER_DIR) && protoc --go_out=pkg/gen --go_opt=paths=source_relative \
		--go-grpc_out=pkg/gen --go-grpc_opt=paths=source_relative \
		api/proto/*.proto
	@echo "3️⃣ Go 서버 빌드 및 시작..."
	@mkdir -p $(BINARY_DIR)
	cd $(SERVER_DIR) && go build -o ../$(CHAT_SERVER_BINARY) ./cmd/server
	@echo "🚀 gRPC 서버 시작 (포트 8081)..."
	$(CHAT_SERVER_BINARY) &
	@sleep 3
	@echo "🌐 WebSocket Gateway 시작 (포트 8080)..."
	$(CHAT_SERVER_BINARY) web &
	@sleep 2
	@echo "4️⃣ Next.js 클라이언트 시작..."
	cd $(CLIENT_DIR) && npm install > /dev/null 2>&1 && npm run dev &
	@echo ""
	@echo "✅ 시스템이 시작되었습니다!"
	@echo "🌐 접속 URL:"
	@echo "  - Next.js 클라이언트: http://localhost:3000"
	@echo "  - Kafka UI: http://localhost:8088"
	@echo ""
	@echo "💡 시스템을 중지하려면: make stop"

# 프로젝트 전체 중단
stop: ## 전체 프로젝트 중단
	@echo "⏹️ CloudClub Chat 전체 시스템 중지..."
	-pkill -f "chat-server"
	-pkill -f "next-server"
	-pkill -f "npm run dev"
	cd $(INFRA_DIR) && docker-compose down
	@echo "✅ 모든 서비스가 중지되었습니다"

# Go 서버만 시작
server: ## Go 서버만 시작
	@echo "🚀 Go 채팅 서버 시작..."
	@mkdir -p $(BINARY_DIR)
	cd $(SERVER_DIR) && protoc --go_out=pkg/gen --go_opt=paths=source_relative \
		--go-grpc_out=pkg/gen --go-grpc_opt=paths=source_relative \
		api/proto/*.proto > /dev/null 2>&1
	cd $(SERVER_DIR) && go build -o ../$(CHAT_SERVER_BINARY) ./cmd/server
	@echo "✅ 서버가 시작되었습니다 (포트 8080, 8081)"
	$(CHAT_SERVER_BINARY)

# Go 서버만 중단
server-stop: ## Go 서버만 중단
	@echo "⏹️ Go 채팅 서버 중지..."
	-pkill -f "chat-server"
	@echo "✅ Go 서버가 중지되었습니다"

.DEFAULT_GOAL := help 