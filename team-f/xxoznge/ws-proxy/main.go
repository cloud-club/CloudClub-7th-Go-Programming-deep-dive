package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"example.com/golang-grpc-chat/grpcapi"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var (
	clients      = make(map[string]*websocket.Conn) // 닉네임 → 연결
	clientsMu    sync.Mutex
	seenMessages = make(map[string]bool) // ✅ 중복 메시지 추적
	seenMu       sync.Mutex              // ✅ 동시 접근 방지
)

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	fmt.Println("WebSocket Proxy running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func broadcastOnlineCount() {
	clientsMu.Lock()
	count := len(clients)
	clientsMu.Unlock()

	msg := map[string]interface{}{
		"type":  "online-count",
		"count": count,
	}
	data, _ := json.Marshal(msg)

	clientsMu.Lock()
	defer clientsMu.Unlock()
	for nickname, conn := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Println("접속자 수 전송 실패:", nickname, err)
			conn.Close()
			delete(clients, nickname)
		}
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket 업그레이드 실패:", err)
		return
	}

	var initMsg grpcapi.ChatMessage
	if err := conn.ReadJSON(&initMsg); err != nil {
		log.Println("닉네임 수신 실패:", err)
		conn.Close()
		return
	}
	nickname := initMsg.User

	grpcConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Println("gRPC 연결 실패:", err)
		conn.Close()
		return
	}
	client := grpcapi.NewChatServiceClient(grpcConn)
	stream, err := client.Chat(context.Background())
	if err != nil {
		log.Println("gRPC 스트림 생성 실패:", err)
		conn.Close()
		grpcConn.Close()
		return
	}

	clientsMu.Lock()
	if oldConn, exists := clients[nickname]; exists {
		oldConn.Close()
	}
	clients[nickname] = conn
	clientsMu.Unlock()
	broadcastOnlineCount()

	if err := stream.Send(&initMsg); err != nil {
		log.Println("입장 메시지 gRPC 전송 실패:", err)
	}

	defer func() {
		clientsMu.Lock()
		if c, ok := clients[nickname]; ok && c == conn {
			delete(clients, nickname)
		}
		clientsMu.Unlock()
		broadcastOnlineCount()
		conn.Close()
		grpcConn.Close()
	}()

	// ✅ gRPC 수신 → WebSocket 브로드캐스트 (중복 방지 추가)
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				log.Println("gRPC 수신 에러:", err)
				return
			}

			// 중복 메시지 필터링
			key := fmt.Sprintf("%s|%s|%d", msg.User, msg.Content, msg.Timestamp)
			seenMu.Lock()
			if seenMessages[key] {
				seenMu.Unlock()
				continue
			}
			seenMessages[key] = true
			seenMu.Unlock()

			log.Println("gRPC 수신:", msg)

			clientsMu.Lock()
			for targetNick, wsConn := range clients {
				if err := wsConn.WriteJSON(msg); err != nil {
					log.Println("메시지 전송 실패:", targetNick, err)
					wsConn.Close()
					delete(clients, targetNick)
				}
			}
			clientsMu.Unlock()
		}
	}()

	// WebSocket 수신 → gRPC 전송
	for {
		var msg grpcapi.ChatMessage
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("WebSocket 수신 에러:", err)
			break
		}
		log.Println("WebSocket → gRPC 전송:", msg)
		if err := stream.Send(&msg); err != nil {
			log.Println("gRPC 전송 실패:", err)
			break
		}
	}
}
