package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"example.com/grpc-chat-app/internal/config"
	pb "example.com/grpc-chat-app/pkg/gen"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// WebMessage: WebSocket을 통해 주고받는 메시지 구조체
type WebMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WebServer: 웹 게이트웨이 서버
type WebServer struct {
	config   *config.Config
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]*WebClient
	grpcConn *grpc.ClientConn
	client   pb.ChatServiceClient
}

// WebClient: 개별 웹 클라이언트 정보
type WebClient struct {
	conn       *websocket.Conn
	user       *pb.User
	roomID     string
	stream     pb.ChatService_JoinRoomClient
	writeMux   sync.Mutex // WebSocket 쓰기 동기화
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewWebServer: 새로운 웹 서버 생성
func NewWebServer(cfg *config.Config) *WebServer {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 개발용: 모든 오리진 허용
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return &WebServer{
		config:   cfg,
		upgrader: upgrader,
		clients:  make(map[*websocket.Conn]*WebClient),
	}
}

// Start: 웹 서버 시작
func (ws *WebServer) Start() error {
	// gRPC 클라이언트 연결
	if err := ws.connectToGRPCServer(); err != nil {
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	// HTTP 라우터 설정
	router := mux.NewRouter()

	// CORS 미들웨어
	router.Use(ws.corsMiddleware)

	// API 엔드포인트 (구체적인 경로 먼저)
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/rooms", ws.handleListRooms).Methods("GET")
	api.HandleFunc("/rooms", ws.handleCreateRoom).Methods("POST")

	// WebSocket 엔드포인트
	router.HandleFunc("/ws", ws.handleWebSocket)

	// 정적 파일 서빙 (마지막에 catch-all)
	router.PathPrefix("/").Handler(ws.staticFileHandler())

	// HTTP 서버 시작
	server := &http.Server{
		Addr:         ws.config.HTTPPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("Web server starting on %s", ws.config.HTTPPort)
	return server.ListenAndServe()
}

// connectToGRPCServer: gRPC 서버에 연결
func (ws *WebServer) connectToGRPCServer() error {
	grpcAddr := fmt.Sprintf("localhost%s", ws.config.GRPCPort)
	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server at %s: %w", grpcAddr, err)
	}

	ws.grpcConn = conn
	ws.client = pb.NewChatServiceClient(conn)

	log.Printf("Connected to gRPC server at %s", grpcAddr)
	return nil
}

// staticFileHandler: 정적 파일 핸들러
func (ws *WebServer) staticFileHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(ws.config.WebRoot, "index.html"))
			return
		}

		// 정적 파일 서빙
		fs := http.FileServer(http.Dir(ws.config.WebRoot))
		fs.ServeHTTP(w, r)
	})
}

