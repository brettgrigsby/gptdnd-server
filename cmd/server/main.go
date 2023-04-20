package main

import (
	"encoding/json"
	"fmt"
	"gptdnd-server/internal/libs"
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
	ID   string `json:"id"`
	Name string `json:"name"`
	Conn *websocket.Conn `json:"-"`
}

type WSMessage struct {
	Action string `json:"action"`
	Payload string `json:"payload"`
}

var rooms = make(map[string]*Room)
var roomsLock sync.Mutex

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/rooms/{room_id}/messages", getRoomMessages).Methods("GET")
	router.HandleFunc("/rooms/{room_id}", getRoomData).Methods("GET")
	router.HandleFunc("/rooms/{room_id}/join", joinRoom).Methods("POST")
	router.HandleFunc("/rooms/{room_id}/players/{player_id}/ws", connectToRoom).Methods("GET")
	router.HandleFunc("/rooms/create", createRoom).Methods("POST")

	corsOptions := handlers.CORS(
		handlers.AllowedHeaders([]string{"Content-Type", "Accept"}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "UPGRADE", "WEBSOCKET"}),
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

func connectToRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, ok := vars["room_id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	playerID, ok := vars["player_id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	room, ok := rooms[roomID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var player *Player
	for _, p := range room.Players {
		if p.ID == playerID {
			player = p
			break
		}
	}
	if player == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	player.Conn = conn
	go handlePlayer(room, player)
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
		fmt.Println("Error parsing request body:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// conn, err := upgrader.Upgrade(w, r, nil)
	// if err != nil {
	// 	log.Println("Upgrade error:", err)
	// 	return
	// }

	player := &Player{
		ID:   utils.CreateUUID(),
		Name: requestBody.Name,
	}
	room.Players = append(room.Players, player)
	json.NewEncoder(w).Encode(player)
	// go handlePlayer(room, player)
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

		// log message as string
		log.Println("message from player: ", string(message))

		msg := openai.ChatCompletionMessage{
			Content: player.Name + ": " + string(message),
			Role: openai.ChatMessageRoleUser,
		}

		sendMessageToRoom(room, msg)

		// send message to openai
		responseMsg, err := libs.GetChatCompletion(room.Messages)
		if err != nil {
			log.Println("Error getting chat completion:", err)
			return
		}

		sendMessageToRoom(room, responseMsg)
	}
}

func sendMessageToRoom(room *Room, message openai.ChatCompletionMessage) {
	roomsLock.Lock()
	room.Messages = append(room.Messages, message)
	roomsLock.Unlock()

	// marshal message to json
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Println("Error marshalling message:", err)
		return
	}

	for _, player := range room.Players {
		err := player.Conn.WriteMessage(websocket.TextMessage, []byte(jsonMessage))
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