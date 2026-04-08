package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/ifuryst/llm-wiki/internal/mcpserver"
)

func registerMCPRoutes(engine *gin.Engine, manager *mcpserver.Manager) {
	engine.Any("/mcp", gin.WrapH(mcpserver.StreamableHTTPHandler(manager)))
	engine.Any("/sse", gin.WrapH(mcpserver.SSEHandler(manager)))
}
