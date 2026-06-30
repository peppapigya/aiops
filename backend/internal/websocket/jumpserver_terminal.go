package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"devops-console-backend/internal/services/jumpserver"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// JumpserverTerminalHandler WebSocket SSH 终端处理器
type JumpserverTerminalHandler struct {
	sshProxy *jumpserver.SSHProxy
}

// NewJumpserverTerminalHandler 创建终端处理器
func NewJumpserverTerminalHandler(proxy *jumpserver.SSHProxy) *JumpserverTerminalHandler {
	return &JumpserverTerminalHandler{sshProxy: proxy}
}

// JSTerminalMessage WebSocket 消息格式
type JSTerminalMessage struct {
	Type string `json:"type"` // "stdin", "resize", "ping", "disconnect"
	Data string `json:"data,omitempty"`
	Rows uint16 `json:"rows,omitempty"`
	Cols uint16 `json:"cols,omitempty"`
}

// wsConn 包装 WebSocket 连接，保证写并发安全
type wsConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *wsConn) writeJSON(v interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteJSON(v)
}

// HandleWebSocket 处理 WebSocket 连接
func (h *JumpserverTerminalHandler) HandleWebSocket(c *gin.Context) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "升级WebSocket失败"})
		return
	}
	defer conn.Close()

	wc := &wsConn{conn: conn}

	// 获取参数
	sessionID := c.Query("sessionId")
	if sessionID == "" {
		sendJSError(wc, "缺少 sessionId 参数")
		return
	}

	// 获取活跃会话
	active := h.sshProxy.GetActiveSession(sessionID)
	if active == nil {
		sendJSError(wc, "会话不存在或已过期")
		return
	}

	// 发送连接成功消息
	wc.writeJSON(map[string]interface{}{
		"type":      "connected",
		"sessionId": active.SessionID,
		"hostName":  active.HostName,
		"hostIP":    active.HostIP,
	})

	// 启动 stdout 接收协程（从自己独立的 listener channel 读取，不会与 parseCommands 竞争）
		stdoutCh := active.AddStdoutListener()
		go func() {
			for data := range stdoutCh {
				// 录像
				active.Recorder.RecordOutput(data)
				// 发送到前端
				wc.writeJSON(map[string]interface{}{
					"type": "stdout",
					"data": string(data),
				})
			}
		}()

	// 读取前端消息
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg JSTerminalMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "stdin":
			active.StdinPipe.Write([]byte(msg.Data))
		case "command":
			// 前端发来的完整命令（用户按下回车时发送）
			active.RecordCommand(msg.Data)
		case "resize":
			if msg.Rows > 0 && msg.Cols > 0 {
				active.SSHSession.WindowChange(int(msg.Rows), int(msg.Cols))
				active.Recorder.Resize(int(msg.Cols), int(msg.Rows))
			}
		case "ping":
			wc.writeJSON(map[string]interface{}{"type": "pong"})
		case "disconnect":
			h.sshProxy.Disconnect(sessionID)
			return
		}
	}

	// 连接断开，清理
	h.sshProxy.Disconnect(sessionID)
}

// sendJSError 发送错误消息
func sendJSError(wc *wsConn, message string) {
	wc.writeJSON(map[string]interface{}{
		"type":  "error",
		"error": message,
	})
}