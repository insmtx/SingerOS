# Client WebSocket Connector

## Overview
The Client WebSocket Connector provides real-time communication between SingerOS and client applications. It enables bidirectional messaging for:

1. User command submission
2. Agent execution process information 
3. Result delivery
4. Detailed execution logs
5. System notifications

## Routes

### WebSocket Endpoint
```
GET /ws/client
```
- Query param: `client_id` (optional) - unique identifier for the client. If not provided, a random ID is generated.

### Status Endpoint
```
GET /api/client/status
```
Returns connection statistics:
```json
{
  "connected_clients": 2,
  "status": "active",
  "timestamp": 1234567890
}
```

## Client Message Types

Clients can send messages with the following types:

- `ping` - Heartbeat message, expects `pong` response
- `user_command` - Submit a command/task to SingerOS agents
- `subscribe_to_agent` - Subscribe to updates from a specific agent task

## Server Message Types

Clients will receive messages with the following types:

- `welcome` - Initial connection confirmation
- `pong` - Response to `ping` message
- `agent_status_update` - Status updates for agent tasks
- `agent_step_update` - Step-by-step progress during agent execution
- `agent_result` - Final result of agent processing
- `agent_log` - Detailed log messages during execution
- `system_notification` - System-wide notifications

## Using the Client Manager

From your services (agents, etc.), you can send messages to connected clients:

```go
import "github.com/insmtx/SingerOS/backend/clientmgr"

// Get the manager singleton
manager := clientmgr.GetDefaultManager()

// Check if initialized before using
if manager.IsInitialized() {
    // Send status update to specific client
    err := manager.SendAgentStatusUpdate("client123", "task456", "processing", "Working on your request...")
    
    // Send step update
    err = manager.SendAgentStepUpdate("client123", "task456", "analyzing", "Analyzing the provided data...")
    
    // Send detailed log
    err = manager.SendLogMessage("client123", "task456", "info", "Successfully parsed input data")
    
    // Send final result
    err = manager.SendAgentResult("client123", "task456", "success", "Operation completed successfully")
}
```

## Message Format

All messages follow this structure:
```json
{
  "type": "message_type",
  "payload": { /* message-specific data */ },
  "id": "optional_message_id",
  "timestamp": "2023-01-01T00:00:00Z"
}
```

## Broadcasting Messages

To send a message to all connected clients, use an empty client ID:
```go
err := manager.SendMessage("", "system_notification", map[string]interface{}{
    "message": "System maintenance in 1 hour",
    "level": "warning"
})
```