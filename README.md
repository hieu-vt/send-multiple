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
```
curl --location --request GET 'localhost:8083/v1/noti/notification/streaming-data/account?account=108723&organisation_code=TPBRSE' \
--header 'Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ2dW9uZy5uZ3V5ZW5Abm92dXMtZmludGVjaC5jb20iLCJzdWIiOiJlcTE2NTU3OTc3MTM1NDkiLCJleHAiOjE2ODA4MzU5OTguNDY0LCJkZXZpY2VfaWQiOiI0YzUyMDQyOS04ZTVmLTRlYjItODM4ZS00NmE1ODUyZjc3MjgiLCJpYXQiOjE2ODA4MzQ3OTh9.KeQ6AZM8KO_YTX523kDjUoAW6kF-LIblNOxXgnyaQBv_e4bnPORaKIKxkHUtFEdn3MxR_Y_rpgOBIMh1wefGkIspeTya0ADn3KYatSyLTeE7nGNil-enNv0ECdhgPr7gQPKdFwXsvBSxMHEHi3w5NfNschAr_yEIQyFZ9YBaONHfkSkOuRrC-JtmQvpGUs3Boolke9pPKMXvXlNC-oFkXamO22_pAlbjzKehyRpc3BNzAXEgDHwGcvove0oECx2WPHVhDUqq5XuQ9nvOSv6rpQClT2rP4KI2VZsgcDX0DvfpTl4P-CHp-PrgTEZBzhR0vJKSRy8dOU70NNdcJgdfV4Q-fS6os9OznGo4uarPq2Q6-3oJzzPADojDU_UVNJZbCnVgCHBCmH624TTuWOifE1HL4u53aqmfTHxBHW2El6jCMIjMIn5rLHBiBPIyx6-I8MT7v2eNoNsxSQ6FULPtVyDEzM-u5xsc6SM_J8TmYkvd6iAgZNcJroGsK8E8nTCNIefcsgelo8Da5KTISeKFwzbTHhPKJ_oGBqYC545BfQWTro2HOirt4aQRN-rnc9fAUts0pH2Io1WZTm1ndJ2am2Gs6sCx6kZvCXlcWRVm64qS4wk-Xde1uRBlAYRvbcxNfs9LzVomhGFzR0P_iRF06ECrtPHUkRpP14ZIlph14YQ'
```