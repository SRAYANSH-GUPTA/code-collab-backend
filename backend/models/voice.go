package models

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

// VoiceMessage represents WebSocket messages for voice operations
// @Description Voice message structure for WebSocket communication in voice calls
type VoiceMessage struct {
	Action    string                     `json:"action" example:"join" enums:"join,leave,offer,ice_candidate,mute,unmute"`                         // Action type: join, leave, offer, ice_candidate, mute, unmute
	RoomID    string                     `json:"roomId" example:"voice-room-1"`                                                                   // Voice channel/room identifier
	UserID    string                     `json:"userId" example:"user-123"`                                                                       // User identifier (set by server)
	TargetID  string                     `json:"targetId,omitempty" example:"user-456"`                                                           // Target user for P2P signaling
	SDP       *webrtc.SessionDescription `json:"sdp,omitempty" swaggerignore:"true"`                                                              // WebRTC Session Description (offer/answer)
	Candidate *webrtc.ICECandidateInit   `json:"candidate,omitempty" swaggerignore:"true"`                                                        // ICE candidate for connection establishment
	IsMuted   bool                       `json:"isMuted,omitempty" example:"false"`                                                               // Mute status
	Timestamp time.Time                  `json:"timestamp" example:"2025-01-15T10:30:00Z"`                                                        // Message timestamp
}

// VoiceResponse represents responses sent to clients for voice operations
// @Description Voice response structure sent by server for voice-related events
type VoiceResponse struct {
	Type      string                     `json:"type" example:"joined" enums:"joined,left,answer,ice_candidate,muted,unmuted,user_joined,user_left,error"` // Response type
	RoomID    string                     `json:"roomId" example:"voice-room-1"`                                                                              // Voice room identifier
	UserID    string                     `json:"userId" example:"user-123"`                                                                                  // User identifier
	Users     []VoiceUser                `json:"users,omitempty"`                                                                                            // List of users in the room
	SDP       *webrtc.SessionDescription `json:"sdp,omitempty" swaggerignore:"true"`                                                                         // WebRTC Session Description (answer)
	Candidate *webrtc.ICECandidateInit   `json:"candidate,omitempty" swaggerignore:"true"`                                                                   // ICE candidate from server
	Message   string                     `json:"message,omitempty" example:"Successfully joined room"`                                                       // Success message
	Error     string                     `json:"error,omitempty" example:"Room not found"`                                                                   // Error message
}

// VoiceUser represents a user in a voice room
// @Description User information in a voice room
type VoiceUser struct {
	UserID     string    `json:"userId" example:"user-123"`          // User identifier
	IsMuted    bool      `json:"isMuted" example:"false"`            // Mute status
	IsSpeaking bool      `json:"isSpeaking" example:"true"`          // Voice activity detection status
	JoinedAt   time.Time `json:"joinedAt" example:"2025-01-15T10:30:00Z"` // Time when user joined the room
	Quality    string    `json:"quality" example:"good" enums:"good,fair,poor"` // Connection quality indicator
}


type VoiceRoom struct {
	ID             string
	Name           string
	Peers          map[string]*VoicePeer 
	CreatedAt      time.Time
	MaxParticipants int
	mu             sync.RWMutex
}


type VoicePeer struct {
	UserID         string
	PeerConnection *webrtc.PeerConnection
	WSConnection   *websocket.Conn
	IsMuted        bool
	IsSpeaking     bool
	JoinedAt       time.Time
	LastActivity   time.Time


	LocalTracks    []*webrtc.TrackLocalStaticRTP
	RemoteTracks   []*webrtc.TrackRemote


	PacketsLost    uint32
	Jitter         float64
	RTT            time.Duration

	// Flag to indicate if renegotiation is needed when connection becomes stable
	NeedsRenegotiation bool

	Mu             sync.RWMutex
	wsMu           sync.Mutex // protects concurrent WebSocket writes
}

// SendJSON safely writes a JSON message to the peer's WebSocket connection.
// It serializes concurrent writes to avoid corrupting WebSocket frames.
func (p *VoicePeer) SendJSON(v interface{}) error {
	p.wsMu.Lock()
	defer p.wsMu.Unlock()
	return p.WSConnection.WriteJSON(v)
}

// SendMessage safely writes a raw WebSocket message to the peer's connection.
func (p *VoicePeer) SendMessage(messageType int, data []byte) error {
	p.wsMu.Lock()
	defer p.wsMu.Unlock()
	return p.WSConnection.WriteMessage(messageType, data)
}


func NewVoiceRoom(id, name string) *VoiceRoom {
	return &VoiceRoom{
		ID:              id,
		Name:            name,
		Peers:           make(map[string]*VoicePeer),
		CreatedAt:       time.Now(),
		MaxParticipants: 50, 
	}
}


func (r *VoiceRoom) AddPeer(peer *VoicePeer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Peers) >= r.MaxParticipants {
		return ErrRoomFull
	}

	r.Peers[peer.UserID] = peer
	return nil
}


func (r *VoiceRoom) RemovePeer(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if peer, exists := r.Peers[userID]; exists {
		
		if peer.PeerConnection != nil {
			peer.PeerConnection.Close()
		}
		delete(r.Peers, userID)
	}
}


func (r *VoiceRoom) GetPeer(userID string) (*VoicePeer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	peer, exists := r.Peers[userID]
	return peer, exists
}


func (r *VoiceRoom) GetAllPeers() []*VoicePeer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	peers := make([]*VoicePeer, 0, len(r.Peers))
	for _, peer := range r.Peers {
		peers = append(peers, peer)
	}
	return peers
}


func (r *VoiceRoom) GetPeerCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.Peers)
}


func (r *VoiceRoom) GetVoiceUsers() []VoiceUser {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]VoiceUser, 0, len(r.Peers))
	for _, peer := range r.Peers {
		quality := "good"
		if peer.PacketsLost > 100 {
			quality = "poor"
		} else if peer.PacketsLost > 50 {
			quality = "fair"
		}

		users = append(users, VoiceUser{
			UserID:     peer.UserID,
			IsMuted:    peer.IsMuted,
			IsSpeaking: peer.IsSpeaking,
			JoinedAt:   peer.JoinedAt,
			Quality:    quality,
		})
	}
	return users
}


var (
	ErrRoomFull        = &VoiceError{Code: "ROOM_FULL", Message: "Voice room has reached maximum capacity"}
	ErrRoomNotFound    = &VoiceError{Code: "ROOM_NOT_FOUND", Message: "Voice room not found"}
	ErrPeerNotFound    = &VoiceError{Code: "PEER_NOT_FOUND", Message: "Peer not found in room"}
	ErrInvalidSDP      = &VoiceError{Code: "INVALID_SDP", Message: "Invalid SDP format"}
	ErrConnectionFailed = &VoiceError{Code: "CONNECTION_FAILED", Message: "Failed to establish WebRTC connection"}
)


type VoiceError struct {
	Code    string
	Message string
}

func (e *VoiceError) Error() string {
	return e.Message
}
