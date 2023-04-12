package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"go-streaming/model"
	"log"
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

func getPrefixStreaming(arrParams []string) (prefix string, key string) {
	var lastIndex = len(arrParams) - 1
	key = arrParams[lastIndex]
	arrayx := remove(arrParams, lastIndex)
	prefix = strings.Join(arrayx[:], ":")
	return prefix, key
}

func GetPrefixStreamingByGinContext(c *gin.Context) (prefix string, key string) {
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
	if len(c.Param("p6")) > 0 {
		arrParam = append(arrParam, c.Param("p6"))
	}
	var arrCode []string
	if len(c.Query("organisation_code")) > 0 {
		orgs := strings.Split(c.Query("organisation_code"), ",")
		for i := 0; i < len(orgs); i++ {
			arrCode = append(arrCode, fmt.Sprintf("org_%s", orgs[i]))
		}
	}
	if len(c.Query("branch_code")) > 0 {
		branchs := strings.Split(c.Query("branch_code"), ",")
		for i := 0; i < len(branchs); i++ {
			arrCode = append(arrCode, fmt.Sprintf("branch_%s", branchs[i]))
		}
	}
	if len(c.Query("advisor_code")) > 0 {
		advisors := strings.Split(c.Query("advisor_code"), ",")
		for i := 0; i < len(advisors); i++ {
			arrCode = append(arrCode, fmt.Sprintf("advisor_%s", advisors[i]))
		}
	}
	if len(c.Query("account")) > 0 {
		accounts := strings.Split(c.Query("account"), ",")
		arrCode = append(arrCode, accounts...)
	}
	if len(c.Query("account_id")) > 0 {
		accounts := strings.Split(c.Query("account_id"), ",")
		arrCode = append(arrCode, accounts...)
	}
	if len(arrCode) <= 0 {
		return getPrefixStreaming(arrParam)
	}
	prefix = strings.Join(arrParam[:], ":")
	key = strings.Join(arrCode[:], ",")
	return prefix, key
}

func LoadConfiguration(fileConfig string) (model.Config, error) {
	cloudPath := fmt.Sprintf("/etc/equix/%s/%s", fileConfig, fileConfig)
	localPath := fmt.Sprintf("config/%s.json", fileConfig)
	file := cloudPath
	_, error := os.Stat(cloudPath)
	if error != nil {
		file = localPath
	}
	var config model.Config
	configFile, err := os.Open(file)
	if err != nil {
		return config, err
	}
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	defer configFile.Close()
	log.Printf("config %s, %v", file, config)
	return config, err
}

func LoadFile(fileConfig string) (string, error) {
	cloudPath := fmt.Sprintf("/etc/equix/%s/%s", fileConfig, fileConfig)
	localPath := fmt.Sprintf("config/%s", fileConfig)
	file := cloudPath
	_, error := os.Stat(cloudPath)
	if error != nil {
		file = localPath
	}
	dat, err := os.ReadFile(file)
	content := string(dat)
	log.Printf("config %s, %v", file, content)
	return content, err
}

func SendData(channelManager *model.ChannelManager, msg *redis.Message, prefix string) {
	channel := msg.Channel
	if len(prefix) != 0 {
		channel = strings.Replace(msg.Channel, fmt.Sprintf("%s:", prefix), "", 1)
	}
	channelManager.Submit(channel, fmt.Sprintf(" %s", msg.Payload))
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
	channelManager.Submit("PING", fmt.Sprintf(" %s", string(s)))
}

func SendStatus(channelManager *model.ChannelManager, sseInstanceId string, rdb *redis.Client, prefix string) {
	content := fmt.Sprintf("%v", channelManager.Channels)
	log.Printf("Channels: %s", content)
	path := "streaming:status"
	if len(prefix) > 0 {
		path = fmt.Sprintf("%v", channelManager)
	}
	rdb.Publish(context.Background(), path, content).Err()
}
