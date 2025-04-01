package service

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"synapse/common"
	"synapse/common/config"
	"synapse/common/enum"
	"synapse/common/log"
	"synapse/common/utils"
	"synapse/worker/constants"
	"synapse/worker/repository/repo"
	entity "synapse/worker/repository/types"
	"synapse/worker/types"
	"time"
)

type GpuResourceService struct {
	db *gorm.DB
}

func NewGpuResourceService(db *gorm.DB) *GpuResourceService {
	return &GpuResourceService{db: db}
}

func (svc *GpuResourceService) ListAgentNodeResources(ctx context.Context, pageMark int64, pageSize int) (*types.PageResponse[types.GpuResourceItem], error) {
	repo := repo.NewAgentNodeResourceRepository(svc.db.WithContext(ctx))

	resources, hasMore, err := repo.List(pageMark, pageSize, enum.SortAsc)
	if err != nil {
		return nil, err
	}

	items := make([]types.GpuResourceItem, len(resources))
	res := types.PageResponse[types.GpuResourceItem]{
		Items:    items,
		PageSize: pageSize,
		HasMore:  hasMore,
		PageMark: constants.DefaultPageMark,
	}
	// TODO copy to items
	if len(resources) > 0 {
		for i, resource := range resources {
			err = copier.Copy(&items[i], &resource)
			if err != nil {
				return nil, err
			}
		}
		res.PageMark = resources[len(resources)-1].ID
	}

	return &res, nil
}

func (svc *GpuResourceService) ApplyGpuResource(ctx context.Context, req *types.ApplyGpuResourceRequest) error {
	log.Log.Infof("apply gpu resource: %v", req)

	lock := svc.createGpuResourceRedisLock(req.SassPodID)
	if locked, err := lock.TryLock(ctx); err != nil {
		log.Log.Error("apply gpu resource try lock error.", zap.Int64("SassPodID", req.SassPodID), zap.Error(err))
		return err
	} else if !locked {
		log.Log.Error("apply gpu resource try lock failed.", zap.Int64("SassPodID", req.SassPodID))
		return common.ErrRequestLimit
	}
	defer lock.Unlock(ctx)

	agentServiceRepo := repo.NewAgentServiceRepository(svc.db.WithContext(ctx))
	if _, exists, err := agentServiceRepo.FindBySaasPodID(req.SassPodID); err != nil {
		return fmt.Errorf("apply gpu resource failed to query agent service: %w", err)
	} else if exists {
		return common.ErrSaasPodIDExists
	}

	if err := svc.checkProcessingRequest(ctx, req.SassPodID); err != nil {
		return err
	}

	now := time.Now().UTC().UnixMilli()
	serviceInfo, operateLog := svc.buildApplyInfo(req, now)

	// Save using repository
	scheduleRepo := repo.NewScheduleRepository(svc.db.WithContext(ctx))
	//TODO send request to agent
	return scheduleRepo.SaveSchedule(ctx, serviceInfo, operateLog)
}

func (svc *GpuResourceService) GetPodStatus(ctx context.Context, SassPodID int64) (*types.GpuStatusResponse, error) {
	agentServiceRepo := repo.NewAgentServiceRepository(svc.db.WithContext(ctx))
	existedServiceInfo, exists, err := agentServiceRepo.FindBySaasPodID(SassPodID)
	if err != nil {
		return nil, fmt.Errorf("apply gpu resource failed to query agent service: %w", err)
	} else if !exists {
		return nil, common.ErrSaasPodIDNotExist
	}

	//existedServiceInfo
	// mock data
	// TODO: remove mock data, get data from existedServiceInfo
	res := types.GpuStatusResponse{
		Image:                 "ubuntu:20.04",
		ResourceType:          1,
		GpuType:               "H100",
		GpuCount:              10,
		SingleCardVram:        40,
		SingleCardRam:         8,
		SingleCardVcpu:        4,
		Location:              "shanghai",
		Region:                "cn-shanghai",
		ContainerVolume:       50,
		PersistentVolume:      50,
		PersistentVolumeId:    100,
		PersistentMountPath:   "/data",
		NetworkUpload:         100,
		NetworkDownload:       100,
		DiskReadSpeed:         100,
		DiskWriteSpeed:        100,
		SingleCardPrice:       3.5,
		ContainerVolumePrice:  0.03,
		PersistentVolumePrice: 0.02,
		InitCommand:           "",
		EnvVars: []types.KeyValuePair[string]{
			{Key: "gpuType", Value: "H100"},
			{Key: "ram", Value: "40G"},
		},
		Expose: []types.ExposedPort{
			{PortMapping: types.PortMapping{Port: 6379, Protocol: "tcp"}, ProxyPort: 16379, Host: "localhost"},
			{PortMapping: types.PortMapping{Port: 8080, Protocol: "http"}, ProxyPort: 18080, Host: "localhost"},
		},
		SshUser: "ubuntu",
		SshHost: "localhost",
		SshPort: "10022",
		Status:  existedServiceInfo.Status,
	}
	return &res, nil
}

