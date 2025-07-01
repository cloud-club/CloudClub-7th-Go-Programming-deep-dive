// App.js
import React, { useState } from "react";
import "./App.css";
import ChatRoom from "./pages/ChatRoom";

function App() {
  const [nickname, setNickname] = useState("");
  const [joined, setJoined] = useState(false);
  const [rooms] = useState(["📢 공지사항", "💬 자유게시판", "❓ 질문방"]);
  const [selectedRoom, setSelectedRoom] = useState(null);

  // ✅ 누락된 함수 추가
  const handleJoin = () => {
    if (!nickname.trim()) return;
    setJoined(true);
  };

  const handleRoomSelect = (room) => {
    setSelectedRoom(room);
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
        <ChatRoom
          nickname={nickname}
          room={selectedRoom}
          onLeave={() => {
            setSelectedRoom(null); 
          }}
        />
      )}
    </div>
  );
}

export default App;
