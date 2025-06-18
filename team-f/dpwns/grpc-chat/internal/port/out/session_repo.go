package out

import "grpc-chat/internal/domain"

// SessionRepository defines methods to manage sessions
type SessionRepository interface {
    Add(user domain.User, sendFunc func(msg domain.Message) error)
    Remove(userID string)
    List() map[string]func(msg domain.Message) error
    SendTo(userID string, msg domain.Message) error
    ListUsersExcept(excludeID string) []string
}

