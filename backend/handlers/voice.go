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
	voiceLogger.Info("🚀 handleVoiceJoin - User %s joining room %s", msg.UserID, msg.RoomID)

	if msg.RoomID == "" {
		voiceLogger.Error("❌ Join failed: Missing roomId for user %s", msg.UserID)
		sendVoiceError(ws, "Missing roomId")
		metrics.RecordVoiceOperation("join", "error")
		return
	}

	voiceLogger.Info("📞 Calling sfuService.HandleJoin for user %s in room %s", msg.UserID, msg.RoomID)
	if err := sfuService.HandleJoin(msg.RoomID, msg.UserID, ws); err != nil {
		voiceLogger.Error("❌ Failed to join room %s for user %s: %v", msg.RoomID, msg.UserID, err)
		sendVoiceError(ws, "Failed to join room: "+err.Error())
		metrics.RecordVoiceOperation("join", "error")
		return
	}

	voiceLogger.Info("✅ sfuService.HandleJoin succeeded for user %s", msg.UserID)

	stats, err := sfuService.GetRoomStats(msg.RoomID)
	if err != nil {
		voiceLogger.Error("⚠️  Failed to get room stats: %v", err)
	} else {
		voiceLogger.Info("📊 Room stats retrieved: %d users in room %s", stats["peerCount"], msg.RoomID)
	}

	response := models.VoiceResponse{
		Type:   "joined",
		RoomID: msg.RoomID,
		UserID: msg.UserID,
		Users:  stats["users"].([]models.VoiceUser),
	}

	voiceLogger.Info("📤 Sending 'joined' response to user %s with %d users", msg.UserID, len(response.Users))
	if err := sfuService.SendToUser(msg.RoomID, msg.UserID, response); err != nil {
		voiceLogger.Error("❌ Failed to send join response to user %s: %v", msg.UserID, err)
	} else {
		voiceLogger.Info("✅ Successfully sent 'joined' response to user %s", msg.UserID)
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
	voiceLogger.Info("📥 handleVoiceOffer - Received offer from user %s in room %s", msg.UserID, msg.RoomID)

	if msg.RoomID == "" || msg.SDP == nil {
		voiceLogger.Error("❌ Offer validation failed - roomId: %s, SDP: %v", msg.RoomID, msg.SDP != nil)
		sendVoiceError(ws, "Missing roomId or SDP")
		return
	}

	voiceLogger.Info("📝 Processing offer - SDP type: %s, SDP length: %d bytes", msg.SDP.Type, len(msg.SDP.SDP))
	answer, err := sfuService.HandleOffer(msg.RoomID, msg.UserID, *msg.SDP)
	if err != nil {
		voiceLogger.Error("❌ Failed to handle offer from user %s: %v", msg.UserID, err)
		sendVoiceError(ws, "Failed to process offer: "+err.Error())
		return
	}

	voiceLogger.Info("✅ Created answer for user %s - SDP type: %s", msg.UserID, answer.Type)

	response := models.VoiceResponse{
		Type:   "answer",
		RoomID: msg.RoomID,
		UserID: msg.UserID,
		SDP:    answer,
	}

	voiceLogger.Info("📤 Sending answer to user %s", msg.UserID)
	if err := sfuService.SendToUser(msg.RoomID, msg.UserID, response); err != nil {
		voiceLogger.Error("❌ Failed to send answer to user %s: %v", msg.UserID, err)
	} else {
		voiceLogger.Info("✅ Successfully sent answer to user %s", msg.UserID)
	}
}


func handleVoiceAnswer(ws *websocket.Conn, msg models.VoiceMessage) {
	voiceLogger.Info("🔵 handleVoiceAnswer called for user %s in room %s", msg.UserID, msg.RoomID)

	if msg.RoomID == "" || msg.SDP == nil {
		voiceLogger.Warn("Missing roomId or SDP - roomId: %s, SDP: %v", msg.RoomID, msg.SDP)
		sendVoiceError(ws, "Missing roomId or SDP")
		return
	}

	voiceLogger.Info("📥 Processing answer with SDP type: %s", msg.SDP.Type)

	if err := sfuService.HandleAnswer(msg.RoomID, msg.UserID, *msg.SDP); err != nil {
		voiceLogger.Error("Failed to handle answer: %v", err)
		sendVoiceError(ws, "Failed to process answer: "+err.Error())
		return
	}

	voiceLogger.Info("✅ Successfully processed answer from user %s", msg.UserID)
}



func handleVoiceICECandidate(ws *websocket.Conn, msg models.VoiceMessage) {
	voiceLogger.Info("🧊 handleVoiceICECandidate - Received from user %s in room %s", msg.UserID, msg.RoomID)

	if msg.RoomID == "" || msg.Candidate == nil {
		voiceLogger.Error("❌ ICE candidate validation failed - roomId: %s, candidate: %v", msg.RoomID, msg.Candidate != nil)
		sendVoiceError(ws, "Missing roomId or candidate")
		return
	}

	voiceLogger.Info("📝 ICE candidate details - Candidate: %s, SDPMid: %s, SDPMLineIndex: %d",
		msg.Candidate.Candidate, stringOrEmpty(msg.Candidate.SDPMid), intOrZero(msg.Candidate.SDPMLineIndex))

	if err := sfuService.HandleICECandidate(msg.RoomID, msg.UserID, *msg.Candidate); err != nil {
		voiceLogger.Error("❌ Failed to handle ICE candidate for user %s: %v", msg.UserID, err)
		sendVoiceError(ws, "Failed to process ICE candidate: "+err.Error())
		return
	}

	voiceLogger.Info("✅ Successfully added ICE candidate for user %s", msg.UserID)
}

func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func intOrZero(i *uint16) uint16 {
	if i == nil {
		return 0
	}
	return *i
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
