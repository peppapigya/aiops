package system

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"

	"github.com/gin-gonic/gin"
)

// ModuleConfigController 模块配置控制器
type ModuleConfigController struct {
	moduleConfigMapper *mapper.SysModuleConfigMapper
}

func NewModuleConfigController(mcm *mapper.SysModuleConfigMapper) *ModuleConfigController {
	return &ModuleConfigController{moduleConfigMapper: mcm}
}

// List 获取所有模块配置
func (c *ModuleConfigController) List(ctx *gin.Context) {
	list, err := c.moduleConfigMapper.ListAll()
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": list})
}

// Toggle 切换模块启用状态
func (c *ModuleConfigController) Toggle(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	cfg, err := c.moduleConfigMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "模块配置不存在")
		return
	}
	// 切换状态
	newEnabled := !cfg.IsEnabled
	if err := c.moduleConfigMapper.UpdateEnabled(id, newEnabled); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	action := "已启用"
	if !newEnabled {
		action = "已停用"
	}
	common.Success(ctx, gin.H{"data": gin.H{"id": id, "isEnabled": newEnabled, "message": cfg.ModuleName + action}})
}