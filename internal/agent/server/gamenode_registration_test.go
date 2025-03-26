package server

import (
	"context"
	"testing"

	"github.com/open-beagle/beagle-wind-game/internal/agent/proto"
	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	mockManager := new(MockGameNodeManager)
	server := NewAgentServer(DefaultServerOptions, mockManager)

	tests := []struct {
		name    string
		req     *proto.RegisterRequest
		wantErr bool
	}{
		{
			name: "正常注册",
			req: &proto.RegisterRequest{
				NodeId: "test-node-1",
				NodeInfo: &proto.NodeInfo{
					Hostname: "test-host",
					Ip:       "192.168.1.1",
				},
			},
			wantErr: false,
		},
		{
			name: "节点ID为空",
			req: &proto.RegisterRequest{
				NodeId: "",
				NodeInfo: &proto.NodeInfo{
					Hostname: "test-host",
					Ip:       "192.168.1.1",
				},
			},
			wantErr: true,
		},
		{
			name: "节点管理器未初始化",
			req: &proto.RegisterRequest{
				NodeId: "test-node-2",
				NodeInfo: &proto.NodeInfo{
					Hostname: "test-host",
					Ip:       "192.168.1.1",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				server.nodeManager = nil
			} else {
				server.nodeManager = mockManager
				mockManager.On("GetNode", tt.req.NodeId).Return(&models.GameNode{
					ID:   tt.req.NodeId,
					Name: tt.req.NodeInfo.Hostname,
				}, nil)
				mockManager.On("UpdateNodeStatus", tt.req.NodeId, string(models.GameNodeStateReady)).Return(nil)
			}

			resp, err := server.Register(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.req.NodeId == "" {
					assert.Contains(t, resp.Message, "节点ID不能为空")
				} else {
					assert.Contains(t, err.Error(), "节点管理器未初始化")
				}
			} else {
				assert.NoError(t, err)
				assert.True(t, resp.Success)
				assert.NotEmpty(t, resp.SessionId)
				mockManager.AssertExpectations(t)
			}
		})
	}
}

func TestHandleNodeRegistration(t *testing.T) {
	mockManager := new(MockGameNodeManager)
	server := NewAgentServer(DefaultServerOptions, mockManager)

	tests := []struct {
		name    string
		nodeID  string
		info    *proto.NodeInfo
		wantErr bool
	}{
		{
			name:   "正常注册处理",
			nodeID: "test-node-1",
			info: &proto.NodeInfo{
				Hostname: "test-host",
				Ip:       "192.168.1.1",
			},
			wantErr: false,
		},
		{
			name:    "节点管理器未初始化",
			nodeID:  "test-node-2",
			info:    &proto.NodeInfo{},
			wantErr: true,
		},
		{
			name:    "节点不存在",
			nodeID:  "test-node-3",
			info:    &proto.NodeInfo{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.name == "节点管理器未初始化" {
					server.nodeManager = nil
				} else {
					server.nodeManager = mockManager
					mockManager.On("GetNode", tt.nodeID).Return(nil, assert.AnError)
				}
			} else {
				server.nodeManager = mockManager
				mockManager.On("GetNode", tt.nodeID).Return(&models.GameNode{
					ID:   tt.nodeID,
					Name: tt.info.Hostname,
				}, nil)
				mockManager.On("UpdateNodeStatus", tt.nodeID, string(models.GameNodeStateReady)).Return(nil)
			}

			err := server.handleNodeRegistration(tt.nodeID, tt.info)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.name == "节点不存在" {
					assert.Contains(t, err.Error(), "获取节点信息失败")
				} else if tt.name == "节点管理器未初始化" {
					assert.Contains(t, err.Error(), "节点管理器未初始化")
				}
			} else {
				assert.NoError(t, err)
				mockManager.AssertExpectations(t)
			}
		})
	}
}

func TestHandleNodeDisconnection(t *testing.T) {
	mockManager := new(MockGameNodeManager)
	server := NewAgentServer(DefaultServerOptions, mockManager)

	tests := []struct {
		name    string
		nodeID  string
		wantErr bool
	}{
		{
			name:    "正常断开连接",
			nodeID:  "test-node-1",
			wantErr: false,
		},
		{
			name:    "节点管理器未初始化",
			nodeID:  "test-node-2",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				server.nodeManager = nil
			} else {
				server.nodeManager = mockManager
				mockManager.On("UpdateNodeOnlineStatus", tt.nodeID, false).Return(nil)
			}

			err := server.handleNodeDisconnection(tt.nodeID)
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
