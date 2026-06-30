package jumpserver

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	req "devops-console-backend/internal/dal/request/jumpserver"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CredentialController 凭证管理控制器
type CredentialController struct {
	credentialMapper *mapper.JumpserverCredentialMapper
}

func NewCredentialController(cm *mapper.JumpserverCredentialMapper) *CredentialController {
	return &CredentialController{credentialMapper: cm}
}

// ListCredentials 凭证列表
func (c *CredentialController) ListCredentials(ctx *gin.Context) {
	var req req.CredentialPageReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	total, list, err := c.credentialMapper.ListPage(req.Page, req.PageSize, req.Name, req.Type, req.Username)
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"total": total, "list": list}})
}

// ListAllCredentials 所有凭证
func (c *CredentialController) ListAllCredentials(ctx *gin.Context) {
	list, err := c.credentialMapper.ListAll()
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": list})
}

// CreateCredential 创建凭证
func (c *CredentialController) CreateCredential(ctx *gin.Context) {
	var req req.CredentialCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	now := time.Now()

	// 获取当前用户ID
	userID := getUserID(ctx)

	cred := &model.JumpserverCredential{
		Name:      req.Name,
		Type:      req.Type,
		Username:  req.Username,
		Protocol:  req.Protocol,
		Priority:  req.Priority,
		IsGlobal:  req.IsGlobal,
		CreatedBy: &userID,
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	// 加密敏感字段
	if req.Password != "" {
		encrypted, err := mapper.EncryptPassword(req.Password)
		if err != nil {
			common.FailWithMsg(ctx, "加密密码失败: "+err.Error())
			return
		}
		cred.Password = &encrypted
	}
	if req.PrivateKey != "" {
		encrypted, err := mapper.EncryptPassword(req.PrivateKey)
		if err != nil {
			common.FailWithMsg(ctx, "加密私钥失败: "+err.Error())
			return
		}
		cred.PrivateKey = &encrypted
	}
	if req.Passphrase != "" {
		encrypted, err := mapper.EncryptPassword(req.Passphrase)
		if err != nil {
			common.FailWithMsg(ctx, "加密私钥密码失败: "+err.Error())
			return
		}
		cred.Passphrase = &encrypted
	}

	remark := req.Remark
	if remark != "" {
		cred.Remark = &remark
	}

	if err := c.credentialMapper.Create(cred); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"id": cred.ID}})
}

// UpdateCredential 更新凭证
func (c *CredentialController) UpdateCredential(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req req.CredentialUpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	cred, err := c.credentialMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "凭证不存在")
		return
	}
	now := time.Now()
	cred.Name = req.Name
	cred.Type = req.Type
	cred.Username = req.Username
	cred.Protocol = req.Protocol
	cred.Priority = req.Priority
	cred.IsGlobal = req.IsGlobal
	cred.UpdatedAt = &now

	if req.Password != "" {
		encrypted, err := mapper.EncryptPassword(req.Password)
		if err != nil {
			common.FailWithMsg(ctx, "加密密码失败")
			return
		}
		cred.Password = &encrypted
	}
	if req.PrivateKey != "" {
		encrypted, err := mapper.EncryptPassword(req.PrivateKey)
		if err != nil {
			common.FailWithMsg(ctx, "加密私钥失败")
			return
		}
		cred.PrivateKey = &encrypted
	}
	if req.Passphrase != "" {
		encrypted, err := mapper.EncryptPassword(req.Passphrase)
		if err != nil {
			common.FailWithMsg(ctx, "加密私钥密码失败")
			return
		}
		cred.Passphrase = &encrypted
	}
	remark := req.Remark
	if remark != "" {
		cred.Remark = &remark
	}

	if err := c.credentialMapper.Update(cred); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

// DeleteCredential 删除凭证
func (c *CredentialController) DeleteCredential(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if err := c.credentialMapper.SoftDelete(id); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

// GetCredential 获取凭证详情
func (c *CredentialController) GetCredential(ctx *gin.Context) {
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	cred, err := c.credentialMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "凭证不存在")
		return
	}
	common.Success(ctx, gin.H{"data": cred})
}

// ==================== 辅助函数 ====================

func ParseUint64Param(ctx *gin.Context, key string) (uint64, error) {
	return strconv.ParseUint(ctx.Param(key), 10, 64)
}

func parseUint64Param(ctx *gin.Context, key string) (uint64, error) {
	return ParseUint64Param(ctx, key)
}

func getUserID(ctx *gin.Context) uint64 {
	claims, exists := ctx.Get("claims")
	if !exists {
		return 0
	}
	// 使用类型断言获取用户ID
	type Claims interface {
		GetUserId() int64
	}
	if c, ok := claims.(Claims); ok {
		return uint64(c.GetUserId())
	}
	return 0
}