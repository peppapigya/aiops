package jumpserver

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	req "devops-console-backend/internal/dal/request/jumpserver"
	"time"

	"github.com/gin-gonic/gin"
)

// PermissionController 权限管理控制器
type PermissionController struct {
	permissionMapper *mapper.JumpserverAssetPermissionMapper
	hostMapper       *mapper.AssetHostMapper
}

func NewPermissionController(pm *mapper.JumpserverAssetPermissionMapper, hm *mapper.AssetHostMapper) *PermissionController {
	return &PermissionController{permissionMapper: pm, hostMapper: hm}
}

// ListPermissions 权限规则列表
func (c *PermissionController) ListPermissions(ctx *gin.Context) {
	var req req.PermissionPageReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	total, list, err := c.permissionMapper.ListPage(req.Page, req.PageSize, req.Name, req.IsActive)
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"total": total, "list": list}})
}

// GetPermission 获取权限规则详情
func (c *PermissionController) GetPermission(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	p, err := c.permissionMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "权限规则不存在")
		return
	}
	common.Success(ctx, gin.H{"data": p})
}

// CreatePermission 创建权限规则
func (c *PermissionController) CreatePermission(ctx *gin.Context) {
	var req req.PermissionCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	now := time.Now()
	userID := getUserID(ctx)

	p := &model.JumpserverAssetPermission{
		Name:               req.Name,
		UserIDs:            req.UserIDs,
		RoleIDs:            req.RoleIDs,
		HostIDs:            req.HostIDs,
		HostGroupIDs:       req.HostGroupIDs,
		CredentialIDs:      req.CredentialIDs,
		Protocols:          req.Protocols,
		IsActive:           req.IsActive,
		MaxSessionDuration: req.MaxSessionDuration,
		NeedApproval:       req.NeedApproval,
		ApproverIDs:        req.ApproverIDs,
		CreatedBy:          &userID,
		CreatedAt:          &now,
		UpdatedAt:          &now,
	}

	if req.DateStart != "" {
		t, err := time.Parse("2006-01-02 15:04:05", req.DateStart)
		if err == nil {
			p.DateStart = &t
		}
	}
	if req.DateExpired != "" {
		t, err := time.Parse("2006-01-02 15:04:05", req.DateExpired)
		if err == nil {
			p.DateExpired = &t
		}
	}
	if req.TimeStart != "" {
		p.TimeStart = &req.TimeStart
	}
	if req.TimeEnd != "" {
		p.TimeEnd = &req.TimeEnd
	}
	if req.Remark != "" {
		p.Remark = &req.Remark
	}

	if err := c.permissionMapper.Create(p); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"id": p.ID}})
}

// UpdatePermission 更新权限规则
func (c *PermissionController) UpdatePermission(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req req.PermissionUpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	p, err := c.permissionMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "权限规则不存在")
		return
	}
	now := time.Now()
	p.Name = req.Name
	p.UserIDs = req.UserIDs
	p.RoleIDs = req.RoleIDs
	p.HostIDs = req.HostIDs
	p.HostGroupIDs = req.HostGroupIDs
	p.CredentialIDs = req.CredentialIDs
	p.Protocols = req.Protocols
	p.IsActive = req.IsActive
	p.MaxSessionDuration = req.MaxSessionDuration
	p.NeedApproval = req.NeedApproval
	p.ApproverIDs = req.ApproverIDs
	p.UpdatedAt = &now

	if req.DateStart != "" {
		t, err := time.Parse("2006-01-02 15:04:05", req.DateStart)
		if err == nil {
			p.DateStart = &t
		}
	}
	if req.DateExpired != "" {
		t, err := time.Parse("2006-01-02 15:04:05", req.DateExpired)
		if err == nil {
			p.DateExpired = &t
		}
	}
	if req.TimeStart != "" {
		p.TimeStart = &req.TimeStart
	}
	if req.TimeEnd != "" {
		p.TimeEnd = &req.TimeEnd
	}
	if req.Remark != "" {
		p.Remark = &req.Remark
	}

	if err := c.permissionMapper.Update(p); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

// DeletePermission 删除权限规则
func (c *PermissionController) DeletePermission(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if err := c.permissionMapper.SoftDelete(id); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

// CheckPermission 检查用户对主机的权限
func (c *PermissionController) CheckPermission(ctx *gin.Context) {
	hostIDStr := ctx.Query("hostId")
	hostID, err := parseUint64ParamFromStr(hostIDStr)
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	userID := getUserID(ctx)
	allowed, creds, needApproval, err := c.permissionMapper.CheckPermission(userID, hostID)
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"allowed": allowed, "credentialIds": creds, "needApproval": needApproval}})
}

func parseUint64ParamFromStr(s string) (uint64, error) {
	val := uint64(0)
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, common.BadRequest
		}
		val = val*10 + uint64(c-'0')
	}
	return val, nil
}