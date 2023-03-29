package handlers

import (
	"fmt"
	"net/http"
)

func RoomHandler(w http.ResponseWriter, r *http.Request) {
	// get room id from request
	roomID := r.URL.Path[len("/rooms/"):]
	fmt.Println("Room ID:", roomID)
}
