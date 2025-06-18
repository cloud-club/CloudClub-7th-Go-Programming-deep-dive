# 🚀 실시간 채팅 시스템 완벽 분석

## 📋 프로젝트 개요

이 프로젝트는 **Go + gRPC + Kafka**를 활용한 **분산 실시간 채팅 시스템**입니다. 확장성과 고가용성을 고려한 마이크로서비스 아키텍처로 설계되었으며, 다음과 같은 핵심 기술들을 활용합니다:

- **gRPC**: 고성능 양방향 스트리밍 통신
- **Apache Kafka**: 분산 메시지 브로커 시스템
- **Protocol Buffers**: 효율적인 데이터 직렬화
- **Docker Compose**: 컨테이너 기반 인프라 구성
- **Go Goroutines**: 동시성 프로그래밍

## 🏗️ 전체 시스템 아키텍처

### 1. 아키텍처 다이어그램

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   gRPC Client   │    │   gRPC Client   │    │   gRPC Client   │
│     (Alice)     │    │      (Bob)      │    │    (Charlie)    │
│   Port: Dynamic │    │   Port: Dynamic │    │   Port: Dynamic │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          │ ① CreateStream()     │ ② BroadcastMessage() │ ③ Recv Messages
          │   BroadcastMessage() │                      │
          ▼                      ▼                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                    gRPC Chat Server                             │
│                   (localhost:8081)                              │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │  Connection     │  │  Kafka Producer │  │ MessageProcessor│ │
│  │     Pool        │  │                 │  │   (내장형)     │ │
│  │                 │  │ - Send to       │  │                 │ │
│  │ - Alice Stream  │  │   "chatting"    │  │ - Consumer Group│ │
│  │ - Bob Stream    │  │   Topic         │  │ - Pool 참조     │ │
│  │ - Charlie Stream│  │ - User Events   │  │ - Broadcast     │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────┬───────────────────────────────────┬─────────┘
                  │ ④ Produce Messages                │ ⑥ Consume &
                  │   to Kafka                        │   Process Messages
                  ▼                                   ▲
┌─────────────────────────────────────────────────────────────────┐
│                       Kafka Cluster                             │
│                    (3-Broker Setup)                             │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   kafka-01      │  │   kafka-02      │  │   kafka-03      │ │
│  │   :9092         │  │   :9093         │  │   :9094         │ │
│  │   Broker ID: 1  │  │   Broker ID: 2  │  │   Broker ID: 3  │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
│                                                                 │
│  ⑤ Topics & Partitions:                                        │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ chatting Topic:                                         │   │
│  │ ├── Partition 0 (Replica: kafka-01, kafka-02)         │   │
│  │ └── Partition 1 (Replica: kafka-02, kafka-03)         │   │
│  │                                                         │   │
│  │ user-connections Topic:                                 │   │
│  │ ├── Partition 0 (Replica: kafka-01, kafka-02)         │   │
│  │ ├── Partition 1 (Replica: kafka-02, kafka-03)         │   │
│  │ └── Partition 2 (Replica: kafka-03, kafka-01)         │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
                    ┌─────────────────────────────┐
                    │      Kafka UI & Monitoring  │
                    │      (localhost:8080)       │
                    └─────────────────────────────┘
```

### 2. 메시지 플로우 상세 분석

#### ① 클라이언트 연결 과정

```
Alice Client → gRPC Server
├── CreateStream(Connect{UserID: "alice", Active: true})
├── gRPC Server: Connection Pool에 Alice Stream 추가
├── Kafka Producer: user-connections 토픽에 연결 이벤트 발행
└── MessageProcessor: 사용자 연결 이벤트 처리
```

#### ② 메시지 전송 과정

```
Alice: "Hello World!" 입력
├── BroadcastMessage(Message{ID: "alice", Content: "Hello World!"})
├── gRPC Server: Kafka Producer를 통해 "chatting" 토픽에 메시지 발행
├── Kafka: 메시지를 파티션에 저장 (Key 기반 파티셔닝)
├── MessageProcessor: 메시지 소비 및 Pool 참조를 통해 로컬 클라이언트들에 전달
└── All Clients: Alice의 메시지 실시간 수신
```

## 💻 Go 언어 활용 방식 상세 분석

### 1. gRPC 서버 구현 패턴

#### Protocol Buffers 정의 (`proto/chat.proto`)

```protobuf
syntax = "proto3";
package chat;

// 사용자 정보
message User {
  string id = 1;
  string name = 2;
}

// 채팅 메시지
message Message {
  string id = 1;
  string content = 2;
  google.protobuf.Timestamp timestamp = 3;
}

