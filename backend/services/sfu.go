package services

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"codecollab/models"
	"codecollab/utils"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

var sfuLogger = utils.NewLogger("sfu")



type SFUService struct {
	roomManager *RoomManager
	config      webrtc.Configuration
}


func NewSFUService(roomManager *RoomManager, stunServers, turnServers []string, turnUsername, turnPassword string) *SFUService {
	
	iceServers := []webrtc.ICEServer{}

	
	for _, stun := range stunServers {
		iceServers = append(iceServers, webrtc.ICEServer{
			URLs: []string{stun},
		})
	}

	
	for _, turn := range turnServers {
		iceServers = append(iceServers, webrtc.ICEServer{
			URLs:       []string{turn},
			Username:   turnUsername,
			Credential: turnPassword,
		})
	}

	return &SFUService{
		roomManager: roomManager,
		config: webrtc.Configuration{
			ICEServers: iceServers,
			
			SDPSemantics: webrtc.SDPSemanticsUnifiedPlan,
		},
	}
}



func (s *SFUService) HandleJoin(roomID, userID string, ws *websocket.Conn) error {
	sfuLogger.Info("🚀 HandleJoin - User %s joining room %s", userID, roomID)

	room := s.roomManager.GetOrCreateRoom(roomID, roomID)
	sfuLogger.Info("📦 Room %s obtained/created (current peer count: %d)", roomID, room.GetPeerCount())

	sfuLogger.Info("🔧 Creating peer connection for user %s with %d ICE servers", userID, len(s.config.ICEServers))
	peerConnection, err := webrtc.NewPeerConnection(s.config)
	if err != nil {
		sfuLogger.Error("❌ Failed to create peer connection for user %s: %v", userID, err)
		return err
	}

	sfuLogger.Info("✅ Created peer connection for user %s in room %s", userID, roomID)

	peer := &models.VoicePeer{
		UserID:         userID,
		PeerConnection: peerConnection,
		WSConnection:   ws,
		IsMuted:        false,
		IsSpeaking:     false,
		JoinedAt:       time.Now(),
		LastActivity:   time.Now(),
		LocalTracks:    make([]*webrtc.TrackLocalStaticRTP, 0),
		RemoteTracks:   make([]*webrtc.TrackRemote, 0),
	}

	sfuLogger.Info("👤 Adding peer %s to room %s", userID, roomID)
	if err := room.AddPeer(peer); err != nil {
		sfuLogger.Error("❌ Failed to add peer %s to room: %v", userID, err)
		peerConnection.Close()
		return err
	}

	sfuLogger.Info("🎧 Setting up peer handlers for user %s", userID)
	s.setupPeerHandlers(peer, room)

	sfuLogger.Info("✅ User %s successfully joined room %s (total: %d users)", userID, roomID, room.GetPeerCount())

	sfuLogger.Info("📢 Broadcasting room state for user_joined event")
	s.broadcastRoomState(room, userID, "user_joined")

	return nil
}



