// Package api는 Tiny Prometheus의 HTTP API 서버를 구현합니다.
// 메트릭 쿼리와 시스템 상태를 위한 엔드포인트를 제공합니다.
package api

import (
	"net/http"
)

// Server는 API 서버를 나타냅니다
type Server struct {
	// TODO: 스토리지 인터페이스 추가
	// TODO: 설정 추가
}

// NewServer는 새로운 API 서버 인스턴스를 생성합니다
func NewServer() *Server {
	return &Server{}
}

// Start는 API 서버를 시작합니다
func (s *Server) Start(addr string) error {
	// TODO: 서버 시작 구현
	return nil
}

// Stop은 API 서버를 정상적으로 중지합니다
func (s *Server) Stop() error {
	// TODO: 서버 중지 구현
	return nil
}

// handleQuery는 메트릭 쿼리 요청을 처리합니다
func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	// TODO: 쿼리 핸들러 구현
}

// handleHealth는 헬스 체크 요청을 처리합니다
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// TODO: 헬스 체크 구현
}
