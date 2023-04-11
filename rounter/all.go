package rounter

import (
	"go-streaming/model"
	"io"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AllHandler(checkJwt bool, jwtToken model.JWT, channelMan *model.ChannelManager) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		prefix, keys := "ALL", "ALL"
		sseId := uuid.New() // ID of sse connection
		log.Printf("Connect SSE | %s | %s | %s", sseId, prefix, keys)
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
				log.Printf("Disonnect SSE | %s | %s | %s", sseId, prefix, keys)
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
