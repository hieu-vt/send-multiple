package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/penglongli/gin-metrics/ginmetrics"
	"go-streaming/engine"
	"go-streaming/model"
	"go-streaming/rounter"
	"go-streaming/utils"
	"log"
	"math/rand"
	"time"
)

var jwtToken model.JWT
var sseInstanceId string = uuid.New().String() // uuid of see service
var channelManager *model.ChannelManager       // channel manager

func RandomSymbol() string {
	rand.Seed(time.Now().UnixNano())

	// Define the alphabet
	alphabet := "ABC"

	// Shuffle the alphabet
	runes := []rune(alphabet)
	rand.Shuffle(len(runes), func(i, j int) {
		runes[i], runes[j] = runes[j], runes[i]
	})

	// Create a string with the first three characters
	str := string(runes[:3])
	return "ASX:" + str
}

func main() {
	log.SetFlags(log.Ldate | log.Lmsgprefix | log.Ltime | log.Lshortfile)
	gin.SetMode(gin.ReleaseMode)                                    // Set release mode
	config, _ := utils.LoadConfiguration("config-go-streaming-api") // Get config streaming
	prvKey, _ := utils.LoadFile("config-jwtrs256-key")
	pubKey, _ := utils.LoadFile("config-jwtrs256-key-pem")
	jwtToken = model.NewJWT([]byte(prvKey), []byte(pubKey))

	log.Printf("Redis enpoint: %s", config.Redis.Host)
	channelManager = model.NewChannelManager() // Init channel manager
	channelEngine := engine.NewChannelEngine()
	// Init redis connection
	//rdb := redis.NewClient(&redis.Options{
	//	Addr:     config.Redis.Host + ":" + config.Redis.Port, // Address or redis. Should separate redis to improve performance
	//	Password: config.Redis.Password,                       // Password
	//	DB:       config.Redis.Database,                       // Database
	//})
	// Get message from redis pub/sub
	go func() {
		time.Sleep(time.Second * 5)

		for {
			time.Sleep(time.Millisecond)
			cId := RandomSymbol()
			log.Println("ChannelId redis:", cId)
			channelEngine.Send(cId, engine.Message{Text: "Hello subscrible", ChannelId: cId})
		}
	}()

	go func() {
		time.Sleep(time.Second * 5)

		for {
			time.Sleep(time.Millisecond)
			cId := RandomSymbol()
			log.Println("ChannelId redis:", cId)
			channelEngine.Send(cId, engine.Message{Text: "Hello subscrible", ChannelId: cId})
		}
	}()

	go func() {
		time.Sleep(time.Second * 5)

		for {
			time.Sleep(time.Millisecond)
			cId := RandomSymbol()
			log.Println("ChannelId redis:", cId)
			channelEngine.Send(cId, engine.Message{Text: "Hello subscrible", ChannelId: cId})
		}
	}()
	go func() {
		time.Sleep(time.Second * 5)

		for {
			time.Sleep(time.Millisecond)
			cId := RandomSymbol()
			log.Println("ChannelId redis:", cId)
			channelEngine.Send(cId, engine.Message{Text: "Hello subscrible", ChannelId: cId})
		}
	}()
	go func() {
		time.Sleep(time.Second * 5)

		for {
			time.Sleep(time.Millisecond)
			cId := RandomSymbol()
			log.Println("ChannelId redis:", cId)
			channelEngine.Send(cId, engine.Message{Text: "Hello subscrible", ChannelId: cId})
		}
	}()
	go func() {
		time.Sleep(time.Second * 5)

		for {
			time.Sleep(time.Millisecond)
			cId := RandomSymbol()
			log.Println("ChannelId redis:", cId)
			channelEngine.Send(cId, engine.Message{Text: "Hello subscrible", ChannelId: cId})
		}
	}()
	go func() {
		time.Sleep(time.Second * 5)

		for {
			time.Sleep(time.Millisecond)
			cId := RandomSymbol()
			log.Println("ChannelId redis:", cId)
			channelEngine.Send(cId, engine.Message{Text: "Hello subscrible", ChannelId: cId})
		}
	}()
	go func() {
		time.Sleep(time.Second * 5)

		for {
			time.Sleep(time.Millisecond)
			cId := RandomSymbol()
			log.Println("ChannelId redis:", cId)
			channelEngine.Send(cId, engine.Message{Text: "Hello subscrible", ChannelId: cId})
		}
	}()
	go func() {
		time.Sleep(time.Second * 5)

		for {
			time.Sleep(time.Millisecond)
			cId := RandomSymbol()
			log.Println("ChannelId redis:", cId)
			channelEngine.Send(cId, engine.Message{Text: "Hello subscrible", ChannelId: cId})
		}
	}()
	go func() {
		time.Sleep(time.Second * 5)
		for {
			time.Sleep(time.Millisecond)
			cId := RandomSymbol()
			log.Println("ChannelId redis:", cId)
			channelEngine.Send(cId, engine.Message{Text: "Hello subscrible", ChannelId: cId})
		}
	}()
	go func() {
		time.Sleep(time.Second * 5)
		for {
			time.Sleep(time.Millisecond)
			cId := RandomSymbol()
			log.Println("ChannelId redis:", cId)
			channelEngine.Send(cId, engine.Message{Text: "Hello subscrible", ChannelId: cId})
		}
	}()
	go func() {
		time.Sleep(time.Second * 5)
		for {
			time.Sleep(time.Millisecond)
			cId := RandomSymbol()
			log.Println("ChannelId redis:", cId)
			channelEngine.Send(cId, engine.Message{Text: "Hello subscrible", ChannelId: cId})
		}
	}()
	//// Heartbeat
	//go func() {
	//	ticker := time.Tick(time.Duration(10000 * time.Millisecond/100)) // Interval 10s send heartbeat
	//	for {
	//		<-ticker
	//		utils.SendPing(channelManager, sseInstanceId) // Send heartbeat
	//	}
	//}()
	//// Status
	//go func() {
	//	ticker := time.Tick(time.Duration(20000 * time.Millisecond/100)) // Interval 10s send status
	//	for {
	//		<-ticker
	//		utils.SendStatus(channelManager.Count, sseInstanceId, rdb, prefix) // Send status
	//	}
	//}()
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
	//router.GET("/:p1/:p2/:p3/:p4/:p5/:p6", rounter.SseHandler(config.CheckJwt, jwtToken, channelManager)) // Sse
	//router.GET("/:p1/:p2/:p3/:p4/:p5", rounter.SseHandler(config.CheckJwt, jwtToken, channelManager))     // Sse
	//router.GET("/:p1/:p2/:p3/:p4", rounter.SseHandler(config.CheckJwt, jwtToken, channelManager))         // Sse
	//router.GET("/:p1/:p2/:p3", rounter.SseHandler(config.CheckJwt, jwtToken, channelManager))             // Sse
	router.GET("/:p1/:p2", rounter.SseHandler(channelEngine)) // Sse
	// Generality Options
	router.OPTIONS("/:p1/:p2/:p3/:p4/:p5/p6", rounter.Options) // Options
	router.OPTIONS("/:p1/:p2/:p3/:p4/:p5", rounter.Options)    // Options
	router.OPTIONS("/:p1/:p2/:p3/:p4", rounter.Options)        // Options
	router.OPTIONS("/:p1/:p2/:p3", rounter.Options)            // Options
	router.OPTIONS("/:p1/:p2", rounter.Options)                // Options
	// Generality Websocket
	router.GET("/ws/:p1/:p2/:p3/:p4/:p5/:p6", rounter.WebsocketHandler(config.CheckJwt, jwtToken, channelManager)) // Websocket
	router.GET("/ws/:p1/:p2/:p3/:p4/:p5", rounter.WebsocketHandler(config.CheckJwt, jwtToken, channelManager))     // Websocket
	router.GET("/ws//:p1/:p2/:p3/:p4", rounter.WebsocketHandler(config.CheckJwt, jwtToken, channelManager))        // Websocket
	router.GET("/ws//:p1/:p2/:p3", rounter.WebsocketHandler(config.CheckJwt, jwtToken, channelManager))            // Websocket
	router.GET("/ws/:p1/:p2", rounter.WebsocketHandler(config.CheckJwt, jwtToken, channelManager))                 // Sse
	// Start service
	router.GET("/status", rounter.StatusHandler(channelManager)) // Sse
	log.Printf("listen port: %s", config.Port)
	router.Run(":" + config.Port)
}
