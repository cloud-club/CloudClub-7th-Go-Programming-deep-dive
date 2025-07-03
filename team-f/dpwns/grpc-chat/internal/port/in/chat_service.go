package in

import "grpc-chat/internal/domain"

// ChatService defines application usecases
type ChatService interface {
    Register(user domain.User) error
    Broadcast(msg domain.Message) error
    SendTo(msg domain.Message) error
}


