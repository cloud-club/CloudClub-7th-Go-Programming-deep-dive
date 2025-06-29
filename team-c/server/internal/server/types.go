package server

import (
	"context"
	"sync"
	"time"

	pb "example.com/grpc-chat-app/pkg/gen"
)

// Connection: 개별 클라이언트 연결을 나타내는 구조체
type Connection struct {
	Stream     pb.ChatService_JoinRoomServer
	User       *pb.User
	RoomID     string
	Active     bool
	Error      chan error
	Context    context.Context
	CancelFunc context.CancelFunc
	CreatedAt  time.Time
}

// Room: 채팅방 정보
type Room struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	Connections []*Connection
	Mutex       sync.RWMutex
}
