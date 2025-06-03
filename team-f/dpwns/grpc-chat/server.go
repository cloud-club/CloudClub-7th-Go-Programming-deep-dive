package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"time"

	pb "grpc-chat/chatpb" // chatpb는 go_package로 설정한 경로에 따라 변경

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

type chatServer struct {
	pb.UnimplementedChatServiceServer
}

func (s *chatServer) ChatStream(stream pb.ChatService_ChatStreamServer) error {
	log.Println("📡 새 채팅 스트림 수신")

	for {
		// 클라이언트 메시지 수신
		msg, err := stream.Recv()
		if err == io.EOF {
			log.Println("✅ 클라이언트 연결 종료")
			return nil
		}
		if err != nil {
			log.Printf("❌ 메시지 수신 오류: %v", err)
			return err
		}

		log.Printf("[%s]: %s", msg.User, msg.Message)

		// 응답 메시지 전송
		resp := &pb.ChatMessage{
			User:      "Server",
			Message:   "Echo: " + msg.Message,
			Timestamp: time.Now().Unix(),
		}
		if err := stream.Send(resp); err != nil {
			log.Printf("❌ 메시지 전송 오류: %v", err)
			return err
		}
	}
}

func main() {
	// 1. Prometheus /metrics HTTP 서버 시작
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("📈 Prometheus 메트릭 노출: http://localhost:2112/metrics")
		log.Fatal(http.ListenAndServe(":2112", nil))
	}()

	// 2. gRPC 서버 초기화
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("포트 열기 실패: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)

	// 3. 서비스 등록 및 메트릭 등록
	pb.RegisterChatServiceServer(grpcServer, &chatServer{})
	grpc_prometheus.Register(grpcServer)

	log.Println("🚀 gRPC 채팅 서버 시작 (50051)")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("서버 종료: %v", err)
	}
}

