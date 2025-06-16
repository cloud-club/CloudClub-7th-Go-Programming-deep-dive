package main

import (
	"log"
	"net"

	"golang-grpc-chat/grpcapi"
	"golang-grpc-chat/handler"

	"google.golang.org/grpc"

	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Prometheus exporter 시작
	startMetricsEndpoint()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}

	server := grpc.NewServer()
	grpcapi.RegisterChatServiceServer(server, handler.NewChatHandler())

	log.Println("gRPC chat server running on :50051")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func startMetricsEndpoint() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil) // Prometheus가 수집할 포트
	}()
}
