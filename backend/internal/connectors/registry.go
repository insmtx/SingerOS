package connectors

import "github.com/gin-gonic/gin"

// Registry is a registry of connectors
type Registry struct {
	connectors map[string]Connector
}

// NewRegistry creates a new connector registry
func NewRegistry() *Registry {
	return &Registry{
		connectors: make(map[string]Connector),
	}
}

// Register registers a connector
func (r *Registry) Register(c Connector) {
	r.connectors[c.ChannelCode()] = c
}

// RegisterRoutes registers routes for all connectors
func (r *Registry) RegisterRoutes(router gin.IRouter) {
	for _, c := range r.connectors {
		c.RegisterRoutes(router)
	}
}
