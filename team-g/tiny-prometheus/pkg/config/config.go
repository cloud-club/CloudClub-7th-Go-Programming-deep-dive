// Package config는 Tiny Prometheus의 설정 관리를 구현합니다.
// YAML 파일에서 설정을 로드하고 검증하는 기능을 제공합니다.
package config

import (
	"time"
)

// Config는 애플리케이션 설정을 나타냅니다
type Config struct {
	// 스크래퍼 설정
	ScrapeInterval time.Duration `yaml:"scrape_interval"`
	Targets        []string      `yaml:"targets"`

	// 스토리지 설정
	RetentionPeriod time.Duration `yaml:"retention_period"`
	MaxDataPoints   int           `yaml:"max_data_points"`

	// API 설정
	ListenAddr string `yaml:"listen_addr"`
}

// LoadConfig는 파일에서 설정을 로드합니다
func LoadConfig(path string) (*Config, error) {
	// TODO: 설정 로드 구현
	return &Config{}, nil
}

// Validate는 설정이 유효한지 검사합니다
func (c *Config) Validate() error {
	// TODO: 설정 검증 구현
	return nil
}
