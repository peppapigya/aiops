package jumpserver

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	req "devops-console-backend/internal/dal/request/jumpserver"
	"devops-console-backend/internal/services/jumpserver"
	"os"

	"github.com/gin-gonic/gin"
)

// SessionController 会话管理控制器
type SessionController struct {
	sessionMapper *mapper.JumpserverSessionMapper
	commandMapper *mapper.JumpserverCommandMapper
	sshProxy      *jumpserver.SSHProxy
}

func NewSessionController(sm *mapper.JumpserverSessionMapper, cm *mapper.JumpserverCommandMapper, proxy *jumpserver.SSHProxy) *SessionController {
	return &SessionController{
		sessionMapper: sm,
		commandMapper: cm,
		sshProxy:      proxy,
	}
}

// ListSessions 会话列表
func (c *SessionController) ListSessions(ctx *gin.Context) {
	var req req.SessionPageReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	total, list, err := c.sessionMapper.ListPage(req.Page, req.PageSize, req.UserID, req.HostID, req.Status, req.RiskLevel, req.DateFrom, req.DateTo, req.Keyword)
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"total": total, "list": list}})
}

// GetSession 获取会话详情
func (c *SessionController) GetSession(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	session, err := c.sessionMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "会话不存在")
		return
	}
	common.Success(ctx, gin.H{"data": session})
}

// GetSessionCommands 获取会话命令列表
func (c *SessionController) GetSessionCommands(ctx *gin.Context) {
	var req req.CommandPageReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	// 从路由参数获取 session ID
	req.SessionID = ctx.Param("id")
	total, list, err := c.commandMapper.ListPage(req.Page, req.PageSize, req.SessionID, req.IsRisky, req.Keyword)
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"total": total, "list": list}})
}

// GetSessionRecording 获取会话录像文件（用于回放）
func (c *SessionController) GetSessionRecording(ctx *gin.Context) {
	sessionID := ctx.Param("sessionId")
	path := c.sshProxy.GetRecordingPath(sessionID)
	if path == "" {
		common.FailWithMsg(ctx, "录像文件不存在")
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		common.FailWithMsg(ctx, "录像文件不存在")
		return
	}

	// 不设置 Content-Disposition: attachment，让浏览器直接返回内容
	// asciinema-player 需要能直接读取 cast 文件内容
	ctx.File(path)
}

// DownloadRecording 下载录像文件
func (c *SessionController) DownloadRecording(ctx *gin.Context) {
	sessionID := ctx.Param("sessionId")
	path := c.sshProxy.GetRecordingPath(sessionID)
	if path == "" {
		common.FailWithMsg(ctx, "录像文件不存在")
		return
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		common.FailWithMsg(ctx, "录像文件不存在")
		return
	}
	ctx.File(path)
}

// DeleteSession 删除会话
func (c *SessionController) DeleteSession(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if err := c.sessionMapper.SoftDelete(id); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

// GetSessionStats 会话统计
func (c *SessionController) GetSessionStats(ctx *gin.Context) {
	stats, err := c.sessionMapper.GetTodayStats()
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	onlineCount, _ := c.sessionMapper.GetOnlineCount()
	stats["online"] = onlineCount
	common.Success(ctx, gin.H{"data": stats})
}

// Connect 创建连接令牌
func (c *SessionController) Connect(ctx *gin.Context) {
	var req req.ConnectReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}

	// 从上下文获取用户信息
	userID := getUserID(ctx)
	username := getUsername(ctx)
	clientIP := ctx.ClientIP()

	active, err := c.sshProxy.Connect(userID, username, req.HostID, req.CredentialID, int(req.Width), int(req.Height), clientIP)
	if err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}

	common.Success(ctx, gin.H{"data": gin.H{
		"sessionId": active.SessionID,
		"token":     active.SessionID,
		"hostName":  active.HostName,
		"hostIp":    active.HostIP,
	}})
}

// Disconnect 断开会话
func (c *SessionController) Disconnect(ctx *gin.Context) {
	sessionID := ctx.Param("sessionId")
	if err := c.sshProxy.Disconnect(sessionID); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

func getUsername(ctx *gin.Context) string {
	claims, exists := ctx.Get("claims")
	if !exists {
		return "unknown"
	}
	type Claims interface {
		GetUserName() string
	}
	if c, ok := claims.(Claims); ok {
		return c.GetUserName()
	}
	return "unknown"
}