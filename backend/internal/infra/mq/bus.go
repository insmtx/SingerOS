// mq 包提供消息队列的抽象和实现
//
// 该包定义了事件发布者和订阅者的标准接口，以及基于 RabbitMQ 的实现。
// 支持多种事件总线实现，如 RabbitMQ、Redis 等。
package mq

import "context"

// Publisher 是事件发布者接口，定义了向指定主题发布事件的方法
type Publisher interface {
	// Publish 向指定主题发布事件
	Publish(ctx context.Context, topic string, event any) error
}

// Subscriber 是事件订阅者接口，定义了订阅指定主题事件的方法
type Subscriber interface {
	// Subscribe 订阅指定主题的事件，并使用提供的处理函数处理收到的事件
	Subscribe(ctx context.Context, topic string, handler func(event any)) error
}

// EventBus 组合了事件发布者和订阅者，提供完整的事件总线功能
type EventBus struct {
	publisher  Publisher  // 事件发布者
	subscriber Subscriber // 事件订阅者
}

// NewEventBus 创建一个新的事件总线实例
func NewEventBus(publisher Publisher, subscriber Subscriber) *EventBus {
	return &EventBus{
		publisher:  publisher,
		subscriber: subscriber,
	}
}
