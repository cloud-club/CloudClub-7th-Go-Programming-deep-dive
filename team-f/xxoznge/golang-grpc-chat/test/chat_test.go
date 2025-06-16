package test

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"golang-grpc-chat/grpcapi"
	"golang-grpc-chat/handler"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	grpcapi.RegisterChatServiceServer(s, handler.NewChatHandler())

	go func() {
		log.Println("🟢 gRPC server starting...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()
}

func TestTwoUsers_Chat(t *testing.T) {
	ctx := context.Background()

	conn1, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("conn1 failed: %v", err)
	}
	client1 := grpcapi.NewChatServiceClient(conn1)
	stream1, err := client1.Chat(ctx)
	if err != nil {
		t.Fatalf("stream1 failed: %v", err)
	}

	conn2, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("conn2 failed: %v", err)
	}
	client2 := grpcapi.NewChatServiceClient(conn2)
	stream2, err := client2.Chat(ctx)
	if err != nil {
		t.Fatalf("stream2 failed: %v", err)
	}

	recvChan := make(chan *grpcapi.ChatMessage, 1)
	go func() {
		msg, err := stream2.Recv()
		if err != nil {
			t.Logf("유저2 Recv 에러: %v", err)
			recvChan <- nil
			return
		}
		recvChan <- msg
	}()

	time.Sleep(200 * time.Millisecond) // ensure stream2 is ready

	sendMsg := &grpcapi.ChatMessage{
		User:      "user1",
		Content:   "hello, User2 !!!!!",
		Timestamp: time.Now().Unix(),
	}

	t.Logf("user1 보낸 메시지: %s", sendMsg.Content)

	err = stream1.Send(sendMsg)
	if err != nil {
		t.Fatalf("유저1 메시지 전송 실패: %v", err)
	}

	select {
	case msg := <-recvChan:
		if msg == nil {
			t.Error("유저2 메시지 수신 실패 (nil)")
		} else {
			t.Logf("user2 받은 메시지: %s", msg.Content)
		}
	case <-time.After(3 * time.Second):
		t.Error("유저2가 메시지를 못 받음 (timeout)")
	}

	stream1.CloseSend()
	stream2.CloseSend()
	conn1.Close()
	conn2.Close()
}
