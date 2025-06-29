"use client";

import { PaperAirplaneIcon } from "@heroicons/react/24/outline";
import { KeyboardEvent, useRef, useState } from "react";

interface MessageInputProps {
  onSendMessage: (content: string) => void;
  disabled?: boolean;
  placeholder?: string;
}

export const MessageInput = ({
  onSendMessage,
  disabled = false,
  placeholder = "메시지를 입력하세요...",
}: MessageInputProps) => {
  const [message, setMessage] = useState("");
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    sendMessage();
  };

  const sendMessage = () => {
    const trimmedMessage = message.trim();
    if (!trimmedMessage || disabled) return;

    onSendMessage(trimmedMessage);
    setMessage("");

    // 텍스트 영역 높이 리셋
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
    }

    // 포커스 유지
    textareaRef.current?.focus();
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter") {
      if (e.shiftKey) {
        // Shift + Enter는 줄바꿈
        return;
      } else {
        // Enter만 누르면 메시지 전송
        e.preventDefault();
        sendMessage();
      }
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setMessage(e.target.value);

    // 자동 높이 조절
    const textarea = e.target;
    textarea.style.height = "auto";
    textarea.style.height = Math.min(textarea.scrollHeight, 120) + "px";
  };

  return (
    <div className="border-t border-gray-200 p-4 bg-white">
      <form onSubmit={handleSubmit} className="flex items-end space-x-3">
        <div className="flex-1 relative">
          <textarea
            ref={textareaRef}
            value={message}
            onChange={handleInputChange}
            onKeyDown={handleKeyDown}
            placeholder={disabled ? "채팅방을 선택하세요..." : placeholder}
            disabled={disabled}
            rows={1}
            className="w-full px-4 py-3 pr-12 border border-gray-300 rounded-lg resize-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 disabled:bg-gray-100 disabled:cursor-not-allowed transition-colors text-gray-900 placeholder-gray-500"
            style={{ maxHeight: "120px" }}
          />
          <div className="absolute bottom-2 right-2 text-xs text-gray-400">
            {message.length > 0 && (
              <span className={message.length > 1000 ? "text-red-500" : ""}>
                {message.length}/1000
              </span>
            )}
          </div>
        </div>

        <button
          type="submit"
          disabled={disabled || !message.trim() || message.length > 1000}
          className="p-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors flex-shrink-0"
          title="메시지 전송 (Enter)"
        >
          <PaperAirplaneIcon className="w-5 h-5" />
        </button>
      </form>

      <div className="mt-2 text-xs text-gray-500 text-center">
        <span className="inline-block">Enter로 전송, Shift+Enter로 줄바꿈</span>
      </div>
    </div>
  );
};
