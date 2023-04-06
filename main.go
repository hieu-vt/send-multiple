package main

import (
	"context"
	"fmt"
	"go-streaming/model"
	"go-streaming/utils"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/penglongli/gin-metrics/ginmetrics"
)

var jwtToken model.JWT
var sseInstanceId string = uuid.New().String() // uuid of see service
var channelManager *model.ChannelManager       // channel manager

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 1024 * 1024,
	WriteBufferSize: 1024 * 1024 * 1024,
	//Solving cross-domain problems
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	log.SetPrefix("GO-STREAMING: ")
	log.SetFlags(log.Ldate | log.Lmsgprefix | log.Ltime | log.Lshortfile)
	gin.SetMode(gin.ReleaseMode)                                      // Set release mode
	config, _ := utils.LoadConfiguration("config/streaming-api.json") // Get config streaming
	prvKey, _ := utils.LoadFile("config/config-jwtrs256-key")
	pubKey, _ := utils.LoadFile("config/config-jwtrs256-key-pem")
	jwtToken = model.NewJWT([]byte(prvKey), []byte(pubKey))

	log.Printf(config.Redis.Host)
	channelManager = model.NewChannelManager() // Init channel manager
	prefix := config.PrefixChannel
	// Config redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Host + ":" + config.Redis.Port, // Address or redis. Should separate redis to improve performance
		Password: config.Redis.Password,                       // Password
		DB:       config.Redis.Database,                       // Database
	})
	channel := "*"
	if len(prefix) > 0 {
		channel = fmt.Sprintf("%s:*", prefix)
	}
	subscribeRedis := rdb.PSubscribe(context.Background(), channel) // Subscribe all channel in redis pubsub
	go func() {
		// Get message from redis pub/sub
		for msg := range subscribeRedis.Channel() { // Listen redis pubsub
			go utils.SendData(channelManager, msg, prefix) // Send data to sse
		}
	}()
	go func() {
		// Heartbeat
		ticker := time.Tick(time.Duration(10000 * time.Millisecond)) // Interval 10s send heartbeat
		for {
			<-ticker
			go utils.SendPing(channelManager, sseInstanceId) // Send heartbeat
		}
	}()
	go func() {
		// Status
		ticker := time.Tick(time.Duration(20000 * time.Millisecond)) // Interval 10s send status
		for {
			<-ticker
			go utils.SendStatus(channelManager, sseInstanceId, rdb, prefix) // Send status
		}
	}()

	router := gin.New() // Init GIN rounter
	// Metris
	metrics := ginmetrics.GetMonitor()
	metrics.SetMetricPath("/metrics")
	metrics.SetSlowTime(2)
	metrics.SetDuration([]float64{0.1, 0.3, 1.2, 5, 10})
	metrics.Use(router)
	// Custom Logger
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s |%s %d %s| %s |%s %s %s %s | %s | %s | %s\n",
			param.TimeStamp.Format(time.RFC1123),
			param.StatusCodeColor(),
			param.StatusCode,
			param.ResetColor(),
			param.ClientIP,
			param.MethodColor(),
			param.Method,
			param.ResetColor(),
			param.Path,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	// generality SSE
	router.GET("/:p1/:p2/:p3/:p4/:p5", sseStreaming) // Ping sse web
	router.OPTIONS("/:p1/:p2/:p3/:p4/:p5", options)  // sse web OPTIONS
	router.GET("/:p1/:p2/:p3/:p4", sseStreaming)     // Ping sse web
	router.OPTIONS("/:p1/:p2/:p3/:p4", options)      // sse web OPTIONS
	router.GET("/:p1/:p2/:p3", sseStreaming)         // Ping sse web
	router.OPTIONS("/:p1/:p2/:p3", options)          // sse web OPTIONS
	// generality Websocket
	router.GET("/ws/:p1/:p2/:p3/:p4/:p5", websocketStreaming) // Ping sse web
	router.GET("/ws//:p1/:p2/:p3/:p4", websocketStreaming)    // Ping sse web
	router.GET("/ws//:p1/:p2/:p3", websocketStreaming)        // Ping sse web
	// status
	router.GET("/status", statusSreaming) // status streaming
	log.Printf("listen port: %s", config.Port)
	router.Run(":" + config.Port)
}

