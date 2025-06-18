package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

// MessageProcessor: Kafka 메시지를 처리하고 서버들에 전달하는 구조체
type MessageProcessor struct {
	consumer        sarama.ConsumerGroup
	servers         map[string]*ServerConnection // 서버 ID -> 연결 정보
	serversMutex    sync.RWMutex
	kafkaConfig     *KafkaConfig
	consumerGroupID string
	ctx             context.Context
	cancel          context.CancelFunc
	pool            interface{} // Pool 참조 (순환 참조 방지를 위해 interface{} 사용)
}

// ServerConnection: 각 gRPC 서버의 연결 정보
type ServerConnection struct {
	ServerID    string
	Address     string
	Users       map[string]bool // 해당 서버에 연결된 사용자들
	LastSeen    time.Time
	UsersMutex  sync.RWMutex
}

// NewMessageProcessor: 새로운 MessageProcessor 인스턴스 생성
func NewMessageProcessor(kafkaConfig *KafkaConfig, consumerGroupID string) (*MessageProcessor, error) {
	// Kafka Consumer 설정
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest // 최신 메시지부터 읽기 시작
	config.Consumer.Group.Session.Timeout = 10 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second

	// Consumer Group 생성
	consumer, err := sarama.NewConsumerGroup(kafkaConfig.Brokers, consumerGroupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &MessageProcessor{
		consumer:        consumer,
		servers:         make(map[string]*ServerConnection),
		kafkaConfig:     kafkaConfig,
		consumerGroupID: consumerGroupID,
		ctx:             ctx,
		cancel:          cancel,
	}, nil
}

// Start: 메시지 처리 시작
func (mp *MessageProcessor) Start() error {
	log.Printf("Starting message processor with consumer group: %s", mp.consumerGroupID)

	// 각 토픽별로 별도의 고루틴에서 처리
	var wg sync.WaitGroup

	// 채팅 메시지 처리
	wg.Add(1)
	go func() {
		defer wg.Done()
		mp.consumeMessages([]string{mp.kafkaConfig.Topic})
	}()

	// 사용자 연결 정보 처리
	wg.Add(1)
	go func() {
		defer wg.Done()
		mp.consumeUserConnections([]string{"user-connections"})
	}()

	// 정리 작업을 위한 고루틴
	wg.Add(1)
	go func() {
		defer wg.Done()
		mp.cleanupInactiveServers()
	}()

	wg.Wait()
	return nil
}

// Stop: 메시지 처리 중지
func (mp *MessageProcessor) Stop() error {
	log.Println("Stopping message processor...")
	mp.cancel()
	return mp.consumer.Close()
}

// consumeMessages: 채팅 메시지 토픽에서 메시지를 소비
func (mp *MessageProcessor) consumeMessages(topics []string) {
	handler := &ChatMessageHandler{
		processor: mp,
	}

	for {
		select {
		case <-mp.ctx.Done():
			log.Println("Chat message consumer stopped")
			return
		default:
			if err := mp.consumer.Consume(mp.ctx, topics, handler); err != nil {
				log.Printf("Error consuming chat messages: %v", err)
				time.Sleep(time.Second) // 에러 발생 시 잠시 대기
			}
		}
	}
}

// consumeUserConnections: 사용자 연결 정보 토픽에서 메시지를 소비
func (mp *MessageProcessor) consumeUserConnections(topics []string) {
	handler := &UserConnectionHandler{
		processor: mp,
	}

	for {
		select {
		case <-mp.ctx.Done():
			log.Println("User connection consumer stopped")
			return
		default:
			if err := mp.consumer.Consume(mp.ctx, topics, handler); err != nil {
				log.Printf("Error consuming user connections: %v", err)
				time.Sleep(time.Second) // 에러 발생 시 잠시 대기
			}
		}
	}
}

// cleanupInactiveServers: 비활성 서버 정리
func (mp *MessageProcessor) cleanupInactiveServers() {
	ticker := time.NewTicker(30 * time.Second) // 30초마다 정리
	defer ticker.Stop()

	for {
		select {
		case <-mp.ctx.Done():
			log.Println("Server cleanup stopped")
			return
		case <-ticker.C:
			mp.removeInactiveServers()
		}
	}
}

// removeInactiveServers: 일정 시간 동안 활동이 없는 서버 제거
func (mp *MessageProcessor) removeInactiveServers() {
	mp.serversMutex.Lock()
	defer mp.serversMutex.Unlock()

	cutoff := time.Now().Add(-2 * time.Minute) // 2분 이상 활동이 없는 서버 제거

	for serverID, server := range mp.servers {
		if server.LastSeen.Before(cutoff) {
			log.Printf("Removing inactive server: %s (last seen: %v)", serverID, server.LastSeen)
			delete(mp.servers, serverID)
		}
	}
}

// ChatMessageHandler: 채팅 메시지를 처리하는 핸들러
type ChatMessageHandler struct {
	processor *MessageProcessor
}

// Setup: Consumer Group 설정 시 호출
func (h *ChatMessageHandler) Setup(sarama.ConsumerGroupSession) error {
	log.Println("Chat message handler setup completed")
	return nil
}

// Cleanup: Consumer Group 정리 시 호출
func (h *ChatMessageHandler) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("Chat message handler cleanup completed")
	return nil
}

