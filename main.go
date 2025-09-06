package main

import (
	"fmt"
	"net/http"

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
	SocketID   string
	Connection *websocket.Conn
}

type Room struct {
	RoomID  string
	Players []Player
}

func wsConnectPlayer(ctx *gin.Context) {

	RoomID := ctx.Query("RoomID")

	if RoomID == "" {
		fmt.Println("No RoomID")
		ctx.JSON(400, gin.H{
			"error": "RoomID missing",
		})
		return
	}

	//Connection Should not be closed
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		fmt.Println("Upgrade failed", err)
		return
	}

	room := rooms[RoomID]
	newPlayer := Player{SocketID: uuid.NewString(), Connection: conn}
	players = append(players, newPlayer)
	room.Players = append(room.Players, newPlayer)

	if len(room.Players) == 2 {

	}

}

var players []Player
var rooms map[string]*Room

func main() {
	router := gin.Default()

	router.GET("/connect_player", wsConnectPlayer)

	//Create Room, should not be WS
	router.GET("/create_room", func(ctx *gin.Context) {
		newRoomID := uuid.NewString()
		rooms[newRoomID] = &Room{RoomID: newRoomID}
		ctx.JSON(200, gin.H{"RoomID": newRoomID})
	})

	router.Run(":8080")
}
