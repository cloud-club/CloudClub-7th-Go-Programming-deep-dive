### 주제
- gRPC 기반 채팅 시스템 구현
___
### 기술 스택 
- gRPC
___

### 아키텍처

<img width="997" alt="스크린샷 2025-05-31 오후 9 48 41" src="https://github.com/user-attachments/assets/5bca8a04-b373-4110-bbd3-84f16c3a2f6a" />

#### 모니터링 지표 [Go Metrics #10826 ](https://github.com/aaukhatov/grafana-dashboards)

##### 📊 Go 런타임 메트릭 상세 설명
| 메트릭 이름                            | 설명                                                |
| --------------------------------- | ------------------------------------------------- |
| `go_goroutines`                   | 현재 실행 중인 고루틴 수. 고루틴 누수 또는 동시성 병목 감지에 유용           |
| `go_gc_duration_seconds`          | GC(가비지 컬렉션)가 소요한 시간. 히스토그램 형태로 제공되며, 평균/최댓값 확인 가능 |
| `go_memstats_alloc_bytes`         | 현재 애플리케이션이 **할당한 총 메모리 바이트 수** (GC 후에도 유지됨)       |
| `go_memstats_alloc_bytes_total`   | 프로그램 실행 이후 할당된 메모리 총량 (GC 포함한 누적값)                |
| `go_memstats_sys_bytes`           | Go 런타임이 OS로부터 **요청한 전체 메모리 바이트 수** (할당과는 다름)      |
| `go_memstats_heap_alloc_bytes`    | GC 대상인 **힙 메모리에서 사용 중인 바이트 수** (현재 살아있는 객체들이 차지)  |
| `go_memstats_heap_sys_bytes`      | 힙을 위해 **OS로부터 요청한 총 메모리 바이트 수**                   |
| `go_memstats_heap_idle_bytes`     | 힙 중에서 **사용되지 않고 대기 상태인 바이트 수** (OS에 반환 안 됨)       |
| `go_memstats_heap_inuse_bytes`    | 힙 중에서 **현재 actively 사용 중인 바이트 수**                 |
| `go_memstats_stack_inuse_bytes`   | **고루틴 스택에서 현재 사용 중인 메모리 바이트 수**                   |
| `go_memstats_stack_sys_bytes`     | 스택 용도로 **OS로부터 요청한 메모리 총량**                       |
| `go_memstats_mspan_inuse_bytes`   | GC 내부 구조체 중 하나인 **mspan**이 사용하는 메모리               |
| `go_memstats_mspan_sys_bytes`     | mspan을 위해 시스템에 요청한 메모리 양                          |
| `go_memstats_mcache_inuse_bytes`  | mcache (per-P 할당 캐시)가 실제 사용하는 메모리                 |
| `go_memstats_mcache_sys_bytes`    | mcache 용도로 시스템에 요청한 메모리 양                         |
| `go_memstats_buck_hash_sys_bytes` | 맵(Hashmap)에서 사용하는 bucket hash 구조에 할당된 메모리         |
| `go_memstats_gc_sys_bytes`        | **GC 작업을 위해 Go 런타임이 예약한 메모리 총량**                  |

##### 🧠 관찰 포인트 및 이상 징후 탐지 팁

| 지표                              | 유의 상황        | 관찰 팁                 |
| ------------------------------- | ------------ | -------------------- |
| `go_goroutines`                 | 고루틴 누수, 병목   | 지속 증가하면 고루틴 누수 가능성   |
| `go_gc_duration_seconds`        | GC가 자주/오래 걸림 | 평균 시간 증가하면 힙 증가 의심   |
| `go_memstats_heap_alloc_bytes`  | 메모리 사용 증가    | GC 후에도 계속 증가하면 누수 가능 |
| `go_memstats_heap_idle_bytes`   | 사용하지 않는 힙    | 높은 경우 GC tuning 고려   |
| `go_memstats_alloc_bytes_total` | 누적 할당량       | GC 빈도 추정용 보조 지표      |


___

### Goroutine 스케줄러
Go Application에 대한 Observability를 위한 기본 Goroutine 아키텍처 이해

<img width="813" alt="스크린샷 2025-05-31 오후 10 06 41" src="https://github.com/user-attachments/assets/93fec7a2-077e-4633-a62b-be4871d2ab78" />

- G(Goroutine): Goroutine는 말그대로 고루틴 의미하며, 고루틴을 구성하는 논리적 구조체의 구현체
  - Go 런타임이 고루틴을 관리하기 위해서 사용, 컨텍스트 스위칭을 위해 스택 포인터, 고루틴의 상태 정보 등을 가지고 있다
  - G는 LRQ에서 대기
- M(Machine): Machine는 OS 쓰레드를 의미하며, 실제 OS 쓰레드가 아닌 논리적 구현체로 표준 POSIX 쓰레드를 따름
  - M은 P로 부터 G를 할당받아 실행, 고루틴과 OS 쓰레드를 연결하므로 쓰레드 핸들 정보, 실행중인 고루틴, P의 포인터를 가지고 있음
- P(Processor): Processor는 프로세서를 의미하며, 실제 물리적 프로세서를 말하는게 아니라 논리적인 프로세서로 정확히는 스케줄링과 관련된 Context 정보를 가지고 있음
  - 런타임 시 Go 환경변수인 최대 GOMAXPROCS 설정 값만큼의 개수로 프로세서를 가질 수 있다
  - P는 컨텍스트 정보를 담고 있으며, 1개의 P당 1개의 LRQ를 가지고 있음 G를 M에 할당하는 역할을 수행
- LRQ(LocalRunQueue): P에 종속되어 있는 Run Queue, 이 LRQ에 실행 가능한 고루틴들이 적재
- GRQ(GlobalRunQueue): LRQ에 할당되지 못한 고루틴을 관리하는 Run Queue, LRQ 적재되지 못한 고루틴들이 이 GRQ에 들어가 관리된다고 보면 됨
