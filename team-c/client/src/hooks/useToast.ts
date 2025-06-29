"use client";

import { ToastType } from "@/types";
import { useCallback, useState } from "react";

interface Toast {
  id: string;
  message: string;
  type: ToastType;
}

export const useToast = () => {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const generateId = useCallback(() => {
    return Math.random().toString(36).substr(2, 9) + Date.now().toString(36);
  }, []);

  const addToast = useCallback(
    (message: string, type: ToastType = "info") => {
      const id = generateId();
      const newToast: Toast = { id, message, type };

      setToasts((prev) => [...prev, newToast]);

      return id;
    },
    [generateId]
  );

  const removeToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((toast) => toast.id !== id));
  }, []);

  const clearAllToasts = useCallback(() => {
    setToasts([]);
  }, []);

  const showSuccess = useCallback(
    (message: string) => {
      return addToast(message, "success");
    },
    [addToast]
  );

  const showError = useCallback(
    (message: string) => {
      return addToast(message, "error");
    },
    [addToast]
  );

  const showWarning = useCallback(
    (message: string) => {
      return addToast(message, "warning");
    },
    [addToast]
  );

  const showInfo = useCallback(
    (message: string) => {
      return addToast(message, "info");
    },
    [addToast]
  );

  return {
    toasts,
    addToast,
    removeToast,
    clearAllToasts,
    showSuccess,
    showError,
    showWarning,
    showInfo,
  };
};
