export const config = {
  API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080",
  WS_URL: process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080/ws",
} as const;
