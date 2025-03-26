package gamenode

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
)

// GameNodeEventType 定义游戏节点事件类型
const (
	GameNodeEventTypeContainer = "container"
	GameNodeEventTypePipeline  = "pipeline"
	GameNodeEventTypeNode      = "node"
)

// GameNodeEventSubscriber 定义游戏节点事件订阅者
type GameNodeEventSubscriber struct {
	types []string
	ch    chan *pb.Event
}

// GameNodeEventManager 管理游戏节点事件系统
type GameNodeEventManager struct {
	sync.RWMutex

	subscribers map[string][]*GameNodeEventSubscriber
	stopCh      chan struct{}
}

// NewGameNodeEventManager 创建新的游戏节点事件管理器
func NewGameNodeEventManager() *GameNodeEventManager {
	return &GameNodeEventManager{
		subscribers: make(map[string][]*GameNodeEventSubscriber),
		stopCh:      make(chan struct{}),
	}
}

// Start 启动事件管理器
func (em *GameNodeEventManager) Start(ctx context.Context) {
	go em.run(ctx)
}

// Stop 停止事件管理器
func (em *GameNodeEventManager) Stop() {
	close(em.stopCh)
}

// Subscribe 订阅事件
func (em *GameNodeEventManager) Subscribe(types []string) *GameNodeEventSubscriber {
	em.Lock()
	defer em.Unlock()

	subscriber := &GameNodeEventSubscriber{
		types: types,
		ch:    make(chan *pb.Event, 100),
	}

	for _, t := range types {
		em.subscribers[t] = append(em.subscribers[t], subscriber)
	}

	return subscriber
}

// Unsubscribe 取消订阅
func (em *GameNodeEventManager) Unsubscribe(subscriber *GameNodeEventSubscriber) {
	em.Lock()
	defer em.Unlock()

	for _, t := range subscriber.types {
		subscribers := em.subscribers[t]
		for i, s := range subscribers {
			if s == subscriber {
				em.subscribers[t] = append(subscribers[:i], subscribers[i+1:]...)
				break
			}
		}
	}
	close(subscriber.ch)
}

// Publish 发布事件
func (em *GameNodeEventManager) Publish(event *pb.Event) {
	em.RLock()
	defer em.RUnlock()

	subscribers := em.subscribers[event.Type]
	for _, subscriber := range subscribers {
		select {
		case subscriber.ch <- event:
		default:
			// 如果通道已满，跳过该事件
		}
	}
}

// run 运行事件管理器
func (em *GameNodeEventManager) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-em.stopCh:
			return
		}
	}
}

// NewGameNodeEvent 创建新的游戏节点事件
func NewGameNodeEvent(eventType, nodeID, entityID, status, message string) *pb.Event {
	return &pb.Event{
		Type:      eventType,
		Source:    nodeID,
		Timestamp: timestamppb.Now(),
		Data: map[string]string{
			"entity_id": entityID,
			"status":    status,
			"message":   message,
		},
	}
}

// NewGameNodeContainerEvent 创建新的游戏节点容器事件
func NewGameNodeContainerEvent(nodeID, containerID, status, message string) *pb.Event {
	return NewGameNodeEvent(GameNodeEventTypeContainer, nodeID, containerID, status, message)
}

// NewGameNodePipelineEvent 创建新的游戏节点流水线事件
func NewGameNodePipelineEvent(nodeID, pipelineID, status, message string) *pb.Event {
	return NewGameNodeEvent(GameNodeEventTypePipeline, nodeID, pipelineID, status, message)
}

// NewGameNodeNodeEvent 创建新的游戏节点事件
func NewGameNodeNodeEvent(nodeID, status, message string) *pb.Event {
	return NewGameNodeEvent(GameNodeEventTypeNode, nodeID, "", status, message)
}

// GameNodeEventStream 游戏节点事件流
type GameNodeEventStream struct {
	subscriber *GameNodeEventSubscriber
}

// NewGameNodeEventStream 创建新的游戏节点事件流
func NewGameNodeEventStream(em *GameNodeEventManager, types []string) *GameNodeEventStream {
	return &GameNodeEventStream{
		subscriber: em.Subscribe(types),
	}
}

// Recv 接收事件
func (s *GameNodeEventStream) Recv() (*pb.Event, error) {
	event, ok := <-s.subscriber.ch
	if !ok {
		return nil, fmt.Errorf("event stream closed")
	}
	return event, nil
}

// Close 关闭事件流
func (s *GameNodeEventStream) Close() {
	s.subscriber.ch = nil
}
