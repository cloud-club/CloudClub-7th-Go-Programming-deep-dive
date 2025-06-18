package usecase

import (
    "log"

    "grpc-chat/internal/domain"
    "grpc-chat/internal/port/out"
)

// ChatUsecase implements ChatService
type ChatUsecase struct {
    repo out.SessionRepository
}

func NewChatUsecase(repo out.SessionRepository) *ChatUsecase {
    return &ChatUsecase{repo: repo}
}

// 사용자를 세션 저장소에 등록
func (u *ChatUsecase) Register(user domain.User) error {
    u.repo.Add(user, func(msg domain.Message) error {
        return nil // placeholder; actual send is in adapter
    })
    return nil
}

// 세션 저장소에 등록된 사용자에게 메시지 전송
func (u *ChatUsecase) Broadcast(msg domain.Message) error {
    // 🔍 서버 터미널에 수신 메시지 로그 출력
    log.Printf("💬 [%s]: %s", msg.From, msg.Content)
    sessions := u.repo.List()
    for userID, send := range sessions {
        if err := send(msg); err != nil {
            // on error, remove session
            u.repo.Remove(userID)
        }
    }
    return nil
}

func (u *ChatUsecase) SendTo(msg domain.Message) error {
    return u.repo.SendTo(msg.TargetID, msg)
}
