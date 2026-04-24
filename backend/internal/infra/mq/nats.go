// nats 提供基于 NATS JetStream 的事件总线实现
//
// 该部分实现了 mq 包中的 Publisher 和 Subscriber 接口，
// 使用 NATS JetStream 作为消息中间件来实现事件的发布和订阅。
package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/ygpkg/yg-go/logs"
)

// natsPublisher 表示一个 NATS 客户端，实现 Publisher 和 Subscriber 接口
type natsPublisher struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	closed bool
	mu     sync.Mutex
}

// NewPublisher 创建一个新的 NATS JetStream 发布者实例
func NewPublisher(url string) (*natsPublisher, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		logs.Errorf("Failed to connect to NATS: %v", err)
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		logs.Errorf("Failed to create JetStream context: %v", err)
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	publisher := &natsPublisher{
		conn:   conn,
		js:     js,
		closed: false,
	}

	logs.Infof("Successfully connected to NATS at %s with JetStream", url)
	return publisher, nil
}

// PublishWithContext 在给定上下文环境中发布消息到指定主题
func (p *natsPublisher) PublishWithContext(ctx context.Context, topic string, message any) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("NATS client is closed")
	}

	// 将消息序列化为 JSON
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 声明 Stream (如果不存在)
	streamName := streamNameFromTopic(topic)
	_, err = p.js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{topic, topic + ".*"},
		Storage:  nats.FileStorage,
	})
	if err != nil {
		return fmt.Errorf("failed to declare stream '%s': %w", streamName, err)
	}

	// 发布消息
	_, err = p.js.Publish(topic, body, nats.Context(ctx))
	if err != nil {
		return fmt.Errorf("failed to publish message to topic '%s': %w", topic, err)
	}

	return nil
}

// SubscribeWithContext 在给定上下文环境中订阅特定主题的消息
func (p *natsPublisher) SubscribeWithContext(ctx context.Context, topic string, handler func(event any)) error {
	// 声明 Stream (如果不存在)
	streamName := streamNameFromTopic(topic)
	_, err := p.js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{topic, topic + ".*"},
		Storage:  nats.FileStorage,
	})
	if err != nil {
		return fmt.Errorf("failed to declare stream '%s': %w", streamName, err)
	}

	// 创建持久化订阅
	durableName := fmt.Sprintf("%s-subscriber", topic)
	sub, err := p.js.Subscribe(topic, func(msg *nats.Msg) {
		// 解析收到的消息
		var message interface{}
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			logs.ErrorContextf(ctx, "Failed to unmarshal message for topic '%s': %v", topic, err)
			return
		}

		// 调用用户定义的处理函数
		handler(message)

		// 手动确认消息
		if err := msg.Ack(); err != nil {
			logs.ErrorContextf(ctx, "Failed to ack message for topic '%s': %v", topic, err)
		}
	},
		nats.Durable(durableName),
		nats.ManualAck(),
		nats.Context(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic '%s': %w", topic, err)
	}

	// 在 context 取消时清理订阅
	go func() {
		<-ctx.Done()
		if err := sub.Unsubscribe(); err != nil {
			logs.WarnContextf(ctx, "Failed to unsubscribe from topic '%s': %v", topic, err)
		}
		logs.InfoContextf(ctx, "Unsubscribed from topic: %s", topic)
	}()

	return nil
}

// Publish implements the eventbus.Publisher interface
func (p *natsPublisher) Publish(ctx context.Context, topic string, event any) error {
	return p.PublishWithContext(ctx, topic, event)
}

// Subscribe implements the eventbus.Subscriber interface
func (p *natsPublisher) Subscribe(ctx context.Context, topic string, handler func(event any)) error {
	return p.SubscribeWithContext(ctx, topic, handler)
}

// Close 关闭 NATS 连接并释放资源
func (p *natsPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true
	p.conn.Close()

	return nil
}

// streamNameFromTopic 根据 topic 生成 Stream 名称
func streamNameFromTopic(topic string) string {
	return fmt.Sprintf("%s_STREAM", topic)
}
