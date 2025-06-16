// cmd/client.go
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"golang-grpc-chat/grpcapi"

	"google.golang.org/grpc"
)

func main() {
	// 사용자 이름 입력
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your name: ")
	username, _ := reader.ReadString('\n')
	username = username[:len(username)-1]

	// 서버 연결
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := grpcapi.NewChatServiceClient(conn)
	stream, err := client.Chat(context.Background())
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	// 수신 고루틴
	go func() {
		for {
			in, err := stream.Recv()
			if err != nil {
				log.Printf("Error receiving message: %v", err)
				break
			}
			fmt.Printf("[%s] %s\n", in.User, in.Content)
		}
	}()

	// 송신 루프
	for {
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		text = text[:len(text)-1]

		if err := stream.Send(&grpcapi.ChatMessage{
			User:      username,
			Content:   text,
			Timestamp: time.Now().Unix(),
		}); err != nil {
			log.Fatalf("Failed to send message: %v", err)
		}
	}
}
