package main

import (
	"context"
	"fmt"
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

	producerCfg := &core.ProducerConfig{
		Exchanges: []core.ExchangeConfig{
			{
				Name:    "orders",
				Type:    "topic",
				Durable: true,
			},
		},
		Retries:       3,
		RetryInterval: 1 * time.Second,
	}

	producer, err := core.NewProducer(connPool, producerCfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		if err := producer.Publish(ctx, "orders", "order.*", []byte(fmt.Sprintf(`{"id": %d}`, i))); err != nil {
			log.Fatal("Publish failed:", err)
		}
	}

}
