package main

import "github.com/gorilla/websocket"

type Player struct {
	PlayerID   string
	Connection *websocket.Conn
}

type Room struct {
	RoomID       string
	Players      []Player
	PlayerIDs    []string
	RoomQuestion Question
}

type RoomIdentify struct {
	RoomID string
}

type PlayerConnectIdentify struct {
	RoomID   string
	PlayerID string
}

type StringQuestion struct {
	Question []string
}

type RegexQuestion struct {
	Question string
}

// Message and its types
type Status struct {
	Status string
}

type PlayerGuess struct { //Can be regex or string
	PlayerID string
	Guess    string
	Type     string
}

type Question struct {
	QuestionType string
	Pattern      string
	Options      []string
}

type Message[T any] struct {
	Type string
	Data T
}
