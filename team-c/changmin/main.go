package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	pb "example.com/grpc-chat-app/gen" // 생성된 Go 코드 패키지 임포트
	"github.com/IBM/sarama"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// KafkaConfig: Kafka 설정을 위한 구조체
type KafkaConfig struct {
	Brokers []string
	Topic   string
}

// ChatMessage: Kafka에 전송할 메시지 구조체
type ChatMessage struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
}

// UserConnection: 사용자 연결 정보를 Kafka에 저장하기 위한 구조체
type UserConnection struct {
	UserID    string    `json:"user_id"`
	ServerID  string    `json:"server_id"`
	Connected bool      `json:"connected"`
	Timestamp time.Time `json:"timestamp"`
}

// Connection: 개별 클라이언트 연결을 나타내는 구조체
type Connection struct {
	pb.UnimplementedBroadcastServer                                 // 항상 임베드해야 함
	stream                          pb.Broadcast_CreateStreamServer // 클라이언트에게 메시지를 보내기 위한 스트림
	id                              string                          // 연결 ID (예: 사용자 ID)
	active                          bool                            // 연결 활성 상태
	error                           chan error                      // 에러 전파를 위한 채널
}

// Pool: 활성 연결들의 풀(모음)을 관리하는 구조체
type Pool struct {
	pb.UnimplementedBroadcastServer // 항상 임베드해야 함
	Connection                      []*Connection
	Producer                        sarama.SyncProducer // Kafka Producer
	KafkaConfig                     *KafkaConfig
	ServerID                        string            // 현재 서버의 고유 ID
	mutex                           sync.RWMutex      // 연결 풀 보호를 위한 뮤텍스
	MessageProcessor                *MessageProcessor // 임베드된 메시지 프로세서
}

