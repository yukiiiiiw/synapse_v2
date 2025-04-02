package repo

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"synapse/common/log"
	entity "synapse/worker/repository/types"
)

type ScheduleRepository struct {
	db *gorm.DB
}

func NewScheduleRepository(db *gorm.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

func (r *ScheduleRepository) SaveSchedule(ctx context.Context, info *entity.AgentServiceInfo, operateLog *entity.ServiceOperateLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create repositories with transaction
		infoRepo := NewAgentServiceRepository(tx)
		logRepo := NewServiceOperateLogRepository(tx)

		// Save AgentServiceInfo
		if err := infoRepo.Save(info); err != nil {
			log.Log.Errorw("Failed to save agent service info",
				"error", err,
				"saas_pod_id", info.SaasPodID)
			return fmt.Errorf("failed to save agent service info: %w", err)
		}

		// Save ServiceOperateLog
		operateLog.ServiceInfoID = info.ID
		if err := logRepo.Save(operateLog); err != nil {
			log.Log.Errorw("Failed to save service operate log",
				"error", err,
				"service_id", operateLog.ServiceInfoID)
			return fmt.Errorf("failed to save service operate log: %w", err)
		}

		return nil
	})
}

func (r *ScheduleRepository) UpdateScheduleStatus(ctx context.Context, info *entity.AgentServiceInfo, operateLog *entity.ServiceOperateLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		infoRepo := NewAgentServiceRepository(tx)
		logRepo := NewServiceOperateLogRepository(tx)

		updated, err := infoRepo.UpdateAgentServiceInfoStatus(info)
		if err != nil {
			return err
		}
		if !updated {
			return fmt.Errorf("UpdateAgentServiceInfoStatus error")
		}
		operateLog.ServiceInfoID = info.ID
		if err := logRepo.Save(operateLog); err != nil {
			log.Log.Errorw("Failed to save service operate log",
				"error", err,
				"service_id", info.ID)
			return fmt.Errorf("failed to save service operate log: %w", err)
		}

		return nil
	})
}

func (r *ScheduleRepository) UpdateSchedule(ctx context.Context, info *entity.AgentServiceInfo, operateLogs []*entity.ServiceOperateLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create repositories with transaction
		infoRepo := NewAgentServiceRepository(tx)
		logRepo := NewServiceOperateLogRepository(tx)

		// Update AgentServiceInfo with optimistic lock
		info.Version += 1
		if err := infoRepo.VersionSave(info); err != nil {
			log.Log.Errorw("Failed to update agent service info",
				"error", err,
				"saas_pod_id", info.SaasPodID)
			return fmt.Errorf("failed to update agent service info: %w", err)
		}

		// Batch update ServiceOperateLog
		if len(operateLogs) > 0 {
			for _, operateLog := range operateLogs {
				operateLog.Version += 1
			}
			if err := logRepo.BatchVersionSave(operateLogs, "service_operate_log"); err != nil {
				log.Log.Errorw("Failed to update service operate log",
					"error", err,
					"saas_pod_id", info.SaasPodID)
				return fmt.Errorf("failed to update service operate log: %w", err)
			}
		}

		return nil
	})
}
