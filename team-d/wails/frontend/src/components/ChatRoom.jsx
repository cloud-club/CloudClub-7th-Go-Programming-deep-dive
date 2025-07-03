import React, { useState, useEffect, useRef } from 'react';
import '../styles/ChatRoom.css';

function ChatRoom({
  roomName,
  messages,
  onSendMessage,
  onExitRoom,
  userId
}) {
  const [newMessage, setNewMessage] = useState('');
  const messagesEndRef = useRef(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  // 입력값을 상태로 저장
  const handleInputChange = (event) => {
    setNewMessage(event.target.value);
  };

  // 입력값을 전송
  const handleSendMessage = () => {
    if (newMessage.trim() !== '') {
      onSendMessage(newMessage);
      setNewMessage('');
    }
  };

  // 엔터를 누를 경우 메시지 전송
  const handleKeyPress = (event) => {
    if (event.key === 'Enter') {
      event.preventDefault();
      handleSendMessage();
    }
  };

  return (
    <div className="chat-room-container chat-room-main">
      <div className="chat-header">
        <h2>{roomName}</h2>
        <div className="chat-header-buttons">
          {onExitRoom && (
            <button className="btn danger" onClick={onExitRoom}>Leave Room</button>
          )}
        </div>
      </div>
      <div className="chat-messages hide-scrollbar">
        {messages.length > 0 ? (
          messages.map((message, index) =>
            message.sender === "admin" ? (
              <div key={index} className="system">
                {message.content}
              </div>
            ) : (
              <div
                key={index}
                className={`message ${message.sender === userId ? 'sent' : 'received'}`}
              >
                <div className="message-sender">{message.sender}:</div>
                <div className="message-text">{message.content}</div>
              </div>
            )
          )
        ) : (
          <div className="system">Send your first message to connect!</div>
        )}

        <div ref={messagesEndRef} />
      </div>
      <div className="chat-input-area">
        <input
          type="text"
          placeholder="Enter message..."
          value={newMessage}
          onChange={handleInputChange}
          onKeyDown={handleKeyPress}
        />
        <button className="btn primary" onClick={handleSendMessage}>Send</button>
      </div>
    </div>
  );
}

export default ChatRoom;