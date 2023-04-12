package models

import (
	"encoding/json"
	"fmt"
	"gptdnd-server/internal/libs"
	"gptdnd-server/internal/utils"

	"github.com/sashabaranov/go-openai"
)

type Room struct {
	ID   		 string `json:"id"`
	Players  []*Player `json:"players"`
	Messages []openai.ChatCompletionMessage `json:"messages"`
}

type RoomRepository interface {
	Create() *Room
	Find(id string) *Room
	AddPlayer(id string, player *Player) *Room
	RemovePlayer(id string, playerID string) *Room
	PushMessage(id string, message openai.ChatCompletionMessage)
	HandleUserMessage(id string, playerID string, message string)
}

type InMemoryRoomRepository struct {
	rooms map[string]*Room
}

func NewInMemoryRoomRepository() *InMemoryRoomRepository {
	return &InMemoryRoomRepository{
		rooms: make(map[string]*Room),
	}
}

func (r *InMemoryRoomRepository) Create() *Room {
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

func (r *InMemoryRoomRepository) Find(id string) *Room {
	return r.rooms[id]
}

func (r *InMemoryRoomRepository) AddPlayer(id string, player *Player) *Room {
	room := r.rooms[id]
	room.Players = append(room.Players, player)
	return room
}

func (r *InMemoryRoomRepository) RemovePlayer(id string, playerID string) *Room {
	room := r.rooms[id]
	for i, p := range room.Players {
		if p.ID == playerID {
			room.Players = append(room.Players[:i], room.Players[i+1:]...)
		}
	}
	return room
}

func (r *InMemoryRoomRepository) PushMessage(id string, message openai.ChatCompletionMessage) {
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

func (r *InMemoryRoomRepository) HandleUserMessage(id string, playerID string, message string) {
	room := r.rooms[id]

	// check if last message was from the AI
	if room.Messages[len(room.Messages)-1].Role == "user" {
		return
	}

	for _, p := range room.Players {
		if p.ID == playerID {
			message = fmt.Sprintf("%s: %s", p.Name, message)
		}
	}

	// add message to room
	msg := openai.ChatCompletionMessage{
		Role: "user",
		Content: message,
	}
	r.PushMessage(id, msg)

	msg, err := libs.GetChatCompletion(room.Messages)	
	if err != nil {
		fmt.Println("Error getting chat completion:", err)
		return
	}

	r.PushMessage(id, msg)
}

