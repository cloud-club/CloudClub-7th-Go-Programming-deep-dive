// App.js

import React, { useState, useEffect } from "react";
import "./App.css";
import ChatRoom from "./pages/ChatRoom";

function App() {
  const [nickname, setNickname] = useState("");
  const [joined, setJoined] = useState(false);
  const [rooms] = useState(["📢 공지사항", "💬 자유게시판", "❓ 질문방"]);
  const [selectedRoom, setSelectedRoom] = useState(null);
  const [ws, setWs] = useState(null);
  const [messages, setMessages] = useState([]);
  const [message, setMessage] = useState("");

  useEffect(() => {
    if (joined && selectedRoom) {
      const socket = new WebSocket("ws://localhost:8080/ws");
      socket.onmessage = (event) => {
        const msg = JSON.parse(event.data);
        setMessages((prev) => [...prev, msg]);
      };
      setWs(socket);
      return () => socket.close();
    }
  }, [joined, selectedRoom, nickname]);

  const handleJoin = () => {
    if (!nickname.trim()) return;
    setJoined(true);
  };

  const handleRoomSelect = (room) => {
    setSelectedRoom(room);
  };

  const handleSend = (e) => {
    e.preventDefault();
    if (message.trim() && ws?.readyState === WebSocket.OPEN) {
      ws.send(
        JSON.stringify({
          user: nickname,
          content: message,
          timestamp: Date.now(),
        })
      );
      setMessage("");
    }
  };

  return (
    <div className="screen">
      {!joined ? (
        <div className="card center">
          <h1 className="title">🌥️ CloudClub</h1>
          <p className="subtitle">닉네임을 입력해주세요</p>
          <input
            type="text"
            className="input"
            placeholder="닉네임 입력"
            value={nickname}
            onChange={(e) => setNickname(e.target.value)}
          />
          <button className="button" onClick={handleJoin}>
            입장하기
          </button>
        </div>
      ) : !selectedRoom ? (
        <div className="card center">
          <h2>👋 {nickname}님, 환영합니다!</h2>
          <p className="subtitle">채팅방을 선택하세요</p>
          <div className="room-buttons">
            {rooms.map((room) => (
              <button
                key={room}
                className="button room"
                onClick={() => handleRoomSelect(room)}
              >
                {room}
              </button>
            ))}
          </div>
        </div>
      ) : (
        <ChatRoom nickname={nickname} room={selectedRoom} />
      )}
    </div>
  );
}

export default App;


