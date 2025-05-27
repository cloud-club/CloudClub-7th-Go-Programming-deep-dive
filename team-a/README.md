# Project: Swarm
go기반 성능 부하 테스트

## Golang Version

* 1.24.2

## 협업 방식

* `team-a` 브랜치를 생성하여 해당 브랜치에 먼저 코드 병합 후, 검토 및 테스트 후 `main` 브랜치로 병합합니다.

## 설정 파일

* YAML 형식으로 설정을 주입합니다.

## 개발 계획

### 1차 목표

* CLI를 통해 간단한 호스트 부하 테스트 구현
  예시:

  ```bash
  swarm --headless --users 10 --spawn-rate 1 -H http://your-server.com
  ```

### 2차 목표

* YAML 설정을 활용한 부하 동작 테스트 구현

### 3차 목표

* 경로별로 다른 비율로 부하를 분산하여 테스트

## Sprint 일정

| 기간           | 내용                  |
| ------------ | ------------------- |
| 5.23 \~ 6.1  | 최소한의 MVP 구현         |
| 6.6 \~ 6.20  | 동영 빠르게 구현, 창현님 느슨하게 |
| 6.20 \~ 6.27 | 느슨하게 MVP 구현 완료?     |

## Input / Output

* **Input**

    * Host 주소
    * 포트 정보
    * 테스트 기간
    * 유저 수

* **Output**

    * 200\~300번대 성공/실패 비율
    * 2초 이내 응답이 없을 경우 실패로 간주

## 참조 프로젝트

* [grafana/k6](https://github.com/grafana/k6)
* [spf13/cobra](https://github.com/spf13/cobra)
* [tsenart/vegeta](https://github.com/tsenart/vegeta)
* [valyala/fasthttp](https://github.com/valyala/fasthttp)
