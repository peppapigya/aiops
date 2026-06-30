package jumpserver

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	req "devops-console-backend/internal/dal/request/jumpserver"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// ApprovalController 审批管理控制器
type ApprovalController struct {
	approvalMapper   *mapper.JumpserverApprovalMapper
	hostMapper       *mapper.AssetHostMapper
	permissionMapper *mapper.JumpserverAssetPermissionMapper
}

func NewApprovalController(am *mapper.JumpserverApprovalMapper, hm *mapper.AssetHostMapper, pm *mapper.JumpserverAssetPermissionMapper) *ApprovalController {
	return &ApprovalController{approvalMapper: am, hostMapper: hm, permissionMapper: pm}
}

// ListApprovals 审批列表
func (c *ApprovalController) ListApprovals(ctx *gin.Context) {
	var req req.ApprovalPageReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	userID := getUserID(ctx)
	total, list, err := c.approvalMapper.ListPage(req.Page, req.PageSize, req.Status, req.MyApplies, userID)
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"total": total, "list": list}})
}

// GetApproval 获取审批详情
func (c *ApprovalController) GetApproval(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	a, err := c.approvalMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "审批单不存在")
		return
	}
	common.Success(ctx, gin.H{"data": a})
}

// CreateApproval 提交审批
func (c *ApprovalController) CreateApproval(ctx *gin.Context) {
	var req req.ApprovalCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	userID := getUserID(ctx)
	username := getUsername(ctx)

	// 获取主机信息
	host, err := c.hostMapper.GetByID(req.HostID)
	if err != nil {
		common.FailWithMsg(ctx, "主机不存在")
		return
	}

	now := time.Now()
	expiredAt := now.Add(time.Duration(req.Duration) * time.Second)

	approval := &model.JumpserverApproval{
		ApplicantID:   userID,
		ApplicantName: username,
		HostID:        req.HostID,
		HostName:      &host.Name,
		HostIP:        &host.IP,
		CredentialID:  &req.CredentialID,
		Duration:      req.Duration,
		Status:        "pending",
		ExpiredAt:     &expiredAt,
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}
	if req.Reason != "" {
		approval.Reason = &req.Reason
	}

	if err := c.approvalMapper.Create(approval); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"id": approval.ID}})
}

// Approve 审批通过
func (c *ApprovalController) Approve(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req req.ApprovalHandleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	a, err := c.approvalMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "审批单不存在")
		return
	}
	if a.Status != "pending" {
		common.FailWithMsg(ctx, "该审批单已处理")
		return
	}

	// 验证审批人资格
	approverID := getUserID(ctx)
	if err := c.validateApprover(a, approverID); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}

	now := time.Now()
	approverName := getUsername(ctx)
	a.Status = "approved"
	a.ApproverID = &approverID
	a.ApproverName = &approverName
	a.ApprovedAt = &now
	a.ExpiredAt = &now
	expiredAt := now.Add(time.Duration(a.Duration) * time.Second)
	a.ExpiredAt = &expiredAt
	a.UpdatedAt = &now
	if req.Remark != "" {
		a.Remark = &req.Remark
	}

	if err := c.approvalMapper.Update(a); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

// Reject 审批拒绝
func (c *ApprovalController) Reject(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req req.ApprovalHandleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	a, err := c.approvalMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "审批单不存在")
		return
	}
	if a.Status != "pending" {
		common.FailWithMsg(ctx, "该审批单已处理")
		return
	}

	// 验证审批人资格
	approverID := getUserID(ctx)
	if err := c.validateApprover(a, approverID); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}

	now := time.Now()
	approverName := getUsername(ctx)
	a.Status = "rejected"
	a.ApproverID = &approverID
	a.ApproverName = &approverName
	a.ApprovedAt = &now
	a.UpdatedAt = &now
	if req.Remark != "" {
		a.Remark = &req.Remark
	}

	if err := c.approvalMapper.Update(a); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

// validateApprover 校验审批人资格
func (c *ApprovalController) validateApprover(approval *model.JumpserverApproval, approverID uint64) error {
	// 1. 不能审批自己的申请
	if approval.ApplicantID == approverID {
		return fmt.Errorf("不能审批自己的申请")
	}

	// 2. 检查当前用户是否在权限规则的审批人列表中
	approverIDs, err := c.permissionMapper.GetApproverIDsByHost(approval.HostID)
	if err != nil {
		return fmt.Errorf("获取审批人列表失败: %w", err)
	}

	// 如果权限规则没有指定审批人，则默认只有管理员可以审批（这里简化为：未指定审批人时拒绝）
	if len(approverIDs) == 0 {
		return fmt.Errorf("未配置审批人，请联系管理员")
	}

	for _, aid := range approverIDs {
		if aid == approverID {
			return nil
		}
	}
	return fmt.Errorf("您不是该主机的指定审批人，无权审批")
}