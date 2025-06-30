### Fast Run
___
```
- 서버 기동 (Terminal #1)
# go run cmd/server/main.go

- Simulator 기동 (Terminal #2)
# go run cmd/simulator/main.go 
📡 gRPC 부하 시뮬레이터 CLI 시작
📚 사용 가능한 명령:
  spawn <숫자> : 지정된 수만큼 클라이언트를 생성
  broadcast <메시지> <숫자> : 모든 클라이언트가 메시지 전송
  summary    : 현재 연결된 클라이언트 요약 출력
  closeAll   : 현재 연결된 클라이언트 종료
  exit       : 시뮬레이터 종료
sim>  spawn 100
sim> broadcast test-message 10

- go resources check (Terminal #3)
# go run cmd/metric/fetch_go_metrics.go 3.36.108.146 2112
📡 Prometheus 메트릭 조회 중...
go_gc_duration_seconds_sum               0          // GC 총 소요 시간 (초)
go_gc_duration_seconds_count             2          // GC 발생 횟수
go_goroutines                            408        // 현재 실행 중인 고루틴 수
go_memstats_alloc_bytes                  10.52 MB   // 현재 할당된 메모리 바이트 수
go_memstats_heap_alloc_bytes             10.52 MB   // 힙에 할당된 바이트 수
go_memstats_heap_inuse_bytes             11.77 MB   // 사용 중인 힙 메모리
go_memstats_next_gc_bytes                12.74 MB   // 다음 GC 발생까지 남은 바이트 수
go_memstats_stack_inuse_bytes            2.38 MB    // 스택에 사용 중인 바이트 수
go_memstats_sys_bytes                    19.21 MB   // Go가 OS에서 요청한 전체 메모리
```
___

### 디렉토리 구조
```
grpc-chat/
├── go.mod  # go Module 선언

├── cmd/  # build 대상
│   └── server/
│       └── main.go
│   └── client/
│       └── main.go
│   └── metric/
│       └── main.go
│   └── simulator/
│       └── main.go

├── internal/  # 기능적 로직 관리
│   ├── domain/    #Domain: 유저와 메시지를 정의
│   │   └── model.go
│   ├── usecase/    #Usecase: 비즈니스 로직 분리
│   │   └── chat_usecase.go  # client 등록 및 broadcast 등의 기능
│   ├── port/    #Port: ChatService 인터페이스와 세션 저장소 정의
│   │   ├── in/
│   │   │   └── chat_service.go
│   │   └── out/
│   │       └── session_repo.go
│   └── adapter/    #Adapter: gRPC 프레임워크와 Core Usecase 연결
│       └── grpc/
│           └── handler.go
│           └── command_handler.go
│           └── command_parser.go
│   └── simulator/    #Simulator: 부하 기능 정의
│       └── runner.go 

├── infrastructure/    #Infrastructure: 메모리 저장소 구현
│   └── memory/    
│       └── session_repository.go

├── proto/  # proto buffer 정의
│   └── chat.proto
├── gen/
│   └── chat.pb.go            # protoc-gen-go 생성
│   └── chat_grpc.pb.go       # protoc-gen-go-grpc 생성
```


### Grafana

```
# 테스트 시나리오
spawn 100씩 증가
test-message 100회 씩 전송
```

<img width="1343" alt="image" src="https://github.com/user-attachments/assets/88f67a5a-157f-4217-9bd2-f1a84a60e574" />
