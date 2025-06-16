// Package storage는 메트릭 스토리지 기능을 구현합니다.
// 시계열 데이터를 위한 메모리 기반 스토리지 시스템을 제공합니다.
package storage

import (
	"sync"
	"time"

	"github.com/cloudclub-7th/tiny-prometheus/internal/models"
)

// Storage는 메트릭 스토리지 시스템을 나타냅니다
type Storage struct {
	metrics map[string][]models.Metric
	mu      sync.RWMutex
}

// NewStorage는 새로운 스토리지 인스턴스를 생성합니다
func NewStorage() *Storage {
	return &Storage{
		metrics: make(map[string][]models.Metric),
	}
}

// StoreMetric은 메트릭 값을 저장합니다
func (s *Storage) StoreMetric(name string, value float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	metric := models.Metric{
		Name:      name,
		Value:     value,
		Timestamp: time.Now(),
	}

	s.metrics[name] = append(s.metrics[name], metric)
	return nil
}

// QueryMetric은 이름을 기반으로 메트릭 값을 조회합니다
func (s *Storage) QueryMetric(name string) ([]models.Metric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if metrics, ok := s.metrics[name]; ok {
		return metrics, nil
	}
	return nil, nil
}