func (svc *GpuResourceService) EditGpuPod(ctx context.Context, req *types.EditGpuPodRequest) error {
	log.Log.Infof("edit gpu pod: %v", req)
	lock := svc.createGpuResourceRedisLock(req.SassPodID)
	if locked, err := lock.TryLock(ctx); err != nil {
		log.Log.Error("edit gpu pod try lock error.", zap.Int64("SassPodID", req.SassPodID), zap.Error(err))
		return err
	} else if !locked {
		log.Log.Error("edit gpu pod try lock failed.", zap.Int64("SassPodID", req.SassPodID))
		return common.ErrRequestLimit
	}
	defer lock.Unlock(ctx)

	if err := svc.checkProcessingRequest(ctx, req.SassPodID); err != nil {
		return err
	}
	serviceInfoID, err := svc.findServiceInfoID(ctx, req.SassPodID)
	if err != nil {
		return err
	}
	now := time.Now().UTC().UnixMilli()
	// TODO check if the pod is running
	// TODO check if the pod is in error status
	// TODO confirm only save operate log
	_, operateLog := svc.buildEditInfo(req, serviceInfoID, now)

	// Update using repository
	operateLogRepo := repo.NewServiceOperateLogRepository(svc.db.WithContext(ctx))
	//TODO send request to agent
	return operateLogRepo.Save(operateLog)

}

func (svc *GpuResourceService) DoPodAction(ctx context.Context, SassPodID int64, operateType enum.OperateType) error {
	lock := svc.createGpuResourceRedisLock(SassPodID)
	if locked, err := lock.TryLock(ctx); err != nil {
		log.Log.Error("gpu pod action try lock error.", zap.Int64("SassPodID", SassPodID), zap.Int("operateType", int(operateType)), zap.Error(err))
		return err
	} else if !locked {
		log.Log.Error("gpu pod action try lock failed.", zap.Int64("SassPodID", SassPodID), zap.Int("operateType", int(operateType)))
		return common.ErrRequestLimit
	}
	defer lock.Unlock(ctx)

	if err := svc.checkProcessingRequest(ctx, SassPodID); err != nil {
		return err
	}
	serviceInfoID, err := svc.findServiceInfoID(ctx, SassPodID)
	if err != nil {
		return err
	}
	now := time.Now().UTC().UnixMilli()
	// TODO check if the pod is running
	// TODO check if the pod is in error status
	// TODO confirm only save operate log
	// TODO confirm id generation
	_, operateLog := svc.buildGpuActionInfo(operateType, SassPodID, serviceInfoID, now)

	// Update using repository
	operateLogRepo := repo.NewServiceOperateLogRepository(svc.db.WithContext(ctx))
	//TODO send request to agent
	return operateLogRepo.Save(operateLog)
}

func (svc *GpuResourceService) buildApplyInfo(req *types.ApplyGpuResourceRequest, now int64) (*entity.AgentServiceInfo, *entity.ServiceOperateLog) {
	serviceInfo := &entity.AgentServiceInfo{
		UpdatedAt: now,
		Version:   1,
		SaasPodID: req.SassPodID,
		Status:    int(enum.Deployment_Status_Deploying),
		CreatedAt: now,
	}

	// Only one field is missing from ApplyGpuResourceRequest: SassPodID
	operateInfo := map[string]interface{}{
		"image":                 req.Image,
		"region":                req.Region,
		"cloudType":             req.CloudType,
		"resourceType":          req.ResourceType,
		"gpuType":               req.GpuType,
		"gpuCount":              req.GpuCount,
		"singleCardVram":        req.SingleCardVram,
		"singleCardRam":         req.SingleCardRam,
		"singleCardVcpu":        req.SingleCardVcpu,
		"containerVolume":       req.ContainerVolume,
		"persistentVolume":      req.PersistentVolume,
		"persistentVolumeId":    req.PersistentVolumeId,
		"persistentMountPath":   req.PersistentMountPath,
		"singleCardPrice":       req.SingleCardPrice,
		"persistentVolumePrice": req.PersistentVolumePrice,
		"containerVolumePrice":  req.ContainerVolumePrice,
		"initCommand":           req.InitCommand,
		"envVars":               req.EnvVars,
		"expose":                req.Expose,
		"sshUser":               req.SshUser,
		"sshPublicKey":          req.SshPublicKey,
	}

	operateLog := &entity.ServiceOperateLog{
		SaasPodID:    req.SassPodID,
		OperateType:  int(enum.OperateType_Deploy),
		Status:       int(enum.Operate_Status_Init),
		ScheduleInfo: operateInfo,
		CreatedAt:    now,
		UpdatedAt:    now,
		Version:      1,
	}
	return serviceInfo, operateLog
}

