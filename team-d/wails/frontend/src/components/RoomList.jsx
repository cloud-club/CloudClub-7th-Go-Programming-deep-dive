import React from 'react';
import '../styles/RoomList.css';

function RoomList({
  rooms,
  onJoinRoom,
  onCreateRoom,
  onBackToLogin,
  onGetRoomList
}) {
  return (
    <div className="room-list-container room-list-sidebar">
      <div className="room-list-header"> 
         <div>💬 CloudClub</div>
          <button className="refresh-room-btn" onClick={onGetRoomList}>Refresh</button>
      </div>

      <ul className="room-items hide-scrollbar">
        {rooms.length > 0 ? (
          rooms.map((room) => (
            <li key={room.id} className="room-item" onClick={() => onJoinRoom(room.id)}>
              <span className="room-name">{room.name}</span>
              <span className="room-meta">
                <span className="dot">•</span>
              </span>
            </li>
          ))
        ) : (
          <li className="no-rooms">Create a room and<br/> start a new conversation!</li>
        )}
      </ul>

      <div className="room-list-buttons">
        <button className="btn primary create-room-btn" onClick={onCreateRoom}>Create Room</button>
        <button className="btn secondary back-to-login-btn" onClick={onBackToLogin}>Main Menu</button>
      </div>
    </div>
  );
}

export default RoomList;