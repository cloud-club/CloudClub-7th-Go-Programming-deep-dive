// Package storage는 메트릭 스토리지 기능을 구현합니다.
// 시계열 데이터를 위한 메모리 기반 스토리지 시스템을 제공합니다.
package storage

import (
	"time"
)

// Storage는 메트릭 스토리지 시스템을 나타냅니다
type Storage struct {
	// TODO: 스토리지 필드 추가
	// TODO: 동시 접근을 위한 뮤텍스 추가
}

// NewStorage는 새로운 스토리지 인스턴스를 생성합니다
func NewStorage() *Storage {
	return &Storage{}
}

// StoreMetric은 메타데이터와 함께 메트릭 값을 저장합니다
func (s *Storage) StoreMetric(name string, value float64, labels map[string]string, timestamp time.Time) error {
	// TODO: 메트릭 저장 구현
	return nil
}

// QueryMetric은 이름과 선택적 레이블을 기반으로 메트릭 값을 조회합니다
func (s *Storage) QueryMetric(name string, labels map[string]string) ([]float64, error) {
	// TODO: 메트릭 쿼리 구현
	return nil, nil
}
