package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/open-beagle/beagle-wind-game/internal/proto"
)

const (
	// 心跳超时时间
	heartbeatTimeout = 30 * time.Second
	// 每个节点最大 Pipeline 数量
	maxPipelinesPerNode = 2
)

// NodeSession 表示一个节点的会话
type NodeSession struct {
	ID       string
	Pipeline chan *proto.GamePipeline
	Cancel   chan string
	Timeout  chan struct{}
	Sources  []string // 正在运行的 Pipeline IDs，由 Agent 主动报告
	mu       sync.RWMutex
}

// NewNodeSession 创建一个新的节点会话
func NewNodeSession(id string) *NodeSession {
	return &NodeSession{
		ID:       id,
		Pipeline: make(chan *proto.GamePipeline, 1),
		Cancel:   make(chan string, 1),
		Timeout:  make(chan struct{}, 1),
		Sources:  make([]string, 0),
	}
}

// UpdateSources 更新 Pipeline 源列表（由 Agent 主动报告）
func (s *NodeSession) UpdateSources(sources []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Sources = append([]string{}, sources...)
}

// HasSource 检查是否存在指定的 Pipeline 源
func (s *NodeSession) HasSource(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, source := range s.Sources {
		if source == id {
			return true
		}
	}
	return false
}

// GetSources 获取所有 Pipeline 源
func (s *NodeSession) GetSources() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string{}, s.Sources...)
}

// GetSourceCount 获取当前 Pipeline 数量
func (s *NodeSession) GetSourceCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Sources)
}

// PipelineServer 表示 Pipeline 服务器
type PipelineServer struct {
	proto.UnimplementedGamePipelineGRPCServiceServer

	mu     sync.RWMutex
	nodes  map[string]*NodeSession
	logger *logrus.Logger
	stop   chan struct{}
}

// NewPipelineServer 创建一个新的 Pipeline 服务器
func NewPipelineServer(logger *logrus.Logger) *PipelineServer {
	server := &PipelineServer{
		nodes:  make(map[string]*NodeSession),
		logger: logger,
		stop:   make(chan struct{}),
	}

	// 启动事件分发器
	go server.eventDispatcher()

	return server
}

// eventDispatcher 事件分发器
func (s *PipelineServer) eventDispatcher() {
	ticker := time.NewTicker(heartbeatTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.mu.RLock()
			for _, node := range s.nodes {
				select {
				case node.Timeout <- struct{}{}:
				default:
				}
			}
			s.mu.RUnlock()
		}
	}
}

// handleSendError 处理发送错误
func (s *PipelineServer) handleSendError(node *NodeSession, err error, msg string) error {
	s.logger.Errorf("%s: %v", msg, err)
	s.removeNode(node.ID)
	return fmt.Errorf("%s: %w", msg, err)
}

// PipelineStream 处理 Pipeline 流式请求
func (s *PipelineServer) PipelineStream(stream proto.GamePipelineGRPCService_PipelineStreamServer) error {
	var node *NodeSession

	// 处理客户端请求
	for {
		req, err := stream.Recv()
		if err != nil {
			if node != nil {
				s.removeNode(node.ID)
			}
			return err
		}

		// 获取或创建节点
		nodeID := req.GetHeartbeat().NodeId
		node = s.getOrCreateNode(nodeID)

		// 处理心跳消息中的 Pipeline 状态
		if heartbeat := req.GetHeartbeat(); heartbeat != nil {
			// 更新节点的 Pipeline 状态
			node.UpdateSources(heartbeat.PipelineIds)
		}

		// 等待事件
		select {
		case reason := <-node.Cancel:
			// 处理取消命令
			resp := &proto.PipelineStreamResponse{
				Response: &proto.PipelineStreamResponse_Cancel{
					Cancel: &proto.CancelCommand{
						Reason: reason,
					},
				},
			}
			if err := stream.Send(resp); err != nil {
				return s.handleSendError(node, err, "发送取消命令失败")
			}

		case pipeline := <-node.Pipeline:
			// 处理 Pipeline 任务
			resp := &proto.PipelineStreamResponse{
				Response: &proto.PipelineStreamResponse_Pipeline{
					Pipeline: pipeline,
				},
			}
			if err := stream.Send(resp); err != nil {
				return s.handleSendError(node, err, "发送 Pipeline 任务失败")
			}

		case <-node.Timeout:
			// 发送心跳确认
			ack := &proto.PipelineStreamResponse{
				Response: &proto.PipelineStreamResponse_HeartbeatAck{
					HeartbeatAck: &proto.HeartbeatAck{
						Success: true,
					},
				},
			}
			if err := stream.Send(ack); err != nil {
				return s.handleSendError(node, err, "发送心跳确认失败")
			}
		}
	}
}

// getOrCreateNode 获取或创建节点
func (s *PipelineServer) getOrCreateNode(id string) *NodeSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	node, ok := s.nodes[id]
	if !ok {
		node = NewNodeSession(id)
		s.nodes[id] = node
	}
	return node
}

// removeNode 移除节点
func (s *PipelineServer) removeNode(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.nodes, id)
}

// SendPipeline 发送 Pipeline 任务到指定节点
func (s *PipelineServer) SendPipeline(ctx context.Context, nodeID string, pipeline *proto.GamePipeline) error {
	s.mu.RLock()
	node, ok := s.nodes[nodeID]
	s.mu.RUnlock()

	if !ok {
		return nil
	}

	// 检查节点是否达到最大 Pipeline 数量
	if node.GetSourceCount() >= maxPipelinesPerNode {
		return fmt.Errorf("节点 %s 已达到最大 Pipeline 数量限制 (%d)", nodeID, maxPipelinesPerNode)
	}

	// 发送 Pipeline 任务
	node.Pipeline <- pipeline
	return nil
}

// SendCancel 发送取消命令到指定节点
func (s *PipelineServer) SendCancel(ctx context.Context, nodeID string, reason string) error {
	s.mu.RLock()
	node, ok := s.nodes[nodeID]
	s.mu.RUnlock()

	if !ok {
		return nil
	}

	// 发送取消命令
	node.Cancel <- reason
	return nil
}

// UpdatePipelineStatus 更新 Pipeline 状态
func (s *PipelineServer) UpdatePipelineStatus(ctx context.Context, req *proto.UpdatePipelineStatusRequest) (*proto.UpdatePipelineStatusResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: 实现状态更新逻辑
	return &proto.UpdatePipelineStatusResponse{Success: true}, nil
}

// UpdateStepStatus 更新步骤状态
func (s *PipelineServer) UpdateStepStatus(ctx context.Context, req *proto.UpdateStepStatusRequest) (*proto.UpdateStepStatusResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: 实现步骤状态更新逻辑
	return &proto.UpdateStepStatusResponse{Success: true}, nil
}

// Register 注册 Pipeline 服务到 gRPC 服务器
func (s *PipelineServer) Register(server *grpc.Server) {
	proto.RegisterGamePipelineGRPCServiceServer(server, s)
}

// Stop 停止服务器
func (s *PipelineServer) Stop() {
	close(s.stop)
}
