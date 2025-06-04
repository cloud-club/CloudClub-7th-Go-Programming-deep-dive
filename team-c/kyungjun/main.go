package main

import (
	"log"
	"net"
	"sync"

	pb "kyungjun" // 여기! 로컬 모듈 경로 기준

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type chatServer struct {
	pb.UnimplementedChatServiceServer
	clients map[string]pb.ChatService_ChatServer
	mu      sync.Mutex
}

func (s *chatServer) Chat(stream pb.ChatService_ChatServer) error {
	msg, err := stream.Recv()
	if err != nil {
		return err
	}
	user := msg.User

	s.mu.Lock()
	s.clients[user] = stream
	s.mu.Unlock()

	log.Printf("%s joined the chat", user)

	for {
		msg, err := stream.Recv()
		if err != nil {
			s.mu.Lock()
			delete(s.clients, user)
			s.mu.Unlock()
			log.Printf("%s left the chat", user)
			return err
		}

		s.mu.Lock()
		for u, client := range s.clients {
			if u != user {
				_ = client.Send(&pb.ChatMessage{User: user, Message: msg.Message})
			}
		}
		s.mu.Unlock()
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterChatServiceServer(s, &chatServer{
		clients: make(map[string]pb.ChatService_ChatServer),
	})
	reflection.Register(s)

	log.Println("gRPC Chat server is running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
