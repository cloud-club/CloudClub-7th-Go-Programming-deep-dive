package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	pb "example.com/grpc-chat-app/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ChatClient: gRPC 채팅 클라이언트
type ChatClient struct {
	client pb.BroadcastClient
	conn   *grpc.ClientConn
	userID string
}

// NewChatClient: 새로운 채팅 클라이언트 생성
func NewChatClient(serverAddr, userID string) (*ChatClient, error) {
	// gRPC 서버에 연결
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	client := pb.NewBroadcastClient(conn)

	return &ChatClient{
		client: client,
		conn:   conn,
		userID: userID,
	}, nil
}

// Close: 클라이언트 연결 종료
func (c *ChatClient) Close() error {
	return c.conn.Close()
}

// ConnectAndListen: 서버에 연결하고 메시지 수신 대기
func (c *ChatClient) ConnectAndListen(ctx context.Context) error {
	// 서버에 연결 요청
	connectMsg := &pb.Connect{
		User: &pb.User{
			Id:   c.userID,
			Name: fmt.Sprintf("User-%s", c.userID),
		},
		Active: true,
	}

	// 스트림 생성
	stream, err := c.client.CreateStream(ctx, connectMsg)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	log.Printf("Connected to server as user: %s", c.userID)

	// 메시지 수신 루프
	for {
		select {
		case <-ctx.Done():
			log.Printf("Client %s disconnecting...", c.userID)
			return ctx.Err()
		default:
			msg, err := stream.Recv()
			if err == io.EOF {
				log.Printf("Server closed the stream for user: %s", c.userID)
				return nil
			}
			if err != nil {
				return fmt.Errorf("failed to receive message: %w", err)
			}

			// 수신한 메시지 출력 (개선된 형식)
			timestamp := "unknown"
			if msg.Timestamp != nil {
				timestamp = msg.Timestamp.AsTime().Format("15:04:05")
			}

			// 내가 보낸 메시지인지 확인
			if msg.Id == c.userID {
				fmt.Printf("\r\033[K✓ [%s] You: %s\n[%s] > ", timestamp, msg.Content, c.userID)
			} else {
				fmt.Printf("\r\033[K📩 [%s] %s: %s\n[%s] > ", timestamp, msg.Id, msg.Content, c.userID)
			}
		}
	}
}

// SendMessage: 메시지 전송
func (c *ChatClient) SendMessage(ctx context.Context, content string) error {
	msg := &pb.Message{
		Id:        c.userID,
		Content:   content,
		Timestamp: timestamppb.Now(),
	}

	_, err := c.client.BroadcastMessage(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// 메시지 전송 성공 (서버에서 다시 받아서 표시되므로 여기서는 로그 제거)
	return nil
}

// runClient: 클라이언트 실행 함수
func runClient(userID string) {
	serverAddr := "localhost:8081"

	// 클라이언트 생성
	client, err := NewChatClient(serverAddr, userID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown 처리
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 연결 및 메시지 수신을 별도 고루틴에서 실행
	go func() {
		if err := client.ConnectAndListen(ctx); err != nil {
			if ctx.Err() == nil { // context가 취소되지 않은 상태에서의 에러만 로깅
				log.Printf("Listen error: %v", err)
			}
		}
	}()

	// 잠시 대기하여 연결 안정화
	time.Sleep(1 * time.Second)

	fmt.Printf("\n=== Chat Client Started ===\n")
	fmt.Printf("User: %s\n", userID)
	fmt.Printf("Server: %s\n", serverAddr)
	fmt.Printf("Commands:\n")
	fmt.Printf("  Type message and press Enter to send\n")
	fmt.Printf("  Type '/quit' or press Ctrl+C to exit\n")
	fmt.Printf("===========================\n\n")

	// 사용자 입력을 받기 위한 스캐너
	scanner := bufio.NewScanner(os.Stdin)

	// 사용자 입력 처리 루프
	go func() {
		for {
			fmt.Printf("[%s] > ", userID)

			if !scanner.Scan() {
				// 입력 종료 (EOF)
				cancel()
				return
			}

			input := strings.TrimSpace(scanner.Text())

			// 빈 입력 무시
			if input == "" {
				continue
			}

			// 종료 명령어 처리
			if input == "/quit" || input == "/exit" {
				fmt.Println("Goodbye!")
				cancel()
				return
			}

			// 메시지 전송
			if err := client.SendMessage(ctx, input); err != nil {
				if ctx.Err() == nil { // context가 취소되지 않은 상태에서의 에러만 로깅
					log.Printf("Failed to send message: %v", err)
				}
				continue
			}
		}
	}()

	// 종료 시그널 또는 컨텍스트 취소 대기
	select {
	case <-sigChan:
		fmt.Println("\nReceived interrupt signal. Shutting down...")
		cancel()
	case <-ctx.Done():
		// 컨텍스트 취소됨 (사용자가 /quit 입력하거나 다른 이유)
	}

	// 정리 대기
	time.Sleep(500 * time.Millisecond)
	log.Printf("Client %s disconnected", userID)
}

// main 함수 - 클라이언트를 독립적으로 실행하기 위해 주석 해제됨
func init() {
	// main.go와 함께 빌드될 때는 이 함수가 실행되지 않도록 함
}

// runClientMain: 독립 실행 시 사용
func runClientMain() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run client.go <user_id>")
		fmt.Println("Example: go run client.go alice")
		return
	}

	userID := os.Args[1]
	runClient(userID)
}
