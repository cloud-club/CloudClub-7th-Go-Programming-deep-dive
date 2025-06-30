package main

import (
    "fmt"
    "bufio"
    "os"
    "strings"

    "grpc-chat/internal/simulator"

)

type Command struct {
    Name        string
    Description string
    Usage       string
}

var commands = []Command{
    {"spawn",     "지정된 수만큼 클라이언트를 생성",     "spawn <숫자>"},
    {"broadcast", "모든 클라이언트가 메시지 전송",       "broadcast <메시지> <숫자>"},
    {"summary",   "현재 연결된 클라이언트 요약 출력",     "summary"},
    {"closeAll", "현재 연결된 클라이언트 종료",		"closeAll"},
    {"exit",      "시뮬레이터 종료",                    "exit"},
}

func printHelp() {
    fmt.Println("📚 사용 가능한 명령:")
    for _, cmd := range commands {
        fmt.Printf("  %-10s : %s\n", cmd.Usage, cmd.Description)
    }
}

func main() {

    sim := simulator.NewRunner("localhost:50051")

    fmt.Println("📡 gRPC 부하 시뮬레이터 CLI 시작")
    printHelp()

    scanner := bufio.NewScanner(os.Stdin)
    for {
        fmt.Print("sim> ")
        if !scanner.Scan() {
            break
        }

        line := strings.TrimSpace(scanner.Text())
        if line == "" {
            continue
        }

        parts := strings.SplitN(line, " ", 3)
        cmd := parts[0]

        switch cmd {
        case "spawn":
            if len(parts) < 2 {
                fmt.Println("❗ 사용법: spawn <숫자>")
                continue
            }
            var n int
            fmt.Sscanf(parts[1], "%d", &n)
            sim.Spawn(n)

        case "broadcast":
            if len(parts) < 3 {
                fmt.Println("❗ 사용법: broadcast <메시지> <숫자>")
                continue
            }
	    var n int
            fmt.Sscanf(parts[2], "%d", &n)
	    sim.BroadcastRandomInterval(parts[1], n)

        case "summary":
            res := sim.Summary()
            fmt.Printf("✅ 연결 성공: %d, 실패: %d\n", res.Success, res.Failure)

        case "exit":
	    sim.CloseAll()
            fmt.Println("👋 종료합니다.")
            return
	
        case "closeAll":
	    sim.CloseAll()
	    fmt.Println("🛑 모든 연결 종료")
        default:
            fmt.Println("❓ 알 수 없는 명령:", cmd)
	    printHelp()

        }
    }

}