// 연결 요청
message Connect {
  User user = 1;
  bool active = 2;
}

// 서비스 정의
service Broadcast {
  // 서버 스트리밍: 클라이언트가 연결하면 지속적으로 메시지 수신
  rpc CreateStream(Connect) returns (stream Message);

  // 단일 요청-응답: 메시지 전송
  rpc BroadcastMessage(Message) returns (Close);
}
```

#### gRPC 서버 구현 핵심 구조

```go
// Pool: 연결 관리 및 메시지 브로드캐스팅을 담당하는 핵심 구조체
type Pool struct {
    pb.UnimplementedBroadcastServer  // gRPC 인터페이스 구현
    Connection      []*Connection    // 활성 클라이언트 연결들
    Producer        sarama.SyncProducer  // Kafka Producer
    KafkaConfig     *KafkaConfig
    ServerID        string           // 서버 고유 식별자
    mutex           sync.RWMutex     // 동시성 안전성을 위한 뮤텍스
    MessageProcessor *MessageProcessor // 내장된 메시지 처리기
}

// Connection: 개별 클라이언트 연결을 나타내는 구조체
type Connection struct {
    stream pb.Broadcast_CreateStreamServer  // gRPC 스트림
    id     string                          // 사용자 ID
    active bool                            // 연결 상태
    error  chan error                      // 에러 전파용 채널
}
```

### 2. 동시성 프로그래밍 패턴

#### Goroutine 활용 예시

```go
// CreateStream: 클라이언트 연결 및 스트림 관리
func (p *Pool) CreateStream(pconn *pb.Connect, stream pb.Broadcast_CreateStreamServer) error {
    conn := &Connection{
        stream: stream,
        id:     pconn.User.Id,
        active: true,
        error:  make(chan error),
    }

    // 1. 연결 풀에 안전하게 추가 (뮤텍스 사용)
    p.mutex.Lock()
    p.Connection = append(p.Connection, conn)
    p.mutex.Unlock()

    // 2. 사용자 연결 이벤트를 Kafka에 비동기 발행
    go func() {
        if err := p.publishUserConnection(conn.id, true); err != nil {
            log.Printf("Failed to publish user connection: %v", err)
        }
    }()

    // 3. 에러 채널을 통한 연결 상태 관리
    err := <-conn.error  // 블로킹: 연결이 종료될 때까지 대기

    // 4. 정리 작업
    p.removeConnection(conn.id)
    return err
}
```

#### Channel 패턴 활용

```go
// 클라이언트에서의 Graceful Shutdown 패턴
func runClient(userID string) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 시그널 처리용 채널
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // 메시지 수신을 별도 고루틴에서 처리
    go func() {
        if err := client.ConnectAndListen(ctx); err != nil {
            if ctx.Err() == nil {
                log.Printf("Listen error: %v", err)
            }
        }
    }()

    // 사용자 입력 처리를 별도 고루틴에서 처리
    go func() {
        scanner := bufio.NewScanner(os.Stdin)
        for scanner.Scan() {
            input := strings.TrimSpace(scanner.Text())
            if input == "/quit" {
                cancel()
                return
            }
            client.SendMessage(ctx, input)
        }
    }()

    // 종료 조건 대기 (시그널 또는 컨텍스트 취소)
    select {
    case <-sigChan:
        fmt.Println("Received interrupt signal")
        cancel()
    case <-ctx.Done():
        // 사용자가 /quit 입력
    }
}
```

### 3. Kafka 클라이언트 (Sarama) 활용

#### Producer 설정 및 구현

```go
// Kafka Producer 설정 - 고가용성과 데이터 안전성 중심
func NewPool(kafkaConfig *KafkaConfig, serverID string) (*Pool, error) {
    config := sarama.NewConfig()
    config.Producer.RequiredAcks = sarama.WaitForAll  // 모든 복제본 확인
    config.Producer.Retry.Max = 5                     // 최대 5회 재시도
    config.Producer.Return.Successes = true           // 성공 응답 반환

    producer, err := sarama.NewSyncProducer(kafkaConfig.Brokers, config)
    if err != nil {
        return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
    }

    return &Pool{Producer: producer, ...}, nil
}

