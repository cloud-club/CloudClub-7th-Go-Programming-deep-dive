import React, { useState } from 'react';
import './App.css'; // Keep general app styles if any
import { v4 as uuidv4 } from 'uuid';
import LoginExit from './components/LoginExit';
import RoomList from './components/RoomList';
import ChatRoom from './components/ChatRoom';
import { Quit, EventsOn, EventsOff } from '../wailsjs/runtime/runtime';
import { SetUserId, GetUserId, StartChat, Close, CloseChat, SendChatMessage, GetMessageHistory, GetChannelList } from '../wailsjs/go/client/Client'
import { useEffect } from 'react';

function App() {
  const [currentView, setCurrentView] = useState('login');
  const [userId, setUserId] = useState('');
  const [activeRoom, setActiveRoom] = useState(null);
  const [joinedRooms, setJoinedRooms] = useState([]);
  const [rooms, setRooms] = useState([]);
  const [messages, setMessages] = useState([]);

  useEffect(() => {
    EventsOn("newMessage", handleReceiveMessage);
    return () => {
      EventsOff("newMessage");
    };
  }, [activeRoom]);

  // Handlers
  const handleLogin = async () => {
    await SetUserId()
    const uid = await GetUserId();
    setUserId(uid);
    setCurrentView('chatRoom');
  };

  const handleExit = () => {
    Close();
    Quit();
  };

  const handleBackToLogin = () => {
    setCurrentView('login');
  }; 

  // 새로운 채팅방 생성
  // 리액트 상에서만 생성, 서버와 연결하지 않음
  const handleCreateRoom = () => {
    const newRoomId = uuidv4() // 채팅방ID는 uuid 기반
    const newRoomName = `Chatroom ${newRoomId.slice(0, 8)}...`;
    const newRoom = { id: newRoomId, name: newRoomName };
    setRooms([...rooms, newRoom]);
  };

  // 채팅방 들어가기
  const handleJoinRoom = async (roomId) => {
    // 채팅방이 존재하는지 찾기
    const roomToJoin = rooms.find(room => room.id === roomId);
    if (!roomToJoin) return;
    
    // 이미 가입한 방인지 확인
    const alreadyJoined = joinedRooms.some(room => room.id === roomId);
    if (!alreadyJoined) {
      setJoinedRooms(prev => [
        ...prev,
        { id: roomToJoin.id, name: roomToJoin.name }
      ]);
      setMessages([]);

      await StartChat(roomId);
      
      // 입장 메시지 전송
      // const timestamp = new Date(Date.now()).toISOString();
      // const newMessage = {
      //   channel: roomToJoin.id,
      //   sender: "admin",
      //   content: `${userId} join the chat room!`,
      //   timestamp: timestamp
      // };
      // await SendChatMessage(newMessage)
    } else {
      const history = await GetMessageHistory(roomId);
      setMessages(history || []);
    }

    setActiveRoom(roomToJoin);
    setCurrentView('chatRoom');
  };

  // 메시지 송신
  const handleSendMessage = async (messageText) => {
    if (!activeRoom) return;

    const timestamp = new Date(Date.now()).toISOString();
    const newMessage = {
      channel: activeRoom.id,
      sender: userId,
      content: messageText,
      timestamp: timestamp
    };

    await SendChatMessage(newMessage)
  };

  // 메시지 수신
  const handleReceiveMessage = (message) => {
    setMessages(prev => [...prev, message]);
  };

  // 채팅방 나가기
  const handleExitRoom = async (roomId) => {
    // const timestamp = new Date(Date.now()).toISOString();
    // const newMessage = {
    //   channel: activeRoom.id,
    //   sender: "admin",
    //   content: `${userId} left the chat room!`,
    //   timestamp: timestamp
    // };
    // await SendChatMessage(newMessage)

    await CloseChat(roomId)
    setActiveRoom(null);
    setMessages([]);
    setJoinedRooms(prev => prev.filter(room => room.id !== roomId));
  }

  const handleGetRoomList = async () => {
    const roomList = await GetChannelList()
    if(roomList == null) return;

    const roomObjects = roomList.map(id => ({
        id: id,
        name: `Chatroom ${id.slice(0, 8)}...`,
      }));
      setRooms(roomObjects);
  }

  const renderView = () => {
    switch (currentView) {
      case 'login':
        return <LoginExit onLogin={handleLogin} onExit={handleExit} />;
      case 'chatRoom':
        return (
          <div className="main-layout">
            <RoomList
              rooms={rooms}
              onJoinRoom={handleJoinRoom}
              onCreateRoom={handleCreateRoom}
              onBackToLogin={handleBackToLogin}
              onGetRoomList={handleGetRoomList}
            />
            {activeRoom ? (
              <ChatRoom
                roomName={activeRoom.name}
                messages={messages || []}
                onSendMessage={handleSendMessage}
                onExitRoom={() => handleExitRoom(activeRoom.id)}
                userId={userId}
              />
            ) : (
              <div className="chat-room-no-room">
              </div>
            )}
          </div>
        );
      default:
        return <LoginExit onLogin={handleLogin} onExit={handleExit} />;
    }
  };

  return (
    <div id="App">
      {renderView()}
    </div>
  );
}

export default App;
