package grpc

import (
    "io"
    "log"
    "strings"
    "time"

    pb "grpc-chat/gen"
    "grpc-chat/internal/domain"
    "grpc-chat/internal/port/in"
    "grpc-chat/internal/port/out"
)
// 실제 gRPC 스트림 처리
type ChatHandler struct {
    pb.UnimplementedChatServiceServer
    service    in.ChatService
    sessionRepo out.SessionRepository
    cmdHandler  *CommandHandler
}

func NewChatHandler(s in.ChatService, repo out.SessionRepository) *ChatHandler {
    return &ChatHandler{
	service: s,
	sessionRepo: repo,
    	cmdHandler:  NewCommandHandler(repo), // 명령 처리 핸들러 구성
    }
}

func (h *ChatHandler) ChatStream(stream pb.ChatService_ChatStreamServer) error {
    log.Println("📡 ChatStream 연결 수신")
    
    firstMsg, err := stream.Recv()
    if err != nil {
        return err
    }

    // 클라이언트 접속 시 사용자 등록
    user := domain.User{ID: firstMsg.User, Name: firstMsg.User}
    h.service.Register(user)

    // 사용자가 받을 콜백함수 => 누군가 메시지를 보내면 send 함수가 실행되어 스트림에 전송
    h.sessionRepo.Add(user, func(msg domain.Message) error {
        return stream.Send(&pb.ChatMessage{
            User:    msg.From,
            Message: msg.Content,
            Timestamp: time.Now().Unix(),
        })
    })

    // 첫 메시지도 전송 처리
    h.service.SendTo(domain.Message{
	    From: user.ID,
	    TargetID: firstMsg.TargetId,
	    Content: firstMsg.Message,
    })

    // 현재 대화 대상
    var targetID string

    for {
        msgPb, err := stream.Recv()
        if err == io.EOF {
	    log.Printf("✅ %s 연결 종료", user.ID)
            h.sessionRepo.Remove(user.ID)
            return nil
        }
        if err != nil {
            h.sessionRepo.Remove(user.ID)
            return err
        }
	input := strings.TrimSpace(msgPb.Message)
	if input == "" {
		continue
	}
	cmd, ok := ParseCommand(input)
	if ok {
		if err := h.cmdHandler.Handle(cmd, user, stream); err != nil {
			log.Printf("명령 처리 오류: %v", err)
		}
		// connect이면 현재 대상 갱신
		if cmd.Type == "connect" {
			targetID = cmd.Argument
		}
		continue
	}

	if targetID == "" {
		stream.Send(&pb.ChatMessage{
		User:      "Server",
		Message:   "⚠ 먼저 connect <id> 명령으로 대화 대상을 선택하세요",
		Timestamp: time.Now().Unix(),
	})
		continue
	}

	// 메시지 전송
	msg := domain.Message{
		From:    user.ID,
		TargetID:      targetID,
		Content: input,
	}
	if err := h.service.SendTo(msg); err != nil {
		log.Printf("❌ 메시지 전송 실패: %v", err)
	}
    }
}

