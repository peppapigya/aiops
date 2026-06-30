package mapper

import (
	"devops-console-backend/internal/dal/model"

	"gorm.io/gorm"
)

// SysModuleConfigMapper 模块配置数据访问
type SysModuleConfigMapper struct {
	DB *gorm.DB
}

func NewSysModuleConfigMapper(db *gorm.DB) *SysModuleConfigMapper {
	return &SysModuleConfigMapper{DB: db}
}

// ListAll 获取所有模块配置
func (m *SysModuleConfigMapper) ListAll() ([]model.SysModuleConfig, error) {
	var list []model.SysModuleConfig
	err := m.DB.Order("id ASC").Find(&list).Error
	return list, err
}

// GetByModuleKey 根据模块标识获取配置
func (m *SysModuleConfigMapper) GetByModuleKey(moduleKey string) (*model.SysModuleConfig, error) {
	var cfg model.SysModuleConfig
	err := m.DB.Where("module_key = ?", moduleKey).First(&cfg).Error
	return &cfg, err
}

// GetByID 根据ID获取配置
func (m *SysModuleConfigMapper) GetByID(id uint64) (*model.SysModuleConfig, error) {
	var cfg model.SysModuleConfig
	err := m.DB.Where("id = ?", id).First(&cfg).Error
	return &cfg, err
}

// UpdateEnabled 更新模块启用状态
func (m *SysModuleConfigMapper) UpdateEnabled(id uint64, enabled bool) error {
	return m.DB.Model(&model.SysModuleConfig{}).Where("id = ?", id).Update("is_enabled", enabled).Error
}

// GetDisabledModuleKeys 获取所有已禁用的模块标识列表
func (m *SysModuleConfigMapper) GetDisabledModuleKeys() ([]string, error) {
	var keys []string
	err := m.DB.Model(&model.SysModuleConfig{}).Where("is_enabled = 0").Pluck("module_key", &keys).Error
	return keys, err
}

// GetDisabledMenuParentIDs 获取所有已禁用模块对应的菜单父ID列表
func (m *SysModuleConfigMapper) GetDisabledMenuParentIDs() ([]uint64, error) {
	var ids []uint64
	err := m.DB.Model(&model.SysModuleConfig{}).Where("is_enabled = 0 AND menu_parent_id > 0").Pluck("menu_parent_id", &ids).Error
	return ids, err
}