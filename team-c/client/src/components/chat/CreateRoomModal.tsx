"use client";

import { Modal } from "@/components/ui/Modal";
import { PlusIcon } from "@heroicons/react/24/outline";
import { useState } from "react";

interface CreateRoomModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreateRoom: (name: string, description: string) => Promise<void>;
}

export const CreateRoomModal = ({
  isOpen,
  onClose,
  onCreateRoom,
}: CreateRoomModalProps) => {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const trimmedName = name.trim();

    if (!trimmedName) {
      alert("채팅방 이름을 입력해주세요");
      return;
    }

    setIsLoading(true);

    try {
      await onCreateRoom(trimmedName, description.trim());
      setName("");
      setDescription("");
      onClose();
    } catch (error) {
      console.error("Failed to create room:", error);
      alert("채팅방 생성에 실패했습니다");
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    if (!isLoading) {
      setName("");
      setDescription("");
      onClose();
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="새 채팅방">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label
            htmlFor="room-name"
            className="block text-sm font-medium text-gray-700 mb-2"
          >
            채팅방 이름 *
          </label>
          <input
            type="text"
            id="room-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="채팅방 이름을 입력하세요"
            className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-colors text-gray-900 placeholder-gray-500"
            required
            disabled={isLoading}
            autoFocus
          />
        </div>

        <div>
          <label
            htmlFor="room-description"
            className="block text-sm font-medium text-gray-700 mb-2"
          >
            설명 (선택사항)
          </label>
          <textarea
            id="room-description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="채팅방 설명을 입력하세요"
            rows={3}
            className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 transition-colors resize-none text-gray-900 placeholder-gray-500"
            disabled={isLoading}
          />
        </div>

        <div className="flex space-x-3 pt-4">
          <button
            type="button"
            onClick={handleClose}
            disabled={isLoading}
            className="flex-1 bg-gray-100 text-gray-700 py-3 px-4 rounded-lg font-medium hover:bg-gray-200 focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            취소
          </button>
          <button
            type="submit"
            disabled={isLoading || !name.trim()}
            className="flex-1 bg-indigo-600 text-white py-3 px-4 rounded-lg font-medium hover:bg-indigo-700 focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center"
          >
            {isLoading ? (
              <div className="flex items-center">
                <div className="animate-spin w-5 h-5 border-2 border-white border-t-transparent rounded-full mr-2" />
                생성 중...
              </div>
            ) : (
              <>
                <PlusIcon className="w-5 h-5 mr-2" />
                생성하기
              </>
            )}
          </button>
        </div>
      </form>
    </Modal>
  );
};
