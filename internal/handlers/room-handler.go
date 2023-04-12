package handlers

import (
	"fmt"
	"gptdnd-server/internal/models"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sashabaranov/go-openai"
)

var repo models.RoomRepository = models.NewInMemoryRoomRepository()

type WSAction int
const (
	SendMessage WSAction = iota
	JoinRoom
)

func RoomHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request URL:", r.URL.Path)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	roomID := r.URL.Path[len("/rooms/"):]
	fmt.Println("Room ID:", roomID)
}

func RoomWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()
	
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			return
		}

		// get action and payload from message
		fmt.Println("Message received:", string(p))
	}
}

func handleJoinRoom(roomID string, name string, class models.PlayerClass, race models.PlayerRace, conn *websocket.Conn) {
	player := models.CreatePlayer(name, class, race, conn)
	repo.AddPlayer(roomID, player)
}

func handleNewMessage(roomID string, playerID string, message string) {
	// check if last message was from the AI
	room := repo.Find(roomID)
	lastMsg := room.Messages[len(room.Messages)-1]
	if lastMsg.Role == "user" {
		return
	}

	msg := openai.ChatCompletionMessage{
		Role: "user",
		Content: message,
	}
	repo.PushMessage(roomID, msg)

	// generate response
	// push response to room
	// send response to all players

}