// 메시지 발행 구현
func (p *Pool) BroadcastMessage(ctx context.Context, msg *pb.Message) (*pb.Close, error) {
    // 1. Protocol Buffer를 JSON으로 변환
    chatMsg := ChatMessage{
        ID:        msg.Id,
        Content:   msg.Content,
        Timestamp: msg.Timestamp.AsTime(),
        UserID:    msg.Id,
    }

    data, err := json.Marshal(chatMsg)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal message: %w", err)
    }

    // 2. Kafka 메시지 구성
    kafkaMsg := &sarama.ProducerMessage{
        Topic: p.KafkaConfig.Topic,  // "chatting"
        Key:   sarama.StringEncoder(msg.Id),  // 사용자 ID로 파티셔닝
        Value: sarama.StringEncoder(data),
    }

    // 3. 동기식 전송 (안전성 보장)
    partition, offset, err := p.Producer.SendMessage(kafkaMsg)
    if err != nil {
        return nil, fmt.Errorf("failed to send message to Kafka: %w", err)
    }

    log.Printf("Message sent to Kafka - Topic: %s, Partition: %d, Offset: %d",
        p.KafkaConfig.Topic, partition, offset)

    return &pb.Close{}, nil
}
```

#### Consumer Group 구현

```go
// MessageProcessor: Kafka Consumer Group을 활용한 메시지 처리
type MessageProcessor struct {
    consumer        sarama.ConsumerGroup    // Consumer Group 인스턴스
    servers         map[string]*ServerConnection  // 서버 레지스트리
    serversMutex    sync.RWMutex
    kafkaConfig     *KafkaConfig
    consumerGroupID string                  // "chatting-processor-group"
    ctx             context.Context
    cancel          context.CancelFunc
    pool            interface{}             // Pool 참조 (순환 참조 방지)
}

// Consumer 설정 - 확장성과 장애 복구 중심
func NewMessageProcessor(kafkaConfig *KafkaConfig, consumerGroupID string) (*MessageProcessor, error) {
    config := sarama.NewConfig()
    config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
    config.Consumer.Offsets.Initial = sarama.OffsetNewest  // 최신 메시지부터
    config.Consumer.Group.Session.Timeout = 10 * time.Second
    config.Consumer.Group.Heartbeat.Interval = 3 * time.Second

    consumer, err := sarama.NewConsumerGroup(kafkaConfig.Brokers, consumerGroupID, config)
    return &MessageProcessor{consumer: consumer, ...}, err
}
```

## 🔧 Kafka 아키텍처 및 3개 브로커 사용 이유

### 1. Kafka 클러스터 구성 분석

#### Docker Compose 기반 클러스터 설정

```yaml
# 3-Broker Kafka Cluster + Zookeeper
services:
  zookeeper:
    image: confluentinc/cp-zookeeper:7.4.0
    ports: ["2181:2181"]
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000

  kafka-01:
    image: confluentinc/cp-kafka:7.4.0
    ports: ["9092:9092"]
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: "zookeeper:2181"
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 2

  kafka-02:
    image: confluentinc/cp-kafka:7.4.0
    ports: ["9093:9093"]
    environment:
      KAFKA_BROKER_ID: 2
      KAFKA_ZOOKEEPER_CONNECT: "zookeeper:2181"

  kafka-03:
    image: confluentinc/cp-kafka:7.4.0
    ports: ["9094:9094"]
    environment:
      KAFKA_BROKER_ID: 3
      KAFKA_ZOOKEEPER_CONNECT: "zookeeper:2181"
```

### 2. 3개 브로커 사용 이유 심화 분석

#### ① 고가용성 (High Availability)

```
시나리오: kafka-02 브로커 장애 발생 시

Before (장애 전):
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│  kafka-01   │  │  kafka-02   │  │  kafka-03   │
│  (Leader)   │  │  (Follower) │  │  (Follower) │
│  Partition 0│  │  Partition 0│  │  Partition 1│
│  Partition 1│  │  Partition 1│  │  Partition 0│
└─────────────┘  └─────────────┘  └─────────────┘

After (장애 후 - 자동 리더 선출):
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│  kafka-01   │  │   (DOWN)    │  │  (NEW LEADER)│
│  (Leader)   │  │      X      │  │  Partition 1 │
│  Partition 0│  │      X      │  │  Partition 0│
└─────────────┘  └─────────────┘  └─────────────┘

결과: 서비스 중단 없이 계속 운영 (kafka-03이 Partition 1의 새 리더가 됨)
```

#### ② 복제 전략 (Replication Factor = 2)

```
chatting Topic (2 Partitions, 2 Replicas):

Partition 0:
├── Leader: kafka-01
└── Follower: kafka-02

Partition 1:
├── Leader: kafka-02
└── Follower: kafka-03

user-connections Topic (3 Partitions, 2 Replicas):

Partition 0:
├── Leader: kafka-01
└── Follower: kafka-02

Partition 1:
├── Leader: kafka-02
└── Follower: kafka-03

