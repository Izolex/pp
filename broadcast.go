package main

import (
	"sync"
)

type Broadcast struct {
	sync.RWMutex
	running bool
	history []byte
	stop    chan struct{}
	msg     chan []byte
	sub     chan chan []byte
	unsub   chan chan []byte
}

func NewBroadcast() *Broadcast {
	br := &Broadcast{
		history: make([]byte, 0),
		running: false,
		stop:    make(chan struct{}),
		msg:     make(chan []byte),
		sub:     make(chan chan []byte),
		unsub:   make(chan chan []byte),
	}
	go br.run()
	return br
}

func (b *Broadcast) Subscribe() chan []byte {
	b.RLock()
	defer b.RUnlock()

	ch := make(chan []byte, 1)
	ch <- b.history

	if !b.running {
		close(ch)
	} else {
		b.sub <- ch
	}
	return ch
}

func (b *Broadcast) Unsubscribe(ch chan []byte) {
	b.unsub <- ch
}

func (b *Broadcast) Publish(msg []byte) {
	b.Lock()
	b.history = append(b.history, msg...)
	b.Unlock()

	b.msg <- msg
}

func (b *Broadcast) Start() {
	b.Lock()
	defer b.Unlock()

	b.running = true
}

func (b *Broadcast) Stop() {
	b.Lock()
	defer b.Unlock()

	b.running = false
	b.stop <- struct{}{}
}

func (b *Broadcast) run() {
	subs := map[chan []byte]struct{}{}

	for {
		select {
		case <-b.stop:
			for sub := range subs {
				close(sub)
			}
			subs = map[chan []byte]struct{}{}
		case sub := <-b.sub:
			subs[sub] = struct{}{}
		case sub := <-b.unsub:
			if _, ok := subs[sub]; ok {
				delete(subs, sub)
				close(sub)
			}
		case msg := <-b.msg:
			for ch := range subs {
				ch <- msg
			}
		}
	}
}
