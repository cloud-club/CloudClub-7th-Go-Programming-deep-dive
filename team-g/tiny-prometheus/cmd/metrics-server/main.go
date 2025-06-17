// Package main은 테스트용 메트릭 서버를 구현합니다.
// 이 서버는 /metrics 엔드포인트를 통해 샘플 메트릭 데이터를 제공합니다.
package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	// /metrics 엔드포인트 핸들러 등록
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// 샘플 메트릭 데이터 정의
		metrics := []struct {
			Name  string  `json:"name"`  // 메트릭 이름
			Value float64 `json:"value"` // 메트릭 값
		}{
			{"cpu_usage", 45.5},    // CPU 사용률
			{"memory_usage", 60.2}, // 메모리 사용률
			{"disk_usage", 75.8},   // 디스크 사용률
		}
		
		// JSON 응답 헤더 설정
		w.Header().Set("Content-Type", "application/json")
		
		// 메트릭 데이터를 JSON으로 인코딩하여 응답
		if err := json.NewEncoder(w).Encode(metrics); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// 서버 시작 로그 출력
	log.Println("Starting metrics server on :8080")
	
	// HTTP 서버 시작
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}