Partition 2:
├── Leader: kafka-03
└── Follower: kafka-01
```

**복제 전략의 핵심**:

- **데이터 안전성**: 최소 2개 브로커에 데이터 복제
- **장애 허용**: 1개 브로커 장애 시에도 서비스 지속
- **자동 복구**: 장애 브로커 복구 시 자동으로 동기화

#### ③ 분산 처리 및 로드 밸런싱

```go
// 메시지 파티셔닝 로직
kafkaMsg := &sarama.ProducerMessage{
    Topic: "chatting",
    Key:   sarama.StringEncoder(msg.Id),  // 사용자 ID로 파티셔닝
    Value: sarama.StringEncoder(data),
}

// 파티션 분배 예시:
// Alice (Key: "alice") → hash("alice") % 2 = 0 → Partition 0 (kafka-01)
// Bob   (Key: "bob")   → hash("bob") % 2 = 1   → Partition 1 (kafka-02)
// Charlie (Key: "charlie") → hash("charlie") % 2 = 0 → Partition 0 (kafka-01)
```

**분산 처리 장점**:

- **처리량 향상**: 여러 파티션 동시 처리
- **병렬성**: Consumer Group에서 파티션별 병렬 소비
- **확장성**: 파티션 수 증가로 처리량 선형 확장

### 3. 토픽 설계 전략

#### ① chatting Topic

```bash
# 토픽 생성 명령어 (init-kafka 컨테이너에서 실행)
kafka-topics --create --if-not-exists \
  --bootstrap-server kafka-01:29092 \
  --partitions 2 \
  --replication-factor 2 \
  --topic chatting
```

**설계 이유**:

- **파티션 2개**: 적당한 병렬 처리 (확장 가능)
- **복제 2개**: 1개 브로커 장애 허용
- **키 기반 파티셔닝**: 사용자별 메시지 순서 보장

#### ② user-connections Topic

```bash
kafka-topics --create --if-not-exists \
  --bootstrap-server kafka-01:29092 \
  --partitions 3 \
  --replication-factor 2 \
  --topic user-connections
```

**설계 이유**:

- **파티션 3개**: 연결 이벤트 높은 처리량 대응
- **복제 2개**: 연결 정보 손실 방지
- **분산 저장**: 서버별 연결 정보 분산 관리

## 🏛️ 채팅 서버 아키텍처 심화 분석

### 1. Pool 구조체 - 핵심 연결 관리자

```go
type Pool struct {
    pb.UnimplementedBroadcastServer  // gRPC 서비스 인터페이스
    Connection      []*Connection    // 활성 클라이언트 연결들
    Producer        sarama.SyncProducer  // Kafka 메시지 발행자
    KafkaConfig     *KafkaConfig     // Kafka 설정
    ServerID        string           // 서버 고유 식별자
    mutex           sync.RWMutex     // 연결 풀 동시성 보호
    MessageProcessor *MessageProcessor // 내장 메시지 처리기
}
```

**Pool의 역할**:

1. **연결 관리**: 클라이언트 gRPC 스트림 관리
2. **메시지 라우팅**: 수신 메시지를 Kafka로 발행
3. **브로드캐스팅**: Kafka에서 받은 메시지를 연결된 클라이언트들에게 전달
4. **상태 관리**: 사용자 연결/해제 상태 추적

### 2. Connection 구조체 - 개별 클라이언트 연결

```go
type Connection struct {
    pb.UnimplementedBroadcastServer
    stream pb.Broadcast_CreateStreamServer  // gRPC 서버 스트림
    id     string                          // 사용자 고유 ID
    active bool                           // 연결 활성 상태
    error  chan error                     // 에러 전파 채널
}
```

**Connection의 생명주기**:

```
1. 클라이언트 연결 요청 (CreateStream)
   ↓
2. Connection 객체 생성 및 Pool에 추가
   ↓
3. 사용자 연결 이벤트 Kafka 발행
   ↓
4. 에러 채널 대기 (연결 유지)
   ↓
5. 연결 종료 시 정리 작업
   ↓
6. 사용자 해제 이벤트 Kafka 발행
   ↓
