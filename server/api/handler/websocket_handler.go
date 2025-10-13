package handler

import (
	"eduhub/server/internal/services/notification"

	"github.com/labstack/echo/v4"
)

type WebSocketHandler struct {
	websocketService notification.WebSocketService
}

func NewWebSocketHandler(websocketService notification.WebSocketService) *WebSocketHandler {
	return &WebSocketHandler{
		websocketService: websocketService,
	}
}

// HandleWebSocket upgrades HTTP connection to WebSocket and handles real-time communication
func (h *WebSocketHandler) HandleWebSocket(c echo.Context) error {
	return h.websocketService.HandleWebSocket(c)
}