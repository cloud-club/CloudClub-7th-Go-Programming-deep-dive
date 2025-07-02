package kafka

import (
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type ChatConsumer struct {
	Consumer *kafka.Consumer
	Topic    string
}

func NewChatConsumer(broker, topic, groupID string) (*ChatConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		"group.id":          groupID,
		"auto.offset.reset": "latest",
	})
	if err != nil {
		return nil, err
	}
	if err := c.SubscribeTopics([]string{topic}, nil); err != nil {
		return nil, err
	}
	return &ChatConsumer{Consumer: c, Topic: topic}, nil
}

// 메시지 처리 콜백을 받아서 메시지를 브로드캐스트
func (cc *ChatConsumer) ConsumeLoop(handleMessage func(key, value []byte)) {
	for {
		ev := cc.Consumer.Poll(100)
		if ev == nil {
			continue
		}
		switch e := ev.(type) {
		case *kafka.Message:
			handleMessage(e.Key, e.Value)
		case kafka.Error:
			log.Printf("Kafka error: %v", e)
		}
	}
}
