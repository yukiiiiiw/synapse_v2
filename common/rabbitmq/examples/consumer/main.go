package main

import (
	"context"
	"log"
	"synapse/common/rabbitmq/core"
	"time"

	"synapse/common/rabbitmq/pool"
)

func main() {
	poolCfg := &pool.PoolConfig{
		MaxConnections:     3,
		MaxChannelsPerConn: 10,
		WaitTimeout:        5 * time.Second,
		ReconnectInterval:  10 * time.Second,
	}

	connPool, err := pool.NewConnectionPool(
		"amqp://rabbitmq:rabbitmq@127.0.0.1:5672/testhost",
		poolCfg,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer connPool.Close()

	consumerCfg := &core.ConsumerConfig{
		Queue:         "order_queue",
		Exchange:      "orders",
		RoutingKey:    "order.created",
		ExchangeType:  "topic",
		Durable:       true,
		AutoAck:       false,
		PrefetchCount: 10,
	}

	consumer, err := core.NewConsumer(connPool, consumerCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	if err := consumer.Consume(context.Background(), func(body []byte) error {
		log.Printf("Received order: %s", body)
		// 处理业务逻辑
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	select {}
}
