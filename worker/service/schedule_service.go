package service

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"strconv"
	"synapse/common/enum"
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

func (svc *ScheduleService) ApplyAgentService(ctx context.Context, req *types.ApplyRequest) error {
	now := time.Now().UTC().UnixMilli()
	serviceInfo := svc.buildAgentServiceInfo(req, now)
	operateLog := svc.buildServiceOperateLog(req.SaasPodID, enum.OperateType_Deploy, enum.Operate_Status_Init, req.OperateInfo, now)

	// Save using repository
	scheduleRepo := repo.NewScheduleRepository(svc.db)
	return scheduleRepo.SaveSchedule(ctx, serviceInfo, operateLog)
}

// UpdateAgentService updates an AgentService
func (svc *ScheduleService) UpdateAgentService(ctx context.Context, req *types.UpdateAgentServiceRequest) error {
	// Query for existing operation logs
	logRepo := repo.NewServiceOperateLogRepository(svc.db)
	// TODO ServiceInfoID or SassPodID
	logs, err := logRepo.FindByServiceInfoID(strconv.FormatInt(req.SaasPodID, 10))
	if err != nil {
		return fmt.Errorf("failed to query operate logs: %w", err)
	}

	// Check if there are any ongoing operations
	for _, log := range logs {
		if log.Status == int(enum.Operate_Status_Init) {
			return fmt.Errorf("service %s is already in operation", req.SaasPodID)
		}
	}

	now := time.Now().UTC().UnixMilli()
	serviceInfo := svc.buildAgentServiceInfo(req, now)
	operateLog := svc.buildServiceOperateLog(req.SaasPodID, enum.OperateType_Update, enum.Operate_Status_Init, req.OperateInfo, now)

	// Update using repository
	scheduleRepo := repo.NewScheduleRepository(svc.db)
	return scheduleRepo.UpdateSchedule(ctx, serviceInfo, []*entity.ServiceOperateLog{operateLog})
}

// buildAgentServiceInfo builds an AgentServiceInfo entity
func (svc *ScheduleService) buildAgentServiceInfo(req interface{}, now int64) *entity.AgentServiceInfo {
	serviceInfo := &entity.AgentServiceInfo{
		ID:               util.GetNextId(),
		ServiceName:      req.(interface{ GetServiceName() string }).GetServiceName(),
		ServiceInfo:      req.(interface{ GetServiceInfo() map[string]interface{} }).GetServiceInfo(),
		UpdatedAt:        now,
		ServiceUpdatedAt: now,
		Version:          0,
	}

	switch r := req.(type) {
	case *types.ApplyRequest:
		serviceInfo.SaasPodID = r.SaasPodID
		serviceInfo.Status = int(enum.Deployment_Status_Deploying)
		serviceInfo.CreatedAt = now
	case *types.UpdateAgentServiceRequest:
		serviceInfo.SaasPodID = r.SaasPodID
		serviceInfo.Status = int(enum.Deployment_Status_Updating)
	}

	return serviceInfo
}

// buildServiceOperateLog builds a ServiceOperateLog entity
func (svc *ScheduleService) buildServiceOperateLog(
	saasPodId int64,
	operateType enum.OperateType,
	operateStatus enum.OperateStatus,
	operateInfo map[string]interface{},
	now int64,
) *entity.ServiceOperateLog {
	return &entity.ServiceOperateLog{
		ID:           util.GetNextId(),
		SaasPodID:    saasPodId,
		OperateType:  int(operateType),
		Status:       int(operateStatus),
		ScheduleInfo: operateInfo,
		CreatedAt:    now,
		UpdatedAt:    now,
		Version:      0,
	}
}
