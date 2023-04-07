package main

import (
	"context"
	"fmt"
	"go-streaming/model"
	"go-streaming/rounter"
	"go-streaming/utils"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/penglongli/gin-metrics/ginmetrics"
)

var jwtToken model.JWT
var sseInstanceId string = uuid.New().String() // uuid of see service
var channelManager *model.ChannelManager       // channel manager

func main() {
	log.SetFlags(log.Ldate | log.Lmsgprefix | log.Ltime | log.Lshortfile)
	gin.SetMode(gin.ReleaseMode)                                    // Set release mode
	config, _ := utils.LoadConfiguration("config-go-streaming-api") // Get config streaming
	prvKey, _ := utils.LoadFile("config-jwtrs256-key")
	pubKey, _ := utils.LoadFile("config-jwtrs256-key-pem")
	jwtToken = model.NewJWT([]byte(prvKey), []byte(pubKey))

	log.Printf("Redis enpoint: %s", config.Redis.Host)
	channelManager = model.NewChannelManager() // Init channel manager
	prefix := config.PrefixChannel
	// Init redis connection
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Host + ":" + config.Redis.Port, // Address or redis. Should separate redis to improve performance
		Password: config.Redis.Password,                       // Password
		DB:       config.Redis.Database,                       // Database
	})
	// Get message from redis pub/sub
	go func() {
		channel := "*"
		if len(prefix) > 0 {
			channel = fmt.Sprintf("%s:*", prefix)
		}
		subscribeRedis := rdb.PSubscribe(context.Background(), channel) // Subscribe all channel in redis pubsub
		for msg := range subscribeRedis.Channel() {                     // Listen redis pubsub
			go utils.SendData(channelManager, msg, prefix) // Send data to sse
		}
	}()
	// Heartbeat
	go func() {
		ticker := time.Tick(time.Duration(10000 * time.Millisecond)) // Interval 10s send heartbeat
		for {
			<-ticker
			go utils.SendPing(channelManager, sseInstanceId) // Send heartbeat
		}
	}()
	// Status
	go func() {
		ticker := time.Tick(time.Duration(20000 * time.Millisecond)) // Interval 10s send status
		for {
			<-ticker
			go utils.SendStatus(channelManager, sseInstanceId, rdb, prefix) // Send status
		}
	}()
	// Init GIN rounter
	router := gin.New()
	// Metris
	metrics := ginmetrics.GetMonitor()
	metrics.SetMetricPath("/metrics")
	metrics.SetSlowTime(2)
	metrics.SetDuration([]float64{0.1, 0.3, 1.2, 5, 10})
	metrics.Use(router)
	// Logging gin
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		log.Printf("| %d | %s | %s | %s | %s | %s | %s\n",
			param.StatusCode,
			param.ClientIP,
			param.Method,
			param.Path,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
		return ""
	}))
	// Generality SSE
	router.GET("/:p1/:p2/:p3/:p4/:p5", rounter.SseHandler(config.CheckJwt, jwtToken, channelManager)) // Sse
	router.GET("/:p1/:p2/:p3/:p4", rounter.SseHandler(config.CheckJwt, jwtToken, channelManager))     // Sse
	router.GET("/:p1/:p2/:p3", rounter.SseHandler(config.CheckJwt, jwtToken, channelManager))         // Sse
	// Generality Options
	router.OPTIONS("/:p1/:p2/:p3/:p4/:p5", rounter.Options) // Options
	router.OPTIONS("/:p1/:p2/:p3/:p4", rounter.Options)     // Options
	router.OPTIONS("/:p1/:p2/:p3", rounter.Options)         // Options
	// Generality Websocket
	router.GET("/ws/:p1/:p2/:p3/:p4/:p5", rounter.WebsocketHandler(config.CheckJwt, jwtToken, channelManager)) // Websocket
	router.GET("/ws//:p1/:p2/:p3/:p4", rounter.WebsocketHandler(config.CheckJwt, jwtToken, channelManager))    // Websocket
	router.GET("/ws//:p1/:p2/:p3", rounter.WebsocketHandler(config.CheckJwt, jwtToken, channelManager))        // Websocket
	// Status
	router.GET("/status", rounter.StatusHandler(channelManager)) // Status
	// Test
	router.GET("/test-all", rounter.AllHandler(config.CheckJwt, jwtToken, channelManager)) // Status
	// Start service
	log.Printf("listen port: %s", config.Port)
	router.Run(":" + config.Port)
}
