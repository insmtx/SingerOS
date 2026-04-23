// session 包提供客户端会话管理功能
//
// 该包负责管理 WebSocket 客户端连接和消息发送，
// 是后端与前端实时通信的桥梁。
package session

import (
	"sync"

	"github.com/insmtx/SingerOS/backend/internal/connectors/client"
)

// ClientConnectorInterface defines methods needed from the client connector for messaging
type ClientConnectorInterface interface {
	GetClientMessager() *client.ClientMessager
	GetAllClientIDs() []string
	SendMessageToClient(clientID string, message client.ServerMessage) bool
	BroadcastSend(message client.ServerMessage)
}

// Manager holds a reference to the client connector for messaging
type Manager struct {
	mutex       sync.RWMutex
	connector   ClientConnectorInterface
	initialized bool
}

var defaultManager = &Manager{}

// GetDefaultManager returns the default singleton instance of client manager
func GetDefaultManager() *Manager {
	return defaultManager
}

// SetClientConnector sets the client connector instance that will be used for messaging
func (m *Manager) SetClientConnector(connector ClientConnectorInterface) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.connector = connector
	m.initialized = true
}

// IsInitialized returns true if the client manager has been properly initialized
func (m *Manager) IsInitialized() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.initialized
}

// GetMessager returns an interface for sending messages to clients, or nil if not initialized
func (m *Manager) GetMessager() *client.ClientMessager {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.initialized {
		return nil
	}

	return m.connector.GetClientMessager()
}

// SendMessage broadcasts a message to a specific client or all clients
func (m *Manager) SendMessage(clientID, msgType string, payload map[string]interface{}) error {
	messager := m.GetMessager()
	if messager == nil {
		return nil // or return an error: fmt.Errorf("client manager not initialized")
	}

	dest := client.MessageDestination{ClientID: clientID}
	return messager.SendMessage(dest, msgType, payload)
}

// SendAgentStatusUpdate sends an agent status update to clients
func (m *Manager) SendAgentStatusUpdate(clientID, taskID, status, message string) error {
	messager := m.GetMessager()
	if messager == nil {
		return nil
	}

	return messager.SendAgentStatusUpdate(clientID, taskID, status, message)
}

// SendAgentStepUpdate sends a step-by-step update from an agent during execution
func (m *Manager) SendAgentStepUpdate(clientID, taskID, step, details string) error {
	messager := m.GetMessager()
	if messager == nil {
		return nil
	}

	return messager.SendAgentStepUpdate(clientID, taskID, step, details)
}

// SendAgentResult sends the final result of an agent's work to clients
func (m *Manager) SendAgentResult(clientID, taskID, resultType, result string) error {
	messager := m.GetMessager()
	if messager == nil {
		return nil
	}

	return messager.SendAgentResult(clientID, taskID, resultType, result)
}

// SendLogMessage sends a detailed log message during agent execution
func (m *Manager) SendLogMessage(clientID, taskID, logLevel, message string) error {
	messager := m.GetMessager()
	if messager == nil {
		return nil
	}

	return messager.SendLogMessage(clientID, taskID, logLevel, message)
}

// GetConnectedClients returns a list of all connected client IDs
func (m *Manager) GetConnectedClients() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if !m.initialized {
		return []string{}
	}

	return m.connector.GetAllClientIDs()
}
