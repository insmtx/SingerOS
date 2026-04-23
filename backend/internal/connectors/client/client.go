package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/insmtx/SingerOS/backend/internal/connectors"
	eventbus "github.com/insmtx/SingerOS/backend/internal/infra/mq"
	"github.com/ygpkg/yg-go/logs"
)

const ChannelCodeValue = "client"

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin - adjust as needed for security
		return true
	},
}

// ClientMessage represents possible message types from clients
type ClientMessage struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
	ID      string                 `json:"id,omitempty"`
}

// ServerMessage represents possible message types sent to clients
type ServerMessage struct {
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	ID        string                 `json:"id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// Connection represents a single WebSocket connection
type Connection struct {
	conn     *websocket.Conn
	send     chan ServerMessage
	clientID string
}

// ClientConnector handles client WebSocket connections
type ClientConnector struct {
	publisher   eventbus.Publisher
	connections map[*Connection]bool
	broadcast   chan ServerMessage
	register    chan *Connection
	unregister  chan *Connection
	mu          sync.RWMutex
}

// MessageDestination specifies who should receive the message
type MessageDestination struct {
	ClientID string `json:"client_id"` // Specific client ID, or empty for all clients
	TaskID   string `json:"task_id"`   // Related task ID for filtering
}

// ClientMessager provides interface to send messages to clients
type ClientMessager struct {
	connector *ClientConnector
}

func NewConnector(publisher eventbus.Publisher) connectors.Connector {
	connector := &ClientConnector{
		publisher:   publisher,
		connections: make(map[*Connection]bool),
		broadcast:   make(chan ServerMessage),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
	}

	go connector.run()
	return connector
}

func (c *ClientConnector) ChannelCode() string {
	return ChannelCodeValue
}

func (c *ClientConnector) RegisterRoutes(r gin.IRouter) {
	r.GET("/ws/client", c.handleWebSocket)
	r.GET("/api/client/status", c.getClientStatus)
}

func (c *ClientConnector) handleWebSocket(ctx *gin.Context) {
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logs.ErrorContextf(ctx, "Failed to upgrade connection to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	clientID := ctx.Query("client_id")
	if clientID == "" {
		clientID = fmt.Sprintf("client_%d", time.Now().Unix())
	}

	connection := &Connection{
		conn:     conn,
		send:     make(chan ServerMessage, 256), // Buffered channel to prevent blocking
		clientID: clientID,
	}

	c.register <- connection

	// Start goroutine to read from WebSocket
	go c.readPump(connection)

	// Start goroutine to write to WebSocket
	go c.writePump(connection)
}

func (c *ClientConnector) readPump(conn *Connection) {
	defer func() {
		c.unregister <- conn
		conn.conn.Close()
	}()

	conn.conn.SetReadLimit(512 * 1024) // 512KB max message size
	conn.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.conn.SetPongHandler(func(string) error {
		conn.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := conn.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logs.Errorf("WebSocket error for client %s: %v", conn.clientID, err)
			}
			break
		}

		var clientMsg ClientMessage
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			logs.Errorf("Failed to unmarshal client message: %v", err)
			continue
		}

		// Process client message based on type
		switch clientMsg.Type {
		case "ping":
			response := ServerMessage{
				Type:      "pong",
				Payload:   map[string]interface{}{"message": "pong"},
				ID:        clientMsg.ID,
				Timestamp: time.Now(),
			}
			select {
			case conn.send <- response:
			default:
				logs.Warnf("Dropping message for client %s due to full send buffer", conn.clientID)
			}
		case "user_command":
			// Forward user command to event bus for processing by agents
			err := c.forwardUserCommand(clientMsg, conn.clientID)
			if err != nil {
				logs.Errorf("Failed to forward user command: %v", err)

				errorResponse := ServerMessage{
					Type:      "error",
					Payload:   map[string]interface{}{"error": err.Error()},
					ID:        clientMsg.ID,
					Timestamp: time.Now(),
				}
				select {
				case conn.send <- errorResponse:
				default:
					logs.Warnf("Dropping message for client %s due to full send buffer", conn.clientID)
				}
			}
		case "subscribe_to_agent":
			// Subscribe user to agent's activity
			taskID, ok := clientMsg.Payload["task_id"].(string)
			if !ok {
				logs.Warn("Missing task_id in subscribe_to_agent message")
				continue
			}
			c.subscribeToAgentActivity(taskID, conn)
		default:
			logs.Warnf("Unknown client message type: %s", clientMsg.Type)
		}
	}
}

func (c *ClientConnector) writePump(conn *Connection) {
	ticker := time.NewTicker(54 * time.Second) // Send ping every 54 seconds (slightly less than read timeout)
	defer func() {
		ticker.Stop()
		conn.conn.Close()
	}()

	for {
		select {
		case message, ok := <-conn.send:
			conn.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel closed, send close message and exit
				conn.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			messageBytes, err := json.Marshal(message)
			if err != nil {
				logs.Errorf("Failed to marshal server message: %v", err)
				return
			}
			if _, err := w.Write(messageBytes); err != nil {
				logs.Errorf("Failed to write message: %v", err)
				return
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			conn.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *ClientConnector) run() {
	for {
		select {
		case conn := <-c.register:
			c.mu.Lock()
			c.connections[conn] = true
			c.mu.Unlock()

			// Send welcome message to client
			welcomeMsg := ServerMessage{
				Type:      "welcome",
				Payload:   map[string]interface{}{"client_id": conn.clientID, "message": "Connected to SingerOS client service"},
				ID:        "",
				Timestamp: time.Now(),
			}
			select {
			case conn.send <- welcomeMsg:
			default:
				logs.Warnf("Failed to send welcome message to client %s", conn.clientID)
				close(conn.send)
			}

			logs.Infof("New client connected: %s (total: %d)", conn.clientID, len(c.connections))
		case conn := <-c.unregister:
			c.mu.Lock()
			if _, ok := c.connections[conn]; ok {
				delete(c.connections, conn)
				close(conn.send)
				logs.Infof("Client disconnected: %s (total: %d)", conn.clientID, len(c.connections))
			}
			c.mu.Unlock()
		case message := <-c.broadcast:
			c.mu.RLock()
			for conn := range c.connections {
				select {
				case conn.send <- message:
				default:
					// Remove client if send buffer is full
					logs.Debugf("Removing client %s due to send buffer overflow", conn.clientID)
					go func(connToUnreg *Connection) {
						c.unregister <- connToUnreg
					}(conn)
				}
			}
			c.mu.RUnlock()
		}
	}
}

func (c *ClientConnector) forwardUserCommand(msg ClientMessage, clientID string) error {
	// Publish user command to event bus for processing
	event := map[string]interface{}{
		"type":       "user_command",
		"client_id":  clientID,
		"payload":    msg.Payload,
		"message_id": msg.ID,
		"timestamp":  time.Now().Unix(),
	}

	// Publish to appropriate topic
	err := c.publisher.Publish(context.Background(), "user.commands", event)
	if err != nil {
		return fmt.Errorf("failed to publish user command: %w", err)
	}

	return nil
}

func (c *ClientConnector) subscribeToAgentActivity(taskID string, conn *Connection) {
	// This would normally connect to an agent's activity stream
	// For now we just log the subscription
	logs.Infof("Client %s subscribed to agent activity for task: %s", conn.clientID, taskID)
}

func (c *ClientConnector) getClientStatus(ctx *gin.Context) {
	c.mu.RLock()
	status := map[string]interface{}{
		"connected_clients": len(c.connections),
		"status":            "active",
		"timestamp":         time.Now().Unix(),
	}
	c.mu.RUnlock()

	ctx.JSON(http.StatusOK, status)
}

// GetAllClientIDs returns a list of all connected client IDs
func (c *ClientConnector) GetAllClientIDs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ids := make([]string, 0, len(c.connections))
	for conn := range c.connections {
		ids = append(ids, conn.clientID)
	}
	return ids
}

// SendMessageToClient sends a message to a specific client
func (c *ClientConnector) SendMessageToClient(clientID string, message ServerMessage) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for conn := range c.connections {
		if conn.clientID == clientID {
			select {
			case conn.send <- message:
				return true
			default:
				logs.Warnf("Failed to send message to client %s: send buffer full", clientID)
				return false
			}
		}
	}

	logs.Warnf("Client with ID %s not found", clientID)
	return false
}

// BroadcastSend sends a message to the broadcast channel (to all connected clients)
func (c *ClientConnector) BroadcastSend(message ServerMessage) {
	select {
	case c.broadcast <- message:
	default:
		logs.Warn("Broadcast message dropped due to full broadcast channel")
	}
}

// GetClientMessager returns an interface for sending messages to clients
func (c *ClientConnector) GetClientMessager() *ClientMessager {
	return &ClientMessager{connector: c}
}

// SendMessage broadcasts a message to a specific client or all clients
func (cm *ClientMessager) SendMessage(dest MessageDestination, msgType string, payload map[string]interface{}) error {
	message := ServerMessage{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	if dest.ClientID != "" {
		// Send to specific client
		success := cm.connector.SendMessageToClient(dest.ClientID, message)
		if !success {
			return fmt.Errorf("failed to send message to client %s", dest.ClientID)
		}
	} else {
		// Broadcast to all clients
		cm.connector.BroadcastSend(message)
	}

	return nil
}

// SendAgentStatusUpdate sends an agent status update to clients
func (cm *ClientMessager) SendAgentStatusUpdate(clientID, taskID, status, message string) error {
	payload := map[string]interface{}{
		"task_id":   taskID,
		"status":    status,
		"message":   message,
		"timestamp": time.Now().Unix(),
	}

	dest := MessageDestination{ClientID: clientID}
	return cm.SendMessage(dest, "agent_status_update", payload)
}

// SendAgentStepUpdate sends a step-by-step update from an agent during execution
func (cm *ClientMessager) SendAgentStepUpdate(clientID, taskID, step, details string) error {
	payload := map[string]interface{}{
		"task_id":   taskID,
		"step":      step,
		"details":   details,
		"timestamp": time.Now().Unix(),
	}

	dest := MessageDestination{ClientID: clientID}
	return cm.SendMessage(dest, "agent_step_update", payload)
}

// SendAgentResult sends the final result of an agent's work to clients
func (cm *ClientMessager) SendAgentResult(clientID, taskID, resultType, result string) error {
	payload := map[string]interface{}{
		"task_id":     taskID,
		"result_type": resultType,
		"result":      result,
		"timestamp":   time.Now().Unix(),
	}

	dest := MessageDestination{ClientID: clientID}
	return cm.SendMessage(dest, "agent_result", payload)
}

// SendLogMessage sends a detailed log message during agent execution
func (cm *ClientMessager) SendLogMessage(clientID, taskID, logLevel, message string) error {
	payload := map[string]interface{}{
		"task_id":   taskID,
		"log_level": logLevel,
		"message":   message,
		"timestamp": time.Now().Unix(),
	}

	dest := MessageDestination{ClientID: clientID}
	return cm.SendMessage(dest, "agent_log", payload)
}
