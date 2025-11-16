package handlers

import (
	"encoding/json"
	"time"

	"codecollab/metrics"
	"codecollab/models"
	"codecollab/services"
	"codecollab/utils"

	"github.com/gorilla/websocket"
)

var (
	voiceLogger = utils.NewLogger("voice")
	
	sfuService *services.SFUService
)




func InitVoiceService(stunServers, turnServers []string, turnUsername, turnPassword string) {
	roomManager := services.NewRoomManager()
	sfuService = services.NewSFUService(roomManager, stunServers, turnServers, turnUsername, turnPassword)
	voiceLogger.Info("Voice service initialized with %d STUN servers and %d TURN servers", len(stunServers), len(turnServers))

	
	
	go periodicRoomCleanup(roomManager)
}


func periodicRoomCleanup(rm *services.RoomManager) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cleaned := rm.CleanupEmptyRooms()
		if cleaned > 0 {
			voiceLogger.Info("Cleaned up %d empty rooms", cleaned)
		}
	}
}













func HandleVoiceMessage(ws *websocket.Conn, userID string, messageBytes []byte) {
	var msg models.VoiceMessage
	if err := json.Unmarshal(messageBytes, &msg); err != nil {
		voiceLogger.Error("Failed to parse voice message from user %s: %v", userID, err)
		sendVoiceError(ws, "Invalid message format")
		return
	}

	
	msg.UserID = userID
	msg.Timestamp = time.Now()

	voiceLogger.Info("Voice action '%s' from user %s in room %s", msg.Action, userID, msg.RoomID)


	switch msg.Action {
	case "join":
		handleVoiceJoin(ws, msg)
	case "leave":
		handleVoiceLeave(ws, msg)
	case "offer":
		handleVoiceOffer(ws, msg)
	case "answer":
		handleVoiceAnswer(ws, msg)
	case "ice_candidate":
		handleVoiceICECandidate(ws, msg)
	case "mute":
		handleVoiceMute(ws, msg)
	case "unmute":
		handleVoiceUnmute(ws, msg)
	default:
		voiceLogger.Warn("Unknown voice action: %s", msg.Action)
		sendVoiceError(ws, "Unknown action: "+msg.Action)
	}
}


func handleVoiceJoin(ws *websocket.Conn, msg models.VoiceMessage) {
	if msg.RoomID == "" {
		sendVoiceError(ws, "Missing roomId")
		metrics.RecordVoiceOperation("join", "error")
		return
	}

	if err := sfuService.HandleJoin(msg.RoomID, msg.UserID, ws); err != nil {
		voiceLogger.Error("Failed to join room: %v", err)
		sendVoiceError(ws, "Failed to join room: "+err.Error())
		metrics.RecordVoiceOperation("join", "error")
		return
	}

	
	stats, _ := sfuService.GetRoomStats(msg.RoomID)
	response := models.VoiceResponse{
		Type:   "joined",
		RoomID: msg.RoomID,
		UserID: msg.UserID,
		Users:  stats["users"].([]models.VoiceUser),
	}

	if err := ws.WriteJSON(response); err != nil {
		voiceLogger.Error("Failed to send join response: %v", err)
	}

	
	metrics.RecordVoiceOperation("join", "success")
	updateGlobalVoiceMetrics()
}


func handleVoiceLeave(ws *websocket.Conn, msg models.VoiceMessage) {
	if msg.RoomID == "" {
		sendVoiceError(ws, "Missing roomId")
		metrics.RecordVoiceOperation("leave", "error")
		return
	}

	if err := sfuService.HandleLeave(msg.RoomID, msg.UserID); err != nil {
		voiceLogger.Error("Failed to leave room: %v", err)
		sendVoiceError(ws, "Failed to leave room: "+err.Error())
		metrics.RecordVoiceOperation("leave", "error")
		return
	}

	response := models.VoiceResponse{
		Type:   "left",
		RoomID: msg.RoomID,
		UserID: msg.UserID,
	}

	if err := ws.WriteJSON(response); err != nil {
		voiceLogger.Error("Failed to send leave response: %v", err)
	}

	metrics.RecordVoiceOperation("leave", "success")
	updateGlobalVoiceMetrics()
}



