package models

import (
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
	ID	  string
	Name  string
	Class PlayerClass
	Race  PlayerRace
	Conn  *websocket.Conn
}