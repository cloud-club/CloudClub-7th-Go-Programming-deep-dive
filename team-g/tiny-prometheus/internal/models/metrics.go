// Package models는 애플리케이션 전체에서 사용되는 핵심 데이터 구조를 정의합니다.
package models

import (
	"time"
)

// Metric은 단일 메트릭 데이터 포인트를 나타냅니다
type Metric struct {
	Name      string    `json:"name"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

// QueryResult는 메트릭 쿼리의 결과를 나타냅니다
type QueryResult struct {
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}
