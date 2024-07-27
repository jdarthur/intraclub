package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"time"
)

var WebsocketPingInterval = time.Minute * 20
var upgrader websocket.Upgrader

var connections map[*websocket.Conn]chan bool

func WebsocketInit() {
	connections = make(map[*websocket.Conn]chan bool)
}

func WebsocketReadAudit(conn *websocket.Conn, c chan bool) {
	for {
		err := conn.SetWriteDeadline(time.Now().Add(WebsocketPingInterval))
		if err != nil {
			fmt.Printf("Error setting read deadline: %s\n", err.Error())
			close(c)
			return
		}

		_, _, err = conn.ReadMessage()
		if err != nil {
			close(c)
			return
		}

	}
}

func WebSocketServer(c *gin.Context) {

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("New connection: %s\n", conn.RemoteAddr().String())

	connections[conn] = make(chan bool)

	// start a thread that will try to read a ping off the websocket and close
	// it if we reach a long read timeout
	go WebsocketReadAudit(conn, connections[conn])

	defer CloseConn(conn)
	for range connections[conn] {

		err := conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
		if err != nil {
			fmt.Printf("Error setting write deadline: %s\n", err.Error())
			return
		}

		err = conn.WriteMessage(websocket.TextMessage, []byte("{\"changed\": true}"))
		if err != nil {
			fmt.Printf("Error writing to connection %s: %s\n", conn.RemoteAddr().String(), err.Error())
			return
		}
	}
}

func UpdateOccurred() {
	for _, channel := range connections {
		channel <- true
	}
}

func CloseAllWebsocketConnections() {
	for conn, channel := range connections {
		fmt.Printf("Closing connection %s\n", conn.RemoteAddr().String())
		close(channel)
	}
}

func CloseConn(conn *websocket.Conn) {
	fmt.Printf("Closing connection: %s\n", conn.RemoteAddr().String())

	delete(connections, conn)

	err := conn.Close()
	if err != nil {
		fmt.Printf("Error closing connection: %s\n", err.Error())
	}

}
