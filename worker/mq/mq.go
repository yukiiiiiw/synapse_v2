package mq

import (
	"context"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"synapse/common/log"
	"synapse/common/rabbitmq/core"
	"synapse/common/rabbitmq/pool"
	"synapse/worker/config"
	"sync"
)

var (
	connectionPool *pool.ConnectionPool
	once           sync.Once
	Producer       *core.Producer
)

func initPool(ctx context.Context) *pool.ConnectionPool {
	// Initialize the connection pool
	once.Do(func() {
		amqpConfig := config.Config.RabbitMQConfig.Amqp
		log.Log.Infow("init mq pool", zap.Any("config", amqpConfig))

		var err error
		var poolConfig pool.PoolConfig
		err = convertStruct(amqpConfig.Pool, &poolConfig)
		if err != nil {
			panic(err)
		}

		connectionPool, err = pool.NewConnectionPool(
			amqpConfig.Uri,
			&poolConfig,
		)
		if err != nil {
			panic(errors.New("failed to create RabbitMQ connection pool"))
		}
		go func() {
			select {
			case <-ctx.Done():
				log.Log.Infow("closing mq pool", zap.Any("config", amqpConfig))
				connectionPool.Close()
			}
		}()
	})

	return connectionPool
}

func InitProducer(ctx context.Context, config *config.ProducerConfig) {
	if config == nil {
		return
	}
	log.Log.Infow("init producer", zap.Any("config", config))
	connPool := initPool(ctx)

	var producerConfig core.ProducerConfig
	err := convertStruct(config, &producerConfig)
	if err != nil {
		panic(err)
	}

	Producer, err = core.NewProducer(connPool, &producerConfig)
	if err != nil {
		panic(err)
	}
}

func StartConsumer(ctx context.Context, config *config.ConsumerConfig, handler func(body []byte) error) {
	connPool := initPool(ctx)
	log.Log.Infow("start consumer", zap.Any("config", config))

	var consumerConfig core.ConsumerConfig
	err := convertStruct(config, &consumerConfig)
	if err != nil {
		panic(err)
	}

	consumer, err := core.NewConsumer(connPool, &consumerConfig)
	if err != nil {
		panic(err)
	}
	go func() {
		if handlerErr := consumer.Consume(ctx, handler); handlerErr != nil {
			log.Log.Error("Error starting consumer handler", zap.String("consumer", config.Name), zap.Error(err))
		}
		select {
		case <-ctx.Done():
			log.Log.Infow("consumer stopped", zap.String("consumer", config.Name))
			consumer.Close()
		}
	}()
}

func convertStruct[T any](src any, dest *T) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func ClosePool() {
	if connectionPool != nil {
		connectionPool.Close()
	}
}
