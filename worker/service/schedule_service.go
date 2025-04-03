package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"gorm.io/gorm"
	"synapse/common/enum"
	"synapse/common/log"
	"synapse/worker/config"
	"synapse/worker/repository/repo"
	entity "synapse/worker/repository/types"
	"synapse/worker/types"
	"synapse/worker/util"
	"time"
)

type ScheduleService struct {
	db *gorm.DB
}

func NewScheduleService(db *gorm.DB) *ScheduleService {
	return &ScheduleService{db: db}
}

func (svc *ScheduleService) HandleServiceReport(ctx context.Context, req *types.ServiceReportRequest) {

	agentServiceRepo := repo.NewAgentServiceRepository(svc.db.WithContext(ctx))
	operatorLogRepo := repo.NewServiceOperateLogRepository(svc.db.WithContext(ctx))
	scheduleRepo := repo.NewScheduleRepository(svc.db.WithContext(ctx))

	now := time.Now().UTC().UnixMilli()

	service, exist, _ := agentServiceRepo.FindBySaasPodID(req.Service.ID)
	if !exist {
		log.Log.Errorw("agent service not exists", "SaasPodID", req.Service.ID)
		return
	}
	svc.BuildAgentServiceUpdateInfo(req, service)
	checkMd5 := svc.CheckServiceInfoByMd5(service)
	scheduleTimeOut := svc.CheckScheduleTimeOut(service, now)

	operateLog, exist, _ := operatorLogRepo.FindByServiceInfoID(service.ID)
	if !exist {
		if !checkMd5 {
			log.Log.Errorw("service info md5 mismatch", "serviceID", service.ID)
		}
		return
	}

	// handle final status
	if req.Service.Status == int(enum.Deployment_Status_Failure) ||
		(operateLog.OperateType == int(enum.OperateType_Delete) && req.Service.Status == int(enum.Deployment_Status_Deleted)) {

		service.Status = req.Service.Status
		err := scheduleRepo.UpdateOperateCompleted(service, operateLog)
		if err != nil {
			log.Log.Errorw("failed to update operate success", "error", err)
			return
		}
	}

	if req.Service.Status == int(enum.Deployment_Status_Running) {
		switch enum.OperateType(operateLog.OperateType) {
		case enum.OperateType_Deploy, enum.OperateType_Restart, enum.OperateType_Resume:
			service.Status = int(enum.Deployment_Status_Running)
			err := scheduleRepo.UpdateOperateCompleted(service, operateLog)
			if err != nil {
				log.Log.Errorw("failed to update operate success", "serviceID", service.ID, "operateLogID", operateLog.ID, "error", err)
			}

		case enum.OperateType_Delete, enum.OperateType_Pause:
			if scheduleTimeOut {
				log.Log.Errorw("operation timeout", "serviceID", service.ID, "operateLogID", operateLog.ID)
				// todo alert
			}

		case enum.OperateType_Update:
			if checkMd5 {
				service.Status = int(enum.Deployment_Status_Running)
				err := scheduleRepo.UpdateOperateCompleted(service, operateLog)
				if err != nil {
					log.Log.Errorw("failed to update operate success", "serviceID", service.ID, "operateLogID", operateLog.ID, "error", err)
					return
				}
			}

			if !checkMd5 && scheduleTimeOut {
				log.Log.Errorw("operation timeout", "serviceID", service.ID, "operateLogID", operateLog.ID)
				service.Status = int(enum.Deployment_Status_Deleting)
				service.Version += 1
				operateLog.Version += 1

				// force stop and delete curr service for fast fail
				deleteOperateLog := svc.handleScheduleOperateTimeout(service, now)
				operateLogs := []*entity.ServiceOperateLog{operateLog, deleteOperateLog}
				err := scheduleRepo.UpdateSchedule(ctx, service, operateLogs)
				if err != nil {
					log.Log.Errorw("fail to handle schedule operate timeout", "serviceID", service.ID, "operateLogID", operateLog.ID, "error", err)
					return
				}

				//todo send schedule msg
			}

		}
	}

}