func (s *SFUService) setupPeerHandlers(peer *models.VoicePeer, room *models.VoiceRoom) {
	pc := peer.PeerConnection

	
	
	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		sfuLogger.Info("🎵 OnTrack FIRED - Received %s track from user %s (SSRC: %d, PayloadType: %d)",
			track.Kind(), peer.UserID, track.SSRC(), track.PayloadType())

		codec := track.Codec()
		sfuLogger.Info("📝 Track codec - MimeType: %s, ClockRate: %d, Channels: %d",
			codec.MimeType, codec.ClockRate, codec.Channels)

		peer.Mu.Lock()
		peer.RemoteTracks = append(peer.RemoteTracks, track)
		trackCount := len(peer.RemoteTracks)
		peer.Mu.Unlock()

		sfuLogger.Info("📊 User %s now has %d remote tracks", peer.UserID, trackCount)

		sfuLogger.Info("🔀 Forwarding track from %s to other peers in room %s", peer.UserID, room.ID)
		s.forwardTrackToOthers(peer, track, room)
	})

	
	
	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		sfuLogger.Info("🧊 ICE Connection State Change - User %s: %s", peer.UserID, state.String())

		switch state {
		case webrtc.ICEConnectionStateNew:
			sfuLogger.Info("🆕 User %s ICE connection is new", peer.UserID)
		case webrtc.ICEConnectionStateChecking:
			sfuLogger.Info("🔍 User %s ICE connection is checking", peer.UserID)
		case webrtc.ICEConnectionStateConnected:
			sfuLogger.Info("✅ User %s ICE connection is connected", peer.UserID)
		case webrtc.ICEConnectionStateCompleted:
			sfuLogger.Info("🎉 User %s ICE connection is completed", peer.UserID)
		case webrtc.ICEConnectionStateDisconnected:
			sfuLogger.Warn("⚠️  User %s ICE disconnected", peer.UserID)
		case webrtc.ICEConnectionStateFailed:
			sfuLogger.Error("❌ User %s ICE connection failed", peer.UserID)
			s.HandleLeave(room.ID, peer.UserID)
		case webrtc.ICEConnectionStateClosed:
			sfuLogger.Info("🚪 User %s ICE connection closed", peer.UserID)
		}
	})

	
	
	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			sfuLogger.Info("🏁 ICE gathering complete for user %s (received nil candidate)", peer.UserID)
			return
		}

		sfuLogger.Info("🧊 Generated ICE candidate for user %s - Type: %s, Protocol: %s",
			peer.UserID, candidate.Typ.String(), candidate.Protocol.String())

		candidateInit := candidate.ToJSON()
		response := models.VoiceResponse{
			Type:      "ice_candidate",
			RoomID:    room.ID,
			UserID:    peer.UserID,
			Candidate: &candidateInit,
		}

		sfuLogger.Info("📤 Sending ICE candidate to user %s via WebSocket", peer.UserID)
		if err := peer.WSConnection.WriteJSON(response); err != nil {
			sfuLogger.Error("❌ Failed to send ICE candidate to user %s: %v", peer.UserID, err)
		} else {
			sfuLogger.Info("✅ Successfully sent ICE candidate to user %s", peer.UserID)
		}
	})


	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		sfuLogger.Info("🔌 Peer Connection State Change - User %s: %s", peer.UserID, state.String())

		switch state {
		case webrtc.PeerConnectionStateNew:
			sfuLogger.Info("🆕 User %s peer connection is new", peer.UserID)
		case webrtc.PeerConnectionStateConnecting:
			sfuLogger.Info("🔗 User %s peer connection is connecting", peer.UserID)
		case webrtc.PeerConnectionStateConnected:
			sfuLogger.Info("✅ User %s peer connection is connected", peer.UserID)

			peer.Mu.RLock()
			needsRenegotiation := peer.NeedsRenegotiation
			peer.Mu.RUnlock()

			if needsRenegotiation {
				sfuLogger.Info("⏩ Connection stable for %s, triggering deferred renegotiation", peer.UserID)
				go func() {
					time.Sleep(500 * time.Millisecond)
					s.triggerRenegotiation(peer, room)
				}()
			}
		case webrtc.PeerConnectionStateDisconnected:
			sfuLogger.Warn("⚠️  User %s peer connection disconnected", peer.UserID)
			// Clean up the disconnected peer after a short delay
			go func() {
				time.Sleep(3 * time.Second)
				if peer.PeerConnection.ConnectionState() == webrtc.PeerConnectionStateDisconnected {
					sfuLogger.Warn("🧹 Auto-removing disconnected peer %s from room %s", peer.UserID, room.ID)
					s.HandleLeave(room.ID, peer.UserID)
				}
			}()
		case webrtc.PeerConnectionStateFailed:
			sfuLogger.Error("❌ User %s peer connection failed", peer.UserID)
			// Immediately remove failed connections
			sfuLogger.Info("🧹 Removing failed peer %s from room %s", peer.UserID, room.ID)
			s.HandleLeave(room.ID, peer.UserID)
		case webrtc.PeerConnectionStateClosed:
			sfuLogger.Info("🚪 User %s peer connection closed", peer.UserID)
			// Immediately remove closed connections
			sfuLogger.Info("🧹 Removing closed peer %s from room %s", peer.UserID, room.ID)
			s.HandleLeave(room.ID, peer.UserID)
		}
	})
}



