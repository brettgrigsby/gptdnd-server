package libs

import (
	"gptdnd-server/internal/models"
)

type RoomManager struct {
	Rooms map[string][]*models.Player
}

func NewRoomsManager() *RoomManager {
	return &RoomManager{
		Rooms: make(map[string][]*models.Player),
	}
}

func (rm *RoomManager) AddPlayerToRoom(roomID string, player *models.Player) {
	rm.Rooms[roomID] = append(rm.Rooms[roomID], player)
}

func (rm *RoomManager) RemovePlayerFromRoom(roomID string, player *models.Player) {
	players, ok := rm.Rooms[roomID]
	if (!ok) {
		return
	}

	for i, p := range players {
		if p.ID == player.ID {
			players = append(players[:i], players[i+1:]...)
			break
		}
	}
	rm.Rooms[roomID] = players
}