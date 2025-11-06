package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/codecollab/backend/config"
	"github.com/codecollab/backend/middleware"
	"github.com/codecollab/backend/models"
	"github.com/codecollab/backend/utils"
	"github.com/gorilla/websocket"
)

var (
	connections   = make(map[*websocket.Conn]*models.Connection)
	connectionsMu sync.RWMutex
	upgrader      = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			
			return true
		},
	}
	wsLogger    = utils.NewLogger("websocket")
	rateLimiter = middleware.NewRateLimiter(60, 1*time.Minute)
)


func HandleWebSocket(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		
		token := r.URL.Query().Get("token")
		if token == "" {
			wsLogger.Error("Missing auth token in WebSocket request")
			http.Error(w, "Missing auth token", http.StatusUnauthorized)
			return
		}

		
		userID, err := VerifyToken(token, cfg)
		if err != nil {
			wsLogger.Error("Failed to verify token: %v", err)
			http.Error(w, "Invalid auth token", http.StatusUnauthorized)
			return
		}

		
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			wsLogger.Error("Failed to upgrade connection: %v", err)
			return
		}

		
		connectionsMu.Lock()
		connections[conn] = &models.Connection{
			UserID:   userID,
			LastSeen: time.Now(),
		}
		connectionsMu.Unlock()

		utils.LogConnection("connected", userID)
		wsLogger.Info("New WebSocket connection for user: %s", userID)

		
		go handleConnection(conn, userID, cfg)
	}
}

func handleConnection(conn *websocket.Conn, userID string, cfg *config.Config) {
	defer func() {
		
		connectionsMu.Lock()
		delete(connections, conn)
		connectionsMu.Unlock()

		conn.Close()
		utils.LogConnection("disconnected", userID)
		wsLogger.Info("WebSocket connection closed for user: %s", userID)
	}()

	for {
		
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				wsLogger.Error("WebSocket error for user %s: %v", userID, err)
			}
			break
		}

		
		var request models.AnalyzeRequest
		if err := json.Unmarshal(messageBytes, &request); err != nil {
			wsLogger.Error("Failed to parse request from user %s: %v", userID, err)
			sendError(conn, "Invalid request format")
			continue
		}

		
		if request.Action != "analyze" {
			sendError(conn, "Unknown action: "+request.Action)
			continue
		}

		if request.Language == "" {
			sendError(conn, "Missing language field")
			continue
		}

		if request.Code == "" {
			sendError(conn, "Missing code field")
			continue
		}

		
		if !rateLimiter.CheckRateLimit(userID) {
			wsLogger.Warn("Rate limit exceeded for user: %s", userID)
			sendError(conn, "Rate limit exceeded. Please wait before sending more requests.")
			continue
		}

		
		startTime := time.Now()
		wsLogger.Info("Processing analysis request from user %s for language: %s", userID, request.Language)

		
		errors, err := InvokeLinter(request.Language, request.Code, cfg)
		if err != nil {
			wsLogger.Error("Failed to invoke linter for user %s: %v", userID, err)
			sendError(conn, "Failed to analyze code: "+err.Error())
			continue
		}

		
		executionTime := int(time.Since(startTime).Milliseconds())

		
		response := models.AnalyzeResponse{
			Type:          "analysis_result",
			Errors:        errors,
			ExecutionTime: executionTime,
		}

		if err := conn.WriteJSON(response); err != nil {
			wsLogger.Error("Failed to send response to user %s: %v", userID, err)
			break
		}

		wsLogger.Info("Sent analysis result to user %s: %d errors, %dms", userID, len(errors), executionTime)

		
		connectionsMu.Lock()
		if connInfo, exists := connections[conn]; exists {
			connInfo.LastSeen = time.Now()
		}
		connectionsMu.Unlock()
	}
}

func sendError(conn *websocket.Conn, message string) {
	response := models.AnalyzeResponse{
		Type:         "error",
		ErrorMessage: message,
	}
	conn.WriteJSON(response)
}


func HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	connectionsMu.RLock()
	activeConnections := len(connections)
	connectionsMu.RUnlock()

	response := map[string]interface{}{
		"status":             "healthy",
		"timestamp":          time.Now().Format(time.RFC3339),
		"active_connections": activeConnections,
	}

	json.NewEncoder(w).Encode(response)
}
