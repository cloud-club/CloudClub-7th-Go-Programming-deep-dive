package main

import (
    "context"
    "fmt"
    "log"
    "time"
    "bufio"
    "os"
    "strings"

    pb "grpc-chat/gen"
    "google.golang.org/grpc"
)

func main() {
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("접속 실패: %v", err)
    }
    defer conn.Close()

    client := pb.NewChatServiceClient(conn)
    stream, err := client.ChatStream(context.Background())
    if err != nil {
        log.Fatalf("스트림 생성 실패: %v", err)
    }

    go func() {
        for {
            resp, err := stream.Recv()
            if err != nil {
                log.Println("Recv 종료:", err)
                return
            }
            fmt.Printf("🔁 %s\n", resp.Message)
        }
    }()

    me := fmt.Sprintf("cli-%d", time.Now().Unix())
    stream.Send(&pb.ChatMessage{User: me, Message: "Hello!"})

    scanner := bufio.NewScanner(os.Stdin)

    for {

	fmt.Print("입력> ")
    	if !scanner.Scan() {
        	break
    	}	
	text := strings.TrimSpace(scanner.Text())
    	if text == "" {
        	fmt.Println("❌ 빈 입력 무시됨")
        	continue
    	}
	//fmt.Scanln(&text)
	if text == "" {
    		fmt.Println("❌ [클라이언트] 빈 입력 → 전송 안 함")
    		continue
	}
	fmt.Println("✅ [클라이언트] 메시지 전송:", text)
	if text == "" {
        	continue // 빈 문자열은 보내지 않음!
    	}
    	if text == "exit" {
        	fmt.Println("👋 종료합니다.")
        	stream.CloseSend()
        	break
    	}
    	stream.Send(&pb.ChatMessage{User: me, Message: text})

    }
}

