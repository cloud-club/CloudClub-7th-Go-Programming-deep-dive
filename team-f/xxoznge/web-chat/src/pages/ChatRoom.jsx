import React, { useState, useEffect, useRef } from "react";
import "./ChatRoom.css";

function ChatRoom({ nickname, room, onLeave }) {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState("");
  const socketRef = useRef(null);
  const messagesEndRef = useRef(null);
  const [onlineCount, setOnlineCount] = useState(1);

  useEffect(() => {
    const socket = new WebSocket("ws://localhost:8080/ws");
    socketRef.current = socket;

    socket.onopen = () => {
      socket.send(
        JSON.stringify({
          user: nickname,
          content: `${nickname}님이 입장하셨습니다.`,
          timestamp: Date.now(),
        })
      );
    };

    socket.onmessage = (event) => {
      const msg = JSON.parse(event.data);

      if (msg.type === "online-count") {
        setOnlineCount(msg.count);
        return;
      }

      setMessages((prev) => [...prev, msg]);
    };

    return () => {
      if (socketRef.current?.readyState === WebSocket.OPEN) {
        socketRef.current.send(
          JSON.stringify({
            user: nickname,
            content: `${nickname}님이 퇴장하셨습니다.`,
            timestamp: Date.now(),
          })
        );
        socketRef.current.close();
      }
      socketRef.current = null;
    };
  }, [nickname]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const sendMessage = () => {
    if (
      input.trim() !== "" &&
      socketRef.current?.readyState === WebSocket.OPEN
    ) {
      const message = {
        user: nickname,
        content: input,
        timestamp: Date.now(),
      };
      socketRef.current.send(JSON.stringify(message));
      setInput("");
    }
  };

  const handleKeyPress = (e) => {
    if (e.key === "Enter") sendMessage();
  };

  const handleLeave = () => {
    if (socketRef.current?.readyState === WebSocket.OPEN) {
      socketRef.current.send({
        user: nickname,
        content: `${nickname}님이 퇴장하셨습니다.`,
        timestamp: Date.now(),
      });
      socketRef.current.close();
    }
    socketRef.current = null;
    onLeave(); // 👉 채팅방 선택 화면으로 돌아가기
  };


  return (
    <div className="chat-container">
      <div className="chat-header">
    <button onClick={handleLeave} className="leave-button">나가기</button>
    <div className="chat-room-title"> {room}</div>
    <div className="chat-online-count">👥 {onlineCount}명 접속 중</div>
  </div>

      
      <div className="chat-messages">
        {messages.map((msg, idx) => {
          const isSystem =
            msg.content.includes("입장하셨습니다") ||
            msg.content.includes("퇴장하셨습니다");

          if (isSystem) {
            return (
              <div key={idx} className="system-message">
                {msg.content}
              </div>
            );
          }

          const isMine = msg.user === nickname;

          return (
            <div
              key={idx}
              className={`chat-message-wrapper ${isMine ? "my" : "other"}`}
            >
              <div className="chat-user">{msg.user}</div>
              <div
                className={`chat-bubble ${
                  isMine ? "my-message" : "other-message"
                }`}
              >
                {msg.content}
              </div>
            </div>
          );
        })}
        <div ref={messagesEndRef} />
      </div>

      <div className="chat-input-area">
        <input
          type="text"
          placeholder="메시지를 입력하세요"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyPress}
        />
        <button onClick={sendMessage}>전송</button>
      </div>
    </div>
  );
}

export default ChatRoom;

