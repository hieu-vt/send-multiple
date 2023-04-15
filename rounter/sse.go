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

		s := strings.Split(keys, ",")
		// Wait for close

		stopChan := make(chan bool, len(s))
		listenerCh := make(chan interface{}, len(s))

		defer close(stopChan)

		for i := 0; i < len(s); i++ {
			go func(index int, lisChan chan interface{}, stopCh chan bool) {
				channelId := prefix + ":" + s[index]
				log.Println("ChannelId: ", channelId)
				ch := channelEngine.Listener(channelId)

				if ch != nil {
					for {
						select {
						case mess := <-ch:
							lisChan <- mess
						case isStop := <-stopCh:
							if isStop {
								channelEngine.DeleteChildChannel(ch, channelId)
							}
						}
					}

				}

			}(i, listenerCh, stopChan)
		}

		clientGone := c.Request.Context().Done()

		// Keep connection
		c.Stream(func(w io.Writer) bool {
			select {
			case <-clientGone: // Close connection
				log.Printf("Disconnect SSE | %s | %s | %s | %s | %s", token["iss"], token["device_id"], sseId, prefix, keys)
				for i := 0; i < len(s); i++ {
					stopChan <- true
				}
				return false
			case message := <-listenerCh: // Send message
				log.Println("message: ", message)
				c.SSEvent("", message)
				return true
			}
		})
	}
}
