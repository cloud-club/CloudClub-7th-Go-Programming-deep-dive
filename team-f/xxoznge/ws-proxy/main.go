// ws_proxy/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"example.com/golang-grpc-chat/grpcapi"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 모든 origin 허용 (개발용)
	},
}
var clients = make(map[*websocket.Conn]string)

func main() {
	http.HandleFunc("/ws", handleWebSocket)

	fmt.Println("WebSocket Proxy running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket 업그레이드 실패:", err)
		return
	}
	defer conn.Close()

	// gRPC 서버 연결
	grpcConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Println("gRPC 연결 실패:", err)
		return
	}
	defer grpcConn.Close()
	client := grpcapi.NewChatServiceClient(grpcConn)
	stream, err := client.Chat(context.Background())
	if err != nil {
		log.Println("gRPC 스트림 생성 실패:", err)
		return
	}

	// 수신 루틴
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				log.Println("gRPC 수신 에러:", err)
				break
			}
			log.Println("📩 수신한 gRPC 메시지:", msg)
			conn.WriteJSON(msg)
		}
	}()

	for {
		var msg grpcapi.ChatMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("❌ WebSocket 수신 에러:", err)
			break
		}
		log.Println("📤 WebSocket → gRPC 전송:", msg)
		if err := stream.Send(&msg); err != nil {
			log.Println("❌ gRPC 전송 실패:", err)
			break
		}
	}
}