// triggerRenegotiation creates a new offer for a peer and sends it to the client
func (s *SFUService) triggerRenegotiation(peer *models.VoicePeer, room *models.VoiceRoom) {
	// Check if the connection is in a stable state before renegotiating
	signalingState := peer.PeerConnection.SignalingState()
	if signalingState != webrtc.SignalingStateStable {
		sfuLogger.Info("⏸️  Marking peer %s for deferred renegotiation - signaling state is %s (not stable)",
			peer.UserID, signalingState.String())

		// Mark the peer as needing renegotiation when connection becomes stable
		peer.Mu.Lock()
		peer.NeedsRenegotiation = true
		peer.Mu.Unlock()
		return
	}

	sfuLogger.Info("🔄 Triggering renegotiation for peer %s", peer.UserID)

	// Create a new offer to renegotiate the connection
	offer, err := peer.PeerConnection.CreateOffer(nil)
	if err != nil {
		sfuLogger.Error("Failed to create renegotiation offer for %s: %v", peer.UserID, err)
		return
	}

	if err := peer.PeerConnection.SetLocalDescription(offer); err != nil {
		sfuLogger.Error("Failed to set local description for renegotiation %s: %v", peer.UserID, err)
		return
	}

	// Send the new offer to the client so they can accept the new track
	localDesc := peer.PeerConnection.LocalDescription()
	response := models.VoiceResponse{
		Type:   "offer",
		RoomID: room.ID,
		UserID: peer.UserID,
		SDP:    localDesc,
	}

	sfuLogger.Info("📤 Sending renegotiation offer to peer %s - SDP type: %s, SDP length: %d",
		peer.UserID, localDesc.Type.String(), len(localDesc.SDP))

	if err := peer.WSConnection.WriteJSON(response); err != nil {
		sfuLogger.Error("❌ Failed to send renegotiation offer to %s: %v", peer.UserID, err)
		return
	}

	// Clear the renegotiation flag
	peer.Mu.Lock()
	peer.NeedsRenegotiation = false
	peer.Mu.Unlock()

	sfuLogger.Info("✅ Successfully sent renegotiation offer to peer %s", peer.UserID)
}

func (s *SFUService) addExistingTracksToNewPeer(newPeer *models.VoicePeer, room *models.VoiceRoom) {
	allPeers := room.GetAllPeers()
	sfuLogger.Info("🎯 addExistingTracksToNewPeer: New peer %s joining room with %d total peers",
		newPeer.UserID, len(allPeers))

	tracksAdded := 0
	for _, existingPeer := range allPeers {
		if existingPeer.UserID == newPeer.UserID {
			sfuLogger.Info("⏭️  Skipping self (peer %s)", existingPeer.UserID)
			continue
		}

		// Check if the existing peer's connection is still valid
		connState := existingPeer.PeerConnection.ConnectionState()
		if connState == webrtc.PeerConnectionStateClosed || connState == webrtc.PeerConnectionStateFailed {
			sfuLogger.Warn("⚠️  Skipping tracks from peer %s - connection state is %s", existingPeer.UserID, connState.String())
			continue
		}

		existingPeer.Mu.RLock()
		trackCount := len(existingPeer.LocalTracks)
		sfuLogger.Info("👤 Existing peer %s - LocalTracks: %d, Connection state: %s",
			existingPeer.UserID, trackCount, connState.String())

		if trackCount == 0 {
			sfuLogger.Info("⏭️  Peer %s has no local tracks to add", existingPeer.UserID)
			existingPeer.Mu.RUnlock()
			continue
		}

		for i, localTrack := range existingPeer.LocalTracks {
			sfuLogger.Info("🔀 Attempting to add track %d from %s to new peer %s (Track ID: %s, StreamID: %s)",
				i, existingPeer.UserID, newPeer.UserID, localTrack.ID(), localTrack.StreamID())

			_, err := newPeer.PeerConnection.AddTrack(localTrack)
			if err != nil {
				sfuLogger.Error("❌ Failed to add existing track %d from %s to new peer %s: %v",
					i, existingPeer.UserID, newPeer.UserID, err)
				continue
			}
			tracksAdded++
			sfuLogger.Info("✅ Successfully added track %d from %s to new peer %s (track ID: %s)",
				i, existingPeer.UserID, newPeer.UserID, localTrack.ID())
		}
		existingPeer.Mu.RUnlock()
	}

	if tracksAdded == 0 {
		sfuLogger.Warn("⚠️  No tracks were added to new peer %s (existing peers may not have started sending audio yet)", newPeer.UserID)
	} else {
		sfuLogger.Info("✅ Successfully added %d tracks to new peer %s", tracksAdded, newPeer.UserID)
	}
}



