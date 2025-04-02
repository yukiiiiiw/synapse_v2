package worker

import (
	"context"
	"fmt"
	sentryGin "github.com/cockroachdb/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	idgenerator "synapse/common/id-generator"
	"synapse/common/log"
	"synapse/worker/config"
	"synapse/worker/mq"
)

func Start(ctx context.Context) error {

	if err := idgenerator.InitMultiSnowflakeInstances(
		ctx,
		config.Redis,
		config.SnowflakeNodeIDRedisKeyForNextId,
	); err != nil {
		return errors.WithMessagef(err, "init multi snowflake instances failed")
	}

	engine := gin.New()
	if config.Config.Server.GIN.Mode != "" {
		gin.SetMode(config.Config.Server.GIN.Mode)
	}
	engine.Use(gin.Recovery())

	if config.Config.Sentry.Enabled {
		engine.Use(sentryGin.New(sentryGin.Options{})) // use sentry
	}

	if err := InitRouter(ctx, engine); err != nil {
		return errors.WithMessagef(err, "init router failed")
	}

	go func() {
		if err := http.ListenAndServe(
			fmt.Sprintf("%s:%d", config.Config.Server.Host, config.Config.Server.Port),
			engine,
		); err != nil {
			log.Log.Warn("http server stopped", zap.Error(err))
		}
	}()

	// Initialize the producer
	mq.InitProducer(&config.Config.RabbitMQConfig.Producer)

	// Initialize the consumers
	consumers := config.Config.RabbitMQConfig.Consumers
	for _, consumer := range consumers {
		mq.StartConsumer(ctx, &consumer, func(body []byte) error {
			log.Log.Infof("Received order: %s", body)
			return nil
		})
	}

	return nil
}
