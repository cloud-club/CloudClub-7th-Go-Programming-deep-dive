---
name: First Task
about: 기본 HTTP 메트릭 스크래퍼 구현
title: "feat: 기본 HTTP 메트릭 스크래퍼 구현"
labels: enhancement, good first issue
assignees: ''
---

## 작업 내용
HTTP 엔드포인트에서 간단한 메트릭을 수집하는 기본 스크래퍼 구현

### 구현 사항
- [ ] HTTP 클라이언트를 사용하여 지정된 엔드포인트(/metrics)에서 데이터 수집
- [ ] 수집된 데이터를 메모리 스토리지에 저장
- [ ] 기본적인 에러 처리 및 로깅

### 구현 가이드
1. HTTP 클라이언트 설정
   - timeout 설정 (예: 10초)
   - 기본 HTTP 클라이언트 사용

2. 메트릭 수집 로직
   - GET 요청으로 /metrics 엔드포인트 호출
   - JSON 응답 파싱
   - 기본적인 에러 처리

3. 스토리지 연동
   - 수집된 메트릭을 메모리 스토리지에 저장
   - 동시성 고려 (mutex 사용)

### 테스트 방법
1. 테스트용 메트릭 엔드포인트 실행:
```go
package main

import (
    "encoding/json"
    "net/http"
)

func main() {
    http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
        metrics := []struct {
            Name  string  `json:"name"`
            Value float64 `json:"value"`
        }{
            {"cpu_usage", 45.5},
            {"memory_usage", 60.2},
        }
        json.NewEncoder(w).Encode(metrics)
    })
    http.ListenAndServe(":8080", nil)
}
```

2. 스크래퍼 실행 및 로그 확인

### 팁
- net/http 패키지의 기본 클라이언트 사용
- context를 활용한 timeout 처리 고려
- 에러 발생 시 적절한 로깅 추가
- 동시성 이슈 방지를 위한 mutex 사용
- 단위 테스트 작성 권장

### 참고 자료
- [Go HTTP Client](https://golang.org/pkg/net/http/#Client)
- [Go Mutex](https://golang.org/pkg/sync/#Mutex)
- [Go JSON](https://golang.org/pkg/encoding/json/)

### 예상 소요 시간
- 1-2시간

### 체크리스트
- [ ] HTTP 클라이언트 구현
- [ ] 메트릭 수집 로직 구현
- [ ] 에러 처리 추가
- [ ] 기본 로깅 추가
- [ ] 단위 테스트 작성
- [ ] 테스트용 엔드포인트로 검증 