// ConsumeClaim: 메시지 처리
func (h *ChatMessageHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			h.handleChatMessage(message)
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// handleChatMessage: 개별 채팅 메시지 처리
func (h *ChatMessageHandler) handleChatMessage(msg *sarama.ConsumerMessage) {
	log.Printf("Received chat message: Topic=%s, Partition=%d, Offset=%d", 
		msg.Topic, msg.Partition, msg.Offset)

	var chatMsg ChatMessage
	if err := json.Unmarshal(msg.Value, &chatMsg); err != nil {
		log.Printf("Failed to unmarshal chat message: %v", err)
		return
	}

	log.Printf("Processing chat message: ID=%s, Content=%s, From=%s", 
		chatMsg.ID, chatMsg.Content, chatMsg.UserID)

	// 메시지를 모든 연결된 서버에 전달 (실제 구현에서는 특정 사용자나 채널 기반으로 필터링)
	h.processor.broadcastToAllServers(&chatMsg)
}

// UserConnectionHandler: 사용자 연결 정보를 처리하는 핸들러
type UserConnectionHandler struct {
	processor *MessageProcessor
}

// Setup: Consumer Group 설정 시 호출
func (h *UserConnectionHandler) Setup(sarama.ConsumerGroupSession) error {
	log.Println("User connection handler setup completed")
	return nil
}

// Cleanup: Consumer Group 정리 시 호출
func (h *UserConnectionHandler) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("User connection handler cleanup completed")
	return nil
}

