"use client";

import { ConnectionStatus, Room } from "@/types";
import {
  ArrowPathIcon,
  Bars3Icon,
  ExclamationTriangleIcon,
  ListBulletIcon,
  WifiIcon,
} from "@heroicons/react/24/outline";
import clsx from "clsx";

interface ChatHeaderProps {
  currentRoom: Room | null;
  connectionStatus: ConnectionStatus;
  onToggleSidebar: () => void;
  onReconnect?: () => void;
}

export const ChatHeader = ({
  currentRoom,
  connectionStatus,
  onToggleSidebar,
  onReconnect,
}: ChatHeaderProps) => {
  const getConnectionIcon = () => {
    switch (connectionStatus) {
      case "connected":
        return <WifiIcon className="w-4 h-4 text-green-600" />;
      case "connecting":
        return (
          <ArrowPathIcon className="w-4 h-4 text-yellow-600 animate-spin" />
        );
      case "disconnected":
      case "error":
        return <ExclamationTriangleIcon className="w-4 h-4 text-red-600" />;
      default:
        return <WifiIcon className="w-4 h-4 text-gray-400" />;
    }
  };

  const getConnectionText = () => {
    switch (connectionStatus) {
      case "connected":
        return "연결됨";
      case "connecting":
        return "연결 중...";
      case "disconnected":
        return "연결 끊김";
      case "error":
        return "연결 오류";
      default:
        return "알 수 없음";
    }
  };

  const getConnectionColor = () => {
    switch (connectionStatus) {
      case "connected":
        return "text-green-600";
      case "connecting":
        return "text-yellow-600";
      case "disconnected":
      case "error":
        return "text-red-600";
      default:
        return "text-gray-400";
    }
  };

  return (
    <div className="flex items-center justify-between p-4 border-b border-gray-200 bg-white">
      <div className="flex items-center space-x-3">
        <button
          onClick={onToggleSidebar}
          className="lg:hidden p-2 hover:bg-gray-100 rounded-lg transition-colors"
          aria-label="메뉴 열기"
        >
          <Bars3Icon className="w-5 h-5" />
        </button>

        <div>
          <h1 className="text-lg font-semibold text-gray-900">
            {currentRoom ? currentRoom.name : "채팅방을 선택하세요"}
          </h1>
          {currentRoom?.description && (
            <p className="text-sm text-gray-600 mt-1">
              {currentRoom.description}
            </p>
          )}
        </div>
      </div>

      <div className="flex items-center space-x-3">
        <button
          onClick={onToggleSidebar}
          className="hidden lg:block p-2 hover:bg-gray-100 rounded-lg transition-colors"
          title="채팅방 목록"
        >
          <ListBulletIcon className="w-5 h-5 text-gray-600" />
        </button>

        <div
          className={clsx(
            "flex items-center space-x-2 px-3 py-1 rounded-full text-sm font-medium transition-colors",
            connectionStatus === "connected" && "bg-green-50",
            connectionStatus === "connecting" && "bg-yellow-50",
            (connectionStatus === "disconnected" ||
              connectionStatus === "error") &&
              "bg-red-50 cursor-pointer hover:bg-red-100"
          )}
          onClick={connectionStatus !== "connected" ? onReconnect : undefined}
          title={
            connectionStatus !== "connected" ? "클릭하여 재연결" : undefined
          }
        >
          {getConnectionIcon()}
          <span className={getConnectionColor()}>{getConnectionText()}</span>
        </div>
      </div>
    </div>
  );
};