func (s *SFUService) forwardTrackToOthers(sender *models.VoicePeer, incomingTrack *webrtc.TrackRemote, room *models.VoiceRoom) {
	sfuLogger.Info("🎵 forwardTrackToOthers: Received %s track from %s", incomingTrack.Kind(), sender.UserID)

	localTrack, err := webrtc.NewTrackLocalStaticRTP(
		incomingTrack.Codec().RTPCodecCapability,
		fmt.Sprintf("audio-%s", sender.UserID),
		fmt.Sprintf("stream-%s", sender.UserID),
	)
	if err != nil {
		sfuLogger.Error("Failed to create local track: %v", err)
		return
	}

	sender.Mu.Lock()
	sender.LocalTracks = append(sender.LocalTracks, localTrack)
	sfuLogger.Info("📊 Sender %s now has %d local tracks", sender.UserID, len(sender.LocalTracks))
	sender.Mu.Unlock()

	
	go func() {
		rtpBuf := make([]byte, 1500)
		for {
			n, _, err := incomingTrack.Read(rtpBuf)
			if err != nil {
				if err != io.EOF {
					sfuLogger.Error("Error reading track: %v", err)
				}
				return
			}

			
			if _, err := localTrack.Write(rtpBuf[:n]); err != nil {
				if err != io.ErrClosedPipe {
					sfuLogger.Error("Error writing to track: %v", err)
				}
				return
			}
		}
	}()

	allPeers := room.GetAllPeers()
	sfuLogger.Info("📤 Broadcasting track from %s to %d other peers", sender.UserID, len(allPeers)-1)

	for _, otherPeer := range allPeers {
		if otherPeer.UserID == sender.UserID {
			continue
		}

		// Check if the peer connection is in a valid state before adding track
		connState := otherPeer.PeerConnection.ConnectionState()
		if connState == webrtc.PeerConnectionStateClosed || connState == webrtc.PeerConnectionStateFailed {
			sfuLogger.Warn("⚠️  Skipping peer %s - connection state is %s", otherPeer.UserID, connState.String())
			continue
		}

		rtpSender, err := otherPeer.PeerConnection.AddTrack(localTrack)
		if err != nil {
			sfuLogger.Error("❌ Failed to add track to peer %s (state: %s): %v", otherPeer.UserID, connState.String(), err)
			continue
		}
		sfuLogger.Info("✅ Added track from %s to peer %s", sender.UserID, otherPeer.UserID)

		// Trigger renegotiation for the peer to receive the new track
		go s.triggerRenegotiation(otherPeer, room)

		
		go func(peer *models.VoicePeer) {
			rtcpBuf := make([]byte, 1500)
			for {
				if _, _, err := rtpSender.Read(rtcpBuf); err != nil {
					return
				}
				
				
			}
		}(otherPeer)

		sfuLogger.Info("Forwarding track from %s to %s", sender.UserID, otherPeer.UserID)
	}
}



func (s *SFUService) HandleAnswer(roomID, userID string, answer webrtc.SessionDescription) error {
	sfuLogger.Info("🔵 HandleAnswer called for user %s in room %s with SDP type %s", userID, roomID, answer.Type)

	room, err := s.roomManager.GetRoom(roomID)
	if err != nil {
		sfuLogger.Error("Failed to get room %s: %v", roomID, err)
		return err
	}

	peer, exists := room.GetPeer(userID)
	if !exists {
		sfuLogger.Error("Peer %s not found in room %s", userID, roomID)
		return models.ErrPeerNotFound
	}

	sfuLogger.Info("🔧 Setting remote description for peer %s (current signaling state: %s)", userID, peer.PeerConnection.SignalingState())

	if err := peer.PeerConnection.SetRemoteDescription(answer); err != nil {
		sfuLogger.Error("Failed to set remote description (answer) for %s: %v", userID, err)
		return err
	}

	sfuLogger.Info("✅ Set remote description (answer) for user %s", userID)
	return nil
}

