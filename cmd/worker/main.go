package main

import (
	"context"
	"flag"
	"os/signal"
	"synapse/common"
	commonConfig "synapse/common/config"
	"synapse/common/log"
	"synapse/worker"
	"synapse/worker/config"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	flag.Parse()

	defer log.Log.Sync()
	defer func() {
		if err := recover(); err != nil {
			log.Log.Error("unknown error occurred", zap.Any("err", err))
		}
	}()

	_, err := commonConfig.ReadConfig(common.ServiceWorker, &config.Config)
	if err != nil {
		log.Log.Fatal("read config failed", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	defer stop()

	config.Redis = commonConfig.InitRedis(&config.Config.Redis)

	if err := commonConfig.InitDatasource(ctx, log.Log, &config.Config.Datasource); err != nil {
		log.Log.Fatal("init datasource failed", zap.Error(err))
	}

	if err := worker.Start(ctx); err != nil {
		log.Log.Fatal("start app failed", zap.Error(err))
	}

	<-ctx.Done()
	log.Log.Info("app service stopped")
}
