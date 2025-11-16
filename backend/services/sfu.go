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
	room := s.roomManager.GetOrCreateRoom(roomID, roomID)

	
	peerConnection, err := webrtc.NewPeerConnection(s.config)
	if err != nil {
		sfuLogger.Error("Failed to create peer connection for user %s: %v", userID, err)
		return err
	}

	sfuLogger.Info("Created peer connection for user %s in room %s", userID, roomID)

	
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

	
	if err := room.AddPeer(peer); err != nil {
		peerConnection.Close()
		return err
	}

	
	s.setupPeerHandlers(peer, room)

	sfuLogger.Info("User %s joined room %s (total: %d users)", userID, roomID, room.GetPeerCount())

	
	s.broadcastRoomState(room, userID, "user_joined")

	return nil
}



func (s *SFUService) setupPeerHandlers(peer *models.VoicePeer, room *models.VoiceRoom) {
	pc := peer.PeerConnection

	
	
	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		sfuLogger.Info("Received track from user %s: %s", peer.UserID, track.Kind())

		peer.Mu.Lock()
		peer.RemoteTracks = append(peer.RemoteTracks, track)
		peer.Mu.Unlock()

		
		s.forwardTrackToOthers(peer, track, room)
	})

	
	
	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		sfuLogger.Info("User %s ICE state: %s", peer.UserID, state.String())

		switch state {
		case webrtc.ICEConnectionStateDisconnected:
			sfuLogger.Warn("User %s disconnected", peer.UserID)
		case webrtc.ICEConnectionStateFailed:
			sfuLogger.Error("User %s connection failed", peer.UserID)
			s.HandleLeave(room.ID, peer.UserID)
		case webrtc.ICEConnectionStateClosed:
			sfuLogger.Info("User %s connection closed", peer.UserID)
		}
	})

	
	
	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}

		candidateInit := candidate.ToJSON()
		response := models.VoiceResponse{
			Type:      "ice_candidate",
			RoomID:    room.ID,
			UserID:    peer.UserID,
			Candidate: &candidateInit,
		}

		if err := peer.WSConnection.WriteJSON(response); err != nil {
			sfuLogger.Error("Failed to send ICE candidate to user %s: %v", peer.UserID, err)
		}
	})

	
	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		sfuLogger.Info("User %s connection state: %s", peer.UserID, state.String())
	})
}



func (s *SFUService) addExistingTracksToNewPeer(newPeer *models.VoicePeer, room *models.VoiceRoom) {
	for _, existingPeer := range room.GetAllPeers() {
		if existingPeer.UserID == newPeer.UserID {
			continue 
		}

		
		existingPeer.Mu.RLock()
		for _, localTrack := range existingPeer.LocalTracks {
			_, err := newPeer.PeerConnection.AddTrack(localTrack)
			if err != nil {
				sfuLogger.Error("Failed to add existing track from %s to new peer %s: %v",
					existingPeer.UserID, newPeer.UserID, err)
				continue
			}
			sfuLogger.Info("Added existing track from %s to new peer %s",
				existingPeer.UserID, newPeer.UserID)
		}
		existingPeer.Mu.RUnlock()
	}
}



func (s *SFUService) forwardTrackToOthers(sender *models.VoicePeer, incomingTrack *webrtc.TrackRemote, room *models.VoiceRoom) {
	
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

	
	for _, otherPeer := range room.GetAllPeers() {
		if otherPeer.UserID == sender.UserID {
			continue 
		}

		rtpSender, err := otherPeer.PeerConnection.AddTrack(localTrack)
		if err != nil {
			sfuLogger.Error("Failed to add track to peer %s: %v", otherPeer.UserID, err)
			continue
		}

		
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



func (s *SFUService) HandleOffer(roomID, userID string, offer webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	room, err := s.roomManager.GetRoom(roomID)
	if err != nil {
		return nil, err
	}

	peer, exists := room.GetPeer(userID)
	if !exists {
		return nil, models.ErrPeerNotFound
	}

	
	if err := peer.PeerConnection.SetRemoteDescription(offer); err != nil {
		sfuLogger.Error("Failed to set remote description: %v", err)
		return nil, err
	}

	
	
	s.addExistingTracksToNewPeer(peer, room)

	
	answer, err := peer.PeerConnection.CreateAnswer(nil)
	if err != nil {
		sfuLogger.Error("Failed to create answer: %v", err)
		return nil, err
	}

	
	if err := peer.PeerConnection.SetLocalDescription(answer); err != nil {
		sfuLogger.Error("Failed to set local description: %v", err)
		return nil, err
	}

	sfuLogger.Info("Created answer for user %s with %d existing tracks", userID, len(peer.LocalTracks))
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
