package mq

import (
	"context"
	"errors"
	"github.com/jinzhu/copier"
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

func InitPool() *pool.ConnectionPool {
	// Initialize the connection pool
	once.Do(func() {
		amqpConfig := config.Config.RabbitMQConfig.Amqp
		log.Log.Infow("init mq pool", zap.Any("config", amqpConfig))
		poolConfig := amqpConfig.Pool

		var err error
		connectionPool, err = pool.NewConnectionPool(
			amqpConfig.Uri,
			&pool.PoolConfig{
				MaxConnections:     poolConfig.MaxConnections,
				MaxChannelsPerConn: poolConfig.MaxChannelsPerConn,
				WaitTimeout:        poolConfig.WaitTimeoutSeconds,
				ReconnectInterval:  poolConfig.ReconnectIntervalSeconds,
			},
		)
		if err != nil {
			panic(errors.New("failed to create RabbitMQ connection pool"))
		}
	})

	return connectionPool
}

func InitProducer(config *config.ProducerConfig) {
	if config == nil {
		return
	}

	log.Log.Infow("init producer", zap.Any("config", config))
	connPool := InitPool()

	exchangeCfgs := make([]core.ExchangeConfig, len(config.ExchangeConfig))

	if len(exchangeCfgs) > 0 {
		for i, resource := range config.ExchangeConfig {
			err := copier.Copy(&exchangeCfgs[i], &resource)
			if err != nil {
				panic(err)
			}
		}
	}
	producerCfg := &core.ProducerConfig{
		Exchanges:     exchangeCfgs,
		Retries:       config.Retries,
		RetryInterval: config.RetryIntervalSeconds,
	}

	var err error
	Producer, err = core.NewProducer(connPool, producerCfg)
	if err != nil {
		panic(err)
	}
}

func StartConsumer(ctx context.Context, config *config.ConsumerConfig, handler func(body []byte) error) {
	connPool := InitPool()
	log.Log.Infow("start consumer", zap.Any("config", config))
	consumerConfig := core.ConsumerConfig{
		Queue:         config.Queue,
		Exchange:      config.Exchange,
		RoutingKey:    config.RoutingKey,
		ExchangeType:  config.ExchangeType,
		Durable:       config.Durable,
		AutoAck:       config.AutoAck,
		PrefetchCount: config.PrefetchCount,
	}
	consumer, err := core.NewConsumer(connPool, &consumerConfig)
	if err != nil {
		panic(err)
	}
	go func() {
		if handlerErr := consumer.Consume(ctx, handler); handlerErr != nil {
			log.Log.Error("Error starting consumer handler", zap.String("consumer", config.Name), zap.Error(err))
		}
	}()
}

func ClosePool() {
	if connectionPool != nil {
		connectionPool.Close()
	}
}
