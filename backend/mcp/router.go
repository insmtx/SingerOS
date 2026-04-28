package mcp

import "github.com/gin-gonic/gin"

const routePath = "/mcp"

// RegisterRoutes mounts the SingerOS MCP streamable HTTP endpoint.
func RegisterRoutes(r gin.IRouter, srv *Server) {
	if srv == nil {
		srv = NewServer()
	}

	handlers := []gin.HandlerFunc{requireToken(), gin.WrapH(srv.Handler())}
	r.Any(routePath, handlers...)
	r.Any(routePath+"/*path", handlers...)
}
