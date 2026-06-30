package websocket

import (
	jsRoute "devops-console-backend/internal/routes/jumpserver"

	"github.com/gin-gonic/gin"
)

// RegisterWebSocketRoutes 注册WebSocket路由
func RegisterWebSocketRoutes(r *gin.Engine) {
	ws := r.Group("/ws")

	// Pod日志WebSocket
	ws.GET("/pod/:podname/logs", NewPodLogHandler().HandleWebSocket)

	// Pod终端WebSocket
	ws.GET("/pod/:podname/exec", NewPodExecHandler().HandleWebSocket)

	ws.GET("/executions/:id/logs", HandleLogWebSocket)

	// 跳板机 Web SSH 终端
	ws.GET("/jumpserver/terminal", func(c *gin.Context) {
		proxy := jsRoute.GetSSHProxy()
		if proxy == nil {
			c.JSON(400, gin.H{"error": "SSH Proxy 未初始化"})
			return
		}
		NewJumpserverTerminalHandler(proxy).HandleWebSocket(c)
	})
}
