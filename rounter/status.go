package rounter

import (
	"go-streaming/model"
	"io"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func StatusHandler(channelMan *model.ChannelManager) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		path := "streaming"
		channelId := "status"

		sseId := uuid.New()
		log.Printf("CONNECT SSE | %s | %s/%s", sseId, path, channelId)
		// Create new listener
		listener := channelMan.OpenListener(path, channelId)
		// Wait for close
		defer channelMan.CloseListener(path, channelId, listener)
		clientGone := c.Request.Context().Done()
		// Keep connection
		c.Stream(func(w io.Writer) bool {
			select {
			case <-clientGone: // Close connection
				log.Printf("DISCONNECT SSE | %s | %s/%s", sseId, path, channelId)
				return false
			case message := <-listener: // Send message
				c.SSEvent("", message)
				return true
			}
		})
	}
	return gin.HandlerFunc(fn)
}
