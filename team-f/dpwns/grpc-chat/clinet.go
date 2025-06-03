package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"
	pb "grpc-chat/chatpb" // chatpb는 go_package로 설정한 경로에 따라 변경

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("연결 실패: %v", err)
	}
	defer conn.Close()

	client := pb.NewChatServiceClient(conn)

	stream, err := client.ChatStream(context.Background())
	if err != nil {
		log.Fatalf("스트림 생성 실패: %v", err)
	}

	// 메시지 수신 고루틴
	go func() {
		for {
			resp, err := stream.Recv()
			if err != nil {
				log.Printf("❌ 서버 응답 수신 오류: %v", err)
				break
			}
			fmt.Printf("[서버 응답] %s\n", resp.Message)
		}
	}()

	// 사용자 입력 메시지 전송
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("메시지 입력: ")
		scanner.Scan()
		text := scanner.Text()

		if text == "exit" {
			break
		}

		msg := &pb.ChatMessage{
			User:      "Client1",
			Message:   text,
			Timestamp: time.Now().Unix(),
		}
		if err := stream.Send(msg); err != nil {
			log.Printf("❌ 메시지 전송 오류: %v", err)
			break
		}
	}

	stream.CloseSend()
}

