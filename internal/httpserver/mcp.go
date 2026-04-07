package httpserver

import (
	"github.com/bytedance/docmesh/internal/mcpserver"
	"github.com/gin-gonic/gin"
)

func registerMCPRoutes(engine *gin.Engine, manager *mcpserver.Manager) {
	engine.Any("/mcp", gin.WrapH(mcpserver.StreamableHTTPHandler(manager)))
	engine.Any("/sse", gin.WrapH(mcpserver.SSEHandler(manager)))
}
