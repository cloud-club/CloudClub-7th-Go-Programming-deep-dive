package simulator

import (
    "context"
    "fmt"
    "sync"
    "time"

    pb "grpc-chat/gen"
    "google.golang.org/grpc"
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
        id := fmt.Sprintf("cli-%d", i)
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

func (r *Runner) Broadcast(text string) {
    var wg sync.WaitGroup
    for _, c := range r.clients {
        wg.Add(1)
        go func(ci *ClientInstance) {
            defer wg.Done()
            ci.Stream.Send(&pb.ChatMessage{User: ci.ID, Message: text})
        }(c)
    }
    wg.Wait()
}

type Result struct {
    Success, Failure int
}

func (r *Runner) Summary() Result {
    return Result{Success: len(r.clients), Failure: 0}
}

