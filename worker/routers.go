package worker

import (
	"context"
	"github.com/gin-gonic/gin"
	"synapse/common/config"
	"synapse/worker/controllers"
	"synapse/worker/middleware"
	service "synapse/worker/service"
)

func InitRouter(ctx context.Context, engine *gin.Engine) error {
	// init rpc client
	// Init other services

	var (
		apiGroupAuth = engine.Group("/api/v2", middleware.RequestHeader(), middleware.Authentication())
	)

	{
		// Health check
		ctl := controllers.NewHealthController(nil)
		apiGroupAuth.GET("/health", ctl.Health)
		apiGroupAuth.GET("/status", ctl.Status)
	}

	{
		// gpu resource
		svc := service.NewGpuResourceService(config.DB)
		ctl := controllers.NewGpuController(svc)

		apiGroupAuth.GET("/resource/gpu/infos", ctl.ListAgentNodeResources)
		apiGroupAuth.POST("/resource/gpu/apply", ctl.ApplyGpuResource)
		apiGroupAuth.GET("/resource/gpu/status/:podId", ctl.GetPodStatus)
		apiGroupAuth.PUT("/resource/gpu/:action/:podId", ctl.DoPodAction)
		apiGroupAuth.PUT("/resource/gpu/edit", ctl.EditGpuPod)
	}

	return nil
}
