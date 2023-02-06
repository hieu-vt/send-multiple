# GO STREAMING
- Recieve data from Redis Pub/sub, then push data to client.
- Should separate redis to improve performance
- Support SSE and WEBSOCKET

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

# CONFIG
- Config like: 
```
{
    "port": "8080",
    "redis": {
        "host": "localhost",
        "port": "6379",
        "db": "15",
        "password": "password"
    },
    "hook": ""
}
```

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
- Install package wscat
```
npm install -g wscat
```
- Run Websocket on terminal
```
wscat -c wss://localhost:8080/ws-streaming/price/anz.asx,bhp.asx
```

- Run Curl SSE on terminal 
```
curl http://localhost:8080/streaming/price/anz.asx,bhp.asx
```