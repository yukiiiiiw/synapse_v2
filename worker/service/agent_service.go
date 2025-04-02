package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"strings"
	"synapse/common"
	"synapse/common/enum"
	"synapse/common/log"
	"synapse/common/utils"
	"synapse/worker/config"
	"synapse/worker/repository/repo"
	entity "synapse/worker/repository/types"
	"synapse/worker/types"
	"synapse/worker/util"
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

	now := time.Now().UTC().UnixMilli()
	lastHeartBeat := int64(utils.ParseInt(req.Agent.MetricTimestamp))
	timeOut := config.Config.AgentNode.HeartbeatTimeout
	if now-lastHeartBeat > timeOut.Milliseconds() {
		log.Log.Infow("register expired, ignored:", req)
		return nil
	}

	// Get busy threshold from system config
	busyThreshold := config.Config.AgentNode.BusyThreshold
	if busyThreshold <= 0 {
		busyThreshold = 0.9 // Default threshold
	}

	agentInfoRepo := repo.NewAgentInfoRepository(svc.db.WithContext(ctx))
	nodeRepo := repo.NewAgentNodeRepository(svc.db.WithContext(ctx))
	resourceRepo := repo.NewAgentNodeResourceRepository(svc.db.WithContext(ctx))
	agentRepo := repo.NewAgentRepository(svc.db.WithContext(ctx))

	// Check and update AgentInfo
	agentInfo, exist, _ := agentInfoRepo.FindByAgentID(req.Agent.ID)
	if !exist {
		agentInfo = svc.buildAgentInfo(&req.Agent, now, lastHeartBeat)
	} else {
		agentInfo.LastHeartBeat = lastHeartBeat
		agentInfo.UpdatedAt = now
	}

	// Collect node IDs and calculate resource MD5s
	nodeIDs := make([]string, 0, len(req.Agent.Nodes))
	resourceMD5s := make([]string, 0, len(req.Agent.Nodes))
	nodeByID := make(map[string]*types.Node)
	resourceMD5ByNodeID := make(map[string]string)

	for _, node := range req.Agent.Nodes {
		nodeIDs = append(nodeIDs, node.ID)
		resourceMD5 := svc.CalculateResourceMD5(&req.Agent, &node)
		resourceMD5s = append(resourceMD5s, resourceMD5)
		nodeByID[node.ID] = &node
		resourceMD5ByNodeID[node.ID] = resourceMD5
	}

	// Batch query existing nodes and resources
	existingNodes, err := nodeRepo.FindByAgentIDAndNodeIDs(req.Agent.ID, nodeIDs)
	if err != nil {
		log.Log.Errorw("Error finding nodes", "error", err)
		return err
	}

	existingResources, err := resourceRepo.FindByResourceMD5s(resourceMD5s)
	if err != nil {
		log.Log.Errorw("Error finding resources", "error", err)
		return err
	}

	// Index existing data for quick lookup
	nodeMap := make(map[string]*entity.AgentNodeInfo)
	for _, node := range existingNodes {
		nodeMap[node.AgentNodeID] = node
	}

	resourceMap := make(map[string]*entity.AgentNodeResource)
	for _, resource := range existingResources {
		resourceMap[resource.ResourceMD5] = resource
	}

	// Group nodes by resource MD5
	nodeResourceGroups := make(map[string]*repo.NodeResourceGroup)
	busyNodeCount := 0
	totalNodeCount := len(req.Agent.Nodes)

	// Process each node
	for _, nodeID := range nodeIDs {
		node := nodeByID[nodeID]
		resourceMD5 := resourceMD5ByNodeID[nodeID]

		// Get or create resource info
		var resourceInfo *entity.AgentNodeResource
		if group, exists := nodeResourceGroups[resourceMD5]; exists {
			resourceInfo = group.Resource
		} else if existingResource, exists := resourceMap[resourceMD5]; exists {
			resourceInfo = existingResource
			resourceInfo.UpdatedAt = now
		} else {
			resourceInfo = svc.buildAgentNodeResource(&req.Agent, node, resourceMD5, now)
		}

		// Get or create node info
		var nodeInfo *entity.AgentNodeInfo
		if existingNode, exists := nodeMap[nodeID]; exists {
			nodeInfo = existingNode
			nodeInfo.LastHeartBeat = now
			nodeInfo.UpdatedAt = now
		} else {
			nodeInfo = svc.buildAgentNodeInfo(&req.Agent, node, now, lastHeartBeat, resourceInfo.ID)
		}

		// Calculate node status
		nodeInfo.NodeStatus = svc.CalculateNodeStatus(node, busyThreshold)
		if nodeInfo.NodeStatus == int(enum.Resource_Status_Busy) {
			busyNodeCount++
		}

		// Group nodes by resource MD5
		if group, ok := nodeResourceGroups[resourceMD5]; ok {
			group.Nodes = append(group.Nodes, nodeInfo)
		} else {
			nodeResourceGroups[resourceMD5] = &repo.NodeResourceGroup{
				Nodes:    []*entity.AgentNodeInfo{nodeInfo},
				Resource: resourceInfo,
			}
		}
	}

	// Set agent status based on busy node count
	if totalNodeCount > 0 && busyNodeCount == totalNodeCount {
		agentInfo.Status = int(enum.Resource_Status_Busy)
	}

	// Save or update using transaction
	if err := agentRepo.RegisterAgent(agentInfo, nodeResourceGroups); err != nil {
		log.Log.Errorw("Error registering agent: %v", err)
		return err
	}

	return nil
}

