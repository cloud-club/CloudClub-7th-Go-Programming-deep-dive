// Package main은 Tiny Prometheus 애플리케이션의 진입점입니다.
// 스크래퍼, 스토리지, API 서버를 초기화하고 조정합니다.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cloudclub-7th/tiny-prometheus/pkg/api"
	"github.com/cloudclub-7th/tiny-prometheus/pkg/config"
	"github.com/cloudclub-7th/tiny-prometheus/pkg/scraper"
	"github.com/cloudclub-7th/tiny-prometheus/pkg/storage"
)

func main() {
	// 설정 파일 로드
	configPath := "config.yaml"
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Printf("설정 파일 로드 실패: 기본 설정을 사용합니다. (%v)", err)
		cfg = config.DefaultConfig()
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("설정값이 유효하지 않습니다: %v", err)
	}

	// 스토리지 초기화
	storage := storage.NewStorage()

	// 스크래퍼 초기화 및 시작 (첫 번째 타겟만 사용)
	target := cfg.Targets[0]
	scraper := scraper.NewScraper(target, storage)
	if err := scraper.Start(); err != nil {
		log.Fatalf("스크래퍼 시작 실패: %v", err)
	}

	// API 서버 초기화 및 시작
	server := api.NewServer(storage)
	go func() {
		if err := server.Start(cfg.GetAPIAddress()); err != nil {
			log.Fatalf("API 서버 시작 실패: %v", err)
		}
	}()

	// 종료 시그널 대기
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 정상 종료
	log.Println("애플리케이션 종료 중...")
}
