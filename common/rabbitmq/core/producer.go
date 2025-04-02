package core

import (
	"context"
	"log"
	"synapse/common/rabbitmq"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"synapse/common/rabbitmq/pool"
)

type ProducerConfig struct {
	Exchanges     []ExchangeConfig
	Retries       int
	RetryInterval time.Duration
}

type ExchangeConfig struct {
	Name    string
	Type    string
	Durable bool
}

type Producer struct {
	pool   *pool.ConnectionPool
	config *ProducerConfig
}

func NewProducer(pool *pool.ConnectionPool, config *ProducerConfig) (*Producer, error) {
	ch, err := pool.GetChannel()
	if err != nil {
		return nil, err
	}
	defer pool.ReturnChannel(ch)

	for _, exchange := range config.Exchanges {
		if err := ch.ExchangeDeclare(
			exchange.Name,
			exchange.Type,
			exchange.Durable,
			false,
			false,
			false,
			nil,
		); err != nil {
			panic("declare exchange failed: " + err.Error())
		}
	}

	return &Producer{
		pool:   pool,
		config: config,
	}, nil
}

func (p *Producer) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	for i := 0; i <= p.config.Retries; i++ {
		ch, err := p.pool.GetChannel()
		if err != nil {
			log.Fatalf("get channel failed: %v", err)
			time.Sleep(p.config.RetryInterval)
			continue
		}

		defer p.pool.ReturnChannel(ch)
		err = ch.PublishWithContext(ctx,
			exchange,
			routingKey,
			false,
			false,
			amqp.Publishing{
				ContentType:  "application/json",
				Body:         body,
				DeliveryMode: amqp.Persistent,
				Timestamp:    time.Now(),
			},
		)

		if err == nil {
			return nil
		}

		if i < p.config.Retries {
			time.Sleep(time.Duration(i+1) * p.config.RetryInterval)
		}
	}
	return rabbitmq.ErrMaxRetriesReached
}
