"use client";

import { config } from "@/config/environment";
import { Message, Room, User, WebSocketMessage } from "@/types";
import { useCallback, useEffect, useState } from "react";
import { useWebSocket } from "./useWebSocket";

export const useChat = () => {
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [currentRoom, setCurrentRoom] = useState<Room | null>(null);
  const [rooms, setRooms] = useState<Room[]>([]);
  const [messages, setMessages] = useState<Map<string, Message[]>>(new Map());
  const [isLoading, setIsLoading] = useState(true);

  const handleWebSocketMessage = useCallback(
    (wsMessage: WebSocketMessage) => {
      console.log("Received WebSocket message:", wsMessage);

      switch (wsMessage.type) {
        case "message":
          if (wsMessage.payload && "content" in wsMessage.payload) {
            const message = wsMessage.payload as Message;

            // 다른 사용자의 메시지이거나, 자신의 메시지여도 로컬에 없는 경우 추가
            setMessages((prev) => {
              const roomMessages = prev.get(message.room_id) || [];

              // 이미 존재하는 메시지인지 확인 (ID 기준)
              if (roomMessages.find((m) => m.id === message.id)) {
                console.log("Message already exists, skipping:", message.id);
                return prev;
              }

              // 자신의 메시지인 경우 Optimistic Update된 메시지를 서버 메시지로 교체
              if (currentUser && message.user_id === currentUser.id) {
                console.log(
                  "Replacing optimistic message with server message:",
                  message.id
                );
                // Optimistic 메시지를 찾아서 교체 (같은 내용과 시간대의 메시지 찾기)
                const optimisticMsgIndex = roomMessages.findIndex(
                  (m) =>
                    m.user_id === currentUser.id &&
                    m.content === message.content &&
                    Math.abs(
                      new Date(m.timestamp).getTime() -
                        new Date(message.timestamp).getTime()
                    ) < 5000 // 5초 이내
                );

                if (optimisticMsgIndex !== -1) {
                  // 기존 Optimistic 메시지를 서버 메시지로 교체
                  const newRoomMessages = [...roomMessages];
                  newRoomMessages[optimisticMsgIndex] = message;
                  const newMessages = new Map(prev);
                  newMessages.set(message.room_id, newRoomMessages);
                  return newMessages;
                } else {
                  // Optimistic 메시지를 찾지 못한 경우 그냥 추가
                  console.log(
                    "Optimistic message not found, adding server message"
                  );
                }
              }

              // 새 메시지 추가
              const newMessages = new Map(prev);
              newMessages.set(message.room_id, [...roomMessages, message]);
              return newMessages;
            });
          }
          break;
        case "room_joined":
          // 채팅방 입장 성공 처리
          console.log("Successfully joined room");
          break;
        case "room_left":
          // 채팅방 퇴장 성공 처리
          console.log("Successfully left room");
          break;
        case "user_joined":
          if (
            wsMessage.payload &&
            "user_name" in wsMessage.payload &&
            "room_id" in wsMessage.payload
          ) {
            const payload = wsMessage.payload as {
              user_name: string;
              room_id: string;
            };
            // 시스템 메시지 직접 추가
            const systemMessage: Message = {
              id:
                Math.random().toString(36).substr(2, 9) +
                Date.now().toString(36),
              content: `${payload.user_name}님이 입장하셨습니다`,
              user_id: "system",
              user_name: "시스템",
              room_id: payload.room_id,
              timestamp: new Date().toISOString(),
            };
            setMessages((prev) => {
              const roomMessages = prev.get(payload.room_id) || [];
              const newMessages = new Map(prev);
              newMessages.set(payload.room_id, [
                ...roomMessages,
                systemMessage,
              ]);
              return newMessages;
            });
          }
          break;
        case "user_left":
          if (
            wsMessage.payload &&
            "user_name" in wsMessage.payload &&
            "room_id" in wsMessage.payload
          ) {
            const payload = wsMessage.payload as {
              user_name: string;
              room_id: string;
            };
            // 시스템 메시지 직접 추가
            const systemMessage: Message = {
              id:
                Math.random().toString(36).substr(2, 9) +
                Date.now().toString(36),
              content: `${payload.user_name}님이 퇴장하셨습니다`,
              user_id: "system",
              user_name: "시스템",
              room_id: payload.room_id,
              timestamp: new Date().toISOString(),
            };
            setMessages((prev) => {
              const roomMessages = prev.get(payload.room_id) || [];
              const newMessages = new Map(prev);
              newMessages.set(payload.room_id, [
                ...roomMessages,
                systemMessage,
              ]);
              return newMessages;
            });
          }
          break;
        case "room_created":
          if (wsMessage.payload && "id" in wsMessage.payload) {
            // 방 목록 새로고침 트리거
            setRooms((prev) => [...prev]);
          }
          break;
        case "error":
          if (wsMessage.payload && "message" in wsMessage.payload) {
            console.error("WebSocket error:", wsMessage.payload.message);
          }
          break;
        default:
          console.warn("Unknown message type:", wsMessage.type);
      }
    },
    [currentUser]
  );

  const {
    connectionStatus,
    sendMessage: sendWebSocketMessage,
    hasConnectedOnce,
  } = useWebSocket({
    url: config.WS_URL,
    onMessage: handleWebSocketMessage,
    enabled: !!currentUser, // 사용자가 로그인한 후에만 연결
  });

  const generateId = useCallback(() => {
    return Math.random().toString(36).substr(2, 9) + Date.now().toString(36);
  }, []);

  const addMessage = useCallback((message: Message) => {
    setMessages((prev) => {
      const roomMessages = prev.get(message.room_id) || [];

      // 중복 메시지 방지
      if (roomMessages.find((m) => m.id === message.id)) {
        return prev;
      }

      const newMessages = new Map(prev);
      newMessages.set(message.room_id, [...roomMessages, message]);
      return newMessages;
    });
  }, []);

  const loadRooms = useCallback(async () => {
    try {
      const response = await fetch(`${config.API_BASE_URL}/api/rooms`);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const roomsData = await response.json();
      setRooms(
        roomsData.map((room: Room) => ({
          id: room.id,
          name: room.name,
          description: room.description,
          user_count: room.user_count || 0,
          created_at: room.created_at,
        }))
      );
    } catch (error) {
      console.error("Failed to load rooms:", error);
    }
  }, []);

  const joinRoom = useCallback(
    async (roomId: string) => {
      const room = rooms.find((r) => r.id === roomId);
      if (!room || !currentUser) return;

      // 이전 채팅방에서 나가기
      if (currentRoom && currentRoom.id !== roomId) {
        sendWebSocketMessage({
          type: "leave_room",
          payload: { room_id: currentRoom.id },
        });
      }

      setCurrentRoom(room);

      // 메시지 기록 초기화 (필요시)
      if (!messages.has(roomId)) {
        setMessages((prev) => {
          const newMessages = new Map(prev);
          newMessages.set(roomId, []);
          return newMessages;
        });
      }

      // WebSocket으로 입장 알림
      sendWebSocketMessage({
        type: "join_room",
        payload: {
          room_id: roomId,
          user_id: currentUser.id,
          user_name: currentUser.name,
        },
      });
    },
    [rooms, currentUser, currentRoom, messages, sendWebSocketMessage]
  );

  const sendMessage = useCallback(
    async (content: string) => {
      if (!currentRoom || !currentUser || !content.trim()) return;

      // 고유한 메시지 ID 생성
      const messageId = generateId();

      // 메시지를 즉시 로컬에 추가 (Optimistic Update)
      const localMessage: Message = {
        id: messageId,
        content: content.trim(),
        user_id: currentUser.id,
        user_name: currentUser.name,
        room_id: currentRoom.id,
        timestamp: new Date().toISOString(),
      };
      addMessage(localMessage);

      // 서버로 메시지 전송
      sendWebSocketMessage({
        type: "send_message",
        payload: {
          id: messageId, // 고유 ID 포함
          content: content.trim(),
        },
      });
    },
    [currentRoom, currentUser, sendWebSocketMessage, generateId, addMessage]
  );

  const createRoom = useCallback(
    async (name: string, description: string) => {
      try {
        const response = await fetch(`${config.API_BASE_URL}/api/rooms`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ name, description }),
        });

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const newRoom = await response.json();
        loadRooms(); // 방 목록 새로고침
        return newRoom;
      } catch (error) {
        console.error("Failed to create room:", error);
        throw error;
      }
    },
    [loadRooms]
  );

  const login = useCallback(
    (username: string) => {
      const user: User = {
        id: generateId(),
        name: username,
      };
      setCurrentUser(user);
      setIsLoading(false);
    },
    [generateId]
  );

  const logout = useCallback(() => {
    setCurrentUser(null);
    setCurrentRoom(null);
    setRooms([]);
    setMessages(new Map());
    setIsLoading(true);
  }, []);

  useEffect(() => {
    if (currentUser) {
      loadRooms();
    }
  }, [currentUser, loadRooms]);

  return {
    // State
    currentUser,
    currentRoom,
    rooms,
    messages: currentRoom ? messages.get(currentRoom.id) || [] : [],
    allMessages: messages,
    connectionStatus,
    isLoading,
    hasConnectedOnce,

    // Actions
    login,
    logout,
    joinRoom,
    sendMessage,
    createRoom,
    loadRooms,
  };
};
