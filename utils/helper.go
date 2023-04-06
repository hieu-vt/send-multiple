package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"go-streaming/model"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func GetPrefixStreaming(arrParams []string) (prefix string, key string) {
	var lastIndex = len(arrParams) - 1
	key = arrParams[lastIndex]
	arrayx := remove(arrParams, lastIndex)
	prefix = strings.Join(arrayx[:], ":")
	return prefix, key
}

func LoadConfiguration(file string) (model.Config, error) {
	var config model.Config
	configFile, err := os.Open(file)
	if err != nil {
		return config, err
	}
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	defer configFile.Close()
	return config, err
}

func LoadFile(file string) (string, error) {
	dat, err := os.ReadFile(file)
	return string(dat), err
}

func SendData(channelManager *model.ChannelManager, msg *redis.Message, prefix string) {
	channel := msg.Channel
	if len(prefix) != 0 {
		channel = strings.Replace(msg.Channel, fmt.Sprintf("%s:", prefix), "", 1)
	}
	channelManager.Submit(channel, msg.Payload)
}

func SendDataString(channelManager *model.ChannelManager, channel string, msg string) {
	channelManager.Submit(channel, msg)
}

func SendPing(channelManager *model.ChannelManager, sseInstanceId string) {
	date := time.Now()
	timePing := date.UnixMilli()
	dataPing := model.DataObj{
		Ping: timePing,
	}
	pingObj := model.PingObj{
		Data: dataPing,
		Type: "PING",
		Id:   "PING",
	}
	s, _ := json.Marshal(pingObj)
	channelManager.Submit("PING", string(s))
}

func SendStatus(channelManager *model.ChannelManager, sseInstanceId string, rdb *redis.Client, prefix string) {
	msg := gin.H{
		"Time":       time.Now().Format("2006-01-02 15:04:05"), // server time
		"Server-Id":  sseInstanceId,                            // server uuid
		"SSE-Total":  channelManager.SseTotal,                  // count SSE connections
		"SSE-Closed": channelManager.SseClosed,                 // count SSE closed connection
		"SSE-Live":   channelManager.SseLive,                   // count SSE online connections
		"Messages":   channelManager.TotalMessage,              // count message send to channel
		"WS-Total":   channelManager.WsTotal,                   // count Websocket connections
		"WS-Closed":  channelManager.WsClosed,                  // count Websocket closed connection
		"WS-Live":    channelManager.WsLive,                    // count Websocket online connections
	}
	content := fmt.Sprintf("%#v", msg)
	path := "streaming:status"
	if len(prefix) > 0 {
		path = fmt.Sprintf("%s:streaming:status", prefix)
	}
	rdb.Publish(context.Background(), path, content).Err()
}
