package model

import "time"

const TableNameSysModuleConfig = "sys_module_config"

// SysModuleConfig 模块开关配置
type SysModuleConfig struct {
	ID           uint64     `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	ModuleKey    string     `gorm:"column:module_key;type:varchar(50);not null;uniqueIndex" json:"moduleKey"`
	ModuleName   string     `gorm:"column:module_name;type:varchar(100);not null" json:"moduleName"`
	Description  *string    `gorm:"column:description;type:varchar(255)" json:"description"`
	IsEnabled    bool       `gorm:"column:is_enabled;type:tinyint(1);not null;default:1" json:"isEnabled"`
	MenuParentID uint64     `gorm:"column:menu_parent_id;type:bigint unsigned;not null;default:0" json:"menuParentId"`
	CreatedAt    *time.Time `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
	UpdatedAt    *time.Time `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
}

func (*SysModuleConfig) TableName() string { return TableNameSysModuleConfig }