package system

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	reqSys "devops-console-backend/internal/dal/request/system"
	"devops-console-backend/internal/dal/response"
	"devops-console-backend/pkg/utils"
	"devops-console-backend/pkg/utils/jwt"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SysUserController 用户管理控制器
type SysUserController struct {
	userMapper          *mapper.SysUserMapper
	deptMapper          *mapper.SysDeptMapper
	posMapper           *mapper.SysPositionMapper
	menuMapper          *mapper.SysMenuMapper
	moduleConfigMapper  *mapper.SysModuleConfigMapper
}

func NewSysUserController(um *mapper.SysUserMapper, dm *mapper.SysDeptMapper, pm *mapper.SysPositionMapper, mm *mapper.SysMenuMapper, mcm *mapper.SysModuleConfigMapper) *SysUserController {
	return &SysUserController{userMapper: um, deptMapper: dm, posMapper: pm, menuMapper: mm, moduleConfigMapper: mcm}
}

// List 分页查询用户
func (c *SysUserController) List(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqSys.UserPageRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	total, list, err := c.userMapper.ListPage(req.Page, req.PageSize, req.Username, req.Nickname, req.DeptID, req.Status)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	for i := range list {
		list[i].Roles, _ = c.userMapper.GetUserRoles(list[i].ID)
		list[i].Positions, _ = c.userMapper.GetUserPositions(list[i].ID)
	}
	helper.SuccessWithData("查询成功", "data", response.UserPageVO{Total: total, List: list})
}

// GetByID 查询用户详情
func (c *SysUserController) GetByID(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	u, err := c.userMapper.GetByID(id)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	roles, _ := c.userMapper.GetUserRoles(id)
	positions, _ := c.userMapper.GetUserPositions(id)
	vo := buildUserVO(u, roles, positions)
	helper.SuccessWithData("查询成功", "data", vo)
}

// Create 新建用户
func (c *SysUserController) Create(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	var req reqSys.UserCreateRequest
	if !utils.BindAndValidate(ctx, &req) {
		common.Fail(ctx, common.BadRequest)
		return
	}
	// 检查用户名唯一
	exists, _ := c.userMapper.ExistsByUsername(req.Username, 0)
	if exists {
		helper.Fail(common.NewErrorCode(400, "用户名已存在"))
		return
	}
	nickname := req.Nickname
	email := req.Email
	phone := req.Phone
	remark := req.Remark
	deptID := req.DeptID

	u := &model.SystemUser{
		Username: req.Username,
		// 前端已经 SHA256 加密，直接存储，不再二次哈希
		Password: req.Password,
		Status:   uint32(req.Status),
		Nickname: &nickname,
		Email:    &email,
		Phone:    &phone,
		Remark:   &remark,
		DeptID:   &deptID,
	}
	now := time.Now()
	u.CreatedAt = &now
	u.UpdatedAt = &now

	if err := c.userMapper.Create(u); err != nil {
		helper.InternalError(err.Error())
		return
	}
	// 分配角色和岗位
	if len(req.RoleIDs) > 0 {
		_ = c.userMapper.AssignRoles(uint64(u.ID), req.RoleIDs)
	}
	if len(req.PositionIDs) > 0 {
		_ = c.userMapper.AssignPositions(uint64(u.ID), req.PositionIDs)
	}
	helper.SuccessWithData("创建成功", "data", gin.H{"id": u.ID})
}

// Update 更新用户
func (c *SysUserController) Update(ctx *gin.Context) {

	helper := utils.NewResponseHelper(ctx)
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req reqSys.UserUpdateRequest
	if !utils.BindAndValidate(ctx, &req) {
		common.Fail(ctx, common.BadRequest)
		return
	}
	u, err := c.userMapper.GetByID(id)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	nickname := req.Nickname
	email := req.Email
	phone := req.Phone
	remark := req.Remark
	deptID := req.DeptID

	u.Nickname = &nickname
	u.Email = &email
	u.Phone = &phone
	u.Remark = &remark
	u.DeptID = &deptID
	u.Status = uint32(req.Status)

	if err := c.userMapper.Update(u); err != nil {
		helper.InternalError(err.Error())
		return
	}
	_ = c.userMapper.AssignRoles(id, req.RoleIDs)
	_ = c.userMapper.AssignPositions(id, req.PositionIDs)
	helper.SuccessWithData("更新成功", "data", nil)
}

