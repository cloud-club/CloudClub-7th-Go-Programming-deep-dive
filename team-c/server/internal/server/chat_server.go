package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"example.com/grpc-chat-app/internal/config"
	"example.com/grpc-chat-app/internal/consumer"
	pb "example.com/grpc-chat-app/pkg/gen"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ChatServer: 채팅 서버 구조체
type ChatServer struct {
	pb.UnimplementedChatServiceServer
	config     *config.Config
	rooms      map[string]*Room
	roomsMutex sync.RWMutex
	producer   sarama.SyncProducer
	processor  *consumer.MessageProcessor
	grpcServer *grpc.Server
}

// NewChatServer: 새로운 채팅 서버 생성
func NewChatServer(cfg *config.Config) (*ChatServer, error) {
	// Kafka Producer 설정
	producerConfig := sarama.NewConfig()
	producerConfig.Producer.RequiredAcks = sarama.WaitForAll
	producerConfig.Producer.Retry.Max = 10
	producerConfig.Producer.Return.Successes = true
	producerConfig.Metadata.Retry.Max = 10
	producerConfig.Metadata.Retry.Backoff = 2 * time.Second

	var producer sarama.SyncProducer
	var err error

	// Kafka 연결 재시도 로직
	for i := 0; i < 5; i++ {
		producer, err = sarama.NewSyncProducer(cfg.Kafka.Brokers, producerConfig)
		if err == nil {
			break
		}
		log.Printf("Kafka connection failed (attempt %d/5): %v", i+1, err)
		if i < 4 {
			log.Printf("Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer after 5 attempts: %w", err)
	}

	// MessageProcessor 생성 (재시도 로직)
	var processor *consumer.MessageProcessor
	for i := 0; i < 3; i++ {
		processor, err = consumer.NewMessageProcessor(cfg.Kafka, cfg.Kafka.ConsumerGroup)
		if err == nil {
			break
		}
		log.Printf("MessageProcessor creation failed (attempt %d/3): %v", i+1, err)
		if i < 2 {
			log.Printf("Retrying in 3 seconds...")
			time.Sleep(3 * time.Second)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create message processor after 3 attempts: %w", err)
	}

	server := &ChatServer{
		config:     cfg,
		rooms:      make(map[string]*Room),
		producer:   producer,
		processor:  processor,
		grpcServer: grpc.NewServer(),
	}

	// 기본 채팅방 생성
	defaultRoom := &Room{
		ID:          "general",
		Name:        "공지사항 채널",
		Description: "자유롭게 대화하는 공간입니다",
		CreatedAt:   time.Now(),
		Connections: []*Connection{},
	}
	server.rooms["general"] = defaultRoom

	// MessageProcessor에 서버 참조 설정
	processor.SetServer(server)

	// gRPC 서버 설정
	pb.RegisterChatServiceServer(server.grpcServer, server)
	reflection.Register(server.grpcServer)

	// 죽은 연결 정리 고루틴 시작
	go server.startConnectionCleanup()

	return server, nil
}

// Start: 서버 시작
func (s *ChatServer) Start() error {
	// 메시지 프로세서 시작
	s.startMessageProcessor()

	// gRPC 리스너 생성
	listener, err := net.Listen("tcp", s.config.GRPCPort)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	log.Printf("gRPC server listening on %s", s.config.GRPCPort)

	// gRPC 서버 시작 (블로킹)
	return s.grpcServer.Serve(listener)
}

// Close: 서버 리소스 정리
func (s *ChatServer) Close() error {
	if s.processor != nil {
		s.processor.Stop()
	}
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
	return s.producer.Close()
}

// startMessageProcessor: 메시지 프로세서 시작
func (s *ChatServer) startMessageProcessor() {
	go func() {
		if err := s.processor.Start(); err != nil {
			log.Printf("Message processor error: %v", err)
		}
	}()
}

// ListRooms: 채팅방 목록 조회
func (s *ChatServer) ListRooms(ctx context.Context, req *pb.ListRoomsRequest) (*pb.ListRoomsResponse, error) {
	s.roomsMutex.RLock()
	defer s.roomsMutex.RUnlock()

	rooms := make([]*pb.Room, 0, len(s.rooms))
	for _, room := range s.rooms {
		room.Mutex.RLock()
		userCount := len(room.Connections)
		room.Mutex.RUnlock()

		rooms = append(rooms, &pb.Room{
			Id:          room.ID,
			Name:        room.Name,
			Description: room.Description,
			CreatedAt:   timestamppb.New(room.CreatedAt),
			UserCount:   int32(userCount),
		})
	}

	return &pb.ListRoomsResponse{Rooms: rooms}, nil
}

// CreateRoom: 새로운 채팅방 생성
func (s *ChatServer) CreateRoom(ctx context.Context, req *pb.CreateRoomRequest) (*pb.Response, error) {
	roomID := uuid.New().String()

	newRoom := &Room{
		ID:          roomID,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
		Connections: []*Connection{},
	}

	s.roomsMutex.Lock()
	s.rooms[roomID] = newRoom
	s.roomsMutex.Unlock()

	log.Printf("Room created: %s (%s) by %s", req.Name, roomID, req.Creator.Id)

	return &pb.Response{
		Success: true,
		Message: "채팅방이 생성되었습니다",
		Room: &pb.Room{
			Id:          roomID,
			Name:        req.Name,
			Description: req.Description,
			CreatedAt:   timestamppb.New(newRoom.CreatedAt),
			UserCount:   0,
		},
	}, nil
}

// JoinRoom: 채팅방 입장
func (s *ChatServer) JoinRoom(req *pb.JoinRoomRequest, stream pb.ChatService_JoinRoomServer) error {
	// 채팅방 존재 확인
	s.roomsMutex.RLock()
	room, exists := s.rooms[req.RoomId]
	s.roomsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("room not found: %s", req.RoomId)
	}

	// 연결 객체 생성 (stream의 context 활용)
	ctx := stream.Context()
	conn := &Connection{
		Stream:    stream,
		User:      req.User,
		RoomID:    req.RoomId,
		Active:    true,
		Error:     make(chan error),
		Context:   ctx,
		CreatedAt: time.Now(),
	}

	// 채팅방에 연결 추가
	room.Mutex.Lock()
	room.Connections = append(room.Connections, conn)
	room.Mutex.Unlock()

	log.Printf("User %s joined room %s", req.User.Id, req.RoomId)

	// 사용자 연결 정보를 Kafka에 발행
	if err := s.publishUserConnection(req.User.Id, req.RoomId, true); err != nil {
		log.Printf("Failed to publish user connection: %v", err)
	}

	// 입장 메시지 전송
	joinMsg := &pb.Message{
		Id:        uuid.New().String(),
		Content:   fmt.Sprintf("%s님이 입장하셨습니다", req.User.Name),
		Timestamp: timestamppb.Now(),
		UserId:    "system",
		UserName:  "시스템",
		RoomId:    req.RoomId,
	}
	s.broadcastToRoom(req.RoomId, joinMsg)

	// 에러 채널 또는 context 취소 대기
	var err error
	select {
	case err = <-conn.Error:
		// 에러 발생으로 인한 종료
	case <-ctx.Done():
		// Context 취소로 인한 종료 (클라이언트 연결 해제)
		err = ctx.Err()
	}

	// 퇴장 처리
	s.removeConnection(req.RoomId, req.User.Id)

	// 퇴장 메시지 전송
	leaveMsg := &pb.Message{
		Id:        uuid.New().String(),
		Content:   fmt.Sprintf("%s님이 퇴장하셨습니다", req.User.Name),
		Timestamp: timestamppb.Now(),
		UserId:    "system",
		UserName:  "시스템",
		RoomId:    req.RoomId,
	}
	s.broadcastToRoom(req.RoomId, leaveMsg)

	// 사용자 연결 해제 정보를 Kafka에 발행
	if disconnectErr := s.publishUserConnection(req.User.Id, req.RoomId, false); disconnectErr != nil {
		log.Printf("Failed to publish user disconnection: %v", disconnectErr)
	}

	log.Printf("User %s left room %s", req.User.Id, req.RoomId)
	return err
}

// SendMessage: 메시지 전송
func (s *ChatServer) SendMessage(ctx context.Context, req *pb.Message) (*pb.Response, error) {
	// 메시지 ID가 없는 경우에만 생성
	if req.Id == "" {
		req.Id = uuid.New().String()
	}

	// 현재 시간 설정
	if req.Timestamp == nil {
		req.Timestamp = timestamppb.Now()
	}

	// Kafka에 메시지 발행
	chatMsg := consumer.ChatMessage{
		ID:        req.Id,
		Content:   req.Content,
		Timestamp: req.Timestamp.AsTime(),
		UserID:    req.UserId,
		UserName:  req.UserName,
		RoomID:    req.RoomId,
	}

	data, err := json.Marshal(chatMsg)
	if err != nil {
		return &pb.Response{Success: false, Message: "메시지 처리 중 오류가 발생했습니다"}, err
	}

	kafkaMsg := &sarama.ProducerMessage{
		Topic: s.config.Kafka.Topic,
		Key:   sarama.StringEncoder(req.RoomId), // 채팅방 ID로 파티셔닝
		Value: sarama.StringEncoder(data),
	}

	partition, offset, err := s.producer.SendMessage(kafkaMsg)
	if err != nil {
		log.Printf("Failed to send message to Kafka: %v", err)
		return &pb.Response{Success: false, Message: "메시지 전송에 실패했습니다"}, err
	}

	log.Printf("Message sent to Kafka - Topic: %s, Partition: %d, Offset: %d",
		s.config.Kafka.Topic, partition, offset)

	return &pb.Response{Success: true, Message: "메시지가 전송되었습니다"}, nil
}

// broadcastToRoom: 채팅방의 모든 연결에 메시지 브로드캐스트
func (s *ChatServer) broadcastToRoom(roomID string, msg *pb.Message) {
	s.roomsMutex.RLock()
	room, exists := s.rooms[roomID]
	s.roomsMutex.RUnlock()

	if !exists {
		return
	}

	room.Mutex.RLock()
	connections := make([]*Connection, len(room.Connections))
	copy(connections, room.Connections)
	room.Mutex.RUnlock()

	var wg sync.WaitGroup
	for _, conn := range connections {
		if conn.Active {
			wg.Add(1)
			go func(c *Connection) {
				defer wg.Done()
				if err := c.Stream.Send(msg); err != nil {
					log.Printf("Error sending message to %s: %v", c.User.Id, err)
					c.Active = false
					select {
					case c.Error <- err:
					default:
					}
				}
			}(conn)
		}
	}
	wg.Wait()
}

// removeConnection: 채팅방에서 연결 제거
func (s *ChatServer) removeConnection(roomID, userID string) {
	s.roomsMutex.RLock()
	room, exists := s.rooms[roomID]
	s.roomsMutex.RUnlock()

	if !exists {
		return
	}

	room.Mutex.Lock()
	defer room.Mutex.Unlock()

	for i, conn := range room.Connections {
		if conn.User.Id == userID {
			room.Connections = append(room.Connections[:i], room.Connections[i+1:]...)
			break
		}
	}
}

// publishUserConnection: 사용자 연결/해제 정보를 Kafka에 발행
func (s *ChatServer) publishUserConnection(userID, roomID string, connected bool) error {
	userConn := consumer.UserConnection{
		UserID:    userID,
		RoomID:    roomID,
		ServerID:  s.config.ServerID,
		Connected: connected,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(userConn)
	if err != nil {
		return fmt.Errorf("failed to marshal user connection: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: "user-connections",
		Key:   sarama.StringEncoder(userID),
		Value: sarama.StringEncoder(data),
	}

	_, _, err = s.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send user connection message: %w", err)
	}

	log.Printf("Published user connection event: %s in room %s (connected: %t)", userID, roomID, connected)
	return nil
}

// ProcessIncomingMessage: Kafka에서 받은 메시지를 로컬 클라이언트들에게 전달
func (s *ChatServer) ProcessIncomingMessage(chatMsg *consumer.ChatMessage) {
	pbMsg := &pb.Message{
		Id:        chatMsg.ID,
		Content:   chatMsg.Content,
		Timestamp: timestamppb.New(chatMsg.Timestamp),
		UserId:    chatMsg.UserID,
		UserName:  chatMsg.UserName,
		RoomId:    chatMsg.RoomID,
	}

	s.broadcastToRoom(chatMsg.RoomID, pbMsg)
	log.Printf("Message broadcasted to room %s: %s", chatMsg.RoomID, chatMsg.Content)
}

// startConnectionCleanup: 죽은 연결을 주기적으로 정리
func (s *ChatServer) startConnectionCleanup() {
	ticker := time.NewTicker(30 * time.Second) // 30초마다 정리
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanupDeadConnections()
		}
	}
}

// cleanupDeadConnections: 죽은 연결들을 정리
func (s *ChatServer) cleanupDeadConnections() {
	s.roomsMutex.RLock()
	defer s.roomsMutex.RUnlock()

	for roomID, room := range s.rooms {
		room.Mutex.Lock()
		originalCount := len(room.Connections)

		// 활성화된 연결만 남기기
		activeConnections := []*Connection{}
		for _, conn := range room.Connections {
			// Context가 취소되었거나 5분 이상 된 비활성 연결 제거
			if conn.Context.Err() != nil || (!conn.Active && time.Since(conn.CreatedAt) > 5*time.Minute) {
				log.Printf("Removing dead connection for user %s in room %s", conn.User.Id, roomID)
				continue
			}
			activeConnections = append(activeConnections, conn)
		}

		room.Connections = activeConnections
		cleanedCount := originalCount - len(activeConnections)
		room.Mutex.Unlock()

		if cleanedCount > 0 {
			log.Printf("Cleaned up %d dead connections in room %s", cleanedCount, roomID)
		}
	}
}
