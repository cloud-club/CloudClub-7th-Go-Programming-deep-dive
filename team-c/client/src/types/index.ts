export interface User {
  id: string;
  name: string;
}

export interface Room {
  id: string;
  name: string;
  description: string;
  user_count: number;
  created_at: string;
}

export interface Message {
  id: string;
  content: string;
  user_id: string;
  user_name: string;
  room_id: string;
  timestamp: string;
}

export interface WebSocketMessage {
  type:
    | "message"
    | "send_message"
    | "join_room"
    | "leave_room"
    | "room_joined"
    | "room_left"
    | "user_joined"
    | "user_left"
    | "room_created"
    | "error";
  payload?:
    | Message
    | Room
    | User
    | {
        content?: string;
        message?: string;
        room_id?: string;
        user_id?: string;
        user_name?: string;
        [key: string]: unknown;
      };
}

export type ConnectionStatus =
  | "connected"
  | "disconnected"
  | "connecting"
  | "error";

export type ToastType = "success" | "error" | "warning" | "info";
