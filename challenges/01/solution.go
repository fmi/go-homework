package main

// NewPubSub returns an initialized PubSub
func NewPubSub() *PubSub {
	p := new(PubSub)
	p.input = make(chan string)
	p.subscribers = make([]chan string, 0)
	p.slock = make(chan struct{}, 1)
	go p.broadcast()
	return p
}

// PubSub is used to broadcast messages to multiple subscribers
type PubSub struct {
	input       chan string
	subscribers []chan string
	slock       chan struct{}
}

// Subscribe creates and returns a new subscriber channel
func (b *PubSub) Subscribe() <-chan string {
	member := make(chan string)
	b.lock()
	b.subscribers = append(b.subscribers, member)
	b.unlock()
	return member
}

// Publish sends a message to all registered subscribers
func (b *PubSub) Publish() chan<- string {
	return b.input
}

func (b *PubSub) lock() {
	b.slock <- struct{}{}
}

func (b *PubSub) unlock() {
	<-b.slock
}

func (b *PubSub) broadcast() {
	for message := range b.input {
		b.lock()
		for _, member := range b.subscribers {
			member <- message
		}
		b.unlock()
	}
}
