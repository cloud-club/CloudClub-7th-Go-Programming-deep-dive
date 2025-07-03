package kafka

import (
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type ChatProducer struct {
	Producer *kafka.Producer
	// Event chan kafka.Event
	Topic string
	done  chan struct{}
}

// Producer 생성 및 이벤트 핸들러 고루틴 시작
func NewChatProducer(broker string, topic string) (*ChatProducer, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers":         broker,
		"acks":                      "all",
		"enable.idempotence":        true,
		"retries":                   5,
		"go.delivery.reports":       true,
		"go.delivery.report.fields": "key,value",
	}

	producer, err := kafka.NewProducer(config)
	if err != nil {
		log.Printf("카프카 프로듀서 생성 실패", err)
		return nil, err
	}

	cp := &ChatProducer{
		Producer: producer,
		// Event: make(chan kafka.event, 100),
		Topic: topic,
		done:  make(chan struct{}),
	}

	go cp.handleEvents()

	return cp, nil

}

// 이벤트 핸들러
func (cp *ChatProducer) handleEvents() {
	for {
		select {
		case e := <-cp.Producer.Events():
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("전송 실패: %v", ev.TopicPartition.Error)
				} else {
					log.Printf("전송 성공: %v", ev.TopicPartition)
				}
			}
		case <-cp.done:
			return
		}
	}
}

// 메시지 전송 함수
func (cp *ChatProducer) SendAsyncMessage(key string, message []byte) error {
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &cp.Topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(key),
		Value: message,
	}

	err := cp.Producer.Produce(msg, nil)
	if err != nil {
		log.Printf("메시지 전송 실패: %v", err)
	}
	return err
}

// producer 종료 처리
func (cp *ChatProducer) Close() {
	close(cp.done)               // 이벤트 핸들러 고루틴 종료
	cp.Producer.Flush(15 * 1000) // 15초 대기
	cp.Producer.Close()
}
