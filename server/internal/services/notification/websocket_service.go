package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type WebSocketService interface {
	HandleWebSocket(c echo.Context) error
	BroadcastNotification(ctx context.Context, collegeID int, notification *models.Notification) error
	BroadcastToUser(ctx context.Context, collegeID, userID int, notification *models.Notification) error
	BroadcastToUsers(ctx context.Context, collegeID int, userIDs []int, notification *models.Notification) error
	GetConnectedUsers(collegeID int) []int
}

type websocketService struct {
	clients         map[int]map[int]*websocket.Conn // collegeID -> userID -> connection
	clientsMutex    sync.RWMutex
	notificationRepo repository.NotificationRepository
	upgrader        websocket.Upgrader
}

type WebSocketMessage struct {
	Type         string      `json:"type"`
	Notification *models.Notification `json:"notification,omitempty"`
	Data         interface{} `json:"data,omitempty"`
	Timestamp    time.Time   `json:"timestamp"`
}

func NewWebSocketService(notificationRepo repository.NotificationRepository) WebSocketService {
	return &websocketService{
		clients:         make(map[int]map[int]*websocket.Conn),
		notificationRepo: notificationRepo,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (s *websocketService) HandleWebSocket(c echo.Context) error {
	userID, err := extractUserIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	collegeID, err := extractCollegeIDFromContext(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "college ID required"})
	}

	conn, err := s.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return err
	}

	// Register client
	s.clientsMutex.Lock()
	if s.clients[collegeID] == nil {
		s.clients[collegeID] = make(map[int]*websocket.Conn)
	}
	s.clients[collegeID][userID] = conn
	s.clientsMutex.Unlock()

	defer func() {
		s.clientsMutex.Lock()
		delete(s.clients[collegeID], userID)
		if len(s.clients[collegeID]) == 0 {
			delete(s.clients, collegeID)
		}
		s.clientsMutex.Unlock()
		conn.Close()
	}()

	// Send initial connection confirmation
	welcomeMsg := WebSocketMessage{
		Type:      "connected",
		Data:      map[string]string{"message": "Connected to EduHub notifications"},
		Timestamp: time.Now(),
	}
	if err := conn.WriteJSON(welcomeMsg); err != nil {
		log.Printf("Failed to send welcome message: %v", err)
		return err
	}

	// Keep connection alive and handle incoming messages
	for {
		var msg WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle ping messages to keep connection alive
		if msg.Type == "ping" {
			pongMsg := WebSocketMessage{
				Type:      "pong",
				Timestamp: time.Now(),
			}
			if err := conn.WriteJSON(pongMsg); err != nil {
				log.Printf("Failed to send pong: %v", err)
				break
			}
		}
	}

	return nil
}

func (s *websocketService) BroadcastNotification(ctx context.Context, collegeID int, notification *models.Notification) error {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	if collegeClients, exists := s.clients[collegeID]; exists {
		message := WebSocketMessage{
			Type:         "notification",
			Notification: notification,
			Timestamp:    time.Now(),
		}

		messageBytes, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		for userID, conn := range collegeClients {
			go func(conn *websocket.Conn, userID int) {
				if err := conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
					log.Printf("Failed to send notification to user %d: %v", userID, err)
					// Connection might be dead, but we'll clean it up on next operation
				}
			}(conn, userID)
		}
	}

	return nil
}

func (s *websocketService) BroadcastToUser(ctx context.Context, collegeID, userID int, notification *models.Notification) error {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	if collegeClients, exists := s.clients[collegeID]; exists {
		if conn, userExists := collegeClients[userID]; userExists {
			message := WebSocketMessage{
				Type:         "notification",
				Notification: notification,
				Timestamp:    time.Now(),
			}

			if err := conn.WriteJSON(message); err != nil {
				log.Printf("Failed to send notification to user %d: %v", userID, err)
				return fmt.Errorf("failed to send notification: %w", err)
			}
		}
	}

	return nil
}

func (s *websocketService) BroadcastToUsers(ctx context.Context, collegeID int, userIDs []int, notification *models.Notification) error {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	if collegeClients, exists := s.clients[collegeID]; exists {
		message := WebSocketMessage{
			Type:         "notification",
			Notification: notification,
			Timestamp:    time.Now(),
		}

		messageBytes, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		for _, userID := range userIDs {
			if conn, userExists := collegeClients[userID]; userExists {
				go func(conn *websocket.Conn, userID int) {
					if err := conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
						log.Printf("Failed to send notification to user %d: %v", userID, err)
					}
				}(conn, userID)
			}
		}
	}

	return nil
}

func (s *websocketService) GetConnectedUsers(collegeID int) []int {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	var userIDs []int
	if collegeClients, exists := s.clients[collegeID]; exists {
		for userID := range collegeClients {
			userIDs = append(userIDs, userID)
		}
	}

	return userIDs
}

// Helper functions to extract user and college ID from echo context
// These should be implemented based on your existing middleware
func extractUserIDFromContext(c echo.Context) (int, error) {
	// This should match your existing user ID extraction logic
	userID := c.Get("userID")
	if userID == nil {
		return 0, fmt.Errorf("user ID not found in context")
	}

	if id, ok := userID.(int); ok {
		return id, nil
	}

	return 0, fmt.Errorf("invalid user ID type")
}

func extractCollegeIDFromContext(c echo.Context) (int, error) {
	// This should match your existing college ID extraction logic
	collegeID := c.Get("collegeID")
	if collegeID == nil {
		return 0, fmt.Errorf("college ID not found in context")
	}

	if id, ok := collegeID.(int); ok {
		return id, nil
	}

	return 0, fmt.Errorf("invalid college ID type")
}