package pool

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"synapse/common/rabbitmq"
	"sync"
	"time"
)

type ConnectionWrapper struct {
	conn         *amqp.Connection
	channels     chan *amqp.Channel
	mutex        sync.RWMutex
	lastActive   time.Time
	channelCount int
}

func NewConnectionWrapper(uri string, maxChannels int) (*ConnectionWrapper, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, err
	}

	cw := &ConnectionWrapper{
		conn:       conn,
		channels:   make(chan *amqp.Channel, maxChannels),
		lastActive: time.Now(),
	}

	for i := 0; i < maxChannels; i++ {
		ch, err := cw.createChannel()
		if err != nil {
			continue
		}
		cw.channels <- ch
		cw.channelCount++
	}

	return cw, nil
}

func (cw *ConnectionWrapper) createChannel() (*amqp.Channel, error) {
	cw.mutex.Lock()
	defer cw.mutex.Unlock()

	if cw.conn.IsClosed() {
		return nil, rabbitmq.ErrConnectionClosed
	}

	return cw.conn.Channel()
}

func (cw *ConnectionWrapper) GetChannel() (*amqp.Channel, error) {
	select {
	case ch := <-cw.channels:
		if !ch.IsClosed() {
			return ch, nil
		}
		return cw.createChannel()
	default:
		return cw.createChannel()
	}
}

func (cw *ConnectionWrapper) ReturnChannel(ch *amqp.Channel) {
	if ch.IsClosed() {
		return
	}

	select {
	case cw.channels <- ch:
	default:
		ch.Close()
	}
}

func (cw *ConnectionWrapper) IsAlive() bool {
	return !cw.conn.IsClosed()
}

func (cw *ConnectionWrapper) Close() {
	cw.conn.Close()
	close(cw.channels)
	for ch := range cw.channels {
		ch.Close()
	}
}
