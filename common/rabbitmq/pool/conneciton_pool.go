package pool

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"synapse/common/rabbitmq"
	"sync"
	"time"
)

type PoolConfig struct {
	MaxConnections     int
	MaxChannelsPerConn int
	WaitTimeout        time.Duration
	ReconnectInterval  time.Duration
}

type ConnectionPool struct {
	uri         string
	config      *PoolConfig
	connections []*ConnectionWrapper
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewConnectionPool(uri string, config *PoolConfig) (*ConnectionPool, error) {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &ConnectionPool{
		uri:    uri,
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	if err := pool.initialize(); err != nil {
		return nil, err
	}

	go pool.monitor()
	return pool, nil
}

func (p *ConnectionPool) initialize() error {
	for i := 0; i < p.config.MaxConnections; i++ {
		cw, err := NewConnectionWrapper(p.uri, p.config.MaxChannelsPerConn)
		if err != nil {
			return fmt.Errorf("create connection failed: %w", err)
		}
		p.connections = append(p.connections, cw)
	}
	return nil
}

func (p *ConnectionPool) GetChannel() (*amqp.Channel, error) {
	start := time.Now()

	for {
		p.mutex.RLock()
		for _, cw := range p.connections {
			if ch, err := cw.GetChannel(); err == nil {
				p.mutex.RUnlock()
				return ch, nil
			}
		}
		p.mutex.RUnlock()

		if time.Since(start) > p.config.WaitTimeout {
			return nil, rabbitmq.ErrGetChannelTimeout
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (p *ConnectionPool) ReturnChannel(ch *amqp.Channel) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	for _, cw := range p.connections {
		select {
		case cw.channels <- ch:
			return
		default:
		}
	}
	if err := ch.Close(); err != nil {
		log.Fatalf("Close channel failed: %s", err)
	}
}

func (p *ConnectionPool) monitor() {
	ticker := time.NewTicker(p.config.ReconnectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.checkConnections()
		}
	}
}

func (p *ConnectionPool) checkConnections() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for i, cw := range p.connections {
		if cw.IsAlive() {
			continue
		}

		log.Printf("Reconnecting connection %d...", i)
		newCW, err := NewConnectionWrapper(p.uri, p.config.MaxChannelsPerConn)
		if err != nil {
			log.Printf("Reconnect failed: %v", err)
			continue
		}

		p.connections[i] = newCW
		log.Printf("Connection %d reconnected", i)
	}
}

func (p *ConnectionPool) Close() {
	p.cancel()
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, cw := range p.connections {
		cw.Close()
	}
}
