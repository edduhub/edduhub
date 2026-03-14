package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
	"eduhub/server/internal/services/auth"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type WebSocketService interface {
	HandleWebSocket(c echo.Context) error
	BroadcastNotification(ctx context.Context, collegeID int, notification *models.Notification) error
	BroadcastToUser(ctx context.Context, collegeID, userID int, notification *models.Notification) error
	BroadcastToUsers(ctx context.Context, collegeID int, userIDs []int, notification *models.Notification) error
	GetConnectedUsers(collegeID int) []int

	BroadcastTypingIndicator(ctx context.Context, collegeID, userID int, isTyping bool) error
	BroadcastPresence(ctx context.Context, collegeID, userID int, status string) error
	GetConnectionStats() map[string]any
	Stop()
}

type broadcastTask struct {
	conn    *websocket.Conn
	userID  int
	message []byte
}

type websocketService struct {
	clients          map[int]map[int]*websocket.Conn
	clientsMutex     sync.RWMutex
	connectionTimes  map[int]map[int]time.Time
	notificationRepo repository.NotificationRepository
	upgrader         websocket.Upgrader
	allowedOrigins   map[string]bool

	broadcastQueue chan broadcastTask
	workerPool     []chan struct{}
	stopChan       chan struct{}
	stopped        bool
	stopOnce       sync.Once
	logger         zerolog.Logger
}

type WebSocketMessage struct {
	Type         string               `json:"type"` // notification, typing, presence, ping, pong, connected
	Notification *models.Notification `json:"notification,omitempty"`
	Data         any                  `json:"data,omitempty"`
	Timestamp    time.Time            `json:"timestamp"`
	UserID       int                  `json:"user_id,omitempty"`
	CollegeID    int                  `json:"college_id,omitempty"`
}

func NewWebSocketService(notificationRepo repository.NotificationRepository, allowedOrigins []string) WebSocketService {
	originsMap := make(map[string]bool)
	for _, origin := range allowedOrigins {
		originsMap[origin] = true
	}

	if len(originsMap) == 0 {
		originsMap["http://localhost:3000"] = true
		originsMap["http://localhost:8080"] = true
	}

	numWorkers := 10
	broadcastQueue := make(chan broadcastTask, 1000)
	workerPool := make([]chan struct{}, numWorkers)
	for i := range numWorkers {
		workerPool[i] = make(chan struct{}, 1)
	}

	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}).With().Timestamp().Logger()

	ws := &websocketService{
		clients:          make(map[int]map[int]*websocket.Conn),
		connectionTimes:  make(map[int]map[int]time.Time),
		notificationRepo: notificationRepo,
		allowedOrigins:   originsMap,
		broadcastQueue:   broadcastQueue,
		workerPool:       workerPool,
		stopChan:         make(chan struct{}),
		stopped:          false,
		logger:           logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return false
				}
				return originsMap[origin]
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}

	for i := range numWorkers {
		go ws.broadcastWorker(i)
	}

	go ws.heartbeatMonitor()

	return ws
}

