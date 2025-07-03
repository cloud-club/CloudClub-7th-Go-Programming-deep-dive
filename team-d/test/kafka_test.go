package test

import (
	"testing"

	"gRPC-based-chatting/chatProto"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"google.golang.org/protobuf/proto"
)

func TestKafkaProtobufProduce(t *testing.T) {
	// 1. Kafka 프로듀서 생성
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:10000", // 로컬호스트 주소로
	})
	if err != nil {
		t.Fatalf("Kafka 프로듀서 생성 실패: %v", err)
	}
	defer producer.Close()

	// 2. 메시지 생성 (chatProto.ChatMessage)
	msg := &chatProto.ChatMessage{
		Sender:  "userA",
		Channel: "room1",
		Content: "테스트 메시지입니다!",
	}

	// 3. Protobuf 직렬화
	data, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("Protobuf 직렬화 실패: %v", err)
	}

	// 4. Kafka에 메시지 전송
	topic := "chat-topic"
	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte("room1"),
		Value:          data,
	}, nil)
	if err != nil {
		t.Fatalf("Kafka 메시지 전송 실패: %v", err)
	}

	// 전송 에러 확인
	e := <-producer.Events()
	m, ok := e.(*kafka.Message)
	if ok {
		if m.TopicPartition.Error != nil {
			t.Fatalf("Kafka 메시지 실제 전송 실패: %v", m.TopicPartition.Error)
		} else {
			t.Logf("Kafka 메시지 실제 전송 성공: %v", m.TopicPartition)
		}
	} else {
		t.Fatalf("Delivery report 타입 변환 실패: %v", e)
	}

	// 6. 전송 완료 대기 (추가 안전장치)
	producer.Flush(5000)

	// 7. 결과 로그 출력
	t.Logf("Kafka에 Protobuf 메시지 전송 성공!")
}
