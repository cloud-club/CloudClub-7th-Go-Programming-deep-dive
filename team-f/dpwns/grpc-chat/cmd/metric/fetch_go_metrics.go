package main

import (
    "crypto/tls"
    "fmt"
    "io"
    "net/http"
    "os"
    "strings"
    "strconv"
)

var metricDescriptions = map[string]string{
    "go_goroutines":                 "현재 실행 중인 고루틴 수",
    "go_gc_duration_seconds_sum":   "GC 총 소요 시간 (초)",
    "go_gc_duration_seconds_count": "GC 발생 횟수",
    "go_memstats_alloc_bytes":      "현재 할당된 메모리 바이트 수",
    "go_memstats_heap_alloc_bytes": "힙에 할당된 바이트 수",
    "go_memstats_heap_inuse_bytes": "사용 중인 힙 메모리",
    "go_memstats_stack_inuse_bytes": "스택에 사용 중인 바이트 수",
    "go_memstats_sys_bytes":        "Go가 OS에서 요청한 전체 메모리",
    "go_memstats_next_gc_bytes":    "다음 GC 발생까지 남은 바이트 수",
}

func isTargetMetric(line string) (string, bool) {
    for name := range metricDescriptions {
        if strings.HasPrefix(line, name+" ") { // 공백 필수
            return name, true
        }
    }
    return "", false
}

func formatBytes(bytes float64) string {
    const (
        KB = 1024
        MB = KB * 1024
        GB = MB * 1024
    )

    switch {
    case bytes >= GB:
        return fmt.Sprintf("%.2f GB", bytes/GB)
    case bytes >= MB:
        return fmt.Sprintf("%.2f MB", bytes/MB)
    case bytes >= KB:
        return fmt.Sprintf("%.2f KB", bytes/KB)
    default:
        return fmt.Sprintf("%.0f B", bytes)
    }
}

func main() {
    if len(os.Args) < 3 {
        fmt.Println("❗ 사용법: go run fetch.go <Prometheus_IP> <Port>")
        os.Exit(1)
    }

    ip := os.Args[1]
    port := os.Args[2]
    prometheusURL := fmt.Sprintf("http://%s:%s/metrics", ip, port)

    http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

    fmt.Println("📡 Prometheus 메트릭 조회 중...\n")

    resp, err := http.Get(prometheusURL)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ 요청 실패: %v\n", err)
        os.Exit(1)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        fmt.Fprintf(os.Stderr, "❌ 상태코드 오류: %d\n", resp.StatusCode)
        os.Exit(1)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌ 응답 읽기 실패: %v\n", err)
        os.Exit(1)
    }

    lines := strings.Split(string(body), "\n")
    for _, line := range lines {
        if metricName, ok := isTargetMetric(line); ok {
            fields := strings.Fields(line)
	    if len(fields) >= 2 {
		    valueStr := fields[1]
		    value, err := strconv.ParseFloat(valueStr, 64)
		    if err == nil {
            		var formatted string
            		if strings.Contains(metricName, "_bytes") {
                		formatted = formatBytes(value)
            		} else {
                		formatted = fmt.Sprintf("%.0f", value)
            		}
            		desc := metricDescriptions[metricName]
            		fmt.Printf("%-40s %-10s // %s\n", metricName, formatted, desc)
		}
	    }
	}
    }
}