// corsMiddleware: CORS 미들웨어
func (ws *WebServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleWebSocket: WebSocket 연결 처리
func (ws *WebServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("New WebSocket connection from %s", r.RemoteAddr)

	// Context 생성
	ctx, cancel := context.WithCancel(context.Background())
	client := &WebClient{
		conn:       conn,
		ctx:        ctx,
		cancelFunc: cancel,
	}
	ws.clients[conn] = client

	// 메시지 처리 루프
	for {
		var msg WebMessage
		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		if err := ws.handleWebSocketMessage(client, &msg); err != nil {
			log.Printf("WebSocket message handling error: %v", err)
			ws.sendError(conn, fmt.Sprintf("메시지 처리 중 오류가 발생했습니다: %v", err))
		}
	}

	// 연결 정리
	delete(ws.clients, conn)
	if client.cancelFunc != nil {
		client.cancelFunc() // Context 취소로 gRPC 스트림 강제 종료
	}
	if client.stream != nil {
		client.stream.CloseSend()
	}
	log.Printf("WebSocket connection closed for %s", r.RemoteAddr)
}

// handleWebSocketMessage: WebSocket 메시지 처리
func (ws *WebServer) handleWebSocketMessage(client *WebClient, msg *WebMessage) error {
	switch msg.Type {
	case "join_room":
		return ws.handleJoinRoom(client, msg.Payload)
	case "send_message":
		return ws.handleSendMessage(client, msg.Payload)
	case "leave_room":
		return ws.handleLeaveRoom(client)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handleJoinRoom: 채팅방 입장 처리
func (ws *WebServer) handleJoinRoom(client *WebClient, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal join room payload: %w", err)
	}

	var req struct {
		RoomID   string `json:"room_id"`
		UserID   string `json:"user_id"`
		UserName string `json:"user_name"`
	}

	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("failed to unmarshal join room request: %w", err)
	}

	// 이전 채팅방 스트림이 있다면 정리
	if client.stream != nil && client.roomID != "" {
		log.Printf("User %s leaving previous room %s", req.UserID, client.roomID)

		// 이전 context 취소로 gRPC 스트림 종료
		if client.cancelFunc != nil {
			client.cancelFunc()
		}

		client.stream.CloseSend()
		client.stream = nil
	}

	// 새로운 context 생성
	newCtx, newCancel := context.WithCancel(context.Background())
	client.ctx = newCtx
	client.cancelFunc = newCancel

	// 사용자 정보 설정
	client.user = &pb.User{
		Id:   req.UserID,
		Name: req.UserName,
	}
	client.roomID = req.RoomID

	// gRPC를 통해 채팅방 입장 (새로운 context 사용)
	stream, err := ws.client.JoinRoom(client.ctx, &pb.JoinRoomRequest{
		RoomId: req.RoomID,
		User:   client.user,
	})
	if err != nil {
		return fmt.Errorf("failed to join room via gRPC: %w", err)
	}

	client.stream = stream

	// 입장 성공 응답
	ws.sendMessage(client.conn, "room_joined", map[string]interface{}{
		"room_id":   req.RoomID,
		"user_id":   req.UserID,
		"user_name": req.UserName,
	})

	// 메시지 수신 고루틴 시작
	go ws.receiveMessages(client)

	return nil
}

// handleSendMessage: 메시지 전송 처리
func (ws *WebServer) handleSendMessage(client *WebClient, payload interface{}) error {
	if client.user == nil || client.roomID == "" {
		return fmt.Errorf("user not joined to any room")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal send message payload: %w", err)
	}

	var req struct {
		ID      string `json:"id"`
		Content string `json:"content"`
	}

	if err := json.Unmarshal(data, &req); err != nil {
		return fmt.Errorf("failed to unmarshal send message request: %w", err)
	}

	// gRPC를 통해 메시지 전송
	_, err = ws.client.SendMessage(context.Background(), &pb.Message{
		Id:        req.ID, // 클라이언트에서 받은 ID 사용
		Content:   req.Content,
		Timestamp: timestamppb.Now(),
		UserId:    client.user.Id,
		UserName:  client.user.Name,
		RoomId:    client.roomID,
	})

	if err != nil {
		return fmt.Errorf("failed to send message via gRPC: %w", err)
	}

	return nil
}

// handleLeaveRoom: 채팅방 퇴장 처리
func (ws *WebServer) handleLeaveRoom(client *WebClient) error {
	if client.stream != nil && client.roomID != "" && client.user != nil {
		log.Printf("User %s leaving room %s", client.user.Id, client.roomID)

		// Context 취소를 통해 gRPC 스트림 강제 종료 (서버에서 퇴장 처리됨)
		if client.cancelFunc != nil {
			client.cancelFunc()
		}

		client.stream.CloseSend()
		client.stream = nil
	}

	ws.sendMessage(client.conn, "room_left", map[string]interface{}{
		"room_id": client.roomID,
	})

	client.roomID = ""
	client.user = nil

	return nil
}

// receiveMessages: gRPC 스트림에서 메시지 수신
func (ws *WebServer) receiveMessages(client *WebClient) {
	for {
		msg, err := client.stream.Recv()
		if err != nil {
			log.Printf("Error receiving message from gRPC stream: %v", err)
			break
		}

		// 웹 클라이언트로 메시지 전송
		ws.sendMessage(client.conn, "message", map[string]interface{}{
			"id":        msg.Id,
			"content":   msg.Content,
			"timestamp": msg.Timestamp.AsTime().Format(time.RFC3339),
			"user_id":   msg.UserId,
			"user_name": msg.UserName,
			"room_id":   msg.RoomId,
		})
	}
}

// handleListRooms: 채팅방 목록 API
func (ws *WebServer) handleListRooms(w http.ResponseWriter, r *http.Request) {
	resp, err := ws.client.ListRooms(context.Background(), &pb.ListRoomsRequest{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list rooms: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Rooms)
}

// handleCreateRoom: 채팅방 생성 API
func (ws *WebServer) handleCreateRoom(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		CreatorID   string `json:"creator_id"`
		CreatorName string `json:"creator_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := ws.client.CreateRoom(context.Background(), &pb.CreateRoomRequest{
		Name:        req.Name,
		Description: req.Description,
		Creator: &pb.User{
			Id:   req.CreatorID,
			Name: req.CreatorName,
		},
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create room: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// sendMessage: WebSocket을 통해 메시지 전송
func (ws *WebServer) sendMessage(conn *websocket.Conn, msgType string, payload interface{}) {
	// WebClient 찾기
	var client *WebClient
	for c, cl := range ws.clients {
		if c == conn {
			client = cl
			break
		}
	}

	if client == nil {
		log.Printf("Client not found for connection")
		return
	}

	msg := WebMessage{
		Type:    msgType,
		Payload: payload,
	}

	// 동시 쓰기 방지를 위한 뮤텍스 사용
	client.writeMux.Lock()
	defer client.writeMux.Unlock()

	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Error sending WebSocket message: %v", err)
	}
}

// sendError: WebSocket을 통해 에러 메시지 전송
func (ws *WebServer) sendError(conn *websocket.Conn, message string) {
	ws.sendMessage(conn, "error", map[string]string{
		"message": message,
	})
}