func (svc *ScheduleService) BuildAgentServiceUpdateInfo(req *types.ServiceReportRequest, service *entity.AgentServiceInfo) {
	service.MetricInfo = req
	service.ServiceUpdatedAt = req.Service.StartAt
	service.LastHeartBeat = req.MetricTimestamp
}

func (svc *ScheduleService) CalculateMetricInfoMd5(service *entity.AgentServiceInfo) (string, error) {
	metricInfo := service.MetricInfo.(*types.ServiceReportRequest)

	// Convert MetricInfo to ServiceInfo
	serviceInfo := &types.ServiceInfo{
		AgentID:                metricInfo.AgentID,
		ResourceType:           int(metricInfo.ResourceType),
		GPUType:                metricInfo.GPUType,
		Image:                  metricInfo.Service.Image,
		AllocatedGPUCount:      metricInfo.Service.AllocatedGPUCount,
		AllocatedVCPU:          metricInfo.Service.AllocatedVCPU,
		AllocatedVRAM:          metricInfo.Service.AllocatedVRAM,
		AllocatedRAM:           metricInfo.Service.AllocatedRAM,
		AllocatedCPU:           metricInfo.Service.AllocatedCPU,
		AllocatedMemory:        metricInfo.Service.AllocatedMemory,
		AllocatedStorage:       metricInfo.Service.AllocatedStorage,
		ContainerVolumeMounts:  metricInfo.Service.ContainerVolumeMounts,
		PersistentVolumeMounts: metricInfo.Service.PersistentVolumeMounts,
		ClusterIP:              metricInfo.Service.ClusterIP,
		ExternalIP:             metricInfo.Service.ExternalIP,
		Ingress:                metricInfo.Service.Ingress,
		Ports:                  metricInfo.Service.Ports,
		SSH:                    metricInfo.Service.SSH,
	}

	return svc.CalculateServiceInfoMd5(serviceInfo)
}

func (svc *ScheduleService) CalculateServiceInfoMd5(serviceInfo *types.ServiceInfo) (string, error) {
	jsonBytes, err := json.Marshal(serviceInfo)
	if err != nil {
		log.Log.Errorw("failed to calculate md5", "error", err)
		return "", err
	}

	// Calculate MD5
	hash := md5.New()
	hash.Write(jsonBytes)
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (svc *ScheduleService) CheckServiceInfoByMd5(service *entity.AgentServiceInfo) bool {
	// Calculate md5 for current service info
	serviceInfo := service.ServiceInfo.(*types.ServiceInfo)
	jsonBytes, err := json.Marshal(serviceInfo)
	if err != nil {
		log.Log.Errorw("failed to calculate service info md5", "error", err)
		return false
	}

	hash := md5.New()
	hash.Write(jsonBytes)
	serviceMd5 := hex.EncodeToString(hash.Sum(nil))

	// Calculate md5 for metric info
	metricInfoMd5, err := svc.CalculateMetricInfoMd5(service)
	if err != nil {
		log.Log.Errorw("failed to calculate metric info md5", "error", err)
		return false
	}

	return serviceMd5 == metricInfoMd5
}

func (svc *ScheduleService) CheckScheduleTimeOut(service *entity.AgentServiceInfo, now int64) bool {
	timeout := config.Config.AgentNode.HeartbeatTimeout
	return now-service.UpdatedAt > timeout.Milliseconds()
}

// handleScheduleOperateTimeout creates a new delete operation log
func (svc *ScheduleService) handleScheduleOperateTimeout(service *entity.AgentServiceInfo, now int64) *entity.ServiceOperateLog {
	return &entity.ServiceOperateLog{
		ID:            util.GetNextId(),
		ServiceInfoID: service.ID,
		SaasPodID:     service.SaasPodID,
		OperateType:   int(enum.OperateType_Delete),
		Status:        int(enum.Operate_Status_Init),
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       0,
	}
}