7. Pool에서 Connection 제거
```

### 3. MessageProcessor - 내장형 메시지 처리기

```go
type MessageProcessor struct {
    consumer        sarama.ConsumerGroup
    servers         map[string]*ServerConnection  // 분산 서버 레지스트리
    serversMutex    sync.RWMutex
    kafkaConfig     *KafkaConfig
    consumerGroupID string
    ctx             context.Context
    cancel          context.CancelFunc
    pool            interface{}  // Pool 참조 (순환 참조 방지)
}
```

**MessageProcessor의 멀티 스레드 처리**:

```go
func (mp *MessageProcessor) Start() error {
    var wg sync.WaitGroup

    // 1. 채팅 메시지 처리 고루틴
    wg.Add(1)
    go func() {
        defer wg.Done()
        mp.consumeMessages([]string{mp.kafkaConfig.Topic})
    }()

    // 2. 사용자 연결 정보 처리 고루틴
    wg.Add(1)
    go func() {
        defer wg.Done()
        mp.consumeUserConnections([]string{"user-connections"})
    }()

    // 3. 비활성 서버 정리 고루틴
    wg.Add(1)
    go func() {
        defer wg.Done()
        mp.cleanupInactiveServers()
    }()

    wg.Wait()
    return nil
}
```

### 4. 메시지 브로드캐스팅 메커니즘

```go
func (p *Pool) ProcessIncomingMessage(chatMsg *ChatMessage) {
    // Protocol Buffer로 변환
    pbMsg := &pb.Message{
        Id:        chatMsg.ID,
        Content:   chatMsg.Content,
        Timestamp: timestamppb.New(chatMsg.Timestamp),
    }

    wait := sync.WaitGroup{}
    done := make(chan int)

    // 연결 풀의 스냅샷 생성 (동시성 안전)
    p.mutex.RLock()
    connections := make([]*Connection, len(p.Connection))
    copy(connections, p.Connection)
    p.mutex.RUnlock()

    // 모든 활성 연결에 병렬로 메시지 전송
    for _, conn := range connections {
        wait.Add(1)
        go func(msg *pb.Message, conn *Connection) {
            defer wait.Done()

            if conn.active {
                if err := conn.stream.Send(msg); err != nil {
                    log.Printf("Error sending message to %s: %v", conn.id, err)
                    conn.active = false
                    conn.error <- err  // 연결 종료 신호
                } else {
                    log.Printf("Sent message to %s from %s", conn.id, msg.Id)
                }
            }
        }(pbMsg, conn)
    }

    // 모든 전송 완료 대기
    go func() {
        wait.Wait()
        close(done)
    }()

    <-done
}
```

**브로드캐스팅의 핵심 특징**:

- **병렬 처리**: 각 클라이언트에 동시 전송
- **장애 격리**: 한 클라이언트 오류가 다른 클라이언트에 영향 없음
- **동시성 안전**: RWMutex로 연결 풀 보호
- **자동 정리**: 비활성 연결 자동 제거

## 📊 메시지 플로우 및 실제 동작 과정

### 1. 완전한 메시지 전송 시나리오

#### 시나리오: Alice가 "Hello World!" 메시지를 전송하는 과정

```
[Step 1: 클라이언트 입력]
Alice Terminal: [alice] > Hello World! (Enter)

[Step 2: gRPC 클라이언트 처리]
client.go:
├── SendMessage(ctx, "Hello World!")
├── Message 구조체 생성:
│   ├── Id: "alice"
│   ├── Content: "Hello World!"
│   └── Timestamp: 2025-01-18T20:15:30Z
└── BroadcastMessage() gRPC 호출

[Step 3: gRPC 서버 수신]
main.go - Pool.BroadcastMessage():
├── Protocol Buffer → JSON 변환:
│   {
│     "id": "alice",
│     "content": "Hello World!",
│     "timestamp": "2025-01-18T20:15:30Z",
│     "user_id": "alice"
│   }
├── Kafka Producer Message 생성:
│   ├── Topic: "chatting"
│   ├── Key: "alice" (파티셔닝용)
│   └── Value: JSON 데이터
└── Kafka에 동기 전송 → Partition 0, Offset 127

[Step 4: Kafka 처리]
kafka-01 (Partition 0 Leader):
├── 메시지 수신 및 저장
├── kafka-02 (Follower)에 복제
├── Producer에게 확인 응답
└── Consumer Group에 메시지 전달

[Step 5: MessageProcessor 소비]
consumer.go - ChatMessageHandler:
├── Kafka에서 메시지 수신
├── JSON → ChatMessage 구조체 변환
├── broadcastToAllServers() 호출
└── Pool.ProcessIncomingMessage() 실행

[Step 6: 로컬 브로드캐스팅]
main.go - Pool.ProcessIncomingMessage():
├── ChatMessage → Protocol Buffer 변환
├── 연결 풀 스냅샷 생성 (alice, bob, charlie)
├── 각 연결에 병렬 전송:
│   ├── alice.stream.Send(message) → "✓ [20:15:30] You: Hello World!"
│   ├── bob.stream.Send(message) → "📩 [20:15:30] alice: Hello World!"
│   └── charlie.stream.Send(message) → "📩 [20:15:30] alice: Hello World!"
└── 전송 완료

