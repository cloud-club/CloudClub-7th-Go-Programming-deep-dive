// Package config는 Tiny Prometheus의 설정 관리를 담당합니다.
// YAML 형식의 설정 파일을 로드하고 애플리케이션 전체에서 사용할 수 있도록 합니다.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config는 애플리케이션의 전체 설정을 나타냅니다
type Config struct {
	Scrape  ScrapeConfig  `yaml:"scrape"`
	Storage StorageConfig `yaml:"storage"`
	API     APIConfig     `yaml:"api"`
	Targets []string      `yaml:"targets"`
}

// ScrapeConfig는 스크래핑 관련 설정을 나타냅니다
type ScrapeConfig struct {
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
}

// StorageConfig는 스토리지 관련 설정을 나타냅니다
type StorageConfig struct {
	Type      string        `yaml:"type"`
	Retention time.Duration `yaml:"retention"`
}

// APIConfig는 API 서버 관련 설정을 나타냅니다
type APIConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

// DefaultConfig는 기본 설정값을 반환합니다
func DefaultConfig() *Config {
	return &Config{
		Scrape: ScrapeConfig{
			Interval: 15 * time.Second,
			Timeout:  10 * time.Second,
		},
		Storage: StorageConfig{
			Type:      "memory",
			Retention: 24 * time.Hour,
		},
		API: APIConfig{
			Port: "9090",
			Host: "0.0.0.0",
		},
		Targets: []string{"http://localhost:8080/metrics"},
	}
}

// LoadConfig는 지정된 파일 경로에서 설정을 로드합니다
func LoadConfig(configPath string) (*Config, error) {
	// 기본 설정으로 시작
	config := DefaultConfig()

	// 설정 파일이 존재하는 경우 로드
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("설정 파일 읽기 실패: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("설정 파일 파싱 실패: %w", err)
		}
	}

	return config, nil
}

// SaveConfig는 현재 설정을 파일에 저장합니다
func (c *Config) SaveConfig(configPath string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("설정 마샬링 실패: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("설정 파일 저장 실패: %w", err)
	}

	return nil
}

// Validate는 설정의 유효성을 검사합니다
func (c *Config) Validate() error {
	if c.Scrape.Interval <= 0 {
		return fmt.Errorf("스크래프 간격은 0보다 커야 합니다")
	}

	if c.Scrape.Timeout <= 0 {
		return fmt.Errorf("스크래프 타임아웃은 0보다 커야 합니다")
	}

	if c.Storage.Retention <= 0 {
		return fmt.Errorf("보관 기간은 0보다 커야 합니다")
	}

	if len(c.Targets) == 0 {
		return fmt.Errorf("최소 하나의 타겟이 필요합니다")
	}

	return nil
}

// GetAPIAddress는 API 서버의 전체 주소를 반환합니다
func (c *Config) GetAPIAddress() string {
	return fmt.Sprintf("%s:%s", c.API.Host, c.API.Port)
}
