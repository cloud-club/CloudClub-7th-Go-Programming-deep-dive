package grpc

import (
	"fmt"
	"grpc-chat/internal/domain"
	"grpc-chat/internal/port/out"
	"time"

	pb "grpc-chat/gen"
)

type CommandHandler struct {
	sessionRepo out.SessionRepository
}

func NewCommandHandler(repo out.SessionRepository) *CommandHandler {
	return &CommandHandler{sessionRepo: repo}
}

func (h *CommandHandler) Handle(cmd Command, user domain.User, stream pb.ChatService_ChatStreamServer) error {
	switch cmd.Type {
	case "list":
		others := h.sessionRepo.ListUsersExcept(user.ID)
		msg := "📋 접속자 목록:\n"
		for _, id := range others {
			msg += fmt.Sprintf(" - %s\n", id)
		}
		return stream.Send(&pb.ChatMessage{
			User:      "Server",
			Message:   msg,
			Timestamp: time.Now().Unix(),
		})
	case "connect":
		// 연결 대상 저장 등은 handler가 따로 관리 (이 함수는 알림만 전송)
		return stream.Send(&pb.ChatMessage{
			User:      "Server",
			Message:   fmt.Sprintf("%s님과 대화를 시작합니다.", cmd.Argument),
			Timestamp: time.Now().Unix(),
		})
	default:
		return stream.Send(&pb.ChatMessage{
			User:      "Server",
			Message:   "⚠ 알 수 없는 명령어입니다.",
			Timestamp: time.Now().Unix(),
		})
	}
}

