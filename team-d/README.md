# gRPC-based-chatting
go로 구현한 gRPC 기반 채팅 시스템 

### 기술 스택 
- protocol: gRPC(bi-directional streaming)
- Serialization: Protocol Buffers 
- Frontend/ UI: wails framework(react)
- Language: Go(golang)
- Message broker: kafka
- Containerization: Docker, Docker Compose 

### 아키텍처 
![image](https://github.com/user-attachments/assets/e20e231f-4391-42c7-bc3d-60d4eba86ac9)


### 프로젝트 특징

- Kafka 기반 비동기 메시징 처리 
- DB 없이도 메시지 송수신 가능
    - 카프카의 데이터 임시 저장, 메모리 데이터 저장에 의존
- gRPC 양방향 스트리밍을 통한 다대다 채팅 구현
- wails framework를 통한 데스크탑 애플리케이션 개발

## 클라이언트 측 
### Wails framework

- 웹 브라우저는 gRPC 클라이언트로 동작할 수 없음(HTTP/2 지원 안함)
    - 웹 브라우저가 gRPC 클라이언트로 동작하기 위해서는 gRPC-Web을 추가적으로 사용해야 함
    - 단, gRPC-Web은 양방향 스트리밍을 사용할 수 없기 때문에 사용 시 핵심 기능(양방향 스트리밍을 통한 채팅) 구현 불가
- 따라서 웹 브라우저 대신 데스크탑 애플리케이션 방식으로 구현
- wails framework는 Go 언어와 웹 기술을 결합해 데스크톱 애플리케이션을 개발할 수 있게 해주는 프레임워크로써 사용

### 채팅 메시지 플로우

- 사용자의 데스크탑 애플리케이션 접속
- 사용자의 채팅방 입장 및 gRPC 서버와의 양방향 스트리밍 연결 시작
    - 송수신 고루틴 시작
- 메시지 송신
    - 사용자는 데스크탑 애플리케이션의 UI를 통해 메시지 입력 및 gRPC 클라이언트에 메시지 전송
    - gRPC 클라이언트는 송신 고루틴과 채널을 통해 사용자의 요청을 받아 gRPC 스트림에 메시지 전송
- 메시지 수신
    - 수신 고루틴이 gRPC 스트림에 메시지가 들어온 경우 wails context를 통해 수신 이벤트 발생
    - wails runtime을 통해 React에 이벤트 전달, 수신 메시지를 데스크탑 애플리케이션의 UI로 출력

## 서버 측 

### 서버 아키텍처 및 역할 
- 서버는 gRPC를 통해 클라이언트와 양방향 스트리밍 연결을 유지
- 실시간 메시지 전달은 Kafka를 중간 메시지 브로커로 사용해 처리
- Kafka 브로커는 챗 서버와 독립된 컨테이너 환경에서 구동
- 서버는 stateless하게 설계되어, 여러 chat-server 인스턴스가 동시에 떠 있어도 Kafka를 통해 메시지가 동기화

### 채팅 메시지 플로우 

1. 클라이언트 연결

    - 클라이언트는 ChatStream RPC를 통해 gRPC 서버에 연결 (양방향 스트리밍).
    - 첫 메시지를 통해 userID, channel 정보를 서버에 전달.

2. 메시지 수신 및 Kafka 전송

    - 클라이언트가 보낸 메시지는 Protobuf로 직렬화되어 Kafka의 chat-topic에 발행(produce)
    - 메시지의 key는 채팅방 ID(채널명)로 지정, 동일 채널 메시지는 같은 파티션에 저장

3. Kafka Consumer 동작

    - 서버는 Kafka의 chat-topic을 컨슈머로 구독
    - 수신한 메시지를 역직렬화(Protobuf)하여, 해당 채널에 접속 중인 모든 클라이언트 세션에 전달

4. 브로드캐스트

    - 서버는 해당 channel에 등록된 사용자 스트림 목록을 참조.
    - 각 사용자에게 gRPC 스트림을 통해 메시지 전송.

### 주요 설계 포인트 
- 비동기/내구성: Kafka를 통해 메시지가 일시적으로 저장되므로, 서버가 재시작되어도 메시지 유실 위험이 적음.

- 확장성: 여러 chat-server 인스턴스를 띄워도 Kafka가 메시지를 중개하므로, 서버 간 동기화/브로드캐스트가 자동으로 보장됨.

- Go 고루틴(goroutine) 기반 비동기 브로드캐스트: Kafka 메시지 전송 및 여러 클라이언트로의 브로드캐스트에 고루틴을 사용하여,
서버가 블로킹 없이 실시간으로 대량 메시지를 처리

- 메모리 관리: 채널별로 접속 중인 유저 정보만 메모리에 관리, DB를 사용하지 않아도 높은 성능과 단순성을 유지.
