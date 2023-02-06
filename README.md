# GO STREAMING
Recieve data from Redis Pub/sub, then push data to client.

Support SSE or WEBSOCKET

Library:
- gin: https://github.com/gin-gonic/gin
- redis pub/sub: https://github.com/go-redis/redis/v8
- websocket: https://github.com/gorilla/websocket
- broadcast: https://github.com/dustin/go-broadcast


# STRUCTOR
```
├── Dockerfile
├── README.MD
├── build.sh
├── config
│   └── streaming-api.json
├── deployment
│   ├── config-map-go-streaming.yaml
│   └── deployment.yaml
├── go.mod
├── go.sum
├── main.go
├── model
│   ├── channels.go
│   └── config.go
└── utils
    └── helper.go
```

# MODEL
[PROCESSING] -> [REDIS PUB/SUB] -> [STREAMING] -> [CLIENT]

# INFO
- / : Get service info
- /admin : Get sse service info

# SSE STREAMING
- /streaming/:path/:channel : get sse streaming by path and channel
- /streaming/price/anz.asx,bhp.asx -> recieve all pub/sub to redis like price:anz.asx and price:bhp.asx
- /streaming/account/anz.asx,bhp.asx -> recieve all pub/sub to redis like account:anz.asx and account:bhp.asx

# WS STREAMING
- /ws-streaming/:path/:channel : get sse streaming by path and channel
- /ws-streaming/price/anz.asx,bhp.asx -> recieve all pub/sub to redis like price:anz.asx and price:bhp.asx
- /ws-streaming/account/anz.asx,bhp.asx -> recieve all pub/sub to redis like account:anz.asx and account:bhp.asx


# TESTING
Install package wscat
```
npm install -g wscat
```
Run on terminal
```
wscat -c wss://localhost:8080/ws-streaming/price/anz.asx,bhp.asx
```

```
curl http://localhost:8080/streaming/price/anz.asx,bhp.asx
```