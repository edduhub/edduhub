import type { Notification } from "./types";

export type NotificationAPI = {
  id: number;
  userId?: number | string;
  user_id?: number | string;
  title?: string;
  message?: string;
  type?: Notification["type"];
  category?: string;
  isRead?: boolean;
  is_read?: boolean;
  actionUrl?: string;
  action_url?: string;
  metadata?: Notification["metadata"];
  createdAt?: string;
  created_at?: string;
};

export type NotificationEnvelope = {
  type?: string;
  notification?: NotificationAPI;
  data?: unknown;
  timestamp?: string;
  user_id?: number;
  college_id?: number;
};

const toISODateString = (value?: string): string => value || new Date().toISOString();

export function normalizeNotification(item: NotificationAPI): Notification {
  return {
    id: item.id,
    userId: String(item.userId ?? item.user_id ?? ""),
    title: item.title ?? "",
    message: item.message ?? "",
    type: item.type ?? "info",
    category: item.category ?? item.type ?? "general",
    isRead: item.isRead ?? item.is_read ?? false,
    actionUrl: item.actionUrl ?? item.action_url,
    metadata: item.metadata,
    createdAt: toISODateString(item.createdAt ?? item.created_at),
  };
}

export function parseNotificationEnvelope(payload: string): Notification | null {
  const parsed = JSON.parse(payload) as NotificationEnvelope;
  if (parsed.type !== "notification" || !parsed.notification) {
    return null;
  }

  return normalizeNotification(parsed.notification);
}
