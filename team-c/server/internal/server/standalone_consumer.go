package server

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"example.com/grpc-chat-app/internal/config"
	"example.com/grpc-chat-app/internal/consumer"
)

// StandaloneConsumer: 독립 실행형 컨슈머
type StandaloneConsumer struct {
	config    *config.Config
	processor *consumer.MessageProcessor
}

// NewStandaloneConsumer: 새로운 독립 컨슈머 생성
func NewStandaloneConsumer(cfg *config.Config) (*StandaloneConsumer, error) {
	processor, err := consumer.NewMessageProcessor(cfg.Kafka, cfg.Kafka.ConsumerGroup)
	if err != nil {
		return nil, err
	}

	standaloneConsumer := &StandaloneConsumer{
		config:    cfg,
		processor: processor,
	}

	// 메시지 핸들러 설정 (독립 컨슈머는 메시지를 로그로만 출력)
	processor.SetServer(standaloneConsumer)

	return standaloneConsumer, nil
}

// Start: 독립 컨슈머 시작
func (sc *StandaloneConsumer) Start() error {
	// 종료 시그널 처리
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 메시지 프로세서 시작
	go func() {
		if err := sc.processor.Start(); err != nil {
			log.Printf("Message processor error: %v", err)
		}
	}()

	log.Println("Standalone consumer is running. Press Ctrl+C to stop.")

	// 종료 시그널 대기
	<-sigChan
	log.Println("Shutdown signal received, stopping consumer...")

	// 정리
	sc.processor.Stop()
	log.Println("Standalone consumer stopped")

	return nil
}

// ProcessIncomingMessage: 메시지 처리 (consumer.MessageHandler 인터페이스 구현)
func (sc *StandaloneConsumer) ProcessIncomingMessage(msg *consumer.ChatMessage) {
	// 독립 컨슈머는 메시지를 로그로만 출력
	log.Printf("[STANDALONE] Message received - Room: %s, User: %s (%s), Content: %s",
		msg.RoomID, msg.UserName, msg.UserID, msg.Content)
}
