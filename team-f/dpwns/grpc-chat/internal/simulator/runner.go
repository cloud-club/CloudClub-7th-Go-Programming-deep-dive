package simulator

import (
    "context"
    "sync"
    "math/rand"
    "time"
    "log"
    "fmt"

    pb "grpc-chat/gen"
    "google.golang.org/grpc"
    "github.com/google/uuid"

)


type ClientInstance struct {
    ID     string
    Stream pb.ChatService_ChatStreamClient
    Conn   *grpc.ClientConn
}

type Runner struct {
    addr    string
    clients []*ClientInstance
    mu      sync.Mutex
}

func NewRunner(addr string) *Runner {
    return &Runner{addr: addr}
}

func (r *Runner) Spawn(n int) {
    for i := 0; i < n; i++ {
        id := uuid.New().String()
        conn, err := grpc.Dial(r.addr, grpc.WithInsecure())
        if err != nil {
            continue
        }
        client := pb.NewChatServiceClient(conn)
        stream, err := client.ChatStream(context.Background())
        if err != nil {
            conn.Close()
            continue
        }

        stream.Send(&pb.ChatMessage{User: id, Message: "Hi"})
        r.mu.Lock()
        r.clients = append(r.clients, &ClientInstance{id, stream, conn})
        r.mu.Unlock()
    }
}

func (r *Runner) BroadcastRandomInterval(text string, count int) {
    rand.Seed(time.Now().UnixNano())
    var wg sync.WaitGroup

    for _, c := range r.clients {
        wg.Add(1)
        go func(ci *ClientInstance) {
            defer wg.Done()

            // 각 클라이언트의 메시지 간 전송 간격을 100~1000ms 사이로 랜덤 설정
            delay := time.Duration(rand.Intn(900)+100) * time.Millisecond

            for i := 0; i < count; i++ {
                msg := fmt.Sprintf("[%s] 메시지 #%d", ci.ID, i+1)
		log.Printf("[%s] 메시지: %s", ci.ID, text)
                err := ci.Stream.Send(&pb.ChatMessage{
                    User:    ci.ID,
                    Message: msg,
                })
                if err != nil {
                    log.Printf("❌ [%s] 전송 실패: %v", ci.ID, err)
                    return
                }
                time.Sleep(delay)
            }
        }(c)
    }

    wg.Wait()
}

func (r *Runner) CloseAll() {
    r.mu.Lock()
    defer r.mu.Unlock()
    for _, c := range r.clients {
        c.Stream.CloseSend()
        c.Conn.Close()
    }
    r.clients = nil
}


type Result struct {
    Success, Failure int
}

func (r *Runner) Summary() Result {
    return Result{Success: len(r.clients), Failure: 0}
}

