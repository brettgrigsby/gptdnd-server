package main

import (
	"encoding/json"
	"fmt"
	"gptdnd-server/internal/utils"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	openai "github.com/sashabaranov/go-openai"
)

type Room struct {
	ID       string `json:"id"`
	Messages []openai.ChatCompletionMessage `json:"messages"`
	Players  []*Player `json:"players"`
}

type Player struct {
	ID   int `json:"id"`
	Name string `json:"name"`
	Conn *websocket.Conn `json:"-"`
}

type WSMessage struct {
	Action string `json:"action"`
	Payload string `json:"payload"`
}

var rooms = make(map[string]*Room)
var roomsLock sync.Mutex

var upgrader = websocket.Upgrader{}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/rooms/{room_id}/messages", getRoomMessages).Methods("GET")
	router.HandleFunc("/rooms/{room_id}", getRoomData).Methods("GET")
	router.HandleFunc("/rooms/{room_id}/join", joinRoom).Methods("POST")
	router.HandleFunc("/rooms/create", createRoom).Methods("POST")

	corsOptions := handlers.CORS(
		handlers.AllowedHeaders([]string{"Content-Type", "Accept"}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"http://localhost:3000"}),
	)
	fmt.Println("Server is running on port 8080")

	log.Fatal(http.ListenAndServe(":8080", corsOptions(router)))
}

func getRoomMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, ok := vars["room_id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	room, ok := rooms[roomID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(room.Messages)
}

func getRoomData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, ok := vars["room_id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	room, ok := rooms[roomID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(room)
}

func joinRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, ok := vars["room_id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	room, ok := rooms[roomID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Read and parse the request body
	var requestBody struct {
		Name string `json:"name"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	player := &Player{
		ID:   len(room.Players),
		Conn: conn,
		Name: requestBody.Name,
	}

	room.Players = append(room.Players, player)
	go handlePlayer(room, player)
}

func handlePlayer(room *Room, player *Player) {
	defer player.Conn.Close()
	defer removePlayerFromRoom(room, player)

	for {
		_, message, err := player.Conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		log.Println("message from player: ", message)

		// msg := openai.ChatCompletionMessage{
		// 	Content: string(message),
		// 	Role: "user",
		// }

		// sendMessageToRoom(room, msg)
	}
}

func sendMessageToRoom(room *Room, message openai.ChatCompletionMessage) {
	roomsLock.Lock()
	room.Messages = append(room.Messages, message)
	roomsLock.Unlock()

	for _, player := range room.Players {
		err := player.Conn.WriteMessage(websocket.TextMessage, []byte(message.Content))
		if err != nil {
			log.Println("Write error:", err)
		}
	}
}

func removePlayerFromRoom(room *Room, player *Player) {
	roomsLock.Lock()
	defer roomsLock.Unlock()

	for i, p := range room.Players {
		if p.ID == player.ID {
			// Remove the player from the slice
			room.Players = append(room.Players[:i], room.Players[i+1:]...)
			break
		}
	}
}

func createRoom(w http.ResponseWriter, r *http.Request) {
	roomsLock.Lock()
	defer roomsLock.Unlock()

	roomID := utils.CreateRoomID()
	// Generate a unique room ID
	for {
		if _, exists := rooms[roomID]; !exists {
			break
		}
	}

	firstMsg := openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleAssistant,
		Content: "Your group starts out into the city. There are many opportunities for work, adventure, profit and trouble.",
	}

	// Create the new room
	room := &Room{
		ID: roomID,
		Messages: []openai.ChatCompletionMessage{firstMsg},
		Players: []*Player{},
	}
	rooms[roomID] = room

	// Return message on the newly established websocket?

	// Return the room ID in the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"room_id": roomID})
}