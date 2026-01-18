package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"slices"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins
}

type Player struct {
	PlayerID   string
	Connection *websocket.Conn
}

type Question struct {
	QuestionType string
	Pattern      string
	Options      []string
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

type Message[T any] struct {
	Type string
	Data T
}

func broadcastRoom[T any](room *Room, msg Message[T]) {

	for _, p := range room.Players {
		if p.Connection != nil {
			p.Connection.WriteJSON(msg)
		}
	}
}

func removePlayer(room *Room, playerID string) {
	newPlayers := []Player{}
	for _, p := range room.Players {
		if p.PlayerID != playerID {
			newPlayers = append(newPlayers, p)
		}
	}
	room.Players = newPlayers
}

func wsConnectPlayer(ctx *gin.Context) {
	playerID := ctx.Query("PlayerID")
	roomID := ctx.Query("RoomID")

	if playerID == "" || roomID == "" {
		ctx.JSON(400, gin.H{"error": "Missing IDs"})
		return
	}

	room, exists := rooms[roomID]
	if !exists {
		ctx.JSON(400, gin.H{"error": "InvalidRoomID"})
		return
	}

	if !slices.Contains(room.PlayerIDs, playerID) {
		ctx.JSON(400, gin.H{"error": "PlayerNotAdded"})
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		fmt.Println("Upgrade Error:", err)
		return
	}

	player := Player{PlayerID: playerID, Connection: conn}
	room.Players = append(room.Players, player)

	// Notify others on first join
	var status string
	if len(room.Players) == 1 {
		status = "WAITING"
		msg := Message[Status]{Type: "STATUS", Data: Status{Status: status}}
		broadcastRoom(room, msg)
	} else if len(room.Players) == 2 {
		status = "PLAYER2CONNECTED"
		msg := Message[Status]{Type: "STATUS", Data: Status{Status: status}}
		broadcastRoom(room, msg)

		questionMsg := Message[Question]{
			Type: "QUESTION",
			Data: room.RoomQuestion,
		}
		broadcastRoom(room, questionMsg)
	}

	// Handle read loop (so connection stays alive, read buffer can get full, must be read)
	go func() {
		defer func() {
			removePlayer(room, playerID)
			conn.Close()
			var status string
			if len(room.Players) == 1 {
				status = "WAITING"
			} else if len(room.Players) == 2 {
				status = "PLAYER2CONNECTED"
			}
			msg := Message[Status]{Type: "STATUS", Data: Status{Status: status}}
			broadcastRoom(room, msg)
		}()

		for { // This is the persistent Read
			_, msgBytes, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Read error:", err)
				return
			}
			// fmt.Println(string(msgBytes))

			var playerGuess PlayerGuess
			json.Unmarshal(msgBytes, &playerGuess)

			fmt.Println(playerGuess)

			// data := strings.Split(string(msgBytes), ":")
			// playerIDFromMsg := data[0]
			// guessFromMsg := data[1]

			// var guessMsg = Message[PlayerGuess]{Type: "PLAYERGUESS", Data: PlayerGuess{PlayerID: playerIDFromMsg, Guess: guessFromMsg}}
			// broadcastRoom(room, guessMsg)
		}
	}()
}

var rooms map[string]*Room

func main() {

	regexQuestions := []string{"\\d{3}\\w"}

	stringQuestions := [][]string{{"aa", "bb", "cc"}}

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	rooms = make(map[string]*Room)

	router.GET("/connect_player", wsConnectPlayer)

	//Create Room, should not be WS
	router.GET("/create_room", func(ctx *gin.Context) { //Takes question_type as query parameter
		questionType := ctx.Query("question_type")

		if questionType != "regex" && questionType != "strings" {
			ctx.JSON(400, gin.H{"error": "InvalidOrMissingType"})
			return
		}

		var randomIdx int
		var roomQuestion Question
		if questionType == "regex" {
			randomIdx = rand.Intn(len(regexQuestions))

			roomQuestion.QuestionType = "regex"
			roomQuestion.Pattern = regexQuestions[randomIdx] //\\d{3}\\w
		} else {
			randomIdx = rand.Intn(len(stringQuestions))
			strings_q := stringQuestions[randomIdx] // {"aa", "bb", "cc"}

			roomQuestion.QuestionType = "strings"
			roomQuestion.Options = strings_q
		}

		newRoomID := uuid.NewString()
		rooms[newRoomID] = &Room{RoomID: newRoomID, RoomQuestion: roomQuestion}
		fmt.Println(rooms)
		ctx.JSON(200, gin.H{"RoomID": newRoomID})
	})

	//Add player to created room
	router.POST("/add_player", func(ctx *gin.Context) {
		var roomIdentify RoomIdentify
		if err := ctx.BindJSON(&roomIdentify); err != nil {
			fmt.Println("Binding err", err)
			ctx.JSON(500, gin.H{"error": "BindError"})
			return
		}

		room, exists := rooms[roomIdentify.RoomID]

		if !exists {
			fmt.Println("Invalid RoomID")
			ctx.JSON(400, gin.H{"error": "InvalidRoomIDError"})
			return
		}
		playerID := uuid.NewString()

		noOfPlayers := len(room.PlayerIDs)
		if noOfPlayers == 2 {
			fmt.Println("Room Full")
			ctx.JSON(400, gin.H{"error": "RoomFullError"})
			return
		}
		room.PlayerIDs = append(room.PlayerIDs, playerID)

		ctx.JSON(200, gin.H{"PlayerID": playerID})

	})

	// router.GET("/question", func(ctx *gin.Context) {
	// 	questionType := ctx.Query("type")

	// 	if questionType != "regex" && questionType != "strings" {
	// 		ctx.JSON(400, gin.H{"error": "InvalidOrMissingType"})
	// 		return
	// 	}

	// 	var randomIdx int
	// 	if questionType == "regex" {
	// 		randomIdx = rand.Intn(len(regexQuestions))
	// 		ctx.JSON(200, regexQuestions[randomIdx])
	// 	} else {
	// 		randomIdx = rand.Intn(len(stringQuestions))
	// 		ctx.JSON(200, stringQuestions[randomIdx])
	// 	}

	// })

	router.Run(":8080")
}
