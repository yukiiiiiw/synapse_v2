package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"synapse/common"
	"synapse/common/log"
	"synapse/worker/constants"
	"synapse/worker/mq"
	"synapse/worker/service"
	"synapse/worker/types"
)

type GpuController struct {
	svc *service.GpuResourceService
}

func NewGpuController(svc *service.GpuResourceService) *GpuController {
	return &GpuController{svc: svc}
}

func (ctl *GpuController) ListAgentNodeResources(ctx *gin.Context) {
	pageMark, pageSize, err := queryPageInfo(ctx)

	if err != nil {
		log.Log.Errorw("get query page info error", zap.Error(err))
		common.JSON(ctx, common.HttpOk, common.ErrBadArgument)
		return
	}

	res, err := ctl.svc.ListAgentNodeResources(ctx, pageMark, pageSize)
	if err != nil {
		httpCode, body := common.GuessError(err, func(error) {
			log.Log.Errorw("get gpu resources error", zap.Error(err))
		})
		common.JSON(ctx, httpCode, body)
		return
	}

	common.JSON(ctx, common.HttpOk, common.Ok(res))
}

func (ctl *GpuController) TestProducer(ctx *gin.Context) {

	body, err := ctx.GetRawData()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}
	//log.Log.Infow("test producer", zap.Any("body", string(body)))
	mq.Producer.Publish(ctx, "orders", "order.*", body)
}

func (ctl *GpuController) ApplyGpuResource(ctx *gin.Context) {
	req := types.ApplyGpuResourceRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Log.Errorw("apply gpu resource bind json failed", zap.Error(err))
		common.JSON(ctx, common.HttpOk, common.ErrBadArgument)
		return
	}
	log.Log.Infow("apply gpu resource", zap.Any("request", req))
	if err := common.Validate(&req); err != nil {
		httpCode, body := common.GuessError(err, func(error) {
			log.Log.Errorw("apply gpu resource validate failed", zap.Error(err))
		})
		common.JSON(ctx, httpCode, body)
		return
	}

	err := ctl.svc.ApplyGpuResource(ctx, &req)

	if err != nil {
		httpCode, body := common.GuessError(err, func(error) {
			log.Log.Errorw("apply gpu resource error", zap.Error(err))
		})
		common.JSON(ctx, httpCode, body)
		return
	}

	common.JSON(ctx, common.HttpOk, common.Ok(nil))
}

func (ctl *GpuController) GetPodStatus(ctx *gin.Context) {
	SassPodIDStr := ctx.Param("podId")
	if SassPodIDStr == "" {
		common.JSON(ctx, common.HttpOk, common.ErrBadArgument)
		return
	}
	SassPodID, err := strconv.ParseInt(SassPodIDStr, 10, 64)
	if err != nil {
		common.JSON(ctx, common.HttpOk, common.ErrBadArgument)
		return
	}

	res, err := ctl.svc.GetPodStatus(ctx, SassPodID)
	if err != nil {
		httpCode, body := common.GuessError(err, func(error) {
			log.Log.Errorw("get gpu pod status error", zap.Error(err))
		})
		common.JSON(ctx, httpCode, body)
		return
	}

	common.JSON(ctx, common.HttpOk, common.Ok(res))
}

func (ctl *GpuController) EditGpuPod(ctx *gin.Context) {
	req := types.EditGpuPodRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Log.Errorw("edit gpu pod bind json failed", zap.Error(err))
		common.JSON(ctx, common.HttpOk, common.ErrBadArgument)
		return
	}
	log.Log.Infow("edit gpu pod", zap.Any("request", req))
	if err := common.Validate(&req); err != nil {
		httpCode, body := common.GuessError(err, func(error) {
			log.Log.Errorw("edit gpu pod validate failed", zap.Error(err))
		})
		common.JSON(ctx, httpCode, body)
		return
	}

	if err := ctl.svc.EditGpuPod(ctx, &req); err != nil {
		httpCode, body := common.GuessError(err, func(error) {
			log.Log.Errorw("edit gpu pod error", zap.Error(err))
		})
		common.JSON(ctx, httpCode, body)
		return
	}

	common.JSON(ctx, common.HttpOk, common.Ok(nil))
}

func (ctl *GpuController) DoPodAction(ctx *gin.Context) {
	sassPodIDStr := ctx.Param("podId")
	action := ctx.Param("action")
	log.Log.Infow("do gpu pod action.", zap.String("action", action), zap.String("sassPodId", sassPodIDStr))
	if sassPodIDStr == "" {
		common.JSON(ctx, common.HttpOk, common.ErrBadArgument)
		return
	}
	SassPodID, err := strconv.ParseInt(sassPodIDStr, 10, 64)
	if err != nil {
		common.JSON(ctx, common.HttpOk, common.ErrBadArgument)
		return
	}
	if action == "" {
		common.JSON(ctx, common.HttpOk, common.ErrBadArgument)
		return
	}

	operateTypeDeploymentStatus, isValidAction := constants.CheckAndGetOperateTypeDeploymentStatus(action)
	if !isValidAction {
		common.JSON(ctx, common.HttpOk, common.ErrBadArgument)
		return
	}

	if err := ctl.svc.DoPodAction(ctx, SassPodID, operateTypeDeploymentStatus); err != nil {
		httpCode, body := common.GuessError(err, func(error) {
			log.Log.Errorw("do gpu pod action error", zap.Error(err))
		})
		common.JSON(ctx, httpCode, body)
		return
	}

	common.JSON(ctx, common.HttpOk, common.Ok(nil))
}

// queryPageInfo extracts and validates pagination parameters from gin context.
// Returns:
// - pageNumber: current page number (minimum: 1)
// - pageSize: number of items per page (range: 2-50)
// - error: validation error if parameters are invalid
func queryPageInfo(ctx *gin.Context) (int64, int, error) {
	// Get page number from query params, default to 1
	pageMark, err := strconv.ParseInt(ctx.Query("pageMark"), 10, 64)
	if err != nil {
		return 0, 0, errors.New("pageMark must be a valid integer")
	}

	// Get page size from query params, default to 10
	pageSize, err := strconv.Atoi(ctx.DefaultQuery("pageSize", constants.DefaultPageSize))
	if err != nil {
		return 0, 0, errors.New("pageSize must be a valid integer")
	}

	// Normalize page size to valid ranges
	return pageMark, normalizePageSize(pageSize), nil
}

// normalizePageSize ensures page size is within allowed range
func normalizePageSize(pageSize int) int {
	if pageSize < constants.MinPageSize {
		return constants.MinPageSize
	}
	if pageSize > constants.MaxPageSize {
		return constants.MaxPageSize
	}
	return pageSize
}
