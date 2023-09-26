package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/penglongli/gin-metrics/ginmetrics"
	"go-streaming/engine"
	"go-streaming/model"
	"go-streaming/rounter"
	"go-streaming/utils"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var jwtToken model.JWT
var sseInstanceId string = uuid.New().String() // uuid of see service
var channelManager *model.ChannelManager

// channel manager
type Quote struct {
	OpenPrice  float64
	ClosePrice float64
	LastPrice  float64
	QuoteVol   int
}

type Message struct {
	Symbol string
	Data   Quote
}

// "symbol": "3A.XKLS",
// "instrument_code": "353327422",
// "display_name": "3A",
// "company_name": "THREE-A RESOURCES BHD",
// "currency": "MYR",
// "exchange": "XKLS",
// "market": "MYS",
// "status": "Active",
// "market_status": "out_of_session",
// "country": "MYS"
type Data struct {
	Symbol string
}

type Response struct {
	Data []Data
}

type PriceResponse struct {
	Symbol        string  `json:"symbol"`
	TradePrice    float64 `json:"trade_price"`
	Volume        int     `json:"volume"`
	ChangePoint   int     `json:"change_point"`
	ChangePercent int     `json:"change_percent"`
	Bid           float64 `json:"bid"`
	BidQty        int     `json:"bid_qty"`
	Ask           float64 `json:"ask"`
	AskQty        int     `json:"ask_qty"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	PreClose      float64 `json:"pre_close"`
	Close         float64 `json:"close"`
	MarketStatus  string  `json:"market_status"`
}

func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second / 20

	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		log.Printf("Disconnected due to:%s, will attempt reconnects for %.0fm", err, totalWait.Minutes())
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Printf("Exiting: %v", nc.LastError())
	}))
	return opts
}

func RandomSymbol() string {
	rand.Seed(time.Now().UnixNano())

	// Define the alphabet
	alphabet := "ABCDEFG"

	// Shuffle the alphabet
	runes := []rune(alphabet)
	rand.Shuffle(len(runes), func(i, j int) {
		runes[i], runes[j] = runes[j], runes[i]
	})

	// Create a string with the first three characters
	str := string(runes[:3])
	return "ASX:" + str
}

func GetSymbolDetail(url string) []string {
	var listSymbol []string
	//var responseObject Response
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Print(err.Error())
	}
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIrODQ5NjYxODcyMzIiLCJzdWIiOiIrODQ5NjYxODcyMzIiLCJkZXZpY2VfaWQiOiJhYmMiLCJ1c2VybmFtZSI6InVzZXIyaHNtIiwicGFydHlfaWQiOiIxMDAwMDIiLCJsb2dpbl90eXBlIjoiU1NPIiwidXNlcl9sb2dpbl90eXBlIjoidXNlcl9tb2JpbGUiLCJleHAiOjIwMDc5NTMzMDguOTUzLCJpYXQiOjE2OTI1OTMzMDh9.YVXuTzLYZpV6c2-9A7KuPDxb-Asjckr2blpSRcjKWdgt4Goq2-jF7blL-sPEMkcV4kpa_EgRXW-g_2zgphyzQgfuNTRcvvS89Y5P9Fv-62bmK--Nv24Fsa9WoAaapw9GEmrMl03EWoCzJLII7zGXS7sij8WudAs59rz8BN-Hs8FK_XZLXiKqvcpF6CeHt4MF_YXbG7tYm8dnAc45rFJZP6aGId5Zz0q0u0hpI0r7MP7Fq6irncaj6hfqhXSgR1qymFVjiwtZ2mPmpS50q64PL-jTUeDKeyc5XpzmCK-wYMJ_eQIT353gtmtNsMOYN2iNKqu7mIDsxi871Nn46mISSUs9fIdHkQei_3z8j5yMoLqBHtfA4kC35JYJTlwbInaKbSC3tQo9fnmVc9uf4gVSxil_ntkuvEz6AX4f4Gc1jRFwlfpckH0PqBb1vzhsAUU0QQRFxmZdijCie6hyDkd_1ZKMv4oDlV77pYWDknFeU2voi9TErWnEyU0pyLKYSshd9Nn-_xhqne-Un_njQ79aHyHe6IMIbHyZOGX8SUJRBvkRQ4Kg4T3GRQPFFu447ykh00_Ohe1EVP32Nl0i4MrPLaJiRqdd5R0ASfWhPnS752Xv3Y20pMElZ5JwjHju7AQN2jkr_hN3KcMHo1ToXouhotGbYEadp_yubh6jfujdLkc")
	//json.Unmarshal(responseData, &responseObject)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}

	defer res.Body.Close()
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		fmt.Print(err.Error())
	}

	var symbols Response

	if err := json.Unmarshal([]byte(body), &symbols); err != nil {
		fmt.Println("Error unmarshaling Json: ", err)
	}

	for _, v := range symbols.Data {
		listSymbol = append(listSymbol, v.Symbol)
	}

	return listSymbol
}

func randomPriceResponse(symbol string) PriceResponse {
	// Seed the random number generator with the current time
	rand.Seed(time.Now().UnixNano())

	return PriceResponse{
		Symbol:        symbol,
		TradePrice:    rand.Float64() * 100.0, // Example range, modify as needed
		Volume:        rand.Intn(1000000),     // Example range, modify as needed
		ChangePoint:   rand.Intn(100),         // Example range, modify as needed
		ChangePercent: rand.Intn(100),         // Example range, modify as needed
		Bid:           rand.Float64() * 100.0, // Example range, modify as needed
		BidQty:        rand.Intn(10000),       // Example range, modify as needed
		Ask:           rand.Float64() * 100.0, // Example range, modify as needed
		AskQty:        rand.Intn(10000),       // Example range, modify as needed
		Open:          rand.Float64() * 100.0, // Example range, modify as needed
		High:          rand.Float64() * 100.0, // Example range, modify as needed
		Low:           rand.Float64() * 100.0, // Example range, modify as needed
		PreClose:      rand.Float64() * 100.0, // Example range, modify as needed
		Close:         rand.Float64() * 100.0, // Example range, modify as needed
		MarketStatus:  "main_session",         // Example, modify as needed
	}
}

func sendDataWith(listSymbol []string, engineCn *engine.ChannelEngine) {
	go func() {
		for {
			time.Sleep(time.Second / 20)
			for _, v := range listSymbol {
				priceOfSymbol := randomPriceResponse(v)

				body, _ := json.Marshal(priceOfSymbol)

				engineCn.Send(priceOfSymbol.Symbol, engine.Message{Data: string(body)})
			}
		}
	}()
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
	// Get message from nats pub/sub
	log.Printf("Start subscrible %s\n", nats.DefaultURL)
	url := "https://cgsi-api.equix.app/v1/live/symbol?search=A"

	listSymbol := GetSymbolDetail(url)

	sendDataWith(listSymbol, channelEngine)
	//nc, _ := nats.Connect(nats.DefaultURL, setupConnOptions([]nats.Option{})...)

	//go func() {
	//	for {
	//		time.Sleep(time.Second / 20)
	//		data := RandomQuote()
	//		log.Println("Message: ", data)
	//		result, _ := json.Marshal(data)
	//		channelEngine.Send(data.Symbol, engine.Message{Data: string(result)})
	//	}
	//}()

	//// Heartbeat
	//go func() {
	//	ticker := time.Tick(time.Duration(10000 * time.Second/10/2/100/100)) // Interval 10s send heartbeat
	//	for {
	//		<-ticker
	//		utils.SendPing(channelManager, sseInstanceId) // Send heartbeat
	//	}
	//}()
	//// Status
	//go func() {
	//	ticker := time.Tick(time.Duration(20000 * time.Second/10/2/100/100)) // Interval 10s send status
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
