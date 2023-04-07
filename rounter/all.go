package rounter

import (
	"fmt"
	"go-streaming/model"
	"io"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AllHandler(checkJwt bool, jwtToken model.JWT, channelMan *model.ChannelManager) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		// validate token
		if checkJwt {
			token, err := jwtToken.Validate(c)
			if err != nil {
				c.JSON(401, gin.H{
					"code":    401,
					"message": fmt.Sprintf("%v", err),
				})
				return
			}
			log.Printf("Token :%v", token)
		}
		prefix, keys := "ALL", "ALL"
		sseId := uuid.New() // ID of sse connection
		log.Printf("SSE | %s | %s - %s", sseId, prefix, keys)
		channelMan.SseTotal += 1
		channelMan.SseLive += 1
		// Create new listener
		listener := channelMan.OpenListener(prefix, keys)
		// Wait for close
		defer channelMan.CloseListener(prefix, keys, listener)
		clientGone := c.Request.Context().Done()
		// Keep connection
		c.Stream(func(w io.Writer) bool {
			select {
			case <-clientGone: // Close connection
				log.Printf("DISCONNECT SSE | %s | %s - %s", sseId, prefix, keys)
				channelMan.SseClosed += 1
				channelMan.SseLive -= 1
				return false
			case message := <-listener: // Send message
				c.SSEvent("", message)
				return true
			}
		})
	}
	return gin.HandlerFunc(fn)
}
