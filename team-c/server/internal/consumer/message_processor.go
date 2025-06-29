package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"example.com/grpc-chat-app/internal/config"

	"github.com/IBM/sarama"
)

// ChatMessage: Kafka에서 받은 채팅 메시지
type ChatMessage struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	RoomID    string    `json:"room_id"`
}

// UserConnection: 사용자 연결 정보를 Kafka에 저장하기 위한 구조체
type UserConnection struct {
	UserID    string    `json:"user_id"`
	RoomID    string    `json:"room_id"`
	ServerID  string    `json:"server_id"`
	Connected bool      `json:"connected"`
	Timestamp time.Time `json:"timestamp"`
}

// MessageHandler: 메시지 처리를 위한 인터페이스
type MessageHandler interface {
	ProcessIncomingMessage(msg *ChatMessage)
}

// MessageProcessor: Kafka 메시지를 처리하는 구조체
type MessageProcessor struct {
	config        *config.KafkaConfig
	consumerGroup string
	handler       MessageHandler
	consumer      sarama.ConsumerGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewMessageProcessor: 새로운 메시지 프로세서 생성
func NewMessageProcessor(kafkaConfig *config.KafkaConfig, consumerGroup string) (*MessageProcessor, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Group.Session.Timeout = 10 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumerGroup(kafkaConfig.Brokers, consumerGroup, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	processor := &MessageProcessor{
		config:        kafkaConfig,
		consumerGroup: consumerGroup,
		consumer:      consumer,
		ctx:           ctx,
		cancel:        cancel,
	}

	return processor, nil
}

// SetServer: 메시지 핸들러 설정
func (mp *MessageProcessor) SetServer(handler MessageHandler) {
	mp.handler = handler
}

// Start: 메시지 프로세서 시작
func (mp *MessageProcessor) Start() error {
	if mp.handler == nil {
		return fmt.Errorf("message handler not set")
	}

	log.Printf("Starting message processor with consumer group: %s", mp.consumerGroup)

	// 에러 처리 고루틴
	go func() {
		for {
			select {
			case err := <-mp.consumer.Errors():
				if err != nil {
					log.Printf("Consumer error: %v", err)
				}
			case <-mp.ctx.Done():
				return
			}
		}
	}()

	// 메시지 소비 루프
	for {
		select {
		case <-mp.ctx.Done():
			log.Println("Message processor context cancelled")
			return nil
		default:
			topics := []string{mp.config.Topic}
			err := mp.consumer.Consume(mp.ctx, topics, mp)
			if err != nil {
				log.Printf("Error consuming messages: %v", err)
				time.Sleep(1 * time.Second) // 에러 시 잠시 대기
			}
		}
	}
}

// Stop: 메시지 프로세서 중지
func (mp *MessageProcessor) Stop() {
	log.Println("Stopping message processor...")
	mp.cancel()
	if err := mp.consumer.Close(); err != nil {
		log.Printf("Error closing consumer: %v", err)
	}
}

// Setup: 컨슈머 그룹 세션 설정 (sarama.ConsumerGroupHandler 인터페이스)
func (mp *MessageProcessor) Setup(session sarama.ConsumerGroupSession) error {
	log.Printf("Consumer group session setup - Member ID: %s, Generation ID: %d",
		session.MemberID(), session.GenerationID())
	return nil
}

// Cleanup: 컨슈머 그룹 세션 정리 (sarama.ConsumerGroupHandler 인터페이스)
func (mp *MessageProcessor) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Printf("Consumer group session cleanup - Member ID: %s", session.MemberID())
	return nil
}

// ConsumeClaim: 메시지 소비 (sarama.ConsumerGroupHandler 인터페이스)
func (mp *MessageProcessor) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.Printf("Starting to consume partition %d from offset %d", claim.Partition(), claim.InitialOffset())

	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			mp.processMessage(session, message)

		case <-session.Context().Done():
			log.Printf("Session context cancelled for partition %d", claim.Partition())
			return nil

		case <-mp.ctx.Done():
			log.Printf("Message processor context cancelled for partition %d", claim.Partition())
			return nil
		}
	}
}

// processMessage: 개별 메시지 처리
func (mp *MessageProcessor) processMessage(session sarama.ConsumerGroupSession, message *sarama.ConsumerMessage) {
	var chatMsg ChatMessage
	if err := json.Unmarshal(message.Value, &chatMsg); err != nil {
		log.Printf("Failed to unmarshal message from partition %d, offset %d: %v",
			message.Partition, message.Offset, err)
		session.MarkMessage(message, "")
		return
	}

	log.Printf("Processing message: ID=%s, RoomID=%s, UserID=%s, Content=%s",
		chatMsg.ID, chatMsg.RoomID, chatMsg.UserID, chatMsg.Content)

	// 메시지 핸들러를 통해 처리
	if mp.handler != nil {
		mp.handler.ProcessIncomingMessage(&chatMsg)
	}

	// 메시지 처리 완료 마킹
	session.MarkMessage(message, "")

	log.Printf("Message processed successfully from partition %d, offset %d",
		message.Partition, message.Offset)
}
