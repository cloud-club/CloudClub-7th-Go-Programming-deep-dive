## 채팅 방식

- `Enter your name:` → 사용자 ID 등록
- `<메시지 입력>` → 접속 중인 상대방에게 실시간 전송


## 시스템 구조

```
gRPC Server (Go + prometheus metrics)
        │
        └──> exposes /metrics on :2112
                 │
Prometheus ──────┘
        │
        └──> http://localhost:9090 → 지표 수집
                 │
Grafana ─────────┘
        → http://13.124.97.41:3000 → 시각화
```


## 테스트

- go_goroutines: 현재 실행 중인 고루틴 수
- go_memstats_alloc_bytes: 현재 사용 중인 heap 메모리
- process_cpu_seconds_total: 전체 CPU 사용 시간
- go_threads: 생성된 OS-level 쓰레드 수
- process_resident_memory_bytes: 실제 사용 중인 메모리 (RSS)


<img src="image-1.png" alt="gRPC 모니터링 대시보드" width="600"/>
