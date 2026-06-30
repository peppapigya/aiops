package mapper

import (
	"devops-console-backend/internal/dal/model"
	"devops-console-backend/pkg/configs"
	"devops-console-backend/pkg/utils/aes"
	"fmt"

	"gorm.io/gorm"
)

// ==================== JumpserverCredentialMapper ====================

type JumpserverCredentialMapper struct {
	DB *gorm.DB
}

func NewJumpserverCredentialMapper(db *gorm.DB) *JumpserverCredentialMapper {
	return &JumpserverCredentialMapper{DB: db}
}

func (m *JumpserverCredentialMapper) ListPage(page, pageSize int, name, credType, username string) (int64, []model.JumpserverCredential, error) {
	var total int64
	var list []model.JumpserverCredential
	db := m.DB.Model(&model.JumpserverCredential{}).Where("deleted_at IS NULL")
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if credType != "" {
		db = db.Where("type = ?", credType)
	}
	if username != "" {
		db = db.Where("username LIKE ?", "%"+username+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	return total, list, err
}

func (m *JumpserverCredentialMapper) ListAll() ([]model.JumpserverCredential, error) {
	var list []model.JumpserverCredential
	err := m.DB.Where("deleted_at IS NULL").Order("id ASC").Find(&list).Error
	return list, err
}

func (m *JumpserverCredentialMapper) GetByID(id uint64) (*model.JumpserverCredential, error) {
	var c model.JumpserverCredential
	err := m.DB.Where("id = ? AND deleted_at IS NULL", id).First(&c).Error
	return &c, err
}

func (m *JumpserverCredentialMapper) Create(c *model.JumpserverCredential) error {
	return m.DB.Create(c).Error
}

func (m *JumpserverCredentialMapper) Update(c *model.JumpserverCredential) error {
	return m.DB.Save(c).Error
}

func (m *JumpserverCredentialMapper) SoftDelete(id uint64) error {
	return m.DB.Where("id = ?", id).Delete(&model.JumpserverCredential{}).Error
}

// DecryptPassword 解密凭证密码
func (m *JumpserverCredentialMapper) DecryptPassword(id uint64) (string, error) {
	c, err := m.GetByID(id)
	if err != nil {
		return "", err
	}
	if c.Password == nil || *c.Password == "" {
		return "", nil
	}
	return decrypt(*c.Password)
}

// DecryptPrivateKey 解密凭证私钥
func (m *JumpserverCredentialMapper) DecryptPrivateKey(id uint64) (string, error) {
	c, err := m.GetByID(id)
	if err != nil {
		return "", err
	}
	if c.PrivateKey == nil || *c.PrivateKey == "" {
		return "", nil
	}
	return decrypt(*c.PrivateKey)
}

func decrypt(encrypted string) (string, error) {
	key, err := configs.GetEncryptionKey()
	if err != nil {
		return "", err
	}
	if key == nil {
		return encrypted, nil
	}
	return aes.AESDecrypt(key, encrypted)
}

// ==================== JumpserverHostCredentialMapper ====================

type JumpserverHostCredentialMapper struct {
	DB *gorm.DB
}

func NewJumpserverHostCredentialMapper(db *gorm.DB) *JumpserverHostCredentialMapper {
	return &JumpserverHostCredentialMapper{DB: db}
}

func (m *JumpserverHostCredentialMapper) GetByHostID(hostID uint64) ([]model.JumpserverHostCredential, error) {
	var list []model.JumpserverHostCredential
	err := m.DB.Where("host_id = ?", hostID).Order("priority ASC").Find(&list).Error
	return list, err
}

func (m *JumpserverHostCredentialMapper) BindHostCredentials(hostID uint64, credentialIDs []uint64) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		// 删除旧关联
		if err := tx.Where("host_id = ?", hostID).Delete(&model.JumpserverHostCredential{}).Error; err != nil {
			return err
		}
		// 创建新关联
		for i, cid := range credentialIDs {
			hc := &model.JumpserverHostCredential{
				HostID:       hostID,
				CredentialID: cid,
				Priority:     i,
			}
			if err := tx.Create(hc).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ==================== JumpserverSessionMapper ====================

type JumpserverSessionMapper struct {
	DB *gorm.DB
}

func NewJumpserverSessionMapper(db *gorm.DB) *JumpserverSessionMapper {
	return &JumpserverSessionMapper{DB: db}
}

func (m *JumpserverSessionMapper) ListPage(page, pageSize int, userID, hostID uint64, status, riskLevel, dateFrom, dateTo, keyword string) (int64, []model.JumpserverSession, error) {
	var total int64
	var list []model.JumpserverSession
	db := m.DB.Model(&model.JumpserverSession{})
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if hostID > 0 {
		db = db.Where("host_id = ?", hostID)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if riskLevel != "" {
		db = db.Where("risk_level = ?", riskLevel)
	}
	if dateFrom != "" {
		db = db.Where("started_at >= ?", dateFrom)
	}
	if dateTo != "" {
		db = db.Where("started_at <= ?", dateTo)
	}
	if keyword != "" {
		db = db.Where("host_name LIKE ? OR host_ip LIKE ? OR username LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("started_at DESC").Find(&list).Error
	return total, list, err
}

func (m *JumpserverSessionMapper) GetBySessionID(sessionID string) (*model.JumpserverSession, error) {
	var s model.JumpserverSession
	err := m.DB.Where("session_id = ?", sessionID).First(&s).Error
	return &s, err
}

func (m *JumpserverSessionMapper) GetByID(id uint64) (*model.JumpserverSession, error) {
	var s model.JumpserverSession
	err := m.DB.Where("id = ?", id).First(&s).Error
	return &s, err
}

func (m *JumpserverSessionMapper) Create(s *model.JumpserverSession) error {
	return m.DB.Create(s).Error
}

func (m *JumpserverSessionMapper) Update(s *model.JumpserverSession) error {
	return m.DB.Save(s).Error
}

func (m *JumpserverSessionMapper) SoftDelete(id uint64) error {
	return m.DB.Where("id = ?", id).Delete(&model.JumpserverSession{}).Error
}

func (m *JumpserverSessionMapper) GetActiveSessions() ([]model.JumpserverSession, error) {
	var list []model.JumpserverSession
	err := m.DB.Where("status = 'active'").Find(&list).Error
	return list, err
}

func (m *JumpserverSessionMapper) GetOnlineCount() (int64, error) {
	var count int64
	err := m.DB.Model(&model.JumpserverSession{}).Where("status = 'active'").Count(&count).Error
	return count, err
}

func (m *JumpserverSessionMapper) GetTodayStats() (map[string]interface{}, error) {
	stats := map[string]interface{}{"total": int64(0), "active": int64(0), "avgDuration": int64(0)}
	var total, active int64
	m.DB.Model(&model.JumpserverSession{}).Where("DATE(started_at) = CURDATE()").Count(&total)
	m.DB.Model(&model.JumpserverSession{}).Where("status = 'active'").Count(&active)
	stats["total"] = total
	stats["active"] = active
	var avgDuration float64
	m.DB.Model(&model.JumpserverSession{}).Where("DATE(started_at) = CURDATE() AND status = 'closed'").
		Select("AVG(duration)").Scan(&avgDuration)
	stats["avgDuration"] = int64(avgDuration)
	return stats, nil
}

// ==================== JumpserverCommandMapper ====================

type JumpserverCommandMapper struct {
	DB *gorm.DB
}

func NewJumpserverCommandMapper(db *gorm.DB) *JumpserverCommandMapper {
	return &JumpserverCommandMapper{DB: db}
}

func (m *JumpserverCommandMapper) ListPage(page, pageSize int, sessionID string, isRisky *bool, keyword string) (int64, []model.JumpserverCommand, error) {
	var total int64
	var list []model.JumpserverCommand
	db := m.DB.Model(&model.JumpserverCommand{})
	if sessionID != "" {
		db = db.Where("session_id = ?", sessionID)
	}
	if isRisky != nil {
		db = db.Where("is_risky = ?", *isRisky)
	}
	if keyword != "" {
		db = db.Where("command LIKE ?", "%"+keyword+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("timestamp ASC").Find(&list).Error
	return total, list, err
}

func (m *JumpserverCommandMapper) GetBySessionID(sessionID string) ([]model.JumpserverCommand, error) {
	var list []model.JumpserverCommand
	err := m.DB.Where("session_id = ?", sessionID).Order("timestamp ASC").Find(&list).Error
	return list, err
}

func (m *JumpserverCommandMapper) Create(cmd *model.JumpserverCommand) error {
	return m.DB.Create(cmd).Error
}

func (m *JumpserverCommandMapper) BatchCreate(cmds []model.JumpserverCommand) error {
	if len(cmds) == 0 {
		return nil
	}
	return m.DB.Create(&cmds).Error
}

// ==================== JumpserverAssetPermissionMapper ====================

type JumpserverAssetPermissionMapper struct {
	DB *gorm.DB
}

func NewJumpserverAssetPermissionMapper(db *gorm.DB) *JumpserverAssetPermissionMapper {
	return &JumpserverAssetPermissionMapper{DB: db}
}

func (m *JumpserverAssetPermissionMapper) ListPage(page, pageSize int, name string, isActive *bool) (int64, []model.JumpserverAssetPermission, error) {
	var total int64
	var list []model.JumpserverAssetPermission
	db := m.DB.Model(&model.JumpserverAssetPermission{}).Where("deleted_at IS NULL")
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if isActive != nil {
		db = db.Where("is_active = ?", *isActive)
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	return total, list, err
}

func (m *JumpserverAssetPermissionMapper) GetByID(id uint64) (*model.JumpserverAssetPermission, error) {
	var p model.JumpserverAssetPermission
	err := m.DB.Where("id = ? AND deleted_at IS NULL", id).First(&p).Error
	return &p, err
}

func (m *JumpserverAssetPermissionMapper) Create(p *model.JumpserverAssetPermission) error {
	return m.DB.Create(p).Error
}

func (m *JumpserverAssetPermissionMapper) Update(p *model.JumpserverAssetPermission) error {
	return m.DB.Save(p).Error
}

func (m *JumpserverAssetPermissionMapper) SoftDelete(id uint64) error {
	return m.DB.Where("id = ?", id).Delete(&model.JumpserverAssetPermission{}).Error
}

// CheckPermission 检查用户是否有权限访问指定资产
func (m *JumpserverAssetPermissionMapper) CheckPermission(userID uint64, hostID uint64) (bool, []uint64, bool, error) {
	var permissions []model.JumpserverAssetPermission
	err := m.DB.Where("deleted_at IS NULL AND is_active = 1").Find(&permissions).Error
	if err != nil {
		return false, nil, false, err
	}

	var allowedCredentialIDs []uint64
	needApproval := false
	for _, p := range permissions {
		// 检查用户匹配
		userMatch := false
		if len(p.UserIDs) == 0 {
			userMatch = true // 空=所有用户
		} else {
			for _, uid := range p.UserIDs {
				if uid == userID {
					userMatch = true
					break
				}
			}
		}
		if !userMatch {
			continue
		}

		// 检查主机匹配
		hostMatch := false
		if len(p.HostIDs) == 0 && len(p.HostGroupIDs) == 0 {
			hostMatch = true // 空=所有主机
		} else {
			for _, hid := range p.HostIDs {
				if hid == hostID {
					hostMatch = true
					break
				}
			}
			// 如果主机ID没匹配，检查是否属于授权的分组
			if !hostMatch && len(p.HostGroupIDs) > 0 {
				var host model.AssetHost
				if err := m.DB.Where("id = ?", hostID).First(&host).Error; err == nil {
					for _, gid := range p.HostGroupIDs {
						if host.GroupID == gid {
							hostMatch = true
							break
						}
					}
				}
			}
		}
		if !hostMatch {
			continue
		}

		// 检查是否需要审批
		if p.NeedApproval {
			needApproval = true
		}

		// 收集凭证ID
		allowedCredentialIDs = append(allowedCredentialIDs, p.CredentialIDs...)
	}

	return len(allowedCredentialIDs) > 0, allowedCredentialIDs, needApproval, nil
}

// CheckHostPermissionBatch 批量检查用户对多个主机的权限
func (m *JumpserverAssetPermissionMapper) CheckHostPermissionBatch(userID uint64, hostIDs []uint64) (map[uint64][]uint64, error) {
	result := make(map[uint64][]uint64)
	for _, hostID := range hostIDs {
		allowed, creds, _, err := m.CheckPermission(userID, hostID)
		if err != nil {
			continue
		}
		if allowed {
			result[hostID] = creds
		}
	}
	return result, nil
}

// ==================== JumpserverApprovalMapper ====================

type JumpserverApprovalMapper struct {
	DB *gorm.DB
}

func NewJumpserverApprovalMapper(db *gorm.DB) *JumpserverApprovalMapper {
	return &JumpserverApprovalMapper{DB: db}
}

func (m *JumpserverApprovalMapper) ListPage(page, pageSize int, status string, myApplies *bool, userID uint64) (int64, []model.JumpserverApproval, error) {
	var total int64
	var list []model.JumpserverApproval
	db := m.DB.Model(&model.JumpserverApproval{})
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if myApplies != nil {
		if *myApplies {
			db = db.Where("applicant_id = ?", userID)
		} else {
			db = db.Where("status = 'pending'")
		}
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&list).Error
	return total, list, err
}

func (m *JumpserverApprovalMapper) GetByID(id uint64) (*model.JumpserverApproval, error) {
	var a model.JumpserverApproval
	err := m.DB.Where("id = ?", id).First(&a).Error
	return &a, err
}

func (m *JumpserverApprovalMapper) Create(a *model.JumpserverApproval) error {
	return m.DB.Create(a).Error
}

func (m *JumpserverApprovalMapper) Update(a *model.JumpserverApproval) error {
	return m.DB.Save(a).Error
}

func (m *JumpserverApprovalMapper) GetPendingCount(userID uint64) (int64, error) {
	var count int64
	err := m.DB.Model(&model.JumpserverApproval{}).Where("status = 'pending' AND applicant_id = ?", userID).Count(&count).Error
	return count, err
}

// HasValidApproval 检查用户是否有对某主机的有效审批（已通过且未过期）
func (m *JumpserverApprovalMapper) HasValidApproval(userID uint64, hostID uint64) (bool, error) {
	var count int64
	err := m.DB.Model(&model.JumpserverApproval{}).
		Where("applicant_id = ? AND host_id = ? AND status = 'approved' AND (expired_at IS NULL OR expired_at > NOW())", userID, hostID).
		Count(&count).Error
	return count > 0, err
}

// GetApproverIDsByHost 获取指定主机权限规则中配置的审批人ID列表
func (m *JumpserverAssetPermissionMapper) GetApproverIDsByHost(hostID uint64) ([]uint64, error) {
	var permissions []model.JumpserverAssetPermission
	err := m.DB.Where("deleted_at IS NULL AND is_active = 1 AND need_approval = 1").Find(&permissions).Error
	if err != nil {
		return nil, err
	}

	approverIDSet := make(map[uint64]bool)
	for _, p := range permissions {
		// 检查主机是否匹配
		hostMatch := false
		if len(p.HostIDs) == 0 && len(p.HostGroupIDs) == 0 {
			hostMatch = true
		} else {
			for _, hid := range p.HostIDs {
				if hid == hostID {
					hostMatch = true
					break
				}
			}
			if !hostMatch && len(p.HostGroupIDs) > 0 {
				var host model.AssetHost
				if err := m.DB.Where("id = ?", hostID).First(&host).Error; err == nil {
					for _, gid := range p.HostGroupIDs {
						if host.GroupID == gid {
							hostMatch = true
							break
						}
					}
				}
			}
		}
		if !hostMatch {
			continue
		}

		// 收集审批人
		for _, aid := range p.ApproverIDs {
			approverIDSet[aid] = true
		}
	}

	var result []uint64
	for id := range approverIDSet {
		result = append(result, id)
	}
	return result, nil
}

// ==================== JumpserverAuditLogMapper ====================

type JumpserverAuditLogMapper struct {
	DB *gorm.DB
}

func NewJumpserverAuditLogMapper(db *gorm.DB) *JumpserverAuditLogMapper {
	return &JumpserverAuditLogMapper{DB: db}
}

func (m *JumpserverAuditLogMapper) ListPage(page, pageSize int, userID uint64, action, resourceType, dateFrom, dateTo string) (int64, []model.JumpserverAuditLog, error) {
	var total int64
	var list []model.JumpserverAuditLog
	db := m.DB.Model(&model.JumpserverAuditLog{})
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if action != "" {
		db = db.Where("action = ?", action)
	}
	if resourceType != "" {
		db = db.Where("resource_type = ?", resourceType)
	}
	if dateFrom != "" {
		db = db.Where("created_at >= ?", dateFrom)
	}
	if dateTo != "" {
		db = db.Where("created_at <= ?", dateTo)
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&list).Error
	return total, list, err
}

func (m *JumpserverAuditLogMapper) Create(log *model.JumpserverAuditLog) error {
	return m.DB.Create(log).Error
}

// ==================== JumpserverRiskRuleMapper ====================

type JumpserverRiskRuleMapper struct {
	DB *gorm.DB
}

func NewJumpserverRiskRuleMapper(db *gorm.DB) *JumpserverRiskRuleMapper {
	return &JumpserverRiskRuleMapper{DB: db}
}

func (m *JumpserverRiskRuleMapper) ListPage(page, pageSize int, name, level string, isActive *bool) (int64, []model.JumpserverRiskRule, error) {
	var total int64
	var list []model.JumpserverRiskRule
	db := m.DB.Model(&model.JumpserverRiskRule{})
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if level != "" {
		db = db.Where("level = ?", level)
	}
	if isActive != nil {
		db = db.Where("is_active = ?", *isActive)
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("id ASC").Find(&list).Error
	return total, list, err
}

func (m *JumpserverRiskRuleMapper) GetAllActive() ([]model.JumpserverRiskRule, error) {
	var list []model.JumpserverRiskRule
	err := m.DB.Where("is_active = 1").Order("id ASC").Find(&list).Error
	return list, err
}

func (m *JumpserverRiskRuleMapper) GetByID(id uint64) (*model.JumpserverRiskRule, error) {
	var r model.JumpserverRiskRule
	err := m.DB.Where("id = ?", id).First(&r).Error
	return &r, err
}

func (m *JumpserverRiskRuleMapper) Create(r *model.JumpserverRiskRule) error {
	return m.DB.Create(r).Error
}

func (m *JumpserverRiskRuleMapper) Update(r *model.JumpserverRiskRule) error {
	return m.DB.Save(r).Error
}

func (m *JumpserverRiskRuleMapper) Delete(id uint64) error {
	return m.DB.Where("id = ?", id).Delete(&model.JumpserverRiskRule{}).Error
}

// ==================== JumpserverPlatformMapper ====================

type JumpserverPlatformMapper struct {
	DB *gorm.DB
}

func NewJumpserverPlatformMapper(db *gorm.DB) *JumpserverPlatformMapper {
	return &JumpserverPlatformMapper{DB: db}
}

func (m *JumpserverPlatformMapper) ListPage(page, pageSize int, name, category string) (int64, []model.JumpserverPlatform, error) {
	var total int64
	var list []model.JumpserverPlatform
	db := m.DB.Model(&model.JumpserverPlatform{}).Where("deleted_at IS NULL")
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	if category != "" {
		db = db.Where("category = ?", category)
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Order("id ASC").Find(&list).Error
	return total, list, err
}

func (m *JumpserverPlatformMapper) ListAll() ([]model.JumpserverPlatform, error) {
	var list []model.JumpserverPlatform
	err := m.DB.Where("deleted_at IS NULL AND is_active = 1").Order("id ASC").Find(&list).Error
	return list, err
}

func (m *JumpserverPlatformMapper) GetByID(id uint64) (*model.JumpserverPlatform, error) {
	var p model.JumpserverPlatform
	err := m.DB.Where("id = ? AND deleted_at IS NULL", id).First(&p).Error
	return &p, err
}

func (m *JumpserverPlatformMapper) Create(p *model.JumpserverPlatform) error {
	return m.DB.Create(p).Error
}

func (m *JumpserverPlatformMapper) Update(p *model.JumpserverPlatform) error {
	return m.DB.Save(p).Error
}

func (m *JumpserverPlatformMapper) SoftDelete(id uint64) error {
	return m.DB.Where("id = ?", id).Delete(&model.JumpserverPlatform{}).Error
}

// ==================== 工具函数 ====================

// containsUint64 检查slice是否包含某个值
func containsUint64(slice []uint64, val uint64) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// EncryptPassword 加密密码
func EncryptPassword(plainText string) (string, error) {
	if plainText == "" {
		return "", nil
	}
	key, err := configs.GetEncryptionKey()
	if err != nil {
		return "", fmt.Errorf("获取加密密钥失败: %w", err)
	}
	if key == nil {
		return plainText, nil
	}
	return aes.AESEncrypt(key, plainText)
}