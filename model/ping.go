package model

type DataObj struct {
	Ping int64 `json:"ping"`
}

type PingObj struct {
	Data DataObj `json:"data"`
	Type string  `json:"type"`
	Id   string  `json:"id"`
}