[Step 7: 클라이언트 출력]
Alice Terminal: ✓ [20:15:30] You: Hello World!
Bob Terminal:   📩 [20:15:30] alice: Hello World!
Charlie Terminal: 📩 [20:15:30] alice: Hello World!
```

### 2. 연결 관리 플로우

#### 새 사용자 연결 과정

```
[Bob 연결 시퀀스]

1. Bob Client 시작:
   go run . client bob

2. gRPC 연결 설정:
   NewChatClient("localhost:8081", "bob")
   ├── grpc.Dial() → gRPC 연결 설정
   └── BroadcastClient 생성

3. 스트림 생성:
   CreateStream(Connect{User: {Id: "bob", Name: "User-bob"}, Active: true})

4. 서버 측 처리:
   Pool.CreateStream():
   ├── Connection 객체 생성 (bob stream)
   ├── 연결 풀에 추가 (mutex.Lock())
   ├── Kafka에 연결 이벤트 발행:
   │   Topic: "user-connections"
   │   Data: {"user_id":"bob","server_id":"grpc-server-1234","connected":true}
   └── 에러 채널 대기 시작

5. MessageProcessor 처리:
   UserConnectionHandler:
   ├── 연결 이벤트 수신
   ├── 서버 레지스트리 업데이트
   └── 로그: "User bob connected to server grpc-server-1234"

6. 연결 완료:
   Bob Terminal: "Connected to server as user: bob"
```

### 3. 장애 처리 및 복구 메커니즘

#### Kafka 브로커 장애 시나리오

```
[장애 발생]
kafka-02 브로커 다운 (Partition 1 Leader)

[자동 복구 과정]
1. Zookeeper 감지:
   ├── kafka-02 하트비트 중단 감지
   └── 리더 선출 프로세스 시작

2. 리더 재선출:
   ├── Partition 1: kafka-03이 새 Leader가 됨
   └── ISR(In-Sync Replicas) 업데이트

3. Producer/Consumer 자동 재연결:
   ├── Sarama 클라이언트가 메타데이터 갱신
   ├── 새 Leader(kafka-03)로 자동 연결
   └── 메시지 처리 계속 진행

[결과]
- 서비스 중단: 1-3초 (리더 선출 시간)
- 데이터 손실: 없음 (복제본 존재)
- 자동 복구: 완전 자동화
```

## 🚀 확장성 및 성능 최적화

### 1. 수평 확장 전략

#### 서버 인스턴스 확장

```bash
# 여러 gRPC 서버 인스턴스 실행
# 터미널 1
export SERVER_PORT=8081 && go run . server

# 터미널 2
export SERVER_PORT=8082 && go run . server

# 터미널 3
export SERVER_PORT=8083 && go run . server

# 로드 밸런서 (nginx 예시)
upstream grpc_servers {
    server localhost:8081;
    server localhost:8082;
    server localhost:8083;
}
```

**확장 효과**:

- **연결 분산**: 클라이언트 부하 분산
- **처리량 증가**: 서버별 독립적 처리
- **장애 격리**: 한 서버 장애가 전체에 영향 없음

#### Kafka 파티션 확장

```bash
# chatting 토픽 파티션 확장 (2 → 6)
kafka-topics --alter --bootstrap-server localhost:9092 \
  --topic chatting --partitions 6

# Consumer 확장 효과
Consumer Group "chatting-processor-group":
├── Consumer 1: Partition 0, 1
├── Consumer 2: Partition 2, 3
└── Consumer 3: Partition 4, 5
```

**확장 효과**:

- **병렬 처리**: 파티션별 독립 처리
- **처리량 향상**: 선형적 확장 가능
- **부하 분산**: 메시지 키 기반 분산

### 2. 성능 최적화 기법

#### ① Connection Pool 최적화

```go
// 현재 구현 (슬라이스 기반)
type Pool struct {
    Connection []*Connection  // O(n) 검색, O(n) 삭제
    mutex      sync.RWMutex
}

// 최적화된 구현 (맵 기반)
type OptimizedPool struct {
    Connections map[string]*Connection  // O(1) 검색, O(1) 삭제
    mutex       sync.RWMutex
}

