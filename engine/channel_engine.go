package engine

import (
	"log"
	"sync"
)

type Message struct {
	Data string `json:"data"`
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
		ce.manager[channelId] = append(eCh, ch)
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
			go func(mess Message, index int) {
				chs[index] <- message
			}(message, i)
		}
	}
}

// delete chan when disconnect sse
func (ce *ChannelEngine) DeleteChildChannel(ch chan Message, channelId string) {
	ce.locker.Lock()

	chs, ok := ce.manager[channelId]
	if ok {
		for i := 0; i < len(chs); i++ {
			if chs[i] == ch {
				chs = append(chs[:i], chs[i+1:]...)
				close(ch)
				break
			}
		}
	}

	log.Println("Destroy channel: ", channelId)
	ce.locker.Unlock()
}

func (ce *ChannelEngine) Run() {

}
