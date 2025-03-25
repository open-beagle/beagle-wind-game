package server

import (
	"testing"

	"github.com/open-beagle/beagle-wind-game/internal/agent/proto"
	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNodeManager 模拟节点管理器
type MockNodeManager struct {
	mock.Mock
}

func (m *MockNodeManager) GetNode(id string) (*models.GameNode, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameNode), args.Error(1)
}

func (m *MockNodeManager) UpdateNodeStatus(id string, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockNodeManager) UpdateNodeMetrics(id string, metrics map[string]interface{}) error {
	args := m.Called(id, metrics)
	return args.Error(0)
}

func (m *MockNodeManager) UpdateNodeResources(id string, resources map[string]interface{}) error {
	args := m.Called(id, resources)
	return args.Error(0)
}

func (m *MockNodeManager) UpdateNodeOnlineStatus(id string, online bool) error {
	args := m.Called(id, online)
	return args.Error(0)
}

func TestUpdateNodeStatus(t *testing.T) {
	mockManager := new(MockNodeManager)
	server := &AgentServer{
		nodeManager: mockManager,
	}

	tests := []struct {
		name    string
		nodeID  string
		status  string
		wantErr bool
	}{
		{
			name:    "正常更新状态",
			nodeID:  "test-node-1",
			status:  string(models.GameNodeStateReady),
			wantErr: false,
		},
		{
			name:    "节点管理器未初始化",
			nodeID:  "test-node-2",
			status:  string(models.GameNodeStateReady),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				server.nodeManager = nil
			} else {
				server.nodeManager = mockManager
				mockManager.On("UpdateNodeStatus", tt.nodeID, tt.status).Return(nil)
			}

			err := server.updateNodeStatus(tt.nodeID, tt.status)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "节点管理器未初始化")
			} else {
				assert.NoError(t, err)
				mockManager.AssertExpectations(t)
			}
		})
	}
}

func TestUpdateNodeMetrics(t *testing.T) {
	mockManager := new(MockNodeManager)
	server := &AgentServer{
		nodeManager: mockManager,
	}

	tests := []struct {
		name    string
		nodeID  string
		metrics *proto.NodeMetrics
		wantErr bool
	}{
		{
			name:   "正常更新指标",
			nodeID: "test-node-1",
			metrics: &proto.NodeMetrics{
				CpuUsage:    50.0,
				MemoryUsage: 60.0,
				DiskUsage:   70.0,
				GpuMetrics: []*proto.GpuMetrics{
					{
						Index:       0,
						Usage:       80.0,
						MemoryUsage: 85.0,
						Temperature: 75.0,
					},
				},
				NetworkMetrics: &proto.NetworkMetrics{
					RxBytesPerSec: 1000,
					TxBytesPerSec: 2000,
				},
			},
			wantErr: false,
		},
		{
			name:    "节点管理器未初始化",
			nodeID:  "test-node-2",
			metrics: &proto.NodeMetrics{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				server.nodeManager = nil
			} else {
				server.nodeManager = mockManager
				mockManager.On("UpdateNodeMetrics", tt.nodeID, mock.Anything).Return(nil)
			}

			err := server.updateNodeMetrics(tt.nodeID, tt.metrics)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "节点管理器未初始化")
			} else {
				assert.NoError(t, err)
				mockManager.AssertExpectations(t)
			}
		})
	}
}

func TestUpdateNodeResources(t *testing.T) {
	mockManager := new(MockNodeManager)
	server := &AgentServer{
		nodeManager: mockManager,
	}

	tests := []struct {
		name    string
		nodeID  string
		metrics *proto.NodeMetrics
		wantErr bool
	}{
		{
			name:   "正常更新资源",
			nodeID: "test-node-1",
			metrics: &proto.NodeMetrics{
				CpuUsage:    50.0,
				MemoryUsage: 60.0,
				DiskUsage:   70.0,
				GpuMetrics: []*proto.GpuMetrics{
					{
						Index: 0,
						Usage: 80.0,
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "节点管理器未初始化",
			nodeID:  "test-node-2",
			metrics: &proto.NodeMetrics{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				server.nodeManager = nil
			} else {
				server.nodeManager = mockManager
				mockManager.On("UpdateNodeResources", tt.nodeID, mock.Anything).Return(nil)
			}

			err := server.updateNodeResources(tt.nodeID, tt.metrics)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "节点管理器未初始化")
			} else {
				assert.NoError(t, err)
				mockManager.AssertExpectations(t)
			}
		})
	}
}

func TestHandleNodeHeartbeat(t *testing.T) {
	mockManager := new(MockNodeManager)
	server := &AgentServer{
		nodeManager: mockManager,
	}

	tests := []struct {
		name    string
		nodeID  string
		metrics *proto.NodeMetrics
		wantErr bool
	}{
		{
			name:   "正常心跳处理",
			nodeID: "test-node-1",
			metrics: &proto.NodeMetrics{
				CpuUsage:    50.0,
				MemoryUsage: 60.0,
				DiskUsage:   70.0,
			},
			wantErr: false,
		},
		{
			name:    "节点管理器未初始化",
			nodeID:  "test-node-2",
			metrics: &proto.NodeMetrics{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				server.nodeManager = nil
			} else {
				server.nodeManager = mockManager
				mockManager.On("UpdateNodeOnlineStatus", tt.nodeID, true).Return(nil)
				mockManager.On("UpdateNodeMetrics", tt.nodeID, mock.Anything).Return(nil)
				mockManager.On("UpdateNodeResources", tt.nodeID, mock.Anything).Return(nil)
			}

			err := server.handleNodeHeartbeat(tt.nodeID, tt.metrics)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "节点管理器未初始化")
			} else {
				assert.NoError(t, err)
				mockManager.AssertExpectations(t)
			}
		})
	}
}