// NewPool: 새로운 Pool 인스턴스를 생성하고 Kafka Producer를 초기화
func NewPool(kafkaConfig *KafkaConfig, serverID string) (*Pool, error) {
	// Kafka Producer 설정
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // 모든 복제본으로부터 확인 대기
	config.Producer.Retry.Max = 5                    // 최대 재시도 횟수
	config.Producer.Return.Successes = true          // 성공 응답 반환

	// Kafka Producer 생성
	producer, err := sarama.NewSyncProducer(kafkaConfig.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// MessageProcessor 생성 (Consumer용)
	processor, err := NewMessageProcessor(kafkaConfig, "chatting-processor-group")
	if err != nil {
		return nil, fmt.Errorf("failed to create message processor: %w", err)
	}

	pool := &Pool{
		Connection:       []*Connection{},
		Producer:         producer,
		KafkaConfig:      kafkaConfig,
		ServerID:         serverID,
		MessageProcessor: processor,
	}

	// MessageProcessor에 Pool 참조 설정 (순환 참조 해결을 위해)
	processor.SetPool(pool)

	return pool, nil
}

// Close: Pool 리소스 정리
func (p *Pool) Close() error {
	if p.MessageProcessor != nil {
		p.MessageProcessor.Stop()
	}
	return p.Producer.Close()
}

// StartMessageProcessor: 메시지 프로세서 시작
func (p *Pool) StartMessageProcessor() {
	go func() {
		if err := p.MessageProcessor.Start(); err != nil {
			log.Printf("Message processor error: %v", err)
		}
	}()
}

// CreateStream: 클라이언트가 연결을 요청하고 메시지 스트림을 설정
func (p *Pool) CreateStream(pconn *pb.Connect, stream pb.Broadcast_CreateStreamServer) error {
	conn := &Connection{
		stream: stream,
		id:     pconn.User.Id,
		active: true,
		error:  make(chan error),
	}

	// 연결 풀에 새로운 연결을 안전하게 추가
	p.mutex.Lock()
	p.Connection = append(p.Connection, conn)
	p.mutex.Unlock()

	log.Printf("User %s connected to server %s", conn.id, p.ServerID)

	// 사용자 연결 정보를 Kafka에 전송
	if err := p.publishUserConnection(conn.id, true); err != nil {
		log.Printf("Failed to publish user connection: %v", err)
	}

	// 스트림이 활성 상태인 동안 에러 채널을 통해 대기
	err := <-conn.error

	// 연결 종료 시 사용자 연결 해제 정보를 Kafka에 전송
	if disconnectErr := p.publishUserConnection(conn.id, false); disconnectErr != nil {
		log.Printf("Failed to publish user disconnection: %v", disconnectErr)
	}

	// 연결 풀에서 해당 연결 제거
	p.removeConnection(conn.id)
	log.Printf("User %s disconnected from server %s", conn.id, p.ServerID)

	return err
}

// removeConnection: 연결 풀에서 특정 연결을 제거
func (p *Pool) removeConnection(userID string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for i, conn := range p.Connection {
		if conn.id == userID {
			// 슬라이스에서 해당 요소 제거
			p.Connection = append(p.Connection[:i], p.Connection[i+1:]...)
			break
		}
	}
}

// publishUserConnection: 사용자 연결/해제 정보를 Kafka에 발행
func (p *Pool) publishUserConnection(userID string, connected bool) error {
	userConn := UserConnection{
		UserID:    userID,
		ServerID:  p.ServerID,
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

	_, _, err = p.Producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send user connection message: %w", err)
	}

	log.Printf("Published user connection event: %s (connected: %t)", userID, connected)
	return nil
}

// BroadcastMessage: 메시지를 받아 Kafka에 발행
func (p *Pool) BroadcastMessage(ctx context.Context, msg *pb.Message) (*pb.Close, error) {
	// 현재 시간을 메시지에 설정
	if msg.Timestamp == nil {
		msg.Timestamp = timestamppb.Now()
	}

	// Kafka에 발행할 메시지 구조체 생성
	chatMsg := ChatMessage{
		ID:        msg.Id,
		Content:   msg.Content,
		Timestamp: msg.Timestamp.AsTime(),
		UserID:    msg.Id, // 여기서는 메시지 ID를 사용자 ID로 사용 (실제로는 별도의 필드가 필요)
	}

	// JSON으로 직렬화
	data, err := json.Marshal(chatMsg)
	if err != nil {
		log.Printf("Failed to marshal chat message: %v", err)
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Kafka 메시지 생성
	kafkaMsg := &sarama.ProducerMessage{
		Topic: p.KafkaConfig.Topic,
		Key:   sarama.StringEncoder(msg.Id),
		Value: sarama.StringEncoder(data),
	}

	// Kafka에 메시지 전송
	partition, offset, err := p.Producer.SendMessage(kafkaMsg)
	if err != nil {
		log.Printf("Failed to send message to Kafka: %v", err)
		return nil, fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	log.Printf("Message sent to Kafka - Topic: %s, Partition: %d, Offset: %d",
		p.KafkaConfig.Topic, partition, offset)

	return &pb.Close{}, nil
}

// ProcessIncomingMessage: 다른 서버나 Consumer로부터 받은 메시지를 로컬 클라이언트들에게 전송
func (p *Pool) ProcessIncomingMessage(chatMsg *ChatMessage) {
	// protobuf 메시지로 변환
	pbMsg := &pb.Message{
		Id:        chatMsg.ID,
		Content:   chatMsg.Content,
		Timestamp: timestamppb.New(chatMsg.Timestamp),
	}

	wait := sync.WaitGroup{}
	done := make(chan int)

	// 연결 풀의 모든 활성 연결에 메시지 전송
	p.mutex.RLock()
	connections := make([]*Connection, len(p.Connection))
	copy(connections, p.Connection)
	p.mutex.RUnlock()

	for _, conn := range connections {
		wait.Add(1)

		go func(msg *pb.Message, conn *Connection) {
			defer wait.Done()

			if conn.active {
				if err := conn.stream.Send(msg); err != nil {
					log.Printf("Error sending message to %s: %v", conn.id, err)
					conn.active = false
					conn.error <- err
				} else {
					log.Printf("Sent message to %s from %s", conn.id, msg.Id)
				}
			}
		}(pbMsg, conn)
	}

	// 모든 메시지 전송 완료를 기다리는 고루틴
	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
}

func main() {
	// 명령행 인수에 따라 실행 모드 결정
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case "consumer":
			runConsumer()
			return
		case "client":
			if len(os.Args) < 3 {
				fmt.Println("Usage: go run . client <user_id>")
				fmt.Println("Example: go run . client alice")
				return
			}
			runClient(os.Args[2])
			return
		case "server":
			// 기본 서버 모드로 계속 진행
		default:
			fmt.Println("Available modes:")
			fmt.Println("  server   - Run gRPC chat server (default)")
			fmt.Println("  consumer - Run Kafka message consumer")
			fmt.Println("  client   - Run chat client")
			fmt.Println("")
			fmt.Println("Examples:")
			fmt.Println("  go run . server")
			fmt.Println("  go run . consumer")
			fmt.Println("  go run . client alice")
			return
		}
	}

	// 기본 서버 모드 실행
	runServer()
}

// runServer: gRPC 서버 실행
func runServer() {
	// Kafka 설정
	kafkaConfig := &KafkaConfig{
		Brokers: []string{"localhost:9092", "localhost:9093", "localhost:9094"},
		Topic:   "chatting",
	}

	// 서버 ID 생성 (실제로는 환경변수나 설정 파일에서 읽어올 것)
	serverID := fmt.Sprintf("grpc-server-%d", time.Now().Unix())

	// Pool 생성
	pool, err := NewPool(kafkaConfig, serverID)
	if err != nil {
		log.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	// 메시지 프로세서 시작 (별도 고루틴에서)
	pool.StartMessageProcessor()

	// 새 gRPC 서버 인스턴스 생성
	grpcServer := grpc.NewServer()

	// 생성된 BroadcastServer 인터페이스에 우리 Pool 구현을 등록
	pb.RegisterBroadcastServer(grpcServer, pool)

	// gRPC Reflection 활성화 (Postman에서 테스트하기 위해)
	reflection.Register(grpcServer)

	// TCP 리스너 생성 (포트 변경: 8081로 설정, Kafka UI가 8080 사용)
	listener, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("Error creating TCP listener: %v", err)
	}

	fmt.Printf("gRPC Server (%s) started at port :8081\n", serverID)
	fmt.Println("Kafka Brokers:", kafkaConfig.Brokers)
	fmt.Println("Kafka Topic:", kafkaConfig.Topic)
	fmt.Println("Message Processor: Started (embedded)")

	// gRPC 서버가 리스너를 통해 요청을 받기 시작
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Error serving gRPC requests: %v", err)
	}
}
