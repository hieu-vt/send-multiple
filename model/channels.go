package model

import (
	"go-streaming/broadcast"
	"log"
	"strings"
)

type Message struct {
	ChannelId string
	Text      string
}

type Listener struct {
	ChannelId string
	Chan      chan interface{}
}

type ChannelManager struct {
	channels map[string]broadcast.Broadcaster
	open     chan *Listener
	close    chan *Listener
	delete   chan string
	messages chan *Message
	Count    int
}

func NewChannelManager() *ChannelManager {
	manager := &ChannelManager{
		channels: make(map[string]broadcast.Broadcaster),
		open:     make(chan *Listener, 100),
		close:    make(chan *Listener, 100),
		delete:   make(chan string, 100),
		messages: make(chan *Message, 100),
		Count:    0,
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recover panic, recreate Channel Manager", r)
				manager = NewChannelManager()
			}
		}()
		manager.run()
	}()
	return manager
}

func (m *ChannelManager) run() {
	for {
		select {
		case listener := <-m.open:
			m.register(listener)
		case listener := <-m.close:
			m.deregister(listener)
		case channelId := <-m.delete:
			m.deleteBroadcast(channelId)
		case message := <-m.messages:
			if m.isExistsChannel(message.ChannelId) {
				m.getChannel(message.ChannelId).Submit(message.Text)
			}
		}
	}
}

func (m *ChannelManager) register(listener *Listener) {
	m.getChannel(listener.ChannelId).Register(listener.Chan)
}

func (m *ChannelManager) deregister(listener *Listener) {
	m.getChannel(listener.ChannelId).Unregister(listener.Chan)
	if !IsClosed(listener.Chan) {
		close(listener.Chan)
	}
}

func (m *ChannelManager) deleteBroadcast(channelId string) {
	b, ok := m.channels[channelId]
	if ok {
		b.Close()
		delete(m.channels, channelId)
	}
}

func (m *ChannelManager) isExistsChannel(channelId string) bool {
	_, ok := m.channels[channelId]
	return ok
}

func (m *ChannelManager) getChannel(channelId string) broadcast.Broadcaster {
	b, ok := m.channels[channelId]
	if !ok {
		b = broadcast.NewBroadcaster(10)
		m.channels[channelId] = b
	}
	return b
}

func (m *ChannelManager) OpenListener(prefix string, keys string) chan interface{} {
	m.Count += 1
	// Each channel separate by ,
	s := strings.Split(keys, ",")
	// Add one listener for all channels
	listener := make(chan interface{})
	for i := 0; i < len(s); i++ {
		m.open <- &Listener{
			ChannelId: prefix + ":" + s[i],
			Chan:      listener,
		}
	}
	if prefix == "streaming" && keys == "status" {
		log.Printf("Bypass PING msg to status link")
	} else {
		// Add Ping channel to every listener
		m.open <- &Listener{
			ChannelId: "PING",
			Chan:      listener,
		}
	}
	return listener
}

func (m *ChannelManager) CloseListener(prefix string, keys string, channel chan interface{}) {
	m.Count -= 1
	// Each channel separate by ,
	s := strings.Split(keys, ",")
	for i := 0; i < len(s); i++ {
		m.close <- &Listener{
			ChannelId: prefix + ":" + s[i],
			Chan:      channel,
		}
	}
	// Close Ping channel
	m.close <- &Listener{
		ChannelId: "PING",
		Chan:      channel,
	}
}

func (m *ChannelManager) DeleteBroadcast(prefix string, keys string) {
	// Each channel separate by ,
	s := strings.Split(keys, ",")
	for i := 0; i < len(s); i++ {
		m.delete <- prefix + ":" + s[i]
	}
}

func (m *ChannelManager) Submit(channelId string, text string) {
	var arr []string
	arr = append(arr, channelId)
	// Send message to all listener
	for i := 0; i < len(arr); i++ {
		msg := &Message{
			ChannelId: arr[i],
			Text:      text,
		}
		m.messages <- msg
	}
}

func IsClosed(ch <-chan interface{}) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}
