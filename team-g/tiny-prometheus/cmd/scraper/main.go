// Package main은 메트릭 스크래퍼의 진입점입니다.
// 이 프로그램은 지정된 엔드포인트에서 메트릭을 수집하고 저장합니다.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudclub-7th/tiny-prometheus/internal/scraper"
)

func main() {
	// 새로운 스크래퍼 인스턴스 생성
	// - 타겟 URL: http://localhost:8080/metrics
	// - 수집 주기: 15초
	s := scraper.NewScraper("http://localhost:8080/metrics", 15*time.Second)

	// SIGINT(Ctrl+C) 또는 SIGTERM 시그널을 받으면 취소되는 컨텍스트 생성
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 스크래퍼 시작 (고루틴으로 실행)
	go s.Start(ctx)

	log.Println("스크래퍼가 시작되었습니다. 중지하려면 Ctrl+C를 누르세요.")

	// 컨텍스트 취소 대기 (프로그램 종료 시그널 대기)
	<-ctx.Done()
	log.Println("스크래퍼를 종료합니다...")
}