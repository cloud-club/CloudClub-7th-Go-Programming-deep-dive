package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Metric은 단일 메트릭의 이름과 값을 나타냅니다
type Metric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// Scraper는 대상 엔드포인트에서 메트릭을 수집하는 핸들러입니다
type Scraper struct {
	client    *http.Client
	storage   *Storage
	logger    *logrus.Logger
	targetURL string
	interval  time.Duration
	mu        sync.RWMutex
}

// Storage는 수집된 메트릭을 메모리에 저장합니다
type Storage struct {
	metrics map[string][]Metric
	mu      sync.RWMutex
}

// NewScraper는 새로운 스크래퍼 인스턴스를 생성합니다
func NewScraper(targetURL string, interval time.Duration) *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		storage: &Storage{
			metrics: make(map[string][]Metric),
		},
		logger:    logrus.New(),
		targetURL: targetURL,
		interval:  interval,
	}
}

// Start는 스크래핑 프로세스를 시작합니다
func (s *Scraper) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.scrape(); err != nil {
				s.logger.Errorf("Failed to scrape metrics: %v", err)
			}
		}
	}
}

// scrape는 단일 스크래핑 작업을 수행합니다
func (s *Scraper) scrape() error {
	resp, err := s.client.Get(s.targetURL)
	if err != nil {
		return fmt.Errorf("failed to fetch metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var metrics []Metric
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return fmt.Errorf("failed to decode metrics: %w", err)
	}

	s.storeMetrics(metrics)
	return nil
}

// storeMetrics는 수집된 메트릭을 메모리에 저장합니다
func (s *Scraper) storeMetrics(metrics []Metric) {
	s.storage.mu.Lock()
	defer s.storage.mu.Unlock()

	for _, metric := range metrics {
		s.storage.metrics[metric.Name] = append(s.storage.metrics[metric.Name], metric)
		// 각 메트릭별로 최근 100개의 데이터 포인트만 유지합니다
		if len(s.storage.metrics[metric.Name]) > 100 {
			s.storage.metrics[metric.Name] = s.storage.metrics[metric.Name][1:]
		}
	}
}

// GetMetrics는 저장된 모든 메트릭을 반환합니다
func (s *Scraper) GetMetrics() map[string][]Metric {
	s.storage.mu.RLock()
	defer s.storage.mu.RUnlock()

	// 메트릭 맵의 복사본을 생성합니다
	metrics := make(map[string][]Metric)
	for k, v := range s.storage.metrics {
		metrics[k] = append([]Metric{}, v...)
	}
	return metrics
}