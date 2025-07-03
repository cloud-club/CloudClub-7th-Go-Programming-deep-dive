package main

import (
	"log"
	"net"
	"strings"

	"gRPC-based-chatting/chatProto"
	"gRPC-based-chatting/handler"
	"gRPC-based-chatting/kafka"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func main() {

	// kafka producer 생성
	producer, err := kafka.NewChatProducer("kafka:9092", "chat-topic") // 컨테이너 통신일 때는 서비스명(kafka)로 변경
	if err != nil {
		log.Fatalf("Kafka producer 생성 실패: %v", err)
	}
	defer producer.Close()

	// 핸들러
	chatHandler := handler.NewChatHandler(producer)

	// kafka consumer 생성
	consumer, err := kafka.NewChatConsumer("kafka:9092", "chat-topic", "chat-group-new")
	if err != nil {
		log.Fatalf("Kafka consumer 생성 실패: %v", err)
	}

	// consumer 루프에서 메시지 브로드캐스트
	go consumer.ConsumeLoop(func(key, value []byte) {
		// key에서 채널명(=채팅방명) 추출
		parts := strings.SplitN(string(key), ":", 2)
		channel := parts[0]

		// value를 chatProto.ChatMessage로 역직렬화 (protobuf 사용)
		var msg chatProto.ChatMessage
		if err := proto.Unmarshal(value, &msg); err != nil {
			log.Printf("unmarshal(역직렬화) error: %v", err)
			return
		}

		// 핸들러의 브로드캐스트 함수 호출
		chatHandler.BroadcastFromKafka(channel, &msg)
	})

	// grpc 서버 생성
	grpcServer := grpc.NewServer()
	chatProto.RegisterChatServiceServer(grpcServer, chatHandler)

	// 서버 실행
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("서버 리스닝 실패: %v", err)
	}
	log.Println("gRPC 서버 시작: :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC 서버 종료: %v", err)
	}

}