func (s *websocketService) HandleWebSocket(c echo.Context) error {
	identity, ok := c.Get("identity").(*auth.Identity)
	if !ok || identity == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized: identity not found"})
	}

	userID := identity.UserID
	if userID == 0 {
		if contextUserID, ok := c.Get("user_id").(int); ok {
			userID = contextUserID
		}
	}
	if userID == 0 {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized: user ID not found"})
	}

	collegeID, err := extractCollegeIDFromContext(c)
	if err != nil {
		s.logger.Error().Err(err).Msg("Missing resolved college ID in websocket context")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid college context"})
	}

	conn, err := s.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return err
	}

	s.clientsMutex.Lock()
	if s.clients[collegeID] == nil {
		s.clients[collegeID] = make(map[int]*websocket.Conn)
	}
	s.clients[collegeID][userID] = conn
	if s.connectionTimes[collegeID] == nil {
		s.connectionTimes[collegeID] = make(map[int]time.Time)
	}
	s.connectionTimes[collegeID][userID] = time.Now()
	s.clientsMutex.Unlock()

	s.logger.Info().Int("user_id", userID).Int("college_id", collegeID).Msg("WebSocket connection established")

	defer func() {
		s.clientsMutex.Lock()
		delete(s.clients[collegeID], userID)
		if len(s.clients[collegeID]) == 0 {
			delete(s.clients, collegeID)
			delete(s.connectionTimes, collegeID)
		} else {
			delete(s.connectionTimes[collegeID], userID)
		}
		s.clientsMutex.Unlock()
		conn.Close()
		s.logger.Info().Int("user_id", userID).Int("college_id", collegeID).Msg("WebSocket connection closed")
	}()

	welcomeMsg := newWebSocketMessage("connected", collegeID, userID, nil, map[string]string{"message": "Connected to EduHub notifications"})
	if err := conn.WriteJSON(welcomeMsg); err != nil {
		log.Printf("Failed to send welcome message: %v", err)
		return err
	}

	for {
		var msg WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if msg.Type == "ping" {
			pongMsg := newWebSocketMessage("pong", collegeID, userID, nil, nil)
			if err := conn.WriteJSON(pongMsg); err != nil {
				log.Printf("Failed to send pong: %v", err)
				break
			}
		}
	}

	return nil
}

func (s *websocketService) BroadcastNotification(ctx context.Context, collegeID int, notification *models.Notification) error {
	if s.stopped {
		return fmt.Errorf("websocket service is stopped")
	}

	s.clientsMutex.RLock()
	collegeClients, exists := s.clients[collegeID]
	s.clientsMutex.RUnlock()

	if !exists || len(collegeClients) == 0 {
		return nil
	}

	message := newWebSocketMessage("notification", collegeID, 0, notification, nil)

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	for userID, conn := range collegeClients {
		select {
		case s.broadcastQueue <- broadcastTask{conn: conn, userID: userID, message: messageBytes}:
		default:
			s.logger.Warn().Int("user_id", userID).Msg("Broadcast queue full, dropping message")
		}
	}

	return nil
}

func (s *websocketService) BroadcastToUser(ctx context.Context, collegeID, userID int, notification *models.Notification) error {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	if collegeClients, exists := s.clients[collegeID]; exists {
		if conn, userExists := collegeClients[userID]; userExists {
			message := newWebSocketMessage("notification", collegeID, userID, notification, nil)

			if err := conn.WriteJSON(message); err != nil {
				log.Printf("Failed to send notification to user %d: %v", userID, err)
				return fmt.Errorf("failed to send notification: %w", err)
			}
		}
	}

	return nil
}

