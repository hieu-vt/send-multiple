package rounter

import (
	"go-streaming/engine"
	"go-streaming/utils"
	"io"
	"log"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SseHandler(channelEngine *engine.ChannelEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		// validate token
		stopChan := make(chan bool, 1)
		var token jwt.MapClaims
		//if checkJwt {
		//	tokenOutput, jwt, err := jwtToken.Validate(c)
		//	token = tokenOutput
		//	if err != nil {
		//		log.Printf("Jwt token err [%s] | %s | %s", fmt.Sprintf("%v", err), c.Request.RequestURI, jwt)
		//		c.JSON(401, gin.H{
		//			"code":    401,
		//			"message": fmt.Sprintf("%v", err),
		//		})
		//		return
		//	}
		//}
		// start sse
		prefix, keys := utils.GetPrefixStreamingByGinContext(c)
		sseId := uuid.New() // ID of sse connection
		log.Printf("Connect SSE | %s | %s | %s | %s | %s", token["iss"], token["device_id"], sseId, prefix, keys)
		// Create new listener
		listenerCh := make(chan interface{})

		s := strings.Split(keys, ",")
		// Wait for close
		clientGone := c.Request.Context().Done()

		for i := 0; i < len(s); i++ {
			go func(i int, lisChan chan interface{}) {
				channelId := prefix + ":" + s[i]
				log.Println("ChannelId: ", channelId)
				ch := channelEngine.Listener(channelId)
				if ch != nil {
					for mess := range ch {
						log.Println("mess: ", mess)
						lisChan <- mess
					}
				}
				//if <-stopChan {
				//	channelEngine.DeleteChildChannel(c, cId)
				//}
			}(i, listenerCh)
		}

		// Keep connection
		c.Stream(func(w io.Writer) bool {
			select {
			case <-clientGone: // Close connection
				log.Printf("Disconnect SSE | %s | %s | %s | %s | %s", token["iss"], token["device_id"], sseId, prefix, keys)
				stopChan <- true
				return false
			case message := <-listenerCh: // Send message
				log.Println("message: ", message)
				c.SSEvent("", message)
				return true
			}
		})
	}
}