func handleVoiceOffer(ws *websocket.Conn, msg models.VoiceMessage) {
	if msg.RoomID == "" || msg.SDP == nil {
		sendVoiceError(ws, "Missing roomId or SDP")
		return
	}

	answer, err := sfuService.HandleOffer(msg.RoomID, msg.UserID, *msg.SDP)
	if err != nil {
		voiceLogger.Error("Failed to handle offer: %v", err)
		sendVoiceError(ws, "Failed to process offer: "+err.Error())
		return
	}

	response := models.VoiceResponse{
		Type:   "answer",
		RoomID: msg.RoomID,
		UserID: msg.UserID,
		SDP:    answer,
	}

	if err := ws.WriteJSON(response); err != nil {
		voiceLogger.Error("Failed to send answer: %v", err)
	}
}


func handleVoiceAnswer(ws *websocket.Conn, msg models.VoiceMessage) {
	if msg.RoomID == "" || msg.SDP == nil {
		sendVoiceError(ws, "Missing roomId or SDP")
		return
	}

	if err := sfuService.HandleAnswer(msg.RoomID, msg.UserID, *msg.SDP); err != nil {
		voiceLogger.Error("Failed to handle answer: %v", err)
		sendVoiceError(ws, "Failed to process answer: "+err.Error())
		return
	}

	voiceLogger.Info("Successfully processed answer from user %s", msg.UserID)
}



func handleVoiceICECandidate(ws *websocket.Conn, msg models.VoiceMessage) {
	if msg.RoomID == "" || msg.Candidate == nil {
		sendVoiceError(ws, "Missing roomId or candidate")
		return
	}

	if err := sfuService.HandleICECandidate(msg.RoomID, msg.UserID, *msg.Candidate); err != nil {
		voiceLogger.Error("Failed to handle ICE candidate: %v", err)
		sendVoiceError(ws, "Failed to process ICE candidate: "+err.Error())
		return
	}

	
	voiceLogger.Info("Added ICE candidate for user %s", msg.UserID)
}


func handleVoiceMute(ws *websocket.Conn, msg models.VoiceMessage) {
	if msg.RoomID == "" {
		sendVoiceError(ws, "Missing roomId")
		return
	}

	if err := sfuService.HandleMute(msg.RoomID, msg.UserID, true); err != nil {
		voiceLogger.Error("Failed to mute: %v", err)
		sendVoiceError(ws, "Failed to mute: "+err.Error())
		return
	}

	response := models.VoiceResponse{
		Type:   "muted",
		RoomID: msg.RoomID,
		UserID: msg.UserID,
	}

	if err := ws.WriteJSON(response); err != nil {
		voiceLogger.Error("Failed to send mute response: %v", err)
	}
}


func handleVoiceUnmute(ws *websocket.Conn, msg models.VoiceMessage) {
	if msg.RoomID == "" {
		sendVoiceError(ws, "Missing roomId")
		return
	}

	if err := sfuService.HandleMute(msg.RoomID, msg.UserID, false); err != nil {
		voiceLogger.Error("Failed to unmute: %v", err)
		sendVoiceError(ws, "Failed to unmute: "+err.Error())
		return
	}

	response := models.VoiceResponse{
		Type:   "unmuted",
		RoomID: msg.RoomID,
		UserID: msg.UserID,
	}

	if err := ws.WriteJSON(response); err != nil {
		voiceLogger.Error("Failed to send unmute response: %v", err)
	}
}



func HandleVoiceDisconnect(userID string) {
	
	
	voiceLogger.Info("Cleaning up voice connections for user %s", userID)

	
	
}


func sendVoiceError(ws *websocket.Conn, message string) {
	response := models.VoiceResponse{
		Type:  "error",
		Error: message,
	}
	ws.WriteJSON(response)
}



func GetVoiceStats() map[string]interface{} {
	if sfuService == nil {
		return map[string]interface{}{
			"initialized": false,
		}
	}

	return map[string]interface{}{
		"initialized": true,
		"roomCount":   sfuService.GetRoomManager().GetRoomCount(),
		"totalPeers":  sfuService.GetRoomManager().GetTotalPeerCount(),
	}
}



func GetSFUService() *services.SFUService {
	return sfuService
}



func updateGlobalVoiceMetrics() {
	if sfuService == nil {
		return
	}

	rm := sfuService.GetRoomManager()
	roomCount := rm.GetRoomCount()
	peerCount := rm.GetTotalPeerCount()

	metrics.UpdateVoiceStats(roomCount, peerCount)

	
	for _, room := range rm.GetAllRooms() {
		metrics.VoicePeersPerRoom.Observe(float64(room.GetPeerCount()))
	}
}
