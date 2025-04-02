package service

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"synapse/common/enum"
	"synapse/worker/config"
	"synapse/worker/repository/repo"
	"synapse/worker/service"
	"synapse/worker/types"
	"testing"
	"time"
)

func TestAgentService_Register(t *testing.T) {
	// Setup test environment
	SetupTestEnv(t)

	// Create service instance
	svc := service.NewAgentService(config.DB)

	// Test case: New node registration
	t.Run("New Node Registration", func(t *testing.T) {
		req := &types.AgentRegisterRequest{
			Agent: types.Agent{
				ID:              "test-agent-1",
				Name:            "Test Agent 1",
				Location:        "test-location",
				Region:          "test-region",
				CloudType:       0,
				MetricTimestamp: fmt.Sprintf("%d", time.Now().UTC().UnixMilli()),
				Nodes: []types.Node{
					{
						ID:                        "test-node-1",
						Name:                      "Test Node 1",
						IP:                        "192.168.1.1",
						DriverVersion:             "1.0.0",
						NetworkUpload:             "1000",
						NetworkDownload:           "1000",
						DiskReadSpeed:             "1000",
						DiskWriteSpeed:            "1000",
						RandomType:                "random",
						StorageType:               "ssd",
						TotalGPUCount:             "4",
						TotalVCPU:                 "8",
						TotalVRAM:                 "16000",
						TotalRAM:                  "32000",
						TotalCPU:                  "8",
						TotalMemory:               "32000",
						TotalStorage:              "1000",
						TotalPersistentVolume:     "1000",
						AllocatedGPUCount:         "0",
						AllocatedVCPU:             "0",
						AllocatedVRAM:             "0",
						AllocatedRAM:              "0",
						AllocatedCPU:              "0",
						AllocatedMemory:           "0",
						AllocatedStorage:          "0",
						AllocatedPersistentVolume: "0",
						ResourceType:              "1",
						GPUType:                   "NVIDIA",
					},
				},
			},
		}

		// Execute registration
		err := svc.Register(context.Background(), req)
		if err != nil {
			t.Logf("Registration failed with error: %v", err)
		}
		assert.NoError(t, err)

		// Verify data is correctly written
		agentInfoRepo := repo.NewAgentInfoRepository(config.DB)
		agentInfo, exist, err := agentInfoRepo.FindByAgentID("test-agent-1")
		if err != nil {
			t.Logf("Failed to query agent info: %v", err)
		} else if !exist {
			t.Log("Agent info not found in database")
		} else {
			t.Logf("Agent info retrieved successfully: %+v", agentInfo)
		}
	})
}

