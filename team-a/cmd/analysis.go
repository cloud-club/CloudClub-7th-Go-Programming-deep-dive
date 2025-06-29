package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"sort"
	"swarm/internal/model"
	"time"
)

var inputFile string

// TimeBasedStats 시간 기반 통계를 저장하는 구조체
type TimeBasedStats struct {
	Timestamp    time.Time
	RequestCount int
	SuccessCount int
	FailCount    int
	TotalLatency int64
	MinLatency   int64
	MaxLatency   int64
}

// analyzeTimeBasedStats 시간 기반 통계를 분석하는 함수
func analyzeTimeBasedStats(results []model.Result, interval time.Duration) []TimeBasedStats {
	if len(results) == 0 {
		return nil
	}

	// 시작 시간과 종료 시간 찾기
	startTime := results[0].Timestamp
	endTime := results[len(results)-1].Timestamp

	// 시간 간격으로 버킷 생성
	buckets := make(map[time.Time]*TimeBasedStats)
	current := startTime.Truncate(interval)
	for current.Before(endTime) || current.Equal(endTime) {
		buckets[current] = &TimeBasedStats{
			Timestamp:    current,
			MinLatency:   -1, // 초기값 설정
			MaxLatency:   -1,
		}
		current = current.Add(interval)
	}

	// 결과를 버킷에 분류
	for _, result := range results {
		bucketTime := result.Timestamp.Truncate(interval)
		stats := buckets[bucketTime]
		if stats == nil {
			continue
		}

		stats.RequestCount++
		if result.Error == "" && result.StatusCode < 400 {
			stats.SuccessCount++
		} else {
			stats.FailCount++
		}

		stats.TotalLatency += result.Duration
		if stats.MinLatency == -1 || result.Duration < stats.MinLatency {
			stats.MinLatency = result.Duration
		}
		if stats.MaxLatency == -1 || result.Duration > stats.MaxLatency {
			stats.MaxLatency = result.Duration
		}
	}

	// 결과를 시간순으로 정렬
	var sortedStats []TimeBasedStats
	for _, stats := range buckets {
		sortedStats = append(sortedStats, *stats)
	}
	sort.Slice(sortedStats, func(i, j int) bool {
		return sortedStats[i].Timestamp.Before(sortedStats[j].Timestamp)
	})

	return sortedStats
}

// printTimeBasedAnalysis 시간 기반 분석 결과를 출력하는 함수
func printTimeBasedAnalysis(stats []TimeBasedStats) {
	if len(stats) == 0 {
		return
	}

	fmt.Println("\n📈 Time-based Analysis")
	fmt.Println("────────────────────────────")

	// 전체 통계
	var totalRequests, totalSuccess, totalFail int
	var totalLatency int64
	for _, stat := range stats {
		totalRequests += stat.RequestCount
		totalSuccess += stat.SuccessCount
		totalFail += stat.FailCount
		totalLatency += stat.TotalLatency
	}

	fmt.Printf("Total Duration: %s\n", stats[len(stats)-1].Timestamp.Sub(stats[0].Timestamp))
	fmt.Printf("Average RPS: %.2f\n", float64(totalRequests)/stats[len(stats)-1].Timestamp.Sub(stats[0].Timestamp).Seconds())

	// 시간대별 상세 통계
	fmt.Println("\nTime-based Statistics:")
	fmt.Println("Timestamp\t\tRequests\tSuccess\tFail\tAvg Latency\tMin Latency\tMax Latency")
	fmt.Println("────────────────────────────────────────────────────────────────────────────────────────────")
	
	for _, stat := range stats {
		avgLatency := int64(0)
		if stat.RequestCount > 0 {
			avgLatency = stat.TotalLatency / int64(stat.RequestCount)
		}
		fmt.Printf("%s\t%d\t%d\t%d\t%d ms\t%d ms\t%d ms\n",
			stat.Timestamp.Format("15:04:05"),
			stat.RequestCount,
			stat.SuccessCount,
			stat.FailCount,
			avgLatency,
			stat.MinLatency,
			stat.MaxLatency,
		)
	}

	// RPS 추이 그래프
	fmt.Println("\nRPS Trend:")
	fmt.Println("────────────────────────────")
	maxRPS := 0
	for _, stat := range stats {
		if stat.RequestCount > maxRPS {
			maxRPS = stat.RequestCount
		}
	}

	// ASCII 그래프 생성 (최대 50자 길이)
	for _, stat := range stats {
		barLength := int(float64(stat.RequestCount) / float64(maxRPS) * 50)
		bar := ""
		for i := 0; i < barLength; i++ {
			bar += "█"
		}
		fmt.Printf("%s | %s %d\n", stat.Timestamp.Format("15:04:05"), bar, stat.RequestCount)
	}
}

var analysisCmd = &cobra.Command{
	Use:   "analysis",
	Short: "analysis command",
	Run: func(cmd *cobra.Command, args []string) {
		f, err := os.Open(inputFile)
		if err != nil {
			log.Fatal(err)
			return
		}

		defer f.Close()

		var results []model.Result
		if err := json.NewDecoder(f).Decode(&results); err != nil {
			fmt.Println("Failed to parse JSON:", err)
			return
		}

		total := len(results)
		var success, fail int
		var totalDuration int64
		var durations []int64

		for _, result := range results {
			if result.Error != "" || result.StatusCode >= 400 {
				fail++
			} else {
				success++
			}
			durations = append(durations, result.Duration)
			totalDuration += result.Duration
		}

		sort.Slice(durations, func(i int, j int) bool {
			return durations[i] < durations[j]
		})

		// 익명함수
		percentile := func(p float64) int64 {
			if len(durations) == 0 {
				return 0
			}
			rank := int(float64(len(durations))*p + 0.5)
			if rank >= len(durations) {
				rank = len(durations) - 1
			}
			return durations[rank]
		}
		fmt.Println("📊 Load Test Analysis")
		fmt.Println("────────────────────────────")
		fmt.Printf("Total Requests:        %d\n", total)
		fmt.Printf("Successful (2xx):      %d (%.2f%%)\n", success, float64(success)/float64(total)*100)
		fmt.Printf("Failed (4xx/5xx/errors): %d (%.2f%%)\n", fail, float64(fail)/float64(total)*100)

		if total > 0 {
			fmt.Printf("\nAvg Response Time:     %d ms\n", totalDuration/int64(total))
			fmt.Printf("Min Response Time:     %d ms\n", durations[0])
			fmt.Printf("Max Response Time:     %d ms\n", durations[len(durations)-1])
			fmt.Printf("P90 Response Time:     %d ms\n", percentile(0.90))
			fmt.Printf("P95 Response Time:     %d ms\n", percentile(0.95))
			fmt.Printf("P99 Response Time:     %d ms\n", percentile(0.99))
		}

		// 시간 기반 분석 추가
		timeStats := analyzeTimeBasedStats(results, time.Second) // 1초 간격으로 분석
		printTimeBasedAnalysis(timeStats)
	},
}

func init() {
	analysisCmd.Flags().StringVarP(&inputFile, "input", "i", "results.json", "Path to JSON results file")
	rootCmd.AddCommand(analysisCmd)
}
