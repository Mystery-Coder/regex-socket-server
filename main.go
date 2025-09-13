package main

import (
	"fmt"
	"net/http"
	"slices"
	"time"

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

func wsConnectPlayer(ctx *gin.Context) {

	playerID := ctx.Query("PlayerID")
	roomID := ctx.Query("RoomID")

	if playerID == "" {
		fmt.Println("Missing PlayerID")
		ctx.JSON(400, gin.H{"error": "PlayerIDMissingError"})
		return
	}

	if roomID == "" {
		fmt.Println("Missing RoomID")
		ctx.JSON(400, gin.H{"error": "RoomIDMissingError"})
		return
	}

	room, exists := rooms[roomID]
	if !exists {
		fmt.Println("Invalid RoomID")
		ctx.JSON(400, gin.H{"error": "InvalidRoomIDError"})
		return
	}

	if !slices.Contains(room.PlayerIDs, playerID) {
		fmt.Println("Player not added")
		ctx.JSON(400, gin.H{"error": "PlayerNotAddedError"})
		return
	}

	//Upgrade to WS

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		fmt.Println("Upgrade Error")
		ctx.JSON(500, gin.H{"error": "WebSocketUpgradeError"})
		return
	}
	defer conn.Close()

	room.Players = append(room.Players, Player{PlayerID: playerID, Connection: conn})

	for {
		if len(room.Players) != 2 {
			time.Sleep(time.Second * 2)
			conn.WriteMessage(websocket.TextMessage, []byte("waiting..."))
		} else {
			for _, player := range room.Players {
				player.Connection.WriteMessage(websocket.TextMessage, []byte("2 Players connected"))
			}
			return
		}
	}

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
