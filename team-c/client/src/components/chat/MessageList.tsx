"use client";

import { Message, User } from "@/types";
import { ChatBubbleLeftRightIcon } from "@heroicons/react/24/outline";
import clsx from "clsx";
import { useEffect, useRef } from "react";

interface MessageListProps {
  messages: Message[];
  currentUser: User;
  currentRoomName?: string;
}

export const MessageList = ({
  messages,
  currentUser,
  currentRoomName,
}: MessageListProps) => {
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const formatTime = (timestamp: string) => {
    return new Date(timestamp).toLocaleTimeString("ko-KR", {
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const formatMessageContent = (content: string) => {
    // 간단한 URL 링크 처리
    const urlRegex = /(https?:\/\/[^\s]+)/g;
    return content.replace(
      urlRegex,
      '<a href="$1" target="_blank" rel="noopener noreferrer" class="text-blue-600 hover:underline">$1</a>'
    );
  };

  if (messages.length === 0) {
    return (
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="text-center">
          <ChatBubbleLeftRightIcon className="w-16 h-16 mx-auto mb-4 text-gray-300" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            {currentRoomName
              ? `${currentRoomName}에 오신 것을 환영합니다!`
              : "CloudClub Chat에 오신 것을 환영합니다!"}
          </h3>
          <p className="text-gray-600">
            첫 번째 메시지를 보내서 대화를 시작해보세요.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto p-4 space-y-4">
      {messages.map((message, index) => {
        const isOwn = message.user_id === currentUser.id;
        const isSystem = message.user_id === "system";
        const showAvatar =
          !isOwn &&
          !isSystem &&
          (index === 0 ||
            messages[index - 1].user_id !== message.user_id ||
            new Date(message.timestamp).getTime() -
              new Date(messages[index - 1].timestamp).getTime() >
              5 * 60 * 1000); // 5분 간격

        if (isSystem) {
          return (
            <div key={message.id} className="flex justify-center">
              <div className="bg-gray-100 text-gray-600 px-3 py-1 rounded-full text-sm">
                {message.content}
              </div>
            </div>
          );
        }

        return (
          <div
            key={message.id}
            className={clsx(
              "flex items-end space-x-2",
              isOwn ? "justify-end" : "justify-start"
            )}
          >
            {/* 상대방 아바타 */}
            {!isOwn && (
              <div
                className={clsx(
                  "w-8 h-8 rounded-full bg-gray-300 flex items-center justify-center flex-shrink-0",
                  showAvatar ? "opacity-100" : "opacity-0"
                )}
              >
                <span className="text-xs font-medium text-gray-700">
                  {message.user_name.charAt(0).toUpperCase()}
                </span>
              </div>
            )}

            <div
              className={clsx(
                "max-w-xs lg:max-w-md xl:max-w-lg",
                isOwn ? "order-1" : "order-2"
              )}
            >
              {/* 사용자명과 시간 (상대방 메시지만) */}
              {!isOwn && showAvatar && (
                <div className="flex items-center space-x-2 mb-1">
                  <span className="text-sm font-medium text-gray-900">
                    {message.user_name}
                  </span>
                  <span className="text-xs text-gray-500">
                    {formatTime(message.timestamp)}
                  </span>
                </div>
              )}

              {/* 메시지 내용 */}
              <div
                className={clsx(
                  "px-4 py-2 rounded-2xl",
                  isOwn
                    ? "bg-indigo-600 text-white rounded-br-md"
                    : "bg-gray-100 text-gray-900 rounded-bl-md"
                )}
              >
                <p
                  className="text-sm whitespace-pre-wrap break-words"
                  dangerouslySetInnerHTML={{
                    __html: formatMessageContent(message.content),
                  }}
                />
              </div>

              {/* 내 메시지 시간 */}
              {isOwn && (
                <div className="text-right mt-1">
                  <span className="text-xs text-gray-500">
                    {formatTime(message.timestamp)}
                  </span>
                </div>
              )}
            </div>
          </div>
        );
      })}
      <div ref={messagesEndRef} />
    </div>
  );
};
