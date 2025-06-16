// Package api는 Tiny Prometheus의 HTTP API 서버를 구현합니다.
// 메트릭 쿼리를 위한 엔드포인트를 제공합니다.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/cloudclub-7th/tiny-prometheus/pkg/storage"
)

// Server는 API 서버를 나타냅니다
type Server struct {
	storage *storage.Storage
}

// NewServer는 새로운 API 서버 인스턴스를 생성합니다
func NewServer(storage *storage.Storage) *Server {
	return &Server{
		storage: storage,
	}
}

// Start는 API 서버를 시작합니다
func (s *Server) Start(addr string) error {
	http.HandleFunc("/query", s.handleQuery)
	return http.ListenAndServe(addr, nil)
}

// Stop은 API 서버를 정상적으로 중지합니다
func (s *Server) Stop() error {
	return nil
}

// handleQuery는 메트릭 쿼리 요청을 처리합니다
func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Missing metric name", http.StatusBadRequest)
		return
	}

	metrics, err := s.storage.QueryMetric(name)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
