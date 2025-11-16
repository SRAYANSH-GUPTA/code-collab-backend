package services

import (
	"sync"

	"codecollab/models"
)



type RoomManager struct {
	rooms map[string]*models.VoiceRoom 
	mu    sync.RWMutex
}


func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*models.VoiceRoom),
	}
}



func (rm *RoomManager) CreateRoom(roomID, name string) *models.VoiceRoom {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	
	if room, exists := rm.rooms[roomID]; exists {
		return room
	}

	room := models.NewVoiceRoom(roomID, name)
	rm.rooms[roomID] = room
	return room
}


func (rm *RoomManager) GetRoom(roomID string) (*models.VoiceRoom, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return nil, models.ErrRoomNotFound
	}
	return room, nil
}



func (rm *RoomManager) GetOrCreateRoom(roomID, name string) *models.VoiceRoom {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if room, exists := rm.rooms[roomID]; exists {
		return room
	}

	room := models.NewVoiceRoom(roomID, name)
	rm.rooms[roomID] = room
	return room
}



func (rm *RoomManager) DeleteRoom(roomID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return models.ErrRoomNotFound
	}

	
	for _, peer := range room.GetAllPeers() {
		if peer.PeerConnection != nil {
			peer.PeerConnection.Close()
		}
	}

	delete(rm.rooms, roomID)
	return nil
}



func (rm *RoomManager) CleanupEmptyRooms() int {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	cleaned := 0
	for roomID, room := range rm.rooms {
		if room.GetPeerCount() == 0 {
			delete(rm.rooms, roomID)
			cleaned++
		}
	}
	return cleaned
}


func (rm *RoomManager) GetAllRooms() []*models.VoiceRoom {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rooms := make([]*models.VoiceRoom, 0, len(rm.rooms))
	for _, room := range rm.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}


func (rm *RoomManager) GetRoomCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return len(rm.rooms)
}


func (rm *RoomManager) GetTotalPeerCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	total := 0
	for _, room := range rm.rooms {
		total += room.GetPeerCount()
	}
	return total
}
