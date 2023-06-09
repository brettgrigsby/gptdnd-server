package utils

import (
	"math/rand"
	"time"
)

func CreateRoomID() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, 4)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func CreateUUID() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, 16)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}