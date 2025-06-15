
#### Base Code

Server <-> Client 간 양방향 통신이 아님, Client가 일방적으로 메시지를 전달 -> 메시지 큐의 성격

___

#### 수정 사항

Client에게 메시지를 전달하기 위한 구분자(Primary Key) 필요

양방향 통신을 위해선 Server에서 CLI 기능 추가 필요
```
list: 연결된 client 목록을 보여줌
enter <UUID>: 해당 uuid를 가진 client와 채팅
clear: 정상 종료를 위한 client 커넥션 종료
```

___

기능 추가에 따른 디렉토리 구조 분리
```
grpc-chat/
├── go.mod  # go Module 선언

├── cmd/  # build 대상
│   └── server/
│       └── main.go
│   └── client/
│       └── main.go

├── internal/  # 기능적 로직 관리
│   ├── domain/    #Domain: 유저와 메시지를 정의
│   │   └── model.go
│   ├── server/    #Server: server 기능 정의
│   │   └── server.go
│   ├── client/    #Client: client 기능 정의
│   │   └── client.go
│   ├── usecase/    #Usecase: 비즈니스 로직 분리, 테스트 가능
│   │   └── chat_usecase.go
│   ├── port/    #Port: ChatService 인터페이스와 세션 저장소 정의
│   │   ├── in/
│   │   │   └── chat_service.go
│   │   └── out/
│   │       └── session_repo.go
│   └── adapter/    #Adapter: gRPC 프레임워크와 Core Usecase 연결
│       └── grpc/
│           └── handler.go

├── infrastructure/    #Infrastructure: 메모리 저장소 구현
│   └── memory/    
│       └── session_repository.go

├── proto/  # proto buffer 정의
│   └── chat.proto
├── gen/
│   └── chat.pb.go            # protoc-gen-go 생성
│   └── chat_grpc.pb.go       # protoc-gen-go-grpc 생성
```
