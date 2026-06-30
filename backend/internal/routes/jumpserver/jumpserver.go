package jumpserver

import (
	jumpserverCtrl "devops-console-backend/internal/controllers/jumpserver"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/services/jumpserver"
	"devops-console-backend/pkg/configs"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterJumpserverRouters 注册跳板机路由
func RegisterJumpserverRouters(apiGroup *gin.RouterGroup, externalDB ...*gorm.DB) {
	var db *gorm.DB
	if len(externalDB) > 0 && externalDB[0] != nil {
		db = externalDB[0]
	} else {
		db = configs.NewDB()
	}

	// Mapper
	credentialMapper := mapper.NewJumpserverCredentialMapper(db)
	hostCredentialMapper := mapper.NewJumpserverHostCredentialMapper(db)
	sessionMapper := mapper.NewJumpserverSessionMapper(db)
	commandMapper := mapper.NewJumpserverCommandMapper(db)
	permissionMapper := mapper.NewJumpserverAssetPermissionMapper(db)
	approvalMapper := mapper.NewJumpserverApprovalMapper(db)
	auditLogMapper := mapper.NewJumpserverAuditLogMapper(db)
	riskRuleMapper := mapper.NewJumpserverRiskRuleMapper(db)
	platformMapper := mapper.NewJumpserverPlatformMapper(db)
	hostMapper := mapper.NewAssetHostMapper(db)

	// SSH Proxy
	sshProxy := jumpserver.NewSSHProxy(db, "./storage/replays")

	// Batch Executor
	batchExecutor := jumpserver.NewBatchExecutor(db)

	// Controllers
	credentialCtrl := jumpserverCtrl.NewCredentialController(credentialMapper)
	sessionCtrl := jumpserverCtrl.NewSessionController(sessionMapper, commandMapper, sshProxy)
	permissionCtrl := jumpserverCtrl.NewPermissionController(permissionMapper, hostMapper)
	approvalCtrl := jumpserverCtrl.NewApprovalController(approvalMapper, hostMapper, permissionMapper)
	auditCtrl := jumpserverCtrl.NewAuditController(auditLogMapper, riskRuleMapper)
	batchCtrl := jumpserverCtrl.NewBatchController(batchExecutor, hostMapper)
	platformCtrl := jumpserverCtrl.NewPlatformController(platformMapper)

	jsGroup := apiGroup.Group("/jumpserver")
	{
		// ===== 凭证管理 =====
		credGroup := jsGroup.Group("/credentials")
		{
			credGroup.GET("", credentialCtrl.ListCredentials)
			credGroup.GET("/all", credentialCtrl.ListAllCredentials)
			credGroup.GET("/:id", credentialCtrl.GetCredential)
			credGroup.POST("", credentialCtrl.CreateCredential)
			credGroup.PUT("/:id", credentialCtrl.UpdateCredential)
			credGroup.DELETE("/:id", credentialCtrl.DeleteCredential)
		}

		// ===== 主机-凭证关联 =====
		jsGroup.GET("/host-credentials/:hostId", func(c *gin.Context) {
			id, _ := jumpserverCtrl.ParseUint64Param(c, "hostId")
			list, err := hostCredentialMapper.GetByHostID(id)
			if err != nil {
				c.JSON(200, gin.H{"status": 500, "message": err.Error(), "data": nil})
				return
			}
			c.JSON(200, gin.H{"status": 200, "message": "success", "data": gin.H{"data": list}})
		})
		jsGroup.POST("/host-credentials/bind", func(c *gin.Context) {
			var req struct {
				HostID        uint64   `json:"hostId" binding:"required"`
				CredentialIDs []uint64 `json:"credentialIds" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(200, gin.H{"status": 400, "message": err.Error(), "data": nil})
				return
			}
			if err := hostCredentialMapper.BindHostCredentials(req.HostID, req.CredentialIDs); err != nil {
				c.JSON(200, gin.H{"status": 500, "message": err.Error(), "data": nil})
				return
			}
			c.JSON(200, gin.H{"status": 200, "message": "success", "data": nil})
		})

		// ===== 会话管理 =====
		sessionGroup := jsGroup.Group("/sessions")
		{
			sessionGroup.GET("", sessionCtrl.ListSessions)
			sessionGroup.GET("/stats", sessionCtrl.GetSessionStats)
			sessionGroup.GET("/:id", sessionCtrl.GetSession)
			sessionGroup.GET("/:id/commands", sessionCtrl.GetSessionCommands)
			sessionGroup.DELETE("/:id", sessionCtrl.DeleteSession)
		}

		// ===== 连接管理 =====
		jsGroup.POST("/connect", sessionCtrl.Connect)
		jsGroup.POST("/disconnect/:sessionId", sessionCtrl.Disconnect)

		// ===== 录像回放 =====
		jsGroup.GET("/recordings/:sessionId", sessionCtrl.GetSessionRecording)
		jsGroup.GET("/recordings/:sessionId/download", sessionCtrl.DownloadRecording)

		// ===== 权限管理 =====
		permGroup := jsGroup.Group("/permissions")
		{
			permGroup.GET("", permissionCtrl.ListPermissions)
			permGroup.GET("/check", permissionCtrl.CheckPermission)
			permGroup.GET("/:id", permissionCtrl.GetPermission)
			permGroup.POST("", permissionCtrl.CreatePermission)
			permGroup.PUT("/:id", permissionCtrl.UpdatePermission)
			permGroup.DELETE("/:id", permissionCtrl.DeletePermission)
		}

		// ===== 审批管理 =====
		approvalGroup := jsGroup.Group("/approvals")
		{
			approvalGroup.GET("", approvalCtrl.ListApprovals)
			approvalGroup.GET("/:id", approvalCtrl.GetApproval)
			approvalGroup.POST("", approvalCtrl.CreateApproval)
			approvalGroup.PUT("/:id/approve", approvalCtrl.Approve)
			approvalGroup.PUT("/:id/reject", approvalCtrl.Reject)
		}

		// ===== 审计日志 =====
		jsGroup.GET("/audit-logs", auditCtrl.ListAuditLogs)

		// ===== 危险命令规则 =====
		riskGroup := jsGroup.Group("/risk-rules")
		{
			riskGroup.GET("", auditCtrl.ListRiskRules)
			riskGroup.GET("/:id", auditCtrl.GetRiskRule)
			riskGroup.POST("", auditCtrl.CreateRiskRule)
			riskGroup.PUT("/:id", auditCtrl.UpdateRiskRule)
			riskGroup.DELETE("/:id", auditCtrl.DeleteRiskRule)
		}

		// ===== 批量执行 =====
		jsGroup.POST("/batch-exec", batchCtrl.BatchExec)
		jsGroup.GET("/batch-exec/:taskId", batchCtrl.GetBatchTask)

		// ===== 平台管理 =====
		jsGroup.GET("/platforms", platformCtrl.ListPlatforms)
	}

	// 返回 SSHProxy 实例，供 WebSocket 路由使用
	// (通过包级变量共享)
	SetSSHProxy(sshProxy)
}

// 包级变量用于共享 SSHProxy 实例
var globalSSHProxy *jumpserver.SSHProxy

// SetSSHProxy 设置全局 SSHProxy
func SetSSHProxy(proxy *jumpserver.SSHProxy) {
	globalSSHProxy = proxy
}

// GetSSHProxy 获取全局 SSHProxy
func GetSSHProxy() *jumpserver.SSHProxy {
	return globalSSHProxy
}

// ParseUint64Param 导出解析参数函数
func ParseUint64Param(c *gin.Context, key string) (uint64, error) {
	return jumpserverCtrl.ParseUint64Param(c, key)
}