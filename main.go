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
)

var hook string                                // hook url
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
	gin.SetMode(gin.ReleaseMode)

	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Printf("endpoint %v %v %v %v\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}
	config, _ := utils.LoadConfiguration("config/streaming-api.json")
	fmt.Println(config.Redis.Host)
	hook = config.Hook                         // get hook config
	channelManager = model.NewChannelManager() // Init channel manager
	// Config redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Host + ":" + config.Redis.Port, // Address or redis
		Password: config.Redis.Password,                       // Password
		DB:       config.Redis.Database,                       // Database
	})
	pubsub := rdb.PSubscribe(context.Background(), "*")
	go func() {
		// Get message from redis pub/sub
		for msg := range pubsub.Channel() {
			go utils.SendData(channelManager, msg) // send data to sse
		}
	}()
	go func() {
		// Heartbeat
		ticker := time.Tick(time.Duration(10000 * time.Millisecond))
		i := 0
		for {
			<-ticker
			i++
			go utils.SendPing(channelManager, sseInstanceId, rdb) // Send heartbeat
		}
	}()
	router := gin.Default()
	router.GET("/streaming/:path/:channel", stream)      // sse streaming
	router.GET("/ws-streaming/:path/:channel", wsStream) // websocket streaming
	router.GET("/admin", sseAdmin)                       // admin streaming
	router.GET("/", home)                                // home
	router.Run(":" + config.Port)
}

func stream(c *gin.Context) {
	channelId := c.Param("channel")
	path := c.Param("path")
	go utils.HookUrl(hook, c)
	sseId := uuid.New()
	log.Println("CONNECT SSE |", sseId, "|", path+"/"+channelId)
	channelManager.SseTotal += 1
	channelManager.SseLive += 1
	// Create new listener
	listener := channelManager.OpenListener(path, channelId)
	// Wait for close
	defer channelManager.CloseListener(path, channelId, listener)
	clientGone := c.Request.Context().Done()
	// keep connection
	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			log.Println("DISCONECT SSE |", sseId, "|", path+"/"+channelId)
			channelManager.SseClosed += 1
			channelManager.SseLive -= 1
			go utils.HookUrl(hook, c)
			return false
		case message := <-listener:
			c.SSEvent("", message)
			return true
		}
	})
}

func wsStream(c *gin.Context) {
	w, r := c.Writer, c.Request
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer ws.Close()
	channelId := c.Param("channel")
	path := c.Param("path")
	sseId := uuid.New()
	log.Println("CONNECT WS |", sseId, "|", path+"/"+channelId)
	channelManager.WsTotal += 1
	channelManager.WsLive += 1
	// Create new listener
	listener := channelManager.OpenListener(path, channelId)
	// Wait for close
	defer channelManager.CloseListener(path, channelId, listener)
	// keep connection
	for {
		message := <-listener
		wsRes := fmt.Sprintf("%b", message)
		err = ws.WriteMessage(1, []byte(wsRes))
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
	log.Println("DISCONECT WS |", sseId, "|", path+"/"+channelId)
	channelManager.WsClosed += 1
	channelManager.WsLive -= 1
}

func sseAdmin(c *gin.Context) {
	channelId := "admin"
	path := "sse"
	sseId := uuid.New()
	log.Println("CONNECT SSE |", sseId, "|", path+"/"+channelId)
	// Create new listener
	listener := channelManager.OpenListener(path, channelId)
	// Wait for close
	defer channelManager.CloseListener(path, channelId, listener)
	clientGone := c.Request.Context().Done()
	// keep connection
	c.Stream(func(w io.Writer) bool {
		select {
		case <-clientGone:
			log.Println("DISCONECT SSE |", sseId, "|", path+"/"+channelId)
			return false
		case message := <-listener:
			c.SSEvent("", message)
			return true
		}
	})
}

func home(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Server-Id":  sseInstanceId,               // server uuid
		"SSE-Total":  channelManager.SseTotal,     // count SSE connections
		"SSE-Closed": channelManager.SseClosed,    // count SSE closed connection
		"SSE-Live":   channelManager.SseLive,      // count SSE online connections
		"Messages":   channelManager.TotalMessage, // count message send to channel
		"WS-Total":   channelManager.WsTotal,      // count Websocket connections
		"WS-Closed":  channelManager.WsClosed,     // count Websocket closed connection
		"WS-Live":    channelManager.WsLive,       // count Websocket online connections
	})
}
