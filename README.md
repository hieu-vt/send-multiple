# GO STREAMING
- Recieve data from Redis Pub/sub, then push data to client.
- Should separate redis to improve performance
- Support SSE and WEBSOCKET

Library:
- gin: https://github.com/gin-gonic/gin
- redis pub/sub: https://github.com/go-redis/redis/v8
- websocket: https://github.com/gorilla/websocket
- broadcast: https://github.com/dustin/go-broadcast
- metrics: github.com/penglongli/gin-metrics/ginmetrics

# STRUCTOR
```
├── Dockerfile
├── README.MD
├── build.sh
├── config
│   └── streaming-api.json
│   └── config-jwtrs256-key
│   └── config-jwtrs256-key-pem
├── deployment
│   ├── config-map-go-streaming.yaml
│   └── deployment.yaml
├── go.mod
├── go.sum
├── main.go
├── model
│   ├── channels.go
│   └── config.go
│   └── ping.go
│   └── token.go
├── rounter
│   ├── option.go
│   └── sse.go
│   └── status.go
│   └── websocket.go
└── utils
    └── helper.go
```

# MODEL
[PROCESSING] -> [REDIS PUB/SUB] -> [STREAMING] -> [CLIENT]

# CONFIG
- port: port streaming service listener
- redis: config redis pubsub
    - host: Host redis
    - port: Port redis
    - db: Database
    - password: Password
- prefix_channel: prefix channel to subscribe redis
    Example: subscribe v1:streaming:price:ANZ.ASX -> /streaming/price/ANZ.ASX

```
{
    "port": "8080",
    "redis": {
        "host": "localhost",
        "port": "6379",
        "db": "15",
        "password": "password"
    }
}
```

# INFO
- /metrics : Get metrics prometheus
- /status : Get status streaming of all instances sse running

# SSE STREAMING
- Multil level subscriber from reids pub/sub then send to client
- /:p1/:p2/:p3/:p4/:p5
- Description:
    - /:p1/:p2/:p3/:p4 -> prefix channel in redis (Example: /v1/annoucement/news/asx)
    - /:p5 -> channel ID, separate by comma (Example: ANZ.ASX,BHP.ASX)
    - sub redis channel1: v1:annoucement:news:asx:ANZ.ASX
    - sub redis channel2: v1:annoucement:news:asx:BHP.ASX
    - when have changed in channel1 or channel2, data will send to client

# WS STREAMING
- like sse streaming but using /ws befor sse path
- /ws/:p1/:p2/:p3/:p4/:p5

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