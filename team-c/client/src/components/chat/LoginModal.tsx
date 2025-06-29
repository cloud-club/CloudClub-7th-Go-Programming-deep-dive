"use client";

import { Modal } from "@/components/ui/Modal";
import { UserIcon } from "@heroicons/react/24/outline";
import { useState } from "react";

interface LoginModalProps {
  isOpen: boolean;
  onLogin: (username: string) => void;
}

export const LoginModal = ({ isOpen, onLogin }: LoginModalProps) => {
  const [username, setUsername] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const trimmedUsername = username.trim();

    if (!trimmedUsername) {
      alert("사용자명을 입력해주세요");
      return;
    }

    if (trimmedUsername.length < 2 || trimmedUsername.length > 20) {
      alert("사용자명은 2-20자 사이여야 합니다");
      return;
    }

    setIsLoading(true);

    try {
      onLogin(trimmedUsername);
      setUsername("");
    } catch (error) {
      console.error("Login failed:", error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={() => {}} // 로그인 모달은 강제로 닫을 수 없음
      showCloseButton={false}
    >
      <div className="text-center mb-6">
        <div className="mx-auto w-16 h-16 bg-indigo-100 rounded-full flex items-center justify-center mb-4">
          <UserIcon className="w-8 h-8 text-indigo-600" />
        </div>
        <h2 className="text-2xl font-bold text-gray-900 mb-2">
          CloudClub Chat
        </h2>
        <p className="text-gray-600">채팅을 시작하려면 사용자명을 입력하세요</p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label
            htmlFor="username"
            className="block text-sm font-medium text-gray-700 mb-2"
          >
            사용자명
          </label>
          <input
            type="text"
            id="username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            placeholder="닉네임을 입력하세요"
            className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-colors text-gray-900 placeholder-gray-500"
            required
            minLength={2}
            maxLength={20}
            disabled={isLoading}
            autoFocus
          />
        </div>

        <button
          type="submit"
          disabled={isLoading || !username.trim()}
          className="w-full bg-indigo-600 text-white py-3 px-4 rounded-lg font-medium hover:bg-indigo-700 focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {isLoading ? (
            <div className="flex items-center justify-center">
              <div className="animate-spin w-5 h-5 border-2 border-white border-t-transparent rounded-full mr-2" />
              로그인 중...
            </div>
          ) : (
            "채팅 시작하기"
          )}
        </button>
      </form>
    </Modal>
  );
};
