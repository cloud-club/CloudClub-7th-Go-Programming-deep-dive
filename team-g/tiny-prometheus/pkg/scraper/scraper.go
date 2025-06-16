// Package scraper는 메트릭 수집 기능을 구현합니다.
// 설정된 타겟으로부터 주기적으로 메트릭을 수집하여
// 스토리지 시스템에 저장합니다.
package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudclub-7th/tiny-prometheus/internal/models"
	"github.com/cloudclub-7th/tiny-prometheus/pkg/storage"
)

// Scraper는 타겟으로부터 데이터를 수집하는 메트릭 스크래퍼를 나타냅니다
type Scraper struct {
	storage *storage.Storage
	client  *http.Client
	target  string
}

// NewScraper는 새로운 스크래퍼 인스턴스를 생성합니다
func NewScraper(target string, storage *storage.Storage) *Scraper {
	return &Scraper{
		storage: storage,
		client:  &http.Client{Timeout: 10 * time.Second},
		target:  target,
	}
}

// Start는 스크래핑 프로세스를 시작합니다
func (s *Scraper) Start() error {
	go s.scrape()
	return nil
}

// Stop은 스크래퍼를 정상적으로 중지합니다
func (s *Scraper) Stop() error {
	return nil
}

func (s *Scraper) scrape() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		resp, err := s.client.Get(s.target + "/metrics")
		if err != nil {
			fmt.Printf("스크래핑 실패: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		var metrics []models.Metric
		if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
			fmt.Printf("메트릭 파싱 실패: %v\n", err)
			continue
		}

		for _, metric := range metrics {
			if err := s.storage.StoreMetric(metric.Name, metric.Value); err != nil {
				fmt.Printf("메트릭 저장 실패: %v\n", err)
			}
		}
	}
}
