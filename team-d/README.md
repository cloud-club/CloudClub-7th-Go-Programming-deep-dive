## 주제
- gRPC 통신 프로토콜 사용한 채팅 시스템 

## 기술 스택 
- gRPC
- Kafka 

## 주요 기능 
- 일대일, 다대다 실시간 채팅 구현 
- 메시지 브로커를 통한 비동기 통신 구현 


## 아키텍처 

<img width="514" alt="Image" src="https://github.com/user-attachments/assets/602d3dc0-15da-4804-aa09-b29134319a17" />

## 디렉토리 구조 

```

chat-app/
├── client/              
│   └── client.go        
│
├── server/              
│   └── server.go        
│
├── chatProto/           
│   └── chat.proto       
│
├── kafka/                # Kafka 프로듀서 / 컨슈머 관련 코드
│   ├── producer.go       # Kafka에 메시지를 보내는 로직
│   └── consumer.go       # Kafka로부터 메시지를 받는 로직
│
├── go.mod               
└── go.sum               


```

