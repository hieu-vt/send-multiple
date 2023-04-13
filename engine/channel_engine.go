package engine

import (
	"log"
	"sync"
)

type Message struct {
	Text      string
	ChannelId string
}

type ChannelEngine struct {
	manager map[string][]chan Message
	locker  *sync.RWMutex
}

func NewChannelEngine() *ChannelEngine {
	return &ChannelEngine{
		manager: make(map[string][]chan Message, 100),
		locker:  new(sync.RWMutex),
	}
}

// sse call this method
func (ce *ChannelEngine) Listener(channelId string) chan Message {
	ce.locker.Lock()
	ch := make(chan Message, 100)
	eCh, ok := ce.manager[channelId]

	if ok {
		_ = append(eCh, ch)
	} else {
		ce.manager[channelId] = []chan Message{
			ch,
		}
	}

	ce.locker.Unlock()

	return ch
}

// redis call this method
func (ce *ChannelEngine) Send(channelId string, message Message) {
	log.Println("Send mess: ", message)
	chs, ok := ce.manager[channelId]
	if ok {
		for i := 0; i < len(chs); i++ {
			chs[i] <- message
		}
	}
}

// delete chan when disconnect sse
func (ce *ChannelEngine) DeleteChildChannel(ch <-chan Message, channelId string) {
	chs, ok := ce.manager[channelId]
	if ok {
		for i := 0; i < len(chs); i++ {
			if chs[i] == ch {
				chs = append(chs[:i], chs[i+1:]...)
				break
			}
		}
	}
}
