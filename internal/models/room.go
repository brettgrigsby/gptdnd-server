package models

import (
	"encoding/json"
	"fmt"
	"gptdnd-server/internal/utils"

	"github.com/sashabaranov/go-openai"
)

type Room struct {
	ID   		 string
	Players  []*Player
	Messages []openai.ChatCompletionMessage
}

type RoomRepository interface {
	createRoom() *Room
	findRoom(id string) *Room
	joinRoom(id string, player *Player) *Room
	pushMessageToRoom(id string, message openai.ChatCompletionMessage)
}

type InMemoryRoomRepository struct {
	rooms map[string]*Room
}

func (r *InMemoryRoomRepository) createRoom() *Room {
	// generate a unique room ID
	roomID := utils.CreateRoomID()
	existingRoom := r.rooms[roomID]
	for existingRoom != nil {
		roomID = utils.CreateRoomID()
		existingRoom = r.rooms[roomID]
	}

	room := &Room{
		ID: roomID,
		Players: []*Player{},
	}
	r.rooms[room.ID] = room
	return room
}

func findRoom(r *InMemoryRoomRepository, id string) *Room {
	return r.rooms[id]
}

func joinRoom(r *InMemoryRoomRepository, id string, player *Player) *Room {
	room := r.rooms[id]
	room.Players = append(room.Players, player)
	return room
}

func pushMessageToRoom(r *InMemoryRoomRepository, id string, message openai.ChatCompletionMessage) {
	room := r.rooms[id]

	// add message to room
	room.Messages = append(room.Messages, message)

	// push message to all players in room
	jsonData, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshalling message:", err)
		return
	}
	for _, p := range room.Players {
		p.Conn.WriteMessage(1, []byte(jsonData))
	}
}

