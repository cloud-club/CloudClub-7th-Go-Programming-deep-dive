package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"example.com/grpc-chat-app/internal/config"
	"example.com/grpc-chat-app/internal/server"
)

func main() {
	// 명령행 인수 처리
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "consumer":
			runConsumer()
			return
		case "web":
			runWebServer()
			return
		default:
			fmt.Println("Available modes:")
			fmt.Println("  server   - Run gRPC chat server (default)")
			fmt.Println("  consumer - Run Kafka message consumer")
			fmt.Println("  web      - Run web gateway server")
			return
		}
	}

	// 기본: gRPC 서버 실행
	runGRPCServer()
}

// runGRPCServer: gRPC 서버 실행
func runGRPCServer() {
	cfg := config.NewConfig()

	chatServer, err := server.NewChatServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create chat server: %v", err)
	}
	defer chatServer.Close()

	fmt.Printf("🚀 gRPC Chat Server (%s) starting at port %s\n", cfg.ServerID, cfg.GRPCPort)
	fmt.Printf("📡 Kafka Brokers: %v\n", cfg.Kafka.Brokers)
	fmt.Printf("📂 Kafka Topic: %s\n", cfg.Kafka.Topic)

	if err := chatServer.Start(); err != nil {
		log.Fatalf("Failed to start chat server: %v", err)
	}
}

// runWebServer: 웹 게이트웨이 서버 실행
func runWebServer() {
	cfg := config.NewConfig()

	webServer := server.NewWebServer(cfg)

	fmt.Printf("🌐 Web Gateway Server starting at port %s\n", cfg.HTTPPort)
	fmt.Printf("📁 Static files served from: %s\n", cfg.WebRoot)
	fmt.Printf("🔗 WebSocket endpoint: ws://localhost%s/ws\n", cfg.HTTPPort)

	if err := webServer.Start(); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}

// runConsumer: 독립 컨슈머 실행
func runConsumer() {
	cfg := config.NewConfig()

	consumer, err := server.NewStandaloneConsumer(cfg)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	fmt.Printf("🔄 Standalone Consumer starting\n")
	fmt.Printf("📡 Kafka Brokers: %v\n", cfg.Kafka.Brokers)
	fmt.Printf("👥 Consumer Group: %s\n", cfg.Kafka.ConsumerGroup)

	if err := consumer.Start(); err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}
}
