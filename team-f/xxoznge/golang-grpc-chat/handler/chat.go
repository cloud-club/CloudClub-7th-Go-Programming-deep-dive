package handler

import (
	"io"
	"sync"
	"time"

	"golang-grpc-chat/grpcapi"
)

type ChatHandler struct {
	grpcapi.UnimplementedChatServiceServer
	mu      sync.Mutex
	streams map[string]grpcapi.ChatService_ChatServer
}

func NewChatHandler() *ChatHandler {
	return &ChatHandler{
		streams: make(map[string]grpcapi.ChatService_ChatServer),
	}
}

func (h *ChatHandler) Chat(stream grpcapi.ChatService_ChatServer) error {
	// stream 연결 즉시 temp ID로 등록
	tempID := time.Now().Format("20060102150405.000000")
	h.mu.Lock()
	h.streams[tempID] = stream
	h.mu.Unlock()

	var user string

	for {
		msg, err := stream.Recv()
		if err == io.EOF || err != nil {
			h.mu.Lock()
			delete(h.streams, user)
			h.mu.Unlock()
			return nil
		}

		user = msg.User

		// 첫 메시지 수신 후, 정식 userID로 교체
		h.mu.Lock()
		delete(h.streams, tempID)
		h.streams[user] = stream

		// 브로드캐스트: 나를 제외한 모든 유저에게 전송
		for u, s := range h.streams {
			if u != user {
				s.Send(&grpcapi.ChatMessage{
					User:      msg.User,
					Content:   msg.Content,
					Timestamp: time.Now().Unix(),
				})
			}
		}
		h.mu.Unlock()
	}
}
