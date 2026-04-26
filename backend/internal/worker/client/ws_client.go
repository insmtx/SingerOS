package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ygpkg/yg-go/logs"
)

type WSClient struct {
	conn       *websocket.Conn
	workerID   string
	serverAddr string
	send       chan map[string]interface{}
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewWSClient(serverAddr, workerID string) *WSClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &WSClient{
		workerID:   workerID,
		serverAddr: serverAddr,
		send:       make(chan map[string]interface{}, 256),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (c *WSClient) Connect(ctx context.Context) error {
	wsURL := fmt.Sprintf("ws://%s/ws/worker", c.serverAddr)
	logs.Infof("Connecting to server WebSocket: %s", wsURL)

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to dial WebSocket: %w", err)
	}

	c.conn = conn
	logs.Infof("Connected to server successfully")

	registerMsg := map[string]interface{}{
		"type": "worker_register",
		"payload": map[string]interface{}{
			"worker_id": c.workerID,
			"timestamp": time.Now().Unix(),
		},
	}

	if err := c.sendJSON(registerMsg); err != nil {
		return fmt.Errorf("failed to send registration: %w", err)
	}

	go c.readLoop(ctx)
	go c.writeLoop(ctx)

	return nil
}

func (c *WSClient) readLoop(ctx context.Context) {
	defer func() {
		logs.Info("WebSocket read loop exited")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					logs.Info("WebSocket connection closed")
					return
				}
				logs.Errorf("WebSocket read error: %v", err)
				return
			}

			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err != nil {
				logs.Errorf("Failed to unmarshal message: %v", err)
				continue
			}

			c.handleMessage(msg)
		}
	}
}

func (c *WSClient) writeLoop(ctx context.Context) {
	defer func() {
		logs.Info("WebSocket write loop exited")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.sendJSON(msg); err != nil {
				logs.Errorf("Failed to send message: %v", err)
				return
			}
		}
	}
}

func (c *WSClient) handleMessage(msg map[string]interface{}) {
	msgType, _ := msg["type"].(string)
	switch msgType {
	case "welcome":
		logs.Infof("Received welcome from server")
	case "config_update":
		logs.Infof("Received config update")
	default:
		logs.Debugf("Received message: %s", msgType)
	}
}

func (c *WSClient) sendJSON(msg map[string]interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *WSClient) Close() error {
	c.cancel()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *WSClient) IsConnected() bool {
	return c.conn != nil
}
