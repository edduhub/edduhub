import { normalizeNotification, parseNotificationEnvelope } from "@/lib/notifications";

describe("notification normalization", () => {
  it("normalizes snake_case API fields", () => {
    const notification = normalizeNotification({
      id: 7,
      user_id: 42,
      title: "Demo",
      message: "Hello",
      type: "success",
      is_read: true,
      created_at: "2026-03-07T00:00:00Z",
    });

    expect(notification).toEqual({
      id: 7,
      userId: "42",
      title: "Demo",
      message: "Hello",
      type: "success",
      category: "success",
      isRead: true,
      actionUrl: undefined,
      metadata: undefined,
      createdAt: "2026-03-07T00:00:00Z",
    });
  });

  it("parses websocket notification envelopes and ignores control frames", () => {
    expect(
      parseNotificationEnvelope(
        JSON.stringify({
          type: "notification",
          notification: {
            id: 8,
            user_id: 99,
            title: "Update",
            message: "Body",
            type: "info",
            created_at: "2026-03-07T00:00:00Z",
          },
        })
      )
    ).toMatchObject({
      id: 8,
      userId: "99",
      title: "Update",
      isRead: false,
    });

    expect(parseNotificationEnvelope(JSON.stringify({ type: "connected", data: { ok: true } }))).toBeNull();
    expect(parseNotificationEnvelope(JSON.stringify({ type: "pong" }))).toBeNull();
  });
});
