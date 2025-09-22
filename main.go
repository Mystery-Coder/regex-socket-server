package main

import (
	"fmt"
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

type Room struct {
	RoomID    string
	Players   []Player
	PlayerIDs []string
}

type RoomIdentify struct {
	RoomID string
}

type PlayerConnectIdentify struct {
	RoomID   string
	PlayerID string
}

type RoomStatus struct {
	Status string
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

	// Notify others
	msg := Message[RoomStatus]{Type: "STATUS", Data: RoomStatus{Status: "WAITING"}}
	if len(room.Players) == 2 {
		msg = Message[RoomStatus]{Type: "STATUS", Data: RoomStatus{Status: "PLAYER2CONNECTED"}}
	}

	broadcastRoom(room, msg)

	// Handle read loop (so connection stays alive, read buffer can get full, must be read)
	go func() {
		defer func() {
			removePlayer(room, playerID)
			conn.Close()
			msg := Message[RoomStatus]{Type: "STATUS", Data: RoomStatus{Status: "WAITING"}}
			if len(room.Players) == 2 {
				msg = Message[RoomStatus]{Type: "STATUS", Data: RoomStatus{Status: "PLAYER2CONNECTED"}}
			}
			broadcastRoom(room, msg)
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Read error:", err)
				return
			}
		}
	}()
}

var rooms map[string]*Room

func main() {
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
	router.GET("/create_room", func(ctx *gin.Context) {
		newRoomID := uuid.NewString()
		rooms[newRoomID] = &Room{RoomID: newRoomID}
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

	router.Run(":8080")
}
