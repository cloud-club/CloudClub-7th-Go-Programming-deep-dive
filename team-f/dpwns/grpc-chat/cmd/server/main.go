package main

import (
    "log"
    "net"
    "net/http"
    
    adapter "grpc-chat/internal/adapter/grpc"
    
    "grpc-chat/internal/usecase"
    "grpc-chat/infrastructure/memory"
    
    pb "grpc-chat/gen"
    "github.com/grpc-ecosystem/go-grpc-prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    grpc "google.golang.org/grpc"                   // gRPC 프레임워크용 import 별도 지정
)

func main() {
    repo := memory.NewSessionRepo()	// 메모리 세션 저장소 생성
    uc := usecase.NewChatUsecase(repo)	// 유즈케이스에 저장소 주입
    handler := adapter.NewChatHandler(uc, repo)	// gRPC 핸들러에 유즈케이스 + 저장소 주입

    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
	    log.Fatalf("listen failed: %v", err)
    }

    // 1. Prometheus /metrics HTTP 서버 시작Add commentMore actions
    go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("📈 Prometheus 메트릭 노출: http://localhost:2112/metrics")
		log.Fatal(http.ListenAndServe(":2112", nil))
    }()
    // 프로메테우스 등록
    srv := grpc.NewServer(
        grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
        grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
    )
    pb.RegisterChatServiceServer(srv, handler)
    grpc_prometheus.Register(srv)

    log.Println("gRPC Chat Server start at :50051")
    if err := srv.Serve(lis); err != nil {
        log.Fatalf("server error: %v", err)
    }
}