func (p *OptimizedPool) RemoveConnection(userID string) {
    p.mutex.Lock()
    defer p.mutex.Unlock()
    delete(p.Connections, userID)  // O(1) 삭제
}
```

#### ② Kafka 배치 처리

```go
// 현재: 단일 메시지 전송
func (p *Pool) BroadcastMessage(msg *pb.Message) {
    p.Producer.SendMessage(kafkaMsg)  // 즉시 전송
}

// 최적화: 배치 처리
type BatchProducer struct {
    messages chan *sarama.ProducerMessage
    batch    []*sarama.ProducerMessage
    ticker   *time.Ticker
}

func (bp *BatchProducer) Start() {
    bp.ticker = time.NewTicker(10 * time.Millisecond)  // 10ms 배치
    for {
        select {
        case msg := <-bp.messages:
            bp.batch = append(bp.batch, msg)
        case <-bp.ticker.C:
            if len(bp.batch) > 0 {
                bp.Producer.SendMessages(bp.batch)  // 배치 전송
                bp.batch = bp.batch[:0]
            }
        }
    }
}
```

#### ③ 메모리 풀링

```go
// Protocol Buffer 메시지 풀링
var messagePool = sync.Pool{
    New: func() interface{} {
        return &pb.Message{}
    },
}

func (p *Pool) ProcessIncomingMessage(chatMsg *ChatMessage) {
    // 풀에서 재사용 가능한 객체 가져오기
    pbMsg := messagePool.Get().(*pb.Message)
    defer messagePool.Put(pbMsg)  // 사용 후 풀에 반환

    // 메시지 설정
    pbMsg.Reset()
    pbMsg.Id = chatMsg.ID
    pbMsg.Content = chatMsg.Content
    pbMsg.Timestamp = timestamppb.New(chatMsg.Timestamp)

    // 브로드캐스팅
    p.broadcastToConnections(pbMsg)
}
```

### 3. 모니터링 및 메트릭

#### 핵심 성능 지표

```go
type Metrics struct {
    // 연결 관련
    ActiveConnections    int64    // 현재 활성 연결 수
    TotalConnections     int64    // 총 연결 수
    ConnectionsPerSecond float64  // 초당 연결 수

    // 메시지 관련
    MessagesPerSecond    float64  // 초당 메시지 처리량
    MessageLatency       time.Duration  // 평균 메시지 지연시간
    KafkaProduceLatency  time.Duration  // Kafka 전송 지연시간

    // 에러 관련
    ErrorRate           float64   // 에러율
    FailedConnections   int64     // 실패한 연결 수
    KafkaProduceErrors  int64     // Kafka 전송 실패 수
}

// Prometheus 메트릭 수집
func (p *Pool) updateMetrics() {
    p.mutex.RLock()
    activeConnections := len(p.Connection)
    p.mutex.RUnlock()

    prometheus.activeConnectionsGauge.Set(float64(activeConnections))
    prometheus.messagesPerSecondCounter.Inc()
}
```

## 🧪 실제 실행 및 테스트 방법

### 1. 환경 설정 및 실행

#### Step 1: 전체 환경 구성

```bash
# 1. 의존성 설치
cd team-c/changmin
go mod tidy

# 2. Kafka 클러스터 시작
make kafka-up
# 또는
docker-compose up -d

# 3. Kafka 준비 상태 확인
docker-compose logs init-kafka
# "Topics created successfully!" 메시지 확인

# 4. Kafka UI 접속 (선택사항)
open http://localhost:8080
```

#### Step 2: 서버 실행

```bash
# gRPC 서버 실행 (MessageProcessor 내장)
make server

# 실행 로그 예시:
# gRPC Server (grpc-server-1705665890) started at port :8081
# Kafka Brokers: [localhost:9092 localhost:9093 localhost:9094]
# Kafka Topic: chatting
# Message Processor: Started (embedded)
# Starting message processor with consumer group: chatting-processor-group
```

#### Step 3: 클라이언트 테스트

```bash
# 터미널 1: Alice 클라이언트
make client USER=alice

# 터미널 2: Bob 클라이언트
make client USER=bob

# 터미널 3: Charlie 클라이언트
make client USER=charlie
```

### 2. 실제 채팅 테스트 시나리오

#### 시나리오 1: 기본 채팅 테스트

```bash
# Alice 터미널
[alice] > 안녕하세요! Alice입니다.
✓ [21:30:15] You: 안녕하세요! Alice입니다.

# Bob 터미널
📩 [21:30:15] alice: 안녕하세요! Alice입니다.
[bob] > 반갑습니다 Bob이에요!
✓ [21:30:20] You: 반갑습니다 Bob이에요!

