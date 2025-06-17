// Package models는 애플리케이션 전체에서 사용되는 핵심 데이터 구조를 정의합니다.
package models

import (
	"time"
)

// Metric은 단일 메트릭 데이터 포인트를 나타냅니다
type Metric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels"`
	Timestamp time.Time         `json:"timestamp"`
}

// MetricFamily는 관련된 메트릭들의 그룹을 나타냅니다
type MetricFamily struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Help    string   `json:"help"`
	Metrics []Metric `json:"metrics"`
}

// QueryResult는 메트릭 쿼리의 결과를 나타냅니다
type QueryResult struct {
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}
