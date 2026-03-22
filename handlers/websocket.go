package handlers

import (
	"codecollab/config"
	"codecollab/middleware"
	"codecollab/models"
	"codecollab/utils"
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
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

// HandleWebSocket establishes WebSocket connection for real-time communication
// @Summary WebSocket Connection Endpoint
// @Description Establishes WebSocket connection for code analysis and voice calls. Supports multiple message types including analysis requests and voice operations.
// @Description
// @Description **Message Types:**
// @Description - `analysis`: Request code analysis
// @Description - `voice`: Voice call operations (join, leave, offer, ice_candidate, mute, unmute)
// @Description
// @Description **Voice Actions:**
// @Description - `join`: Join a voice room
// @Description - `leave`: Leave a voice room
// @Description - `offer`: Send WebRTC offer
// @Description - `ice_candidate`: Send ICE candidate for NAT traversal
// @Description - `mute`: Mute microphone
// @Description - `unmute`: Unmute microphone
// @Tags websocket
// @Param token query string true "Authentication token"
// @Success 101 {string} string "Switching Protocols - WebSocket connection established"
// @Failure 401 {string} string "Unauthorized - Invalid or missing token"
// @Failure 429 {string} string "Too Many Requests - Rate limit exceeded"
// @Router /ws [get]
// @Security BearerAuth
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

		HandleVoiceDisconnect(userID)

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

		var actionMsg struct {
			Action string `json:"action"`
		}
		if err := json.Unmarshal(messageBytes, &actionMsg); err != nil {
			wsLogger.Error("Failed to parse request from user %s: %v", userID, err)
			sendError(conn, "Invalid request format")
			continue
		}

		if actionMsg.Action == "join" || actionMsg.Action == "leave" ||
			actionMsg.Action == "offer" || actionMsg.Action == "answer" ||
			actionMsg.Action == "ice_candidate" ||
			actionMsg.Action == "mute" || actionMsg.Action == "unmute" {
			HandleVoiceMessage(conn, userID, messageBytes, cfg)
			continue
		}

		var request models.AnalyzeRequest
		if err := json.Unmarshal(messageBytes, &request); err != nil {
			wsLogger.Error("Failed to parse analyze request from user %s: %v", userID, err)
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

// HandleHealth provides health check endpoint
// @Summary Health Check
// @Description Returns server health status including active connections and voice statistics
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{} "Health status with active connections and voice stats"
// @Router /health [get]
func HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)

	connectionsMu.RLock()
	activeConnections := len(connections)
	connectionsMu.RUnlock()

	response := map[string]interface{}{
		"status":             "healthy",
		"timestamp":          time.Now().Format(time.RFC3339),
		"active_connections": activeConnections,
		"voice":              GetVoiceStats(),
	}

	json.NewEncoder(w).Encode(response)
}