// ConsumeClaim: 사용자 연결 정보 처리
func (h *UserConnectionHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			h.handleUserConnection(message)
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// handleUserConnection: 사용자 연결 정보 처리
func (h *UserConnectionHandler) handleUserConnection(msg *sarama.ConsumerMessage) {
	log.Printf("Received user connection event: Topic=%s, Partition=%d, Offset=%d", 
		msg.Topic, msg.Partition, msg.Offset)

	var userConn UserConnection
	if err := json.Unmarshal(msg.Value, &userConn); err != nil {
		log.Printf("Failed to unmarshal user connection: %v", err)
		return
	}

	log.Printf("Processing user connection: UserID=%s, ServerID=%s, Connected=%t", 
		userConn.UserID, userConn.ServerID, userConn.Connected)

	h.processor.updateServerConnection(userConn)
}

// updateServerConnection: 서버 연결 정보 업데이트
func (mp *MessageProcessor) updateServerConnection(userConn UserConnection) {
	mp.serversMutex.Lock()
	defer mp.serversMutex.Unlock()

	server, exists := mp.servers[userConn.ServerID]
	if !exists {
		// 새로운 서버 정보 생성
		server = &ServerConnection{
			ServerID: userConn.ServerID,
			Address:  fmt.Sprintf("localhost:8082"), // HTTP 내부 API 포트
			Users:    make(map[string]bool),
			LastSeen: userConn.Timestamp,
		}
		mp.servers[userConn.ServerID] = server
		log.Printf("New server registered: %s", userConn.ServerID)
	}

	server.UsersMutex.Lock()
	if userConn.Connected {
		server.Users[userConn.UserID] = true
		log.Printf("User %s connected to server %s", userConn.UserID, userConn.ServerID)
	} else {
		delete(server.Users, userConn.UserID)
		log.Printf("User %s disconnected from server %s", userConn.UserID, userConn.ServerID)
	}
	server.UsersMutex.Unlock()

	server.LastSeen = userConn.Timestamp
}

// SetPool: Pool 참조 설정
func (mp *MessageProcessor) SetPool(pool interface{}) {
	mp.pool = pool
}

// broadcastToAllServers: 모든 서버에 메시지 브로드캐스트 (Pool을 통해 직접 전달)
func (mp *MessageProcessor) broadcastToAllServers(chatMsg *ChatMessage) {
	// Pool이 설정되어 있으면 직접 ProcessIncomingMessage 호출
	if mp.pool != nil {
		if pool, ok := mp.pool.(interface{ ProcessIncomingMessage(*ChatMessage) }); ok {
			log.Printf("Forwarding message directly to local clients: ID=%s", chatMsg.ID)
			pool.ProcessIncomingMessage(chatMsg)
			log.Printf("Message successfully forwarded to local clients")
			return
		}
	}

	// Fallback: 기존 서버 레지스트리 방식 (사용하지 않음)
	mp.serversMutex.RLock()
	defer mp.serversMutex.RUnlock()

	log.Printf("No pool reference available, message not forwarded: %s", chatMsg.ID)
}

// runConsumer: Consumer를 독립 실행파일로 실행하기 위한 함수
func runConsumer() {
	// Kafka 설정
	kafkaConfig := &KafkaConfig{
		Brokers: []string{"localhost:9092", "localhost:9093", "localhost:9094"},
		Topic:   "chatting",
	}

	// Consumer Group ID (여러 Consumer 인스턴스가 같은 그룹에 속하면 파티션별로 분산 처리)
	consumerGroupID := "chatting-processor-group"

	// Message Processor 생성
	processor, err := NewMessageProcessor(kafkaConfig, consumerGroupID)
	if err != nil {
		log.Fatalf("Failed to create message processor: %v", err)
	}

	// 시그널 핸들링을 위한 채널 생성
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	// Message Processor 시작 (별도 고루틴에서)
	go func() {
		if err := processor.Start(); err != nil {
			log.Printf("Message processor error: %v", err)
		}
	}()

	log.Println("Message processor started. Press Ctrl+C to stop.")

	// 종료 신호 대기
	<-sigterm
	log.Println("Received termination signal")

	// Graceful shutdown
	if err := processor.Stop(); err != nil {
		log.Printf("Error stopping message processor: %v", err)
	}

	log.Println("Message processor stopped")
}

// main 함수 - 컨슈머를 독립적으로 실행하기 위해 주석 해제됨
func init() {
	// main.go와 함께 빌드될 때는 이 함수가 실행되지 않도록 함
}

// runConsumerMain: 독립 실행 시 사용
func runConsumerMain() {
	// Consumer만 실행하는 경우
	if len(os.Args) > 1 && strings.ToLower(os.Args[1]) == "consumer" {
		runConsumer()
		return
	}
	
	fmt.Println("Usage: go run consumer.go consumer")
	fmt.Println("Or import this package and use MessageProcessor directly")
} 