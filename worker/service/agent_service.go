package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"log"
	"strconv"
	"strings"
	"synapse/common"
	"synapse/common/enum"
	"synapse/common/utils"
	"synapse/worker/config"
	"synapse/worker/repository/repo"
	entity "synapse/worker/repository/types"
	"synapse/worker/types"
	"time"
)

type AgentService struct {
	db *gorm.DB
}

func NewAgentService(db *gorm.DB) *AgentService {
	return &AgentService{db: db}
}

func (svc *AgentService) Register(ctx context.Context, req *types.AgentRegisterRequest) error {
	// TODO redisLock
	agentInfoRepo := repo.NewAgentInfoRepository(svc.db.WithContext(ctx))
	nodeRepo := repo.NewAgentNodeRepository(svc.db.WithContext(ctx))
	resourceRepo := repo.NewAgentNodeResourceRepository(svc.db.WithContext(ctx))
	agentRepo := repo.NewAgentRepository(svc.db.WithContext(ctx))

	// Process each node
	for _, node := range req.Agent.Nodes {
		// Calculate resource MD5
		resourceMD5 := svc.CalculateResourceMD5(&req.Agent, &node)

		// Check and update AgentInfo
		agentInfo, exist, _ := agentInfoRepo.FindByAgentID(req.Agent.ID)
		now := time.Now().UTC().UnixMilli()
		if !exist {
			agentInfo = svc.buildAgentInfo(&req.Agent, &node, now)
		} else {
			agentInfo.AgentStatus = svc.CalculateAgentStatus(&node)
			agentInfo.LastHeartBeat = now
			agentInfo.UpdatedAt = now
		}

		// Check and update AgentNodeInfo
		nodeInfo, exist, _ := nodeRepo.FindByAgentIDAndNodeID(req.Agent.ID, node.ID)
		if !exist {
			nodeInfo = svc.buildAgentNodeInfo(&req.Agent, &node, now)
		} else {
			nodeInfo.Status = svc.CalculateAgentStatus(&node)
			nodeInfo.LastHeartBeat = now
			nodeInfo.UpdatedAt = now
		}

		// Check and update AgentNodeResource
		resourceInfo, exist, _ := resourceRepo.FindByResourceMD5(resourceMD5)

		if !exist {
			resourceInfo = svc.buildAgentNodeResource(&req.Agent, &node, resourceMD5, now)
		}

		// Save or update using transaction
		if err := agentRepo.RegisterAgent(agentInfo, nodeInfo, resourceInfo); err != nil {
			log.Printf("Error registering agent: %v", err)
			return err
		}
	}

	return nil
}

func (svc *AgentService) buildAgentInfo(agent *types.Agent, node *types.Node, now int64) *entity.AgentInfo {
	return &entity.AgentInfo{
		AgentID:          agent.ID,
		Status:           int(enum.Common_Status_Available),
		AgentStatus:      svc.CalculateAgentStatus(node),
		CloudType:        agent.CloudType,
		MetricInfo:       map[string]interface{}{},
		ResourceFreeInfo: map[string]interface{}{}, // TODO: Calculate node resource free info
		Location:         agent.Location,
		Region:           agent.Region,
		CreatedAt:        now,
		UpdatedAt:        now,
		LastHeartBeat:    now,
		Version:          1,
	}
}

func (svc *AgentService) buildAgentNodeInfo(agent *types.Agent, node *types.Node, now int64) *entity.AgentNodeInfo {
	// Calculate free resources
	totalGPUCount := utils.ParseFloat(node.TotalGPUCount)
	totalVRAM := utils.ParseFloat(node.TotalVRAM)
	totalRAM := utils.ParseFloat(node.TotalRAM)
	totalPersistentVolume := utils.ParseFloat(node.TotalPersistentVolume)

	allocatedGPUCount := utils.ParseFloat(node.AllocatedGPUCount)
	allocatedVRAM := utils.ParseFloat(node.AllocatedVRAM)
	allocatedRAM := utils.ParseFloat(node.AllocatedRAM)
	allocatedPersistentVolume := utils.ParseFloat(node.AllocatedPersistentVolume)

	resourceFreeInfo := map[string]interface{}{
		"gpu_count": totalGPUCount - allocatedGPUCount,
		"vram":      totalVRAM - allocatedVRAM,
		"ram":       totalRAM - allocatedRAM,
		"disk":      totalPersistentVolume - allocatedPersistentVolume,
	}

	return &entity.AgentNodeInfo{
		AgentID:          agent.ID,
		AgentNodeID:      node.ID,
		NodeStatus:       int(enum.Common_Status_Available),
		Status:           svc.CalculateAgentStatus(node),
		ResourceType:     utils.ParseInt(node.ResourceType),
		MetricInfo:       map[string]interface{}{},
		ResourceFreeInfo: resourceFreeInfo,
		Location:         agent.Location,
		Region:           agent.Region,
		GPUType:          node.GPUType,
		TotalGPUCount:    int(totalGPUCount),
		TotalVRAM:        int(totalVRAM),
		TotalRAM:         int(totalRAM),
		TotalVCPU:        utils.ParseInt(node.TotalVCPU),
		CreatedAt:        now,
		UpdatedAt:        now,
		LastHeartBeat:    now,
		Version:          1,
	}
}