// Delete 删除用户
func (c *SysUserController) Delete(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if err := c.userMapper.SoftDelete(id); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("删除成功", "data", nil)
}

// UpdateStatus 修改用户状态
func (c *SysUserController) UpdateStatus(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req reqSys.UserStatusRequest
	if !utils.BindAndValidate(ctx, &req) {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if err := c.userMapper.UpdateFields(id, map[string]interface{}{"status": req.Status}); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("状态更新成功", "data", nil)
}

// ResetPassword 重置密码
func (c *SysUserController) ResetPassword(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	id, err := parseUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req reqSys.UserResetPwdRequest
	if !utils.BindAndValidate(ctx, &req) {
		common.Fail(ctx, common.BadRequest)
		return
	}
	// 前端已经 SHA256 加密，直接存储，不再二次哈希
	if err := c.userMapper.UpdateFields(id, map[string]interface{}{"password": req.Password}); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("密码重置成功", "data", nil)
}

// GetCurrentUserInfo 获取当前登录用户信息（含菜单权限，用于动态路由）
func (c *SysUserController) GetCurrentUserInfo(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	claimsVal, exists := ctx.Get(common.UserInfoKey)
	if !exists {
		common.Fail(ctx, common.UNAUTHORIZED)
		return
	}
	claims, ok := claimsVal.(*jwt.Claims)
	if !ok {
		common.Fail(ctx, common.UNAUTHORIZED)
		return
	}
	userID := uint64(claims.GetUserId())
	u, err := c.userMapper.GetByID(userID)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	roleCodes, _ := c.userMapper.GetUserRoleCodes(userID)
	perms, _ := c.userMapper.GetUserPerms(userID)

	nickname := ""
	if u.Nickname != nil {
		nickname = *u.Nickname
	}
	email := ""
	if u.Email != nil {
		email = *u.Email
	}
	phone := ""
	if u.Phone != nil {
		phone = *u.Phone
	}
	avatar := ""
	if u.Avatar != nil {
		avatar = *u.Avatar
	}

	// 获取用户菜单（type=1或2）
	userMenus, _ := c.menuMapper.ListByUserID(userID)
	// 过滤掉已关闭模块的菜单
	disabledParentIDs, _ := c.moduleConfigMapper.GetDisabledMenuParentIDs()
	if len(disabledParentIDs) > 0 {
		disabledSet := make(map[uint64]bool)
		for _, id := range disabledParentIDs {
			disabledSet[id] = true
		}
		filtered := make([]model.SysMenu, 0, len(userMenus))
		for _, m := range userMenus {
			if disabledSet[m.ParentID] {
				continue
			}
			filtered = append(filtered, m)
		}
		userMenus = filtered
	}
	menuTree := mapper.BuildMenuTree(userMenus, 0)

	info := response.AuthInfoVO{
		UserID:   userID,
		Username: u.Username,
		Nickname: nickname,
		Email:    email,
		Phone:    phone,
		Avatar:   avatar,
		Roles:    roleCodes,
		Perms:    perms,
		Menus:    menuTree,
	}
	helper.SuccessWithData("查询成功", "data", info)
}

// UpdateProfile 修改个人信息
func (c *SysUserController) UpdateProfile(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	claimsVal, exists := ctx.Get(common.UserInfoKey)
	if !exists {
		common.Fail(ctx, common.UNAUTHORIZED)
		return
	}
	claims, ok := claimsVal.(*jwt.Claims)
	if !ok {
		common.Fail(ctx, common.UNAUTHORIZED)
		return
	}
	userID := uint64(claims.GetUserId())

	var req reqSys.ProfileUpdateRequest
	if !utils.BindAndValidate(ctx, &req) {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if err := c.userMapper.UpdateFields(userID, map[string]interface{}{
		"nickname": req.Nickname,
		"email":    req.Email,
		"phone":    req.Phone,
		"avatar":   req.Avatar,
	}); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("更新成功", "data", nil)
}

// UploadAvatar 上传并更新当前用户头像
func (c *SysUserController) UploadAvatar(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	claimsVal, exists := ctx.Get(common.UserInfoKey)
	if !exists {
		common.Fail(ctx, common.UNAUTHORIZED)
		return
	}
	claims, ok := claimsVal.(*jwt.Claims)
	if !ok {
		common.Fail(ctx, common.UNAUTHORIZED)
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		helper.Fail(common.NewErrorCode(http.StatusBadRequest, "请选择要上传的头像"))
		return
	}
	if err := validateAvatarFile(file); err != nil {
		helper.Fail(common.NewErrorCode(http.StatusBadRequest, err.Error()))
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	filename := fmt.Sprintf("%s%s", uuid.NewString(), ext)
	// 保存到 frontend/public/uploads/avatars/，开发和生产环境均可直接访问
	diskPath := filepath.Join("..", "frontend", "public", "uploads", "avatars", filename)
	if err := os.MkdirAll(filepath.Dir(diskPath), 0755); err != nil {
		helper.InternalError(err.Error())
		return
	}
	if err := ctx.SaveUploadedFile(file, diskPath); err != nil {
		helper.InternalError(err.Error())
		return
	}

	avatarURL := "/uploads/avatars/" + filename
	userID := uint64(claims.GetUserId())
	if err := c.userMapper.UpdateFields(userID, map[string]interface{}{"avatar": avatarURL}); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("头像上传成功", "data", gin.H{"url": avatarURL})
}

// ChangePassword 修改密码
func (c *SysUserController) ChangePassword(ctx *gin.Context) {
	helper := utils.NewResponseHelper(ctx)
	claimsVal, _ := ctx.Get(common.UserInfoKey)
	claims := claimsVal.(*jwt.Claims)
	userID := uint64(claims.GetUserId())

	var req reqSys.ChangePasswordRequest
	if !utils.BindAndValidate(ctx, &req) {
		common.Fail(ctx, common.BadRequest)
		return
	}
	u, err := c.userMapper.GetByID(userID)
	if err != nil {
		helper.InternalError(err.Error())
		return
	}
	if u.Password != req.OldPassword {
		helper.Fail(common.UserPasswordError)
		return
	}
	if err := c.userMapper.UpdateFields(userID, map[string]interface{}{"password": req.NewPassword}); err != nil {
		helper.InternalError(err.Error())
		return
	}
	helper.SuccessWithData("密码修改成功", "data", nil)
}

func validateAvatarFile(file *multipart.FileHeader) error {
	const maxAvatarSize = 2 << 20
	if file.Size > maxAvatarSize {
		return fmt.Errorf("头像大小不能超过2MB")
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowed[ext] {
		return fmt.Errorf("头像仅支持 jpg、jpeg、png、gif、webp 格式")
	}
	return nil
}

// ============ 工具函数 ============

func parseUint64Param(ctx *gin.Context, key string) (uint64, error) {
	return strconv.ParseUint(ctx.Param(key), 10, 64)
}

func buildUserVO(u *model.SystemUser, roles []response.RoleSimple, positions []response.PosSimple) response.UserVO {
	vo := response.UserVO{
		ID:        uint64(u.ID),
		Username:  u.Username,
		Status:    uint8(u.Status),
		Roles:     roles,
		Positions: positions,
		CreatedAt: u.CreatedAt,
	}
	if u.Nickname != nil {
		vo.Nickname = *u.Nickname
	}
	if u.Email != nil {
		vo.Email = *u.Email
	}
	if u.Phone != nil {
		vo.Phone = *u.Phone
	}
	if u.Avatar != nil {
		vo.Avatar = *u.Avatar
	}
	if u.DeptID != nil {
		vo.DeptID = *u.DeptID
	}
	if u.Remark != nil {
		vo.Remark = *u.Remark
	}
	return vo
}