func (s *SFUService) HandleOffer(roomID, userID string, offer webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	sfuLogger.Info("📥 HandleOffer - Processing offer from user %s in room %s", userID, roomID)

	room, err := s.roomManager.GetRoom(roomID)
	if err != nil {
		sfuLogger.Error("❌ Failed to get room %s: %v", roomID, err)
		return nil, err
	}

	peer, exists := room.GetPeer(userID)
	if !exists {
		sfuLogger.Error("❌ Peer %s not found in room %s", userID, roomID)
		return nil, models.ErrPeerNotFound
	}

	sfuLogger.Info("🔧 Setting remote description (offer) for peer %s", userID)
	if err := peer.PeerConnection.SetRemoteDescription(offer); err != nil {
		sfuLogger.Error("❌ Failed to set remote description: %v", err)
		return nil, err
	}

	sfuLogger.Info("📦 Adding existing tracks from other peers to new peer %s", userID)
	s.addExistingTracksToNewPeer(peer, room)

	sfuLogger.Info("📝 Creating answer for peer %s", userID)
	answer, err := peer.PeerConnection.CreateAnswer(nil)
	if err != nil {
		sfuLogger.Error("❌ Failed to create answer: %v", err)
		return nil, err
	}

	sfuLogger.Info("🔧 Setting local description (answer) for peer %s", userID)
	if err := peer.PeerConnection.SetLocalDescription(answer); err != nil {
		sfuLogger.Error("❌ Failed to set local description: %v", err)
		return nil, err
	}

	peer.Mu.RLock()
	localTrackCount := len(peer.LocalTracks)
	peer.Mu.RUnlock()

	sfuLogger.Info("✅ Created answer for user %s - Local tracks: %d, Answer SDP type: %s",
		userID, localTrackCount, answer.Type)
	return &answer, nil
}



func (s *SFUService) HandleICECandidate(roomID, userID string, candidate webrtc.ICECandidateInit) error {
	room, err := s.roomManager.GetRoom(roomID)
	if err != nil {
		return err
	}

	peer, exists := room.GetPeer(userID)
	if !exists {
		return models.ErrPeerNotFound
	}

	if err := peer.PeerConnection.AddICECandidate(candidate); err != nil {
		sfuLogger.Error("Failed to add ICE candidate: %v", err)
		return err
	}

	sfuLogger.Info("Added ICE candidate for user %s", userID)
	return nil
}



func (s *SFUService) HandleLeave(roomID, userID string) error {
	room, err := s.roomManager.GetRoom(roomID)
	if err != nil {
		return err
	}

	room.RemovePeer(userID)
	sfuLogger.Info("User %s left room %s (remaining: %d)", userID, roomID, room.GetPeerCount())

	
	s.broadcastRoomState(room, userID, "user_left")

	
	
	if room.GetPeerCount() == 0 {
		s.roomManager.DeleteRoom(roomID)
		sfuLogger.Info("Deleted empty room: %s", roomID)
	}

	return nil
}


func (s *SFUService) HandleMute(roomID, userID string, isMuted bool) error {
	room, err := s.roomManager.GetRoom(roomID)
	if err != nil {
		return err
	}

	peer, exists := room.GetPeer(userID)
	if !exists {
		return models.ErrPeerNotFound
	}

	peer.Mu.Lock()
	peer.IsMuted = isMuted
	peer.Mu.Unlock()

	
	s.broadcastRoomState(room, userID, "user_muted")

	sfuLogger.Info("User %s %s in room %s", userID, map[bool]string{true: "muted", false: "unmuted"}[isMuted], roomID)
	return nil
}



func (s *SFUService) broadcastRoomState(room *models.VoiceRoom, triggerUserID, eventType string) {
	response := models.VoiceResponse{
		Type:   eventType,
		RoomID: room.ID,
		UserID: triggerUserID,
		Users:  room.GetVoiceUsers(),
	}

	jsonData, _ := json.Marshal(response)

	for _, peer := range room.GetAllPeers() {
		if err := peer.WSConnection.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			sfuLogger.Error("Failed to broadcast to user %s: %v", peer.UserID, err)
		}
	}
}



func (s *SFUService) GetRoomStats(roomID string) (map[string]interface{}, error) {
	room, err := s.roomManager.GetRoom(roomID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"roomId":       room.ID,
		"peerCount":    room.GetPeerCount(),
		"users":        room.GetVoiceUsers(),
		"createdAt":    room.CreatedAt,
		"maxParticipants": room.MaxParticipants,
	}, nil
}


func (s *SFUService) GetRoomManager() *RoomManager {
	return s.roomManager
}