func (svc *AgentService) buildAgentNodeResource(agent *types.Agent, node *types.Node, resourceMD5 string, now int64) *entity.AgentNodeResource {
	return &entity.AgentNodeResource{
		Status:              common.StatusSuccess,
		ResourceType:        utils.ParseInt(node.ResourceType),
		Location:            agent.Location,
		Region:              agent.Region,
		GPUType:             node.GPUType,
		DriverVersion:       node.DriverVersion,
		RandomType:          node.RandomType,
		StorageType:         node.StorageType,
		CloudType:           agent.CloudType,
		MaxGPUCount:         utils.ParseInt(node.TotalGPUCount),
		MaxVCPU:             utils.ParseInt(node.TotalVCPU),
		MaxVRAM:             utils.ParseInt(node.TotalVRAM),
		MaxRAM:              utils.ParseInt(node.TotalRAM),
		MaxPersistentVolume: utils.ParseInt(node.TotalPersistentVolume),
		NetworkUpload:       utils.ParseFloat(node.NetworkUpload),
		NetworkDownload:     utils.ParseFloat(node.NetworkDownload),
		DiskReadSpeed:       utils.ParseFloat(node.DiskReadSpeed),
		DiskWriteSpeed:      utils.ParseFloat(node.DiskWriteSpeed),
		ResourceMD5:         resourceMD5,
		CreatedAt:           now,
		UpdatedAt:           now,
		Version:             1,
	}
}

func (svc *AgentService) ListAgentNodeResources(ctx context.Context, pageMark int64, limit int, sort enum.SortType) (*types.AgentNodeResourceListResponse, error) {
	resourceRepo := repo.NewAgentNodeResourceRepository(svc.db.WithContext(ctx))

	resources, hasMore, err := resourceRepo.List(pageMark, limit, sort)
	if err != nil {
		return nil, err
	}

	var response types.AgentNodeResourceListResponse
	response.HasMore = hasMore
	response.Items = make([]types.AgentNodeResourceResponse, len(resources))

	for i, resource := range resources {
		err = copier.Copy(&response.Items[i], &resource)
		if err != nil {
			return nil, err
		}
	}

	if len(resources) > 0 {
		response.PageMark = fmt.Sprintf("%d", resources[len(resources)-1].ID)
	}

	return &response, nil
}

func (svc *AgentService) CalculateResourceMD5(agent *types.Agent, node *types.Node) string {
	// Concatenate fields in specified order
	fields := []string{
		agent.Location,
		agent.Region,
		node.ResourceType,
		node.GPUType,
		node.DriverVersion,
		node.TotalGPUCount,
		node.TotalVCPU,
		node.TotalVRAM,
		node.TotalRAM,
		node.TotalPersistentVolume,
		node.NetworkUpload,
		node.NetworkDownload,
		node.DiskReadSpeed,
		node.RandomType,
		node.StorageType,
		strconv.Itoa(agent.CloudType),
	}

	// Calculate MD5
	hash := md5.New()
	hash.Write([]byte(strings.Join(fields, ":")))
	return hex.EncodeToString(hash.Sum(nil))
}

func (svc *AgentService) CalculateAgentStatus(node *types.Node) int {
	// Get busy threshold from system config
	busyThreshold := config.Config.App.AgentNode.BusyThreshold
	if busyThreshold <= 0 {
		busyThreshold = 0.9 // Default threshold
	}

	// Calculate resource usage ratios
	totalGPUCount := utils.ParseFloat(node.TotalGPUCount)
	totalVRAM := utils.ParseFloat(node.TotalVRAM)
	totalRAM := utils.ParseFloat(node.TotalRAM)
	totalPersistentVolume := utils.ParseFloat(node.TotalPersistentVolume)

	allocatedGPUCount := utils.ParseFloat(node.AllocatedGPUCount)
	allocatedVRAM := utils.ParseFloat(node.AllocatedVRAM)
	allocatedRAM := utils.ParseFloat(node.AllocatedRAM)
	allocatedPersistentVolume := utils.ParseFloat(node.AllocatedPersistentVolume)

	// Calculate usage ratios
	gpuRatio := allocatedGPUCount / totalGPUCount
	vramRatio := allocatedVRAM / totalVRAM
	ramRatio := allocatedRAM / totalRAM
	diskRatio := allocatedPersistentVolume / totalPersistentVolume

	// Check if any resource usage exceeds the threshold
	if gpuRatio >= busyThreshold ||
		vramRatio >= busyThreshold ||
		ramRatio >= busyThreshold ||
		diskRatio >= busyThreshold {
		return int(enum.Resource_Status_Busy)
	}

	return int(enum.Resource_Status_Available)
}
