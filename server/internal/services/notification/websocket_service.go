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
	
	// Enhanced features
	BroadcastTypingIndicator(ctx context.Context, collegeID, userID int, isTyping bool) error
	BroadcastPresence(ctx context.Context, collegeID, userID int, status string) error
	GetConnectionStats() map[string]interface{}
}

type websocketService struct {
	clients          map[int]map[int]*websocket.Conn // collegeID -> userID -> connection
	clientsMutex     sync.RWMutex
	connectionTimes  map[int]map[int]time.Time // Track connection times
	notificationRepo repository.NotificationRepository
	upgrader         websocket.Upgrader
	allowedOrigins   map[string]bool // Allowed origins for CORS
}

type WebSocketMessage struct {
	Type         string               `json:"type"` // notification, typing, presence, ping, pong, connected
	Notification *models.Notification `json:"notification,omitempty"`
	Data         interface{}          `json:"data,omitempty"`
	Timestamp    time.Time            `json:"timestamp"`
	UserID       int                  `json:"user_id,omitempty"`
	CollegeID    int                  `json:"college_id,omitempty"`
}

func NewWebSocketService(notificationRepo repository.NotificationRepository, allowedOrigins []string) WebSocketService {
	// Create allowed origins map for fast lookup
	originsMap := make(map[string]bool)
	for _, origin := range allowedOrigins {
		originsMap[origin] = true
	}

	// Add localhost origins for development if not already included
	if len(originsMap) == 0 {
		originsMap["http://localhost:3000"] = true
		originsMap["http://localhost:8080"] = true
	}

	ws := &websocketService{
		clients:         make(map[int]map[int]*websocket.Conn),
		connectionTimes: make(map[int]map[int]time.Time),
		notificationRepo: notificationRepo,
		allowedOrigins:   originsMap,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				// If no origin header, reject (security best practice)
				if origin == "" {
					return false
				}
				// Check if origin is in allowed list
				return originsMap[origin]
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}

	// Start heartbeat monitor
	go ws.heartbeatMonitor()

	return ws
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
	collegeID := c.Get("college_id")
	if collegeID == nil {
		return 0, fmt.Errorf("college ID not found in context")
	}

	if id, ok := collegeID.(int); ok {
		return id, nil
	}

	return 0, fmt.Errorf("invalid college ID type")
}

// BroadcastTypingIndicator sends typing indicator to users
func (s *websocketService) BroadcastTypingIndicator(ctx context.Context, collegeID, userID int, isTyping bool) error {
	message := WebSocketMessage{
		Type:      "typing",
		UserID:    userID,
		Data:      map[string]bool{"is_typing": isTyping},
		Timestamp: time.Now(),
	}
	
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	
	if collegeClients, exists := s.clients[collegeID]; exists {
		messageBytes, _ := json.Marshal(message)
		for uid, conn := range collegeClients {
			if uid != userID { // Don't send to self
				go conn.WriteMessage(websocket.TextMessage, messageBytes)
			}
		}
	}
	
	return nil
}

// BroadcastPresence broadcasts user presence status (online, away, offline)
func (s *websocketService) BroadcastPresence(ctx context.Context, collegeID, userID int, status string) error {
	message := WebSocketMessage{
		Type:      "presence",
		UserID:    userID,
		Data:      map[string]string{"status": status},
		Timestamp: time.Now(),
	}
	
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	
	if collegeClients, exists := s.clients[collegeID]; exists {
		messageBytes, _ := json.Marshal(message)
		for _, conn := range collegeClients {
			go conn.WriteMessage(websocket.TextMessage, messageBytes)
		}
	}
	
	return nil
}

// GetConnectionStats returns statistics about active connections
func (s *websocketService) GetConnectionStats() map[string]interface{} {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()
	
	totalConnections := 0
	collegeStats := make(map[int]int)
	
	for collegeID, clients := range s.clients {
		count := len(clients)
		totalConnections += count
		collegeStats[collegeID] = count
	}
	
	return map[string]interface{}{
		"total_connections": totalConnections,
		"colleges":          len(s.clients),
		"college_stats":     collegeStats,
	}
}

// heartbeatMonitor sends periodic ping messages to keep connections alive
func (s *websocketService) heartbeatMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		s.clientsMutex.RLock()
		for collegeID, clients := range s.clients {
			for userID, conn := range clients {
				go func(conn *websocket.Conn, cID, uID int) {
					pingMsg := WebSocketMessage{
						Type:      "ping",
						Timestamp: time.Now(),
					}
					if err := conn.WriteJSON(pingMsg); err != nil {
						log.Printf("Heartbeat failed for user %d in college %d: %v", uID, cID, err)
						// Connection is probably dead, will be cleaned up on next operation
					}
				}(conn, collegeID, userID)
			}
		}
		s.clientsMutex.RUnlock()
	}
}