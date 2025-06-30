package memory

import (
    "sync"
    "fmt"
    "grpc-chat/internal/domain"
    "grpc-chat/internal/port/out"
)

type sessionRepo struct {
    mu       sync.Mutex
    sessions map[string]func(msg domain.Message) error
}

func NewSessionRepo() out.SessionRepository {
    return &sessionRepo{
        sessions: make(map[string]func(msg domain.Message) error),
    }
}

func (r *sessionRepo) Add(user domain.User, sendFunc func(msg domain.Message) error) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.sessions[user.ID] = sendFunc
}

func (r *sessionRepo) Remove(userID string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    delete(r.sessions, userID)
}

func (r *sessionRepo) List() map[string]func(msg domain.Message) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    // Return a copy to avoid mutation
    cp := make(map[string]func(msg domain.Message) error, len(r.sessions))
    for k, v := range r.sessions {
        cp[k] = v
    }
    return cp
}

func (r *sessionRepo) SendTo(userID string, msg domain.Message) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    if send, ok := r.sessions[userID]; ok {
        return send(msg)
    }
    return fmt.Errorf("사용자 %s를 찾을 수 없습니다", userID)
}

func (r *sessionRepo) ListUsersExcept(myID string) []string {
    r.mu.Lock()
    defer r.mu.Unlock()
    var result []string
    for id := range r.sessions {
        if id != myID {
            result = append(result, id)
        }
    }
    return result
}
func (r *sessionRepo) Count() int {
    r.mu.Lock()
    defer r.mu.Unlock()
    return len(r.sessions)
}