func options(c *gin.Context) { // options
	c.Writer.WriteHeader(http.StatusNoContent)
}

func sseStreaming(c *gin.Context) { // sse streaming
	// validate token
	token, err := jwtToken.Validate(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(401, gin.H{
			"code":    401,
			"message": "Token invalid",
		})
		return
	}
	log.Printf("Token :%v", token)
	// start sse
	var arrParam []string
	if len(c.Param("p1")) > 0 {
		arrParam = append(arrParam, c.Param("p1"))
	}
	if len(c.Param("p2")) > 0 {
		arrParam = append(arrParam, c.Param("p2"))
	}
	if len(c.Param("p3")) > 0 {
		arrParam = append(arrParam, c.Param("p3"))
	}
	if len(c.Param("p4")) > 0 {
		arrParam = append(arrParam, c.Param("p4"))
	}
	if len(c.Param("p5")) > 0 {
		arrParam = append(arrParam, c.Param("p5"))
	}
	prefix, keys := utils.GetPrefixStreaming(arrParam)
	sseId := uuid.New() // ID of sse connection
	log.Printf("SSE | %s | %s - %s", sseId, prefix, keys)
	channelManager.SseTotal += 1
	channelManager.SseLive += 1
	// Create new listener
	listener := channelManager.OpenListener(prefix, keys)
	// Wait for close
	defer channelManager.CloseListener(prefix, keys, listener)
	clientGone := c.Request.Context().Done()
	// Keep connection
	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone: // Close connection
			log.Printf("DISCONNECT SSE | %s | %s/%s", sseId, prefix, keys)
			channelManager.SseClosed += 1
			channelManager.SseLive -= 1
			return false
		case message := <-listener: // Send message
			c.SSEvent("", message)
			return true
		}
	})
}

func websocketStreaming(c *gin.Context) {
	// validate token
	token, err := jwtToken.Validate(c.GetHeader("Authorization"))
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
	var arrParam []string
	if len(c.Param("p1")) > 0 {
		arrParam = append(arrParam, c.Param("p1"))
	}
	if len(c.Param("p2")) > 0 {
		arrParam = append(arrParam, c.Param("p2"))
	}
	if len(c.Param("p3")) > 0 {
		arrParam = append(arrParam, c.Param("p3"))
	}
	if len(c.Param("p4")) > 0 {
		arrParam = append(arrParam, c.Param("p4"))
	}
	if len(c.Param("p5")) > 0 {
		arrParam = append(arrParam, c.Param("p5"))
	}
	prefix, keys := utils.GetPrefixStreaming(arrParam)
	sseId := uuid.New() // ID of sse connection
	log.Printf("WEBSOCKET | %s | %s/%s", sseId, prefix, keys)
	channelManager.WsTotal += 1
	channelManager.WsLive += 1
	// Create new listener
	listener := channelManager.OpenListener(prefix, keys)
	// Wait for close
	defer channelManager.CloseListener(prefix, keys, listener)
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
	log.Printf("DISCONNECT WEBSOCKET | %s | %s/%s", sseId, prefix, keys)
	channelManager.WsClosed += 1
	channelManager.WsLive -= 1
}

func statusSreaming(c *gin.Context) {
	path := "streaming"
	channelId := "status"

	sseId := uuid.New()
	log.Printf("CONNECT SSE | %s | %s/%s", sseId, path, channelId)
	// Create new listener
	listener := channelManager.OpenListener(path, channelId)
	// Wait for close
	defer channelManager.CloseListener(path, channelId, listener)
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
