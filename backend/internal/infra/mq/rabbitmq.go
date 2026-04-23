// rabbitmq 提供基于 RabbitMQ 的事件总线实现
//
// 该部分实现了 mq 包中的 Publisher 和 Subscriber 接口，
// 使用 RabbitMQ 作为消息中间件来实现事件的发布和订阅。
package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
)

// rabbitmqPublisher 表示一个RabbitMQ客户端，实现原始接口
type rabbitmqPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     config.RabbitMQConfig
	closed  bool
	mu      sync.Mutex
}

// NewPublisher 创建一个新的RabbitMQ发布者的实例（与原始函数签名匹配）
func NewPublisher(cfg config.RabbitMQConfig) (*rabbitmqPublisher, *amqp.Channel, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		logs.Errorf("Failed to connect to RabbitMQ: %v", err)
		return nil, nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		logs.Errorf("Failed to open a channel: %v", err)
		return nil, nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// 设置Qos以公平分配任务
	err = channel.Qos(1, 0, false)
	if err != nil {
		channel.Close()
		conn.Close()
		logs.Errorf("Failed to set QoS: %v", err)
		return nil, nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	publisher := &rabbitmqPublisher{
		conn:    conn,
		channel: channel,
		cfg:     cfg,
		closed:  false,
	}

	logs.Infof("Successfully connected to RabbitMQ at %s", cfg.URL)
	return publisher, channel, nil
}

// PublishWithContext 在给定上下文环境中发布消息到指定主题
func (p *rabbitmqPublisher) PublishWithContext(ctx context.Context, topic string, message any) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("RabbitMQ client is closed")
	}

	// 将消息序列化为JSON
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 声明exchange为topic类型
	err = p.channel.ExchangeDeclare(
		topic,   // exchange name
		"topic", // exchange type - 支持模式匹配
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange '%s': %w", topic, err)
	}

	// 发布消息
	err = p.channel.PublishWithContext(
		ctx,
		topic, // exchange
		"",    // routing key (empty means routing to all bound queues in fanout/exchange scenario)
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message to topic '%s': %w", topic, err)
	}

	return nil
}

// SubscribeWithContext 在给定上下文环境中订阅特定主题的消息
func (p *rabbitmqPublisher) SubscribeWithContext(ctx context.Context, topic string, handler func(event any)) error {
	p.mu.Lock()

	// 声明exchange
	err := p.channel.ExchangeDeclare(
		topic,   // name
		"topic", // type - 使用topic exchange类型
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		p.mu.Unlock()
		return fmt.Errorf("failed to declare exchange '%s': %w", topic, err)
	}

	// 创建一个临时队列（每个订阅者有自己的队列）
	queue, err := p.channel.QueueDeclare(
		"",    // name - 随机名字，让RabbitMQ生成
		false, // durable - 非持久化队列
		false, // delete when unused - 不自动删除
		true,  // exclusive - 独占此频道
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		p.mu.Unlock()
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	// 绑定队列到exchange
	err = p.channel.QueueBind(
		queue.Name, // queue name
		"",         // routing key
		topic,      // exchange
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		p.mu.Unlock()
		return fmt.Errorf("failed to bind a queue '%s' to exchange '%s': %w", queue.Name, topic, err)
	}

	// 注册消费者
	msgs, err := p.channel.Consume(
		queue.Name, // queue name
		"",         // consumer name (server-generated)
		true,       // auto-ack - 自动确认
		false,      // exclusive - 不独占消费者
		false,      // no-local - 不排除本地
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		p.mu.Unlock()
		return fmt.Errorf("failed to register a consumer for queue '%s': %w", queue.Name, err)
	}

	p.mu.Unlock()

	// 启动消息处理goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				logs.InfoContextf(ctx, "Subscription context cancelled for topic: %s", topic)
				return
			case d, ok := <-msgs:
				if !ok {
					logs.WarnContextf(ctx, "Message channel closed for topic: %s", topic)
					return
				}

				// 解析收到的消息
				var message interface{}
				if err := json.Unmarshal(d.Body, &message); err != nil {
					logs.ErrorContextf(ctx, "Failed to unmarshal message for topic '%s': %v", topic, err)
					continue
				}

				// 调用用户定义的处理函数
				handler(message)
			}
		}
	}()

	return nil
}

// Publish implements the eventbus.Publisher interface
func (p *rabbitmqPublisher) Publish(ctx context.Context, topic string, event any) error {
	return p.PublishWithContext(ctx, topic, event)
}

// Subscribe implements the eventbus.Subscriber interface
func (p *rabbitmqPublisher) Subscribe(ctx context.Context, topic string, handler func(event any)) error {
	return p.SubscribeWithContext(ctx, topic, handler)
}

// Close 关闭RabbitMQ连接并释放资源
func (p *rabbitmqPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}

	return nil
}
