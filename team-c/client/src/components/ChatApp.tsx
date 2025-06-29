"use client";

import { useChat } from "@/hooks/useChat";
import { useToast } from "@/hooks/useToast";
import { useEffect, useState } from "react";

import { ChatHeader } from "@/components/chat/ChatHeader";
import { CreateRoomModal } from "@/components/chat/CreateRoomModal";
import { LoginModal } from "@/components/chat/LoginModal";
import { MessageInput } from "@/components/chat/MessageInput";
import { MessageList } from "@/components/chat/MessageList";
import { Sidebar } from "@/components/chat/Sidebar";
import { ToastContainer } from "@/components/ui/Toast";

export default function ChatApp() {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);
  const [isCreateRoomModalOpen, setIsCreateRoomModalOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  const {
    currentUser,
    currentRoom,
    rooms,
    messages,
    connectionStatus,
    hasConnectedOnce,
    login,
    logout,
    joinRoom,
    sendMessage,
    createRoom,
    loadRooms,
  } = useChat();

  const { toasts, removeToast, showSuccess, showError, showWarning, showInfo } =
    useToast();

  // 초기 로딩 효과
  useEffect(() => {
    const timer = setTimeout(() => {
      setIsLoading(false);
    }, 1000);

    return () => clearTimeout(timer);
  }, []);

  // 연결 상태 변화 시 토스트 표시 (최초 연결 제외)
  useEffect(() => {
    // 최초 로딩 중이거나 사용자가 로그인하지 않은 경우 토스트 표시 안함
    if (isLoading || !currentUser) return;

    // 최초 연결 시에는 토스트 표시하지 않음
    if (!hasConnectedOnce && connectionStatus !== "error") return;

    // 연결 상태가 바뀔 때만 토스트 표시 (연결 중 상태는 제외)
    switch (connectionStatus) {
      case "connected":
        // 재연결 시에만 표시
        if (hasConnectedOnce) {
          showSuccess("서버에 연결되었습니다");
        }
        break;
      case "disconnected":
        // 사용자가 의도적으로 로그아웃한 경우가 아닐 때만 표시
        if (currentUser && hasConnectedOnce) {
          showWarning("서버와의 연결이 끊어졌습니다");
        }
        break;
      case "error":
        // 연결 에러는 항상 표시 (최초 연결 실패 포함)
        if (currentUser) {
          showError("연결 오류가 발생했습니다");
        }
        break;
      // "connecting" 상태는 토스트 표시하지 않음
    }
  }, [
    connectionStatus,
    showSuccess,
    showWarning,
    showError,
    isLoading,
    currentUser,
    hasConnectedOnce,
  ]);

  // 채팅방 입장 시 사이드바 닫기 (모바일)
  const handleJoinRoom = async (roomId: string) => {
    await joinRoom(roomId);
    setIsSidebarOpen(false);
  };

  // 채팅방 생성
  const handleCreateRoom = async (name: string, description: string) => {
    try {
      await createRoom(name, description);
      showSuccess("채팅방이 생성되었습니다");
    } catch (error) {
      showError("채팅방 생성에 실패했습니다");
      throw error;
    }
  };

  // 로그인 처리
  const handleLogin = (username: string) => {
    login(username);
    showSuccess(`환영합니다, ${username}님!`);
  };

  // 로그아웃 처리
  const handleLogout = () => {
    logout();
    setIsSidebarOpen(false);
    showInfo("로그아웃되었습니다");
  };

  // 로딩 스크린
  if (isLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-indigo-600 to-purple-700 flex items-center justify-center">
        <div className="text-center text-white">
          <div className="w-16 h-16 mx-auto mb-4 animate-bounce">
            <svg
              className="w-full h-full"
              fill="currentColor"
              viewBox="0 0 24 24"
            >
              <path d="M20 2H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h4l4 4 4-4h4c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2z" />
            </svg>
          </div>
          <h2 className="text-2xl font-bold mb-2">CloudClub Chat</h2>
          <div className="flex items-center justify-center space-x-1">
            <div
              className="w-2 h-2 bg-white rounded-full animate-bounce"
              style={{ animationDelay: "0ms" }}
            />
            <div
              className="w-2 h-2 bg-white rounded-full animate-bounce"
              style={{ animationDelay: "150ms" }}
            />
            <div
              className="w-2 h-2 bg-white rounded-full animate-bounce"
              style={{ animationDelay: "300ms" }}
            />
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="h-screen bg-gray-50 flex overflow-hidden">
      {/* 로그인 모달 */}
      <LoginModal isOpen={!currentUser} onLogin={handleLogin} />

      {/* 채팅방 생성 모달 */}
      <CreateRoomModal
        isOpen={isCreateRoomModalOpen}
        onClose={() => setIsCreateRoomModalOpen(false)}
        onCreateRoom={handleCreateRoom}
      />

      {/* 사이드바 */}
      {currentUser && (
        <Sidebar
          currentUser={currentUser}
          rooms={rooms}
          currentRoom={currentRoom}
          onJoinRoom={handleJoinRoom}
          onCreateRoom={() => setIsCreateRoomModalOpen(true)}
          onRefreshRooms={loadRooms}
          onLogout={handleLogout}
          isMobileOpen={isSidebarOpen}
          onCloseMobile={() => setIsSidebarOpen(false)}
        />
      )}

      {/* 메인 채팅 영역 */}
      {currentUser && (
        <div className="flex-1 flex flex-col min-w-0">
          {/* 채팅 헤더 */}
          <ChatHeader
            currentRoom={currentRoom}
            connectionStatus={connectionStatus}
            onToggleSidebar={() => setIsSidebarOpen(!isSidebarOpen)}
          />

          {/* 메시지 영역 */}
          <MessageList
            messages={messages}
            currentUser={currentUser}
            currentRoomName={currentRoom?.name}
          />

          {/* 메시지 입력 */}
          <MessageInput
            onSendMessage={sendMessage}
            disabled={!currentRoom || connectionStatus !== "connected"}
          />
        </div>
      )}

      {/* 토스트 알림 */}
      <ToastContainer toasts={toasts} onRemove={removeToast} />
    </div>
  );
}