# Charlie 터미널
📩 [21:30:15] alice: 안녕하세요! Alice입니다.
📩 [21:30:20] bob: 반갑습니다 Bob이에요!
[charlie] > 안녕하세요 모두들~ Charlie입니다
✓ [21:30:25] You: 안녕하세요 모두들~ Charlie입니다
```

#### 시나리오 2: 부하 테스트

```bash
# 여러 클라이언트 동시 실행 스크립트
#!/bin/bash
for i in {1..10}; do
    make client USER=user$i &
done
wait
```

#### 시나리오 3: 장애 복구 테스트

```bash
# 1. 정상 채팅 중
[alice] > 테스트 메시지 1

# 2. Kafka 브로커 하나 중단
docker stop kafka-02

# 3. 서비스 계속 작동 확인 (1-3초 지연 후 정상화)
[alice] > 테스트 메시지 2  # 정상 전송됨

# 4. 브로커 복구
docker start kafka-02

# 5. 자동 재연결 및 정상화 확인
[alice] > 테스트 메시지 3  # 정상 전송됨
```

### 3. 디버깅 및 문제 해결

#### 로그 분석 가이드

```bash
# 서버 로그 패턴
✅ 정상: "Message sent to Kafka - Topic: chatting, Partition: 1, Offset: 127"
❌ 에러: "Failed to send message to Kafka: connection refused"

# Consumer 로그 패턴
✅ 정상: "Message successfully forwarded to local clients"
❌ 에러: "No pool reference available, message not forwarded"

# 클라이언트 로그 패턴
✅ 정상: "Connected to server as user: alice"
❌ 에러: "Failed to connect to server: connection refused"
```

#### 일반적인 문제 해결

```bash
# 1. Kafka 연결 실패
# 원인: Kafka 클러스터 미시작
# 해결: make kafka-up

# 2. 클라이언트 연결 실패
# 원인: gRPC 서버 미시작
# 해결: make server

# 3. 메시지 전달 안됨
# 원인: Consumer 별도 실행
# 해결: Consumer 중단, 서버만 실행

# 4. 포트 충돌
# 원인: 이미 사용 중인 포트
# 해결: lsof -ti:8081 | xargs kill -9
```

## 📈 성능 벤치마크 및 한계

### 1. 예상 성능 지표

```
단일 서버 인스턴스 기준:
├── 동시 연결: ~1,000 클라이언트
├── 메시지 처리량: ~10,000 msg/sec
├── 응답 지연시간: ~10ms (P95)
└── 메모리 사용량: ~100MB (1000 연결)

Kafka 클러스터 기준:
├── 처리량: ~50,000 msg/sec (3 브로커)
├── 저장 용량: 제한 없음 (설정에 따라)
├── 복제 지연: ~1ms (동일 데이터센터)
└── 장애 복구: ~1-3초
```

### 2. 확장 한계 및 개선점

#### 현재 구현의 한계

1. **단일 서버 병목**: 연결 풀이 단일 서버에 집중
2. **메모리 증가**: 연결 수에 비례한 메모리 사용
3. **CPU 사용량**: 브로드캐스팅 시 모든 연결에 순차 전송

#### 개선 방안

1. **분산 아키텍처**: 서버간 gRPC 통신 구현
2. **연결 풀 샤딩**: 연결을 여러 풀로 분산
3. **비동기 I/O**: 논블로킹 메시지 전송
4. **캐싱 레이어**: Redis를 활용한 세션 관리

## 🎯 결론 및 핵심 가치

이 실시간 채팅 시스템은 **Go의 동시성 프로그래밍 강점**과 **Kafka의 분산 메시지 처리 능력**을 결합한 확장 가능한 마이크로서비스 아키텍처의 완벽한 예시입니다.

### 핵심 학습 포인트

1. **gRPC 스트리밍**: 실시간 양방향 통신 구현
2. **Kafka 분산 처리**: 고가용성 메시지 브로커 활용
3. **Go 동시성**: Goroutine과 Channel을 활용한 병렬 처리
4. **마이크로서비스**: 컴포넌트 분리 및 독립적 확장
5. **장애 복구**: 자동 장애 감지 및 복구 메커니즘

### 실무 적용 가능성

- **실시간 협업 도구**: Slack, Discord 등
- **게임 채팅 시스템**: 멀티플레이어 게임
- **IoT 메시지 처리**: 센서 데이터 실시간 수집
- **금융 시스템**: 실시간 거래 알림
- **모니터링 시스템**: 실시간 로그 수집

이 프로젝트를 통해 **확장성, 고가용성, 실시간성**을 모두 만족하는 분산 시스템 설계의 핵심 원리를 학습할 수 있습니다. 🚀
