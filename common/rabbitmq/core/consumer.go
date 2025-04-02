package core

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"synapse/common/rabbitmq/pool"
)

type ConsumerConfig struct {
	Queue         string
	Exchange      string
	RoutingKey    string
	ExchangeType  string
	Durable       bool
	AutoAck       bool
	PrefetchCount int
}

type Consumer struct {
	pool   *pool.ConnectionPool
	config *ConsumerConfig
	cancel context.CancelFunc
}

func NewConsumer(pool *pool.ConnectionPool, config *ConsumerConfig) (*Consumer, error) {
	ch, err := pool.GetChannel()
	if err != nil {
		return nil, err
	}
	defer pool.ReturnChannel(ch)

	if err := prepareConsumeChannel(ch, config); err != nil {
		return nil, err
	}

	return &Consumer{
		pool:   pool,
		config: config,
	}, nil
}

func prepareConsumeChannel(ch *amqp.Channel, config *ConsumerConfig) error {
	if err := ch.ExchangeDeclare(
		config.Exchange,
		config.ExchangeType,
		config.Durable,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare exchange failed: %w", err)
	}

	_, err := ch.QueueDeclare(
		config.Queue,
		config.Durable,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare queue failed: %w", err)
	}

	if err := ch.QueueBind(
		config.Queue,
		config.RoutingKey,
		config.Exchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("queue bind failed: %w", err)
	}

	if err := ch.Qos(
		config.PrefetchCount,
		0,
		false,
	); err != nil {
		return fmt.Errorf("set qos failed: %w", err)
	}

	return nil
}

func (c *Consumer) Consume(ctx context.Context, handler func([]byte) error) error {
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	go c.runConsumptionLoop(ctx, handler)
	return nil
}

func (c *Consumer) runConsumptionLoop(ctx context.Context, handler func([]byte) error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			ch, err := c.pool.GetChannel()
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			msgs, err := ch.Consume(
				c.config.Queue,
				"",
				c.config.AutoAck,
				false,
				false,
				false,
				nil,
			)
			if err != nil {
				c.pool.ReturnChannel(ch)
				time.Sleep(1 * time.Second)
				continue
			}

			c.handleMessages(ch, msgs, handler)
		}
	}
}

func (c *Consumer) handleMessages(ch *amqp.Channel, msgs <-chan amqp.Delivery, handler func([]byte) error) {
	defer c.pool.ReturnChannel(ch)

	for msg := range msgs {
		if err := handler(msg.Body); err != nil {
			if !c.config.AutoAck {
				msg.Nack(false, true)
			}
			continue
		}

		if !c.config.AutoAck {
			msg.Ack(false)
		}
	}
}

func (c *Consumer) Close() {
	if c.cancel != nil {
		c.cancel()
	}
}
