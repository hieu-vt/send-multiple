package rounter

import (
	"fmt"
	"go-streaming/model"
	"go-streaming/utils"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
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
		var token jwt.MapClaims
		if checkJwt {
			tokenOutput, jwt, err := jwtToken.Validate(c)
			token = tokenOutput
			if err != nil {
				log.Printf("Jwt token err [%s]: %s", fmt.Sprintf("%v", err), jwt)
				c.JSON(401, gin.H{
					"code":    401,
					"message": fmt.Sprintf("%v", err),
				})
				return
			}
		}
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
		log.Printf("Connect Websocket | %s | %s | %s | %s | %s", token["iss"], token["device_id"], sseId, prefix, keys)
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
		log.Printf("Connect Websocket | %s | %s | %s | %s | %s", token["iss"], token["device_id"], sseId, prefix, keys)
		channelMan.WsClosed += 1
		channelMan.WsLive -= 1
	}
	return gin.HandlerFunc(fn)
}
