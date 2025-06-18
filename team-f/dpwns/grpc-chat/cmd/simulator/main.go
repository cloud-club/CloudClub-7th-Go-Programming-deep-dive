package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    pb "grpc-chat/gen"
    "grpc-chat/internal/simulator"

    "google.golang.org/grpc"
)

func main() {
    sim := simulator.NewRunner("localhost:50051")
    sim.Spawn(100)

    fmt.Println("100개 연결 생성됨")
    time.Sleep(time.Second)

    sim.Broadcast("부하 테스트 메시지")

    counts := sim.Summary()
    fmt.Printf("성공: %d, 실패: %d\n", counts.Success, counts.Failure)
}

