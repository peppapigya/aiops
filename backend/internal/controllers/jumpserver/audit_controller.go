package jumpserver

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	req "devops-console-backend/internal/dal/request/jumpserver"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditController 审计日志控制器
type AuditController struct {
	auditLogMapper *mapper.JumpserverAuditLogMapper
	riskRuleMapper *mapper.JumpserverRiskRuleMapper
}

func NewAuditController(alm *mapper.JumpserverAuditLogMapper, rrm *mapper.JumpserverRiskRuleMapper) *AuditController {
	return &AuditController{auditLogMapper: alm, riskRuleMapper: rrm}
}

// ListAuditLogs 审计日志列表
func (c *AuditController) ListAuditLogs(ctx *gin.Context) {
	var req req.AuditLogPageReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	total, list, err := c.auditLogMapper.ListPage(req.Page, req.PageSize, req.UserID, req.Action, req.ResourceType, req.DateFrom, req.DateTo)
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"total": total, "list": list}})
}

// ==================== 危险命令规则 ====================

// ListRiskRules 危险命令规则列表
func (c *AuditController) ListRiskRules(ctx *gin.Context) {
	var req req.RiskRulePageReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	total, list, err := c.riskRuleMapper.ListPage(req.Page, req.PageSize, req.Name, req.Level, req.IsActive)
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"total": total, "list": list}})
}

// GetRiskRule 获取危险命令规则详情
func (c *AuditController) GetRiskRule(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	r, err := c.riskRuleMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "规则不存在")
		return
	}
	common.Success(ctx, gin.H{"data": r})
}

// CreateRiskRule 创建危险命令规则
func (c *AuditController) CreateRiskRule(ctx *gin.Context) {
	var req req.RiskRuleCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	now := time.Now()
	r := &model.JumpserverRiskRule{
		Name:      req.Name,
		Pattern:   req.Pattern,
		Level:     req.Level,
		Action:    req.Action,
		IsActive:  req.IsActive,
		CreatedAt: &now,
		UpdatedAt: &now,
	}
	if req.Remark != "" {
		r.Remark = &req.Remark
	}
	if err := c.riskRuleMapper.Create(r); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"id": r.ID}})
}

// UpdateRiskRule 更新危险命令规则
func (c *AuditController) UpdateRiskRule(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req req.RiskRuleUpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	r, err := c.riskRuleMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "规则不存在")
		return
	}
	now := time.Now()
	r.Name = req.Name
	r.Pattern = req.Pattern
	r.Level = req.Level
	r.Action = req.Action
	r.IsActive = req.IsActive
	r.UpdatedAt = &now
	if req.Remark != "" {
		r.Remark = &req.Remark
	}
	if err := c.riskRuleMapper.Update(r); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

// DeleteRiskRule 删除危险命令规则
func (c *AuditController) DeleteRiskRule(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if err := c.riskRuleMapper.Delete(id); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}