func TestAgentService_CalculateAgentStatus(t *testing.T) {
	svc := &service.AgentService{}

	tests := []struct {
		name     string
		node     *types.Node
		expected int
	}{
		{
			name: "Available Node",
			node: &types.Node{
				TotalGPUCount:             "4",
				TotalVRAM:                 "16000",
				TotalRAM:                  "32000",
				TotalPersistentVolume:     "1000",
				AllocatedGPUCount:         "0",
				AllocatedVRAM:             "0",
				AllocatedRAM:              "0",
				AllocatedPersistentVolume: "0",
			},
			expected: int(enum.Resource_Status_Available),
		},
		{
			name: "Busy Node",
			node: &types.Node{
				TotalGPUCount:             "4",
				TotalVRAM:                 "16000",
				TotalRAM:                  "32000",
				TotalPersistentVolume:     "1000",
				AllocatedGPUCount:         "4",
				AllocatedVRAM:             "14400",
				AllocatedRAM:              "28800",
				AllocatedPersistentVolume: "900",
			},
			expected: int(enum.Resource_Status_Busy),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			busyThreshold := 0.9
			result := svc.CalculateNodeStatus(tt.node, busyThreshold)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAgentService_CalculateResourceMD5(t *testing.T) {
	svc := &service.AgentService{}

	agent := &types.Agent{
		ID:        "test-agent",
		Location:  "test-location",
		Region:    "test-region",
		CloudType: 0,
	}

	node := &types.Node{
		DriverVersion:         "1.0.0",
		TotalGPUCount:         "4",
		TotalVCPU:             "8",
		TotalVRAM:             "16000",
		TotalRAM:              "32000",
		TotalPersistentVolume: "1000",
		NetworkUpload:         "1000",
		NetworkDownload:       "1000",
		DiskReadSpeed:         "1000",
		RandomType:            "random",
		StorageType:           "ssd",
	}

	// Test MD5 calculation
	md5 := svc.CalculateResourceMD5(agent, node)
	assert.NotEmpty(t, md5)
	assert.Len(t, md5, 32) // MD5 hash should be 32 characters
}

func TestAgentService_ListAgentNodeResources(t *testing.T) {
	// Setup test environment
	SetupTestEnv(t)

	// Create service instance
	svc := service.NewAgentService(config.DB)

	// Register multiple test agents with different resources
	testAgents := []struct {
		agentID  string
		location string
		region   string
		gpuType  string
		gpuCount string
		vcpu     string
		vram     string
		ram      string
		disk     string
	}{
		{
			agentID:  "test-agent-1",
			location: "test-location-1",
			region:   "test-region-1",
			gpuType:  "NVIDIA",
			gpuCount: "4",
			vcpu:     "8",
			vram:     "16000",
			ram:      "32000",
			disk:     "1000",
		},
		{
			agentID:  "test-agent-2",
			location: "test-location-2",
			region:   "test-region-2",
			gpuType:  "NVIDIA",
			gpuCount: "8",
			vcpu:     "16",
			vram:     "32000",
			ram:      "64000",
			disk:     "2000",
		},
		{
			agentID:  "test-agent-3",
			location: "test-location-3",
			region:   "test-region-3",
			gpuType:  "NVIDIA",
			gpuCount: "2",
			vcpu:     "4",
			vram:     "8000",
			ram:      "16000",
			disk:     "500",
		},
	}

	// Register test agents
	for _, agent := range testAgents {
		req := &types.AgentRegisterRequest{
			Agent: types.Agent{
				ID:              agent.agentID,
				Name:            "Test Agent " + agent.agentID,
				Location:        agent.location,
				Region:          agent.region,
				CloudType:       0,
				MetricTimestamp: fmt.Sprintf("%d", time.Now().UTC().UnixMilli()),
				Nodes: []types.Node{
					{
						ID:                        "test-node-" + agent.agentID,
						Name:                      "Test Node " + agent.agentID,
						IP:                        "192.168.1.1",
						DriverVersion:             "1.0.0",
						NetworkUpload:             "1000",
						NetworkDownload:           "1000",
						DiskReadSpeed:             "1000",
						DiskWriteSpeed:            "1000",
						RandomType:                "random",
						StorageType:               "ssd",
						TotalGPUCount:             agent.gpuCount,
						TotalVCPU:                 agent.vcpu,
						TotalVRAM:                 agent.vram,
						TotalRAM:                  agent.ram,
						TotalCPU:                  agent.vcpu,
						TotalMemory:               agent.ram,
						TotalStorage:              agent.disk,
						TotalPersistentVolume:     agent.disk,
						AllocatedGPUCount:         "0",
						AllocatedVCPU:             "0",
						AllocatedVRAM:             "0",
						AllocatedRAM:              "0",
						AllocatedCPU:              "0",
						AllocatedMemory:           "0",
						AllocatedStorage:          "0",
						AllocatedPersistentVolume: "0",
						ResourceType:              "1",
						GPUType:                   agent.gpuType,
					},
				},
			},
		}

		err := svc.Register(context.Background(), req)
		assert.NoError(t, err)
	}

	// Test cases
	tests := []struct {
		name      string
		pageMark  int64
		limit     int
		sort      string
		wantCount int
		wantMore  bool
	}{
		{
			name:      "First page with default sort",
			pageMark:  0,
			limit:     2,
			sort:      "",
			wantCount: 2,
			wantMore:  true,
		},
		{
			name:      "Second page with default sort",
			pageMark:  2,
			limit:     2,
			sort:      "",
			wantCount: 2,
			wantMore:  false,
		},
		{
			name:      "No more data page",
			pageMark:  4,
			limit:     2,
			sort:      "",
			wantCount: 0,
			wantMore:  false,
		},
		{
			name:      "Descending sort by ID",
			pageMark:  0,
			limit:     3,
			sort:      "desc",
			wantCount: 3,
			wantMore:  true,
		},
		{
			name:      "Ascending sort by ID",
			pageMark:  0,
			limit:     3,
			sort:      "asc",
			wantCount: 3,
			wantMore:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.ListAgentNodeResources(context.Background(), tt.pageMark, tt.limit, enum.SortType(tt.sort))
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantCount, len(resp.Items))
			assert.Equal(t, tt.wantMore, resp.HasMore)

			// Verify items are sorted correctly by ID
			if tt.sort == "desc" {
				for i := 1; i < len(resp.Items); i++ {
					assert.GreaterOrEqual(t, resp.Items[i-1].ID, resp.Items[i].ID)
				}
			} else if tt.sort == "asc" {
				for i := 1; i < len(resp.Items); i++ {
					assert.LessOrEqual(t, resp.Items[i-1].ID, resp.Items[i].ID)
				}
			}
		})
	}
}