func (s *websocketService) BroadcastToUsers(ctx context.Context, collegeID int, userIDs []int, notification *models.Notification) error {
	if s.stopped {
		return fmt.Errorf("websocket service is stopped")
	}

	s.clientsMutex.RLock()
	collegeClients, exists := s.clients[collegeID]
	s.clientsMutex.RUnlock()

	if !exists || len(collegeClients) == 0 {
		return nil
	}

	message := newWebSocketMessage("notification", collegeID, 0, notification, nil)

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	for _, userID := range userIDs {
		if conn, userExists := collegeClients[userID]; userExists {
			select {
			case s.broadcastQueue <- broadcastTask{conn: conn, userID: userID, message: messageBytes}:
			default:
				s.logger.Warn().Int("user_id", userID).Msg("Broadcast queue full, dropping message")
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

func (s *websocketService) BroadcastTypingIndicator(ctx context.Context, collegeID, userID int, isTyping bool) error {
	if s.stopped {
		return fmt.Errorf("websocket service is stopped")
	}

	message := newWebSocketMessage("typing", collegeID, userID, nil, map[string]bool{"is_typing": isTyping})

	s.clientsMutex.RLock()
	collegeClients, exists := s.clients[collegeID]
	s.clientsMutex.RUnlock()

	if !exists || len(collegeClients) == 0 {
		return nil
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal typing indicator: %w", err)
	}

	for uid, conn := range collegeClients {
		if uid != userID {
			select {
			case s.broadcastQueue <- broadcastTask{conn: conn, userID: uid, message: messageBytes}:
			default:
				s.logger.Warn().Int("user_id", uid).Msg("Broadcast queue full, dropping typing indicator")
			}
		}
	}

	return nil
}

func (s *websocketService) BroadcastPresence(ctx context.Context, collegeID, userID int, status string) error {
	if s.stopped {
		return fmt.Errorf("websocket service is stopped")
	}

	message := newWebSocketMessage("presence", collegeID, userID, nil, map[string]string{"status": status})

	s.clientsMutex.RLock()
	collegeClients, exists := s.clients[collegeID]
	s.clientsMutex.RUnlock()

	if !exists || len(collegeClients) == 0 {
		return nil
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal presence message: %w", err)
	}

	for uid, conn := range collegeClients {
		select {
		case s.broadcastQueue <- broadcastTask{conn: conn, userID: uid, message: messageBytes}:
		default:
			s.logger.Warn().Int("user_id", uid).Msg("Broadcast queue full, dropping presence update")
		}
	}

	return nil
}

// GetConnectionStats returns statistics about active connections
func (s *websocketService) GetConnectionStats() map[string]any {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	totalConnections := 0
	collegeStats := make(map[int]int)

	for collegeID, clients := range s.clients {
		count := len(clients)
		totalConnections += count
		collegeStats[collegeID] = count
	}

	return map[string]any{
		"total_connections": totalConnections,
		"colleges":          len(s.clients),
		"college_stats":     collegeStats,
	}
}

func (s *websocketService) broadcastWorker(workerID int) {
	for {
		select {
		case <-s.stopChan:
			s.logger.Debug().Int("worker_id", workerID).Msg("Broadcast worker stopping")
			return
		case task := <-s.broadcastQueue:
			select {
			case <-s.stopChan:
				s.logger.Debug().Int("worker_id", workerID).Msg("Broadcast worker stopping (draining queue)")
				return
			default:
			}

			if err := task.conn.WriteMessage(websocket.TextMessage, task.message); err != nil {
				log.Printf("Broadcast worker %d failed to send to user %d: %v", workerID, task.userID, err)
			}
		}
	}
}

func (s *websocketService) Stop() {
	s.stopOnce.Do(func() {
		s.logger.Info().Msg("Stopping WebSocket service...")
		s.stopped = true
		close(s.stopChan)

		s.clientsMutex.Lock()
		for collegeID, clients := range s.clients {
			for userID, conn := range clients {
				conn.Close()
				delete(s.clients[collegeID], userID)
			}
			delete(s.clients, collegeID)
		}
		s.clientsMutex.Unlock()

		s.logger.Info().Msg("WebSocket service stopped")
	})
}

func (s *websocketService) heartbeatMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			s.logger.Debug().Msg("Heartbeat monitor stopping")
			return
		case <-ticker.C:
			s.clientsMutex.RLock()
			for collegeID, clients := range s.clients {
				for userID, conn := range clients {
					go func(conn *websocket.Conn, cID, uID int) {
						pingMsg := newWebSocketMessage("ping", cID, uID, nil, nil)
						if err := conn.WriteJSON(pingMsg); err != nil {
							s.logger.Debug().Int("user_id", uID).Int("college_id", cID).Err(err).Msg("Heartbeat failed")
						}
					}(conn, collegeID, userID)
				}
			}
			s.clientsMutex.RUnlock()
		}
	}
}

func extractCollegeIDFromContext(c echo.Context) (int, error) {
	switch value := c.Get("college_id").(type) {
	case int:
		if value > 0 {
			return value, nil
		}
	case int32:
		if value > 0 {
			return int(value), nil
		}
	case int64:
		if value > 0 {
			return int(value), nil
		}
	case float64:
		if value > 0 {
			return int(value), nil
		}
	}

	return 0, fmt.Errorf("college_id not found in context")
}

func newWebSocketMessage(messageType string, collegeID, userID int, notification *models.Notification, data any) WebSocketMessage {
	return WebSocketMessage{
		Type:         messageType,
		Notification: notification,
		Data:         data,
		Timestamp:    time.Now(),
		UserID:       userID,
		CollegeID:    collegeID,
	}
}
