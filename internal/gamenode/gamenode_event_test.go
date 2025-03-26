package gamenode

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewEventManager 测试创建新的事件管理器
func TestNewEventManager(t *testing.T) {
	em := NewEventManager()
	assert.NotNil(t, em)
	assert.NotNil(t, em.subscribers)
	assert.NotNil(t, em.stopCh)
}

// TestEventManager_StartStop 测试事件管理器的启动和停止
func TestEventManager_StartStop(t *testing.T) {
	em := NewEventManager()
	assert.NotNil(t, em)

	// 测试启动
	ctx := context.Background()
	em.Start(ctx)

	// 测试停止
	em.Stop()
}

// TestEventManager_SubscribeUnsubscribe 测试事件订阅和取消订阅
func TestEventManager_SubscribeUnsubscribe(t *testing.T) {
	em := NewEventManager()
	assert.NotNil(t, em)

	// 订阅事件
	types := []string{EventTypeNode, EventTypePipeline}
	subscriber := em.Subscribe(types)
	assert.NotNil(t, subscriber)
	assert.Equal(t, types, subscriber.types)
	assert.NotNil(t, subscriber.ch)

	// 验证订阅是否成功
	em.RLock()
	for _, eventType := range types {
		subscribers := em.subscribers[eventType]
		assert.Contains(t, subscribers, subscriber)
	}
	em.RUnlock()

	// 取消订阅
	em.Unsubscribe(subscriber)

	// 验证是否取消成功
	em.RLock()
	for _, eventType := range types {
		subscribers := em.subscribers[eventType]
		assert.NotContains(t, subscribers, subscriber)
	}
	em.RUnlock()

	// 验证通道是否已关闭
	_, ok := <-subscriber.ch
	assert.False(t, ok)
}

// TestEventManager_Publish 测试事件发布
func TestEventManager_Publish(t *testing.T) {
	em := NewEventManager()
	assert.NotNil(t, em)

	// 订阅事件
	subscriber := em.Subscribe([]string{EventTypeNode})
	assert.NotNil(t, subscriber)

	// 发布事件
	event := NewNodeEvent("test-node", "started", "节点已启动")
	em.Publish(event)

	// 验证事件接收
	select {
	case receivedEvent := <-subscriber.ch:
		assert.Equal(t, event, receivedEvent)
	case <-time.After(time.Second):
		t.Fatal("未能接收到事件")
	}

	// 取消订阅
	em.Unsubscribe(subscriber)
}

// TestEventManager_PublishMultipleSubscribers 测试多个订阅者的事件发布
func TestEventManager_PublishMultipleSubscribers(t *testing.T) {
	em := NewEventManager()
	assert.NotNil(t, em)

	// 创建多个订阅者
	subscriber1 := em.Subscribe([]string{EventTypeNode})
	subscriber2 := em.Subscribe([]string{EventTypeNode})
	assert.NotNil(t, subscriber1)
	assert.NotNil(t, subscriber2)

	// 发布事件
	event := NewNodeEvent("test-node", "started", "节点已启动")
	em.Publish(event)

	// 验证所有订阅者都收到事件
	subscribers := []*EventSubscriber{subscriber1, subscriber2}
	for _, subscriber := range subscribers {
		select {
		case receivedEvent := <-subscriber.ch:
			assert.Equal(t, event, receivedEvent)
		case <-time.After(time.Second):
			t.Fatal("未能接收到事件")
		}
	}

	// 取消所有订阅
	for _, subscriber := range subscribers {
		em.Unsubscribe(subscriber)
	}
}

// TestEventManager_PublishDifferentTypes 测试不同类型事件的发布
func TestEventManager_PublishDifferentTypes(t *testing.T) {
	em := NewEventManager()
	assert.NotNil(t, em)

	// 订阅不同类型的事件
	nodeSubscriber := em.Subscribe([]string{EventTypeNode})
	pipelineSubscriber := em.Subscribe([]string{EventTypePipeline})
	allSubscriber := em.Subscribe([]string{EventTypeNode, EventTypePipeline})

	// 发布节点事件
	nodeEvent := NewNodeEvent("test-node", "started", "节点已启动")
	em.Publish(nodeEvent)

	// 验证节点事件接收
	select {
	case event := <-nodeSubscriber.ch:
		assert.Equal(t, nodeEvent, event)
	case <-time.After(time.Second):
		t.Fatal("节点订阅者未能接收到节点事件")
	}

	select {
	case event := <-allSubscriber.ch:
		assert.Equal(t, nodeEvent, event)
	case <-time.After(time.Second):
		t.Fatal("全部订阅者未能接收到节点事件")
	}

	// 验证管道订阅者不会收到节点事件
	select {
	case <-pipelineSubscriber.ch:
		t.Fatal("管道订阅者不应该收到节点事件")
	case <-time.After(100 * time.Millisecond):
		// 正常情况
	}

	// 发布管道事件
	pipelineEvent := NewPipelineEvent("test-node", "test-pipeline", "started", "管道已启动")
	em.Publish(pipelineEvent)

	// 验证管道事件接收
	select {
	case event := <-pipelineSubscriber.ch:
		assert.Equal(t, pipelineEvent, event)
	case <-time.After(time.Second):
		t.Fatal("管道订阅者未能接收到管道事件")
	}

	select {
	case event := <-allSubscriber.ch:
		assert.Equal(t, pipelineEvent, event)
	case <-time.After(time.Second):
		t.Fatal("全部订阅者未能接收到管道事件")
	}

	// 验证节点订阅者不会收到管道事件
	select {
	case <-nodeSubscriber.ch:
		t.Fatal("节点订阅者不应该收到管道事件")
	case <-time.After(100 * time.Millisecond):
		// 正常情况
	}

	// 取消所有订阅
	em.Unsubscribe(nodeSubscriber)
	em.Unsubscribe(pipelineSubscriber)
	em.Unsubscribe(allSubscriber)
}
