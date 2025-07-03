package domain

// User represents a connected client
type User struct {
    ID   string
    Name string
}

// Message represents a chat message
type Message struct {
    From    string
    TargetID	string  // 🎯 수신 대상 추가
    Content string
}

