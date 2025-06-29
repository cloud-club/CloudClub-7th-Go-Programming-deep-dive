package config

import (
	"fmt"
	"os"
	"time"
)

// KafkaConfig: Kafka 설정
type KafkaConfig struct {
	Brokers       []string
	Topic         string
	ConsumerGroup string
}

// Config: 애플리케이션 전체 설정
type Config struct {
	ServerID    string
	GRPCPort    string
	HTTPPort    string
	WebRoot     string
	Kafka       *KafkaConfig
	Environment string
}

// NewConfig: 새로운 설정 인스턴스 생성
func NewConfig() *Config {
	return &Config{
		ServerID:    getEnvWithDefault("SERVER_ID", fmt.Sprintf("cloudclub-chat-server-%d", time.Now().Unix())),
		GRPCPort:    getEnvWithDefault("GRPC_PORT", ":8082"),
		HTTPPort:    getEnvWithDefault("HTTP_PORT", ":8080"),
		WebRoot:     getEnvWithDefault("WEB_ROOT", "./web"),
		Environment: getEnvWithDefault("ENVIRONMENT", "development"),
		Kafka: &KafkaConfig{
			Brokers: []string{
				getEnvWithDefault("KAFKA_BROKER_1", "localhost:9092"),
				getEnvWithDefault("KAFKA_BROKER_2", "localhost:9093"),
				getEnvWithDefault("KAFKA_BROKER_3", "localhost:9094"),
			},
			Topic:         getEnvWithDefault("KAFKA_TOPIC", "chatting"),
			ConsumerGroup: getEnvWithDefault("KAFKA_CONSUMER_GROUP", "chatting-processor-group"),
		},
	}
}

// IsDevelopment: 개발 환경인지 확인
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction: 프로덕션 환경인지 확인
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// getEnvWithDefault: 환경변수 값을 가져오되, 없으면 기본값 사용
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
