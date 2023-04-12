package rounter

import (
	"fmt"
	"go-streaming/model"
	"go-streaming/utils"
	"io"
	"log"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SseHandler(checkJwt bool, jwtToken model.JWT, channelMan *model.ChannelManager) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		// validate token
		var token jwt.MapClaims
		if checkJwt {
			tokenOutput, jwt, err := jwtToken.Validate(c)
			token = tokenOutput
			if err != nil {
				log.Printf("Jwt token err [%s] | %s | %s", fmt.Sprintf("%v", err), c.Request.RequestURI, jwt)
				c.JSON(401, gin.H{
					"code":    401,
					"message": fmt.Sprintf("%v", err),
				})
				return
			}
		}
		// start sse
		prefix, keys := utils.GetPrefixStreamingByGinContext(c)
		sseId := uuid.New() // ID of sse connection
		log.Printf("Connect SSE | %s | %s | %s | %s | %s", token["iss"], token["device_id"], sseId, prefix, keys)
		// Create new listener
		listener := channelMan.OpenListener(prefix, keys)
		// Wait for close
		defer channelMan.CloseListener(prefix, keys, listener)
		clientGone := c.Request.Context().Done()
		// Keep connection
		c.Stream(func(w io.Writer) bool {
			select {
			case <-clientGone: // Close connection
				log.Printf("Disconnect SSE | %s | %s | %s | %s | %s", token["iss"], token["device_id"], sseId, prefix, keys)
				return false
			case message := <-listener: // Send message
				c.SSEvent("", message)
				return true
			}
		})
	}
	return gin.HandlerFunc(fn)
}
