
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

