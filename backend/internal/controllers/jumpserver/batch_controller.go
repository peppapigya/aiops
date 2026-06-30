package jumpserver

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	req "devops-console-backend/internal/dal/request/jumpserver"
	"devops-console-backend/internal/services/jumpserver"

	"github.com/gin-gonic/gin"
)

// BatchController 批量执行控制器
type BatchController struct {
	batchExecutor *jumpserver.BatchExecutor
	hostMapper    *mapper.AssetHostMapper
}

func NewBatchController(be *jumpserver.BatchExecutor, hm *mapper.AssetHostMapper) *BatchController {
	return &BatchController{batchExecutor: be, hostMapper: hm}
}

// BatchExec 批量执行命令
func (c *BatchController) BatchExec(ctx *gin.Context) {
	var req req.BatchExecReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	userID := getUserID(ctx)
	username := getUsername(ctx)

	taskID, err := c.batchExecutor.Execute(userID, username, req.HostIDs, req.CredentialIDs, req.Command, req.Timeout)
	if err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"taskId": taskID}})
}

// GetBatchTask 获取批量执行任务状态
func (c *BatchController) GetBatchTask(ctx *gin.Context) {
	taskID := ctx.Param("taskId")
	task := c.batchExecutor.GetTask(taskID)
	if task == nil {
		common.FailWithMsg(ctx, "任务不存在")
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{
		"taskId":   task.TaskID,
		"status":   task.Status,
		"progress": task.Progress,
		"results":  task.Results,
	}})
}

// ==================== 平台管理 ====================

// PlatformController 平台管理控制器
type PlatformController struct {
	platformMapper *mapper.JumpserverPlatformMapper
}

func NewPlatformController(pm *mapper.JumpserverPlatformMapper) *PlatformController {
	return &PlatformController{platformMapper: pm}
}

// ListPlatforms 平台列表
func (c *PlatformController) ListPlatforms(ctx *gin.Context) {
	list, err := c.platformMapper.ListAll()
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": list})
}