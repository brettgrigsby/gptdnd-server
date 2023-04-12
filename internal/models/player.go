package models

import (
	"gptdnd-server/internal/utils"

	"github.com/gorilla/websocket"
)

type PlayerClass int

const (
	Cleric PlayerClass = iota
	Druid
	Fighter
	Rogue
	Sorcerer
	Wizard
)

type PlayerRace int

const (
	Elf PlayerRace = iota
	Dwarf
	Halfling
	Human 
	Orc
)

type Player struct {
	ID	  string `json:"id"`
	Name  string `json:"name"`
	Class PlayerClass `json:"class"`
	Race  PlayerRace `json:"race"`
	Conn  *websocket.Conn `json:"-"`
}

func CreatePlayer(name string, class PlayerClass, race PlayerRace, conn *websocket.Conn) *Player {
	return &Player{
		ID: utils.CreateUUID(),
		Name: name,
		Class: class,
		Race: race,
		Conn: conn,
	}
}