func (svc *AgentService) buildAgentInfo(agent *types.Agent, now int64, lastHeartBeat int64) *entity.AgentInfo {
	return &entity.AgentInfo{
		ID:               util.GetNextId(),
		AgentID:          agent.ID,
		Status:           int(enum.Common_Status_Available),
		CloudType:        agent.CloudType,
		MetricInfo:       agent,
		ResourceFreeInfo: make(map[string]interface{}), // TODO: Calculate node resource free info
		Location:         agent.Location,
		Region:           agent.Region,
		CreatedAt:        now,
		UpdatedAt:        now,
		LastHeartBeat:    lastHeartBeat,
		Version:          0,
	}
}

func (svc *AgentService) buildAgentNodeInfo(agent *types.Agent, node *types.Node, now int64, lastHeartbeat int64, resourceId int64) *entity.AgentNodeInfo {
	// Calculate free resources
	totalGPUCount := utils.ParseFloat(node.TotalGPUCount)
	totalVRAM := utils.ParseFloat(node.TotalVRAM)
	totalRAM := utils.ParseFloat(node.TotalRAM)
	totalPersistentVolume := utils.ParseFloat(node.TotalPersistentVolume)

	allocatedGPUCount := utils.ParseFloat(node.AllocatedGPUCount)
	allocatedVRAM := utils.ParseFloat(node.AllocatedVRAM)
	allocatedRAM := utils.ParseFloat(node.AllocatedRAM)
	allocatedPersistentVolume := utils.ParseFloat(node.AllocatedPersistentVolume)

	resourceFreeInfo := &types.ResourceFreeInfo{
		GPUCount: totalGPUCount - allocatedGPUCount,
		VRAM:     totalVRAM - allocatedVRAM,
		RAM:      totalRAM - allocatedRAM,
		Disk:     totalPersistentVolume - allocatedPersistentVolume,
	}

	return &entity.AgentNodeInfo{
		ID:               util.GetNextId(),
		NodeResourceID:   resourceId,
		AgentID:          agent.ID,
		AgentNodeID:      node.ID,
		NodeStatus:       int(enum.Common_Status_Available),
		ResourceType:     utils.ParseInt(node.ResourceType),
		MetricInfo:       node,
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
		LastHeartBeat:    lastHeartbeat,
		Version:          0,
	}
}

func (svc *AgentService) buildAgentNodeResource(agent *types.Agent, node *types.Node, resourceMD5 string, now int64) *entity.AgentNodeResource {
	return &entity.AgentNodeResource{
		ID:                  util.GetNextId(),
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
		Version:             0,
	}
}

func (svc *AgentService) ListAgentNodeResources(ctx context.Context, pageMark int64, limit int, sort string) (*types.AgentNodeResourceListResponse, error) {
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
		agent.CloudType,
	}

	// Calculate MD5
	hash := md5.New()
	hash.Write([]byte(strings.Join(fields, ":")))
	return hex.EncodeToString(hash.Sum(nil))
}

func (svc *AgentService) CalculateNodeStatus(node *types.Node, busyThreshold float64) int {

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