func (svc *GpuResourceService) buildEditInfo(req *types.EditGpuPodRequest, serviceInfoID int64, now int64) (*entity.AgentServiceInfo, *entity.ServiceOperateLog) {
	// TODO
	serviceInfo := &entity.AgentServiceInfo{
		UpdatedAt: now,
		Version:   1,
		SaasPodID: req.SassPodID,
		Status:    int(enum.Deployment_Status_Deploying),
		CreatedAt: now,
	}

	// Only one field is missing from EditGpuPodRequest: SassPodID
	operateInfo := map[string]interface{}{
		"image":               req.Image,
		"containerVolume":     req.ContainerVolume,
		"persistentVolume":    req.PersistentVolume,
		"persistentVolumeId":  req.PersistentVolumeId,
		"persistentMountPath": req.PersistentMountPath,
		"initCommand":         req.InitCommand,
		"envVars":             req.EnvVars,
		"expose":              req.Expose,
		"sshUser":             req.SshUser,
		"sshPublicKey":        req.SshPublicKey,
	}

	operateLog := &entity.ServiceOperateLog{
		SaasPodID:     req.SassPodID,
		OperateType:   int(enum.OperateType_Update),
		Status:        int(enum.Operate_Status_Init),
		ScheduleInfo:  operateInfo,
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       1,
		ServiceInfoID: serviceInfoID,
	}
	return serviceInfo, operateLog
}

func (svc *GpuResourceService) buildGpuActionInfo(operateType enum.OperateType, sassPodID, serviceInfoID int64, now int64) (*entity.AgentServiceInfo, *entity.ServiceOperateLog) {
	// TODO
	serviceInfo := &entity.AgentServiceInfo{
		UpdatedAt: now,
		Version:   1,
		SaasPodID: sassPodID,
		Status:    int(enum.Deployment_Status_Deploying),
		CreatedAt: now,
	}

	operateLog := &entity.ServiceOperateLog{
		SaasPodID:     sassPodID,
		OperateType:   int(operateType),
		Status:        int(enum.Operate_Status_Init),
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       1,
		ServiceInfoID: serviceInfoID,
	}
	return serviceInfo, operateLog
}

func (svc *GpuResourceService) createGpuResourceRedisLock(saasPodId int64) *utils.Lock {
	lockKey := fmt.Sprintf(constants.GpuResourceLockKey, saasPodId)
	lockToken := fmt.Sprintf("token-%d", time.Now().UnixNano())
	return utils.NewRedisLock(lockKey, lockToken, config.Redis)
}

func (svc *GpuResourceService) checkProcessingRequest(ctx context.Context, sassPodId int64) error {
	operateLogRepo := repo.NewServiceOperateLogRepository(svc.db.WithContext(ctx))
	existedOperateLog, exists, err := operateLogRepo.FindBySaasPodID(sassPodId)
	if err != nil {
		log.Log.Error("failed to query service operate log",
			zap.Int64("SassPodID", sassPodId),
			zap.Error(err))
		return fmt.Errorf("查询操作日志失败: %w", err)
	}

	if exists && existedOperateLog.Status == int(enum.Operate_Status_Init) {
		log.Log.Info("request is being processed",
			zap.Int64("SassPodID", sassPodId))
		return common.ErrSaasPodProcessing
	}
	return nil
}

func (svc *GpuResourceService) checkExistsAgentService(ctx context.Context, sassPodId int64) error {
	_, err := svc.findServiceInfoID(ctx, sassPodId)
	return err
}

func (svc *GpuResourceService) findServiceInfoID(ctx context.Context, sassPodId int64) (int64, error) {
	agentServiceRepo := repo.NewAgentServiceRepository(svc.db.WithContext(ctx))
	existedServiceInfo, exists, err := agentServiceRepo.FindBySaasPodID(sassPodId)
	if err != nil {
		log.Log.Error("failed to query agent service", zap.Int64("SassPodID", sassPodId), zap.Error(err))
		return 0, fmt.Errorf("query agent service: %w", err)
	} else if !exists {
		log.Log.Error("agent service info with SaasPodId not exist", zap.Int64("SassPodID", sassPodId))
		return 0, common.ErrSaasPodIDNotExist
	}
	return existedServiceInfo.ID, nil
}
