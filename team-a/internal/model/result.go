package model

import "time"

type Result struct {
	UserID     int       `json:"user_id"`
	StatusCode int       `json:"status_code"`
	Duration   int64     `json:"duration_ms"` // ms 단위
	Timestamp  time.Time `json:"timestamp"`
	Error      string    `json:"error,omitempty"`
}
