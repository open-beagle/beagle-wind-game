package event

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/open-beagle/beagle-wind-game/internal/proto"
)

// EventType 定义事件类型
const (
	EventTypeContainer = "container" // 容器事件
	EventTypePipeline  = "pipeline"  // 流水线事件
	EventTypeNode      = "node"      // 节点事件
)

// EventSubscriber 事件订阅者
type EventSubscriber struct {
	types []string
	ch    chan *pb.Event
}

// EventManager 事件管理器接口
type EventManager interface {
	// Start 启动事件管理器
	Start(ctx context.Context)
	// Stop 停止事件管理器
	Stop()
	// Subscribe 订阅事件
	Subscribe(types []string) *EventSubscriber
	// Unsubscribe 取消订阅
	Unsubscribe(subscriber *EventSubscriber)
	// Publish 发布事件
	Publish(event *pb.Event)
}

// DefaultEventManager 默认事件管理器实现
type DefaultEventManager struct {
	sync.RWMutex
	subscribers map[string][]*EventSubscriber
	stopCh      chan struct{}
}

// NewDefaultEventManager 创建新的事件管理器
func NewDefaultEventManager() *DefaultEventManager {
	return &DefaultEventManager{
		subscribers: make(map[string][]*EventSubscriber),
		stopCh:      make(chan struct{}),
	}
}

// Start 启动事件管理器
func (em *DefaultEventManager) Start(ctx context.Context) {
	go em.run(ctx)
}

// Stop 停止事件管理器
func (em *DefaultEventManager) Stop() {
	close(em.stopCh)
}

// Subscribe 订阅事件
func (em *DefaultEventManager) Subscribe(types []string) *EventSubscriber {
	em.Lock()
	defer em.Unlock()

	subscriber := &EventSubscriber{
		types: types,
		ch:    make(chan *pb.Event, 100),
	}

	for _, t := range types {
		em.subscribers[t] = append(em.subscribers[t], subscriber)
	}

	return subscriber
}

// Unsubscribe 取消订阅
func (em *DefaultEventManager) Unsubscribe(subscriber *EventSubscriber) {
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
func (em *DefaultEventManager) Publish(event *pb.Event) {
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
func (em *DefaultEventManager) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-em.stopCh:
			return
		}
	}
}

// NewEvent 创建新的事件
func NewEvent(eventType, nodeID, entityID, status, message string) *pb.Event {
	return &pb.Event{
		Type:      eventType,
		Id:        nodeID,
		EntityId:  entityID,
		Status:    status,
		Message:   message,
		Timestamp: timestamppb.Now(),
		Data:      make(map[string]string),
	}
}

// NewContainerEvent 创建新的容器事件
func NewContainerEvent(nodeID, containerID, status, message string) *pb.Event {
	return NewEvent(EventTypeContainer, nodeID, containerID, status, message)
}

// NewPipelineEvent 创建新的流水线事件
func NewPipelineEvent(nodeID, pipelineID, status, message string) *pb.Event {
	return NewEvent(EventTypePipeline, nodeID, pipelineID, status, message)
}

// NewNodeEvent 创建新的节点事件
func NewNodeEvent(nodeID, status, message string) *pb.Event {
	return NewEvent(EventTypeNode, nodeID, "", status, message)
}

// EventStream 事件流
type EventStream struct {
	subscriber *EventSubscriber
}

// NewEventStream 创建新的事件流
func NewEventStream(em EventManager, types []string) *EventStream {
	return &EventStream{
		subscriber: em.Subscribe(types),
	}
}

// Recv 接收事件
func (s *EventStream) Recv() (*pb.Event, error) {
	event, ok := <-s.subscriber.ch
	if !ok {
		return nil, fmt.Errorf("event stream closed")
	}
	return event, nil
}

// Close 关闭事件流
func (s *EventStream) Close() {
	s.subscriber.ch = nil
}
