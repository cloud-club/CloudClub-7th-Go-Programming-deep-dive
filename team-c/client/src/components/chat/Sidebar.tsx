"use client";

import { Room, User } from "@/types";
import {
  ArrowPathIcon,
  ArrowRightOnRectangleIcon,
  ChatBubbleLeftRightIcon,
  PlusIcon,
  UserIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import clsx from "clsx";
import { useState } from "react";

interface SidebarProps {
  currentUser: User;
  rooms: Room[];
  currentRoom: Room | null;
  onJoinRoom: (roomId: string) => void;
  onCreateRoom: () => void;
  onRefreshRooms: () => void;
  onLogout: () => void;
  isMobileOpen: boolean;
  onCloseMobile: () => void;
}

export const Sidebar = ({
  currentUser,
  rooms,
  currentRoom,
  onJoinRoom,
  onCreateRoom,
  onRefreshRooms,
  onLogout,
  isMobileOpen,
  onCloseMobile,
}: SidebarProps) => {
  const [isRefreshing, setIsRefreshing] = useState(false);

  const handleRefresh = async () => {
    setIsRefreshing(true);
    try {
      await onRefreshRooms();
    } finally {
      setIsRefreshing(false);
    }
  };

  const handleRoomClick = (roomId: string) => {
    onJoinRoom(roomId);
    onCloseMobile();
  };

  return (
    <>
      {/* 모바일 오버레이 */}
      {isMobileOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-40 lg:hidden"
          onClick={onCloseMobile}
        />
      )}

      {/* 사이드바 */}
      <div
        className={clsx(
          "fixed inset-y-0 left-0 z-50 w-80 bg-white border-r border-gray-200 transform transition-transform duration-300 ease-in-out lg:translate-x-0 lg:static lg:inset-0",
          isMobileOpen ? "translate-x-0" : "-translate-x-full"
        )}
      >
        <div className="flex flex-col h-full">
          {/* 헤더 */}
          <div className="flex items-center justify-between p-4 border-b border-gray-200">
            <div className="flex items-center space-x-3">
              <div className="w-10 h-10 bg-indigo-600 rounded-full flex items-center justify-center">
                <UserIcon className="w-6 h-6 text-white" />
              </div>
              <div>
                <p className="font-semibold text-gray-900">
                  {currentUser.name}
                </p>
                <p className="text-sm text-green-600 flex items-center">
                  <span className="w-2 h-2 bg-green-500 rounded-full mr-2" />
                  온라인
                </p>
              </div>
            </div>
            <button
              onClick={onCloseMobile}
              className="lg:hidden p-2 hover:bg-gray-100 rounded-lg transition-colors"
            >
              <XMarkIcon className="w-5 h-5" />
            </button>
          </div>

          {/* 채팅방 섹션 */}
          <div className="flex-1 flex flex-col overflow-hidden">
            <div className="flex items-center justify-between p-4 border-b border-gray-200">
              <h3 className="font-semibold text-gray-900 flex items-center">
                <ChatBubbleLeftRightIcon className="w-5 h-5 mr-2" />
                채팅방
              </h3>
              <div className="flex space-x-2">
                <button
                  onClick={handleRefresh}
                  disabled={isRefreshing}
                  className="p-2 hover:bg-gray-100 rounded-lg transition-colors disabled:opacity-50"
                  title="새로고침"
                >
                  <ArrowPathIcon
                    className={clsx(
                      "w-4 h-4 text-gray-600",
                      isRefreshing && "animate-spin"
                    )}
                  />
                </button>
                <button
                  onClick={onCreateRoom}
                  className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                  title="새 채팅방"
                >
                  <PlusIcon className="w-4 h-4 text-gray-600" />
                </button>
              </div>
            </div>

            {/* 채팅방 목록 */}
            <div className="flex-1 overflow-y-auto p-2">
              {rooms.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  <ChatBubbleLeftRightIcon className="w-12 h-12 mx-auto mb-2 opacity-50" />
                  <p>생성된 채팅방이 없습니다</p>
                  <button
                    onClick={onCreateRoom}
                    className="mt-2 text-indigo-600 hover:text-indigo-700 text-sm font-medium"
                  >
                    첫 번째 채팅방 만들기
                  </button>
                </div>
              ) : (
                <div className="space-y-1">
                  {rooms.map((room) => (
                    <button
                      key={room.id}
                      onClick={() => handleRoomClick(room.id)}
                      className={clsx(
                        "w-full text-left p-3 rounded-lg transition-colors",
                        currentRoom?.id === room.id
                          ? "bg-indigo-600 text-white"
                          : "hover:bg-gray-100 text-gray-900"
                      )}
                    >
                      <div className="flex items-center justify-between mb-1">
                        <h4 className="font-medium truncate">{room.name}</h4>
                        <span
                          className={clsx(
                            "text-xs px-2 py-1 rounded-full",
                            currentRoom?.id === room.id
                              ? "bg-white/20 text-white"
                              : "bg-gray-200 text-gray-600"
                          )}
                        >
                          {room.user_count}
                        </span>
                      </div>
                      {room.description && (
                        <p
                          className={clsx(
                            "text-sm truncate",
                            currentRoom?.id === room.id
                              ? "text-white/80"
                              : "text-gray-600"
                          )}
                        >
                          {room.description}
                        </p>
                      )}
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* 푸터 */}
          <div className="p-4 border-t border-gray-200 space-y-2">
            <button
              onClick={handleRefresh}
              disabled={isRefreshing}
              className="w-full flex items-center justify-center px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors disabled:opacity-50"
            >
              <ArrowPathIcon
                className={clsx("w-4 h-4 mr-2", isRefreshing && "animate-spin")}
              />
              {isRefreshing ? "새로고침 중..." : "새로고침"}
            </button>
            <button
              onClick={onLogout}
              className="w-full flex items-center justify-center px-4 py-2 bg-red-100 text-red-700 rounded-lg hover:bg-red-200 transition-colors"
            >
              <ArrowRightOnRectangleIcon className="w-4 h-4 mr-2" />
              나가기
            </button>
          </div>
        </div>
      </div>
    </>
  );
};
