package rounter

import (
	"fmt"
	"go-streaming/model"
	"go-streaming/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 1024 * 1024,
	WriteBufferSize: 1024 * 1024 * 1024,
	//Solving cross-domain problems
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebsocketHandler(checkJwt bool, jwtToken model.JWT, channelMan *model.ChannelManager) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		// validate token
		token, err := jwtToken.Validate(c)
		if err != nil {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "Token invalid",
			})
			return
		}
		log.Printf("Token :%v", token)
		// start ws
		w, r := c.Writer, c.Request
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Err: %v", err)
			return
		}
		defer ws.Close()
		// get parameters
		prefix, keys := utils.GetPrefixStreamingByGinContext(c)
		sseId := uuid.New() // ID of sse connection
		log.Printf("WEBSOCKET | %s | %s/%s", sseId, prefix, keys)
		channelMan.WsTotal += 1
		channelMan.WsLive += 1
		// Create new listener
		listener := channelMan.OpenListener(prefix, keys)
		// Wait for close
		defer channelMan.CloseListener(prefix, keys, listener)
		// Keep connection
		for {
			message := <-listener               // Get message
			wsRes := fmt.Sprintf("%b", message) // Write data
			err = ws.WriteMessage(1, []byte(wsRes))
			if err != nil {
				log.Printf("Write: %v", err)
				break
			}
		}
		log.Printf("DISCONNECT WEBSOCKET | %s | %s - %s", sseId, prefix, keys)
		channelMan.WsClosed += 1
		channelMan.WsLive -= 1
	}
	return gin.HandlerFunc(fn)
}
