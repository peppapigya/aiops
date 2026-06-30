package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// JSONArray 通用 JSON 数组类型，用于存储 []uint64 等
type JSONArray []uint64

func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	b, err := json.Marshal(j)
	return string(b), err
}

func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case string:
		b = []byte(v)
	case []byte:
		b = v
	default:
		return fmt.Errorf("unsupported JSONArray type: %T", value)
	}
	return json.Unmarshal(b, j)
}

// JSONStringArray 通用 JSON 字符串数组
type JSONStringArray []string

func (j JSONStringArray) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	b, err := json.Marshal(j)
	return string(b), err
}

func (j *JSONStringArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case string:
		b = []byte(v)
	case []byte:
		b = v
	default:
		return fmt.Errorf("unsupported JSONStringArray type: %T", value)
	}
	return json.Unmarshal(b, j)
}

// ==================== 平台表 ====================

const TableNameJumpserverPlatform = "jumpserver_platforms"

type JumpserverPlatform struct {
	ID          uint64         `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Category    string         `gorm:"column:category;type:varchar(30);not null;default:host" json:"category"`
	Type        string         `gorm:"column:type;type:varchar(30);not null;default:linux" json:"type"`
	Protocol    string         `gorm:"column:protocol;type:varchar(30);not null;default:ssh" json:"protocol"`
	DefaultPort uint16         `gorm:"column:default_port;type:smallint unsigned;not null;default:22" json:"defaultPort"`
	Charset     string         `gorm:"column:charset;type:varchar(20);default:utf-8" json:"charset"`
	IsActive    bool           `gorm:"column:is_active;type:tinyint(1);not null;default:1" json:"isActive"`
	CreatedAt   *time.Time     `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
	UpdatedAt   *time.Time     `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(3);index" json:"-"`
}

func (*JumpserverPlatform) TableName() string { return TableNameJumpserverPlatform }

// ==================== 认证凭证表 ====================

const TableNameJumpserverCredential = "jumpserver_credentials"

type JumpserverCredential struct {
	ID         uint64         `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	Name       string         `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Type       string         `gorm:"column:type;type:varchar(20);not null;default:password" json:"type"`
	Username   string         `gorm:"column:username;type:varchar(100);not null" json:"username"`
	Password   *string        `gorm:"column:password;type:varchar(500)" json:"-"`
	PrivateKey *string        `gorm:"column:private_key;type:text" json:"-"`
	Passphrase *string        `gorm:"column:passphrase;type:varchar(500)" json:"-"`
	Protocol   string         `gorm:"column:protocol;type:varchar(30);not null;default:ssh" json:"protocol"`
	Priority   int            `gorm:"column:priority;type:int;not null;default:0" json:"priority"`
	IsGlobal   bool           `gorm:"column:is_global;type:tinyint(1);not null;default:0" json:"isGlobal"`
	Remark     *string        `gorm:"column:remark;type:varchar(500)" json:"remark"`
	CreatedBy  *uint64        `gorm:"column:created_by;type:bigint unsigned" json:"createdBy"`
	CreatedAt  *time.Time     `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
	UpdatedAt  *time.Time     `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;type:datetime(3);index" json:"-"`
}

func (*JumpserverCredential) TableName() string { return TableNameJumpserverCredential }

// ==================== 主机-凭证关联表 ====================

const TableNameJumpserverHostCredential = "jumpserver_host_credentials"

type JumpserverHostCredential struct {
	ID           uint64 `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	HostID       uint64 `gorm:"column:host_id;type:bigint unsigned;not null;index" json:"hostId"`
	CredentialID uint64 `gorm:"column:credential_id;type:bigint unsigned;not null;index" json:"credentialId"`
	Priority     int    `gorm:"column:priority;type:int;not null;default:0" json:"priority"`
}

func (*JumpserverHostCredential) TableName() string { return TableNameJumpserverHostCredential }

// ==================== 会话记录表 ====================

const TableNameJumpserverSession = "jumpserver_sessions"

type JumpserverSession struct {
	ID             uint64     `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	SessionID      string     `gorm:"column:session_id;type:varchar(64);not null;uniqueIndex" json:"sessionId"`
	UserID         uint64     `gorm:"column:user_id;type:bigint unsigned;not null;index" json:"userId"`
	Username       string     `gorm:"column:username;type:varchar(100);not null" json:"username"`
	HostID         uint64     `gorm:"column:host_id;type:bigint unsigned;not null;index" json:"hostId"`
	HostName       *string    `gorm:"column:host_name;type:varchar(200)" json:"hostName"`
	HostIP         *string    `gorm:"column:host_ip;type:varchar(50)" json:"hostIp"`
	CredentialID   *uint64    `gorm:"column:credential_id;type:bigint unsigned" json:"credentialId"`
	Protocol       string     `gorm:"column:protocol;type:varchar(20);not null;default:ssh" json:"protocol"`
	LoginFrom      string     `gorm:"column:login_from;type:varchar(20);not null;default:WT" json:"loginFrom"`
	RemoteAddr     *string    `gorm:"column:remote_addr;type:varchar(128)" json:"remoteAddr"`
	Status         string     `gorm:"column:status;type:varchar(20);not null;default:active;index" json:"status"`
	TerminalWidth  uint16     `gorm:"column:terminal_width;type:smallint unsigned;default:80" json:"terminalWidth"`
	TerminalHeight uint16     `gorm:"column:terminal_height;type:smallint unsigned;default:24" json:"terminalHeight"`
	StartedAt      time.Time  `gorm:"column:started_at;type:datetime(3);not null;index" json:"startedAt"`
	EndedAt        *time.Time `gorm:"column:ended_at;type:datetime(3)" json:"endedAt"`
	Duration       uint       `gorm:"column:duration;type:int unsigned;default:0" json:"duration"`
	RecordingPath  *string    `gorm:"column:recording_path;type:varchar(500)" json:"recordingPath"`
	RecordingSize  uint64     `gorm:"column:recording_size;type:bigint unsigned;default:0" json:"recordingSize"`
	CommandCount   uint       `gorm:"column:command_count;type:int unsigned;default:0" json:"commandCount"`
	RiskLevel      string     `gorm:"column:risk_level;type:varchar(10);default:low;index" json:"riskLevel"`
	HasReplay      bool       `gorm:"column:has_replay;type:tinyint(1);not null;default:0" json:"hasReplay"`
	CreatedAt      *time.Time `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
	UpdatedAt      *time.Time `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
}

func (*JumpserverSession) TableName() string { return TableNameJumpserverSession }

// ==================== 命令记录表 ====================

const TableNameJumpserverCommand = "jumpserver_commands"

type JumpserverCommand struct {
	ID        uint64     `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	SessionID string     `gorm:"column:session_id;type:varchar(64);not null;index" json:"sessionId"`
	UserID    uint64     `gorm:"column:user_id;type:bigint unsigned;not null;index" json:"userId"`
	HostID    uint64     `gorm:"column:host_id;type:bigint unsigned;not null;index" json:"hostId"`
	Command   string     `gorm:"column:command;type:text;not null" json:"command"`
	Output    *string    `gorm:"column:output;type:mediumtext" json:"output"`
	ExitCode  *int       `gorm:"column:exit_code;type:int" json:"exitCode"`
	Duration  uint       `gorm:"column:duration;type:int unsigned;default:0" json:"duration"`
	Timestamp float64    `gorm:"column:timestamp;type:double;not null;default:0;index" json:"timestamp"`
	IsRisky   bool       `gorm:"column:is_risky;type:tinyint(1);default:0;index" json:"isRisky"`
	RiskLevel *string    `gorm:"column:risk_level;type:varchar(10)" json:"riskLevel"`
	RiskRule  *string    `gorm:"column:risk_rule;type:varchar(100)" json:"riskRule"`
	CreatedAt *time.Time `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
}

func (*JumpserverCommand) TableName() string { return TableNameJumpserverCommand }

// ==================== 资产权限规则表 ====================

const TableNameJumpserverAssetPermission = "jumpserver_asset_permissions"

type JumpserverAssetPermission struct {
	ID                 uint64          `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	Name               string          `gorm:"column:name;type:varchar(100);not null" json:"name"`
	UserIDs            JSONArray       `gorm:"column:user_ids;type:json" json:"userIds"`
	RoleIDs            JSONArray       `gorm:"column:role_ids;type:json" json:"roleIds"`
	HostIDs            JSONArray       `gorm:"column:host_ids;type:json" json:"hostIds"`
	HostGroupIDs       JSONArray       `gorm:"column:host_group_ids;type:json" json:"hostGroupIds"`
	CredentialIDs      JSONArray       `gorm:"column:credential_ids;type:json" json:"credentialIds"`
	Protocols          JSONStringArray `gorm:"column:protocols;type:json" json:"protocols"`
	IsActive           bool            `gorm:"column:is_active;type:tinyint(1);not null;default:1;index" json:"isActive"`
	DateStart          *time.Time      `gorm:"column:date_start;type:datetime(3)" json:"dateStart"`
	DateExpired        *time.Time      `gorm:"column:date_expired;type:datetime(3)" json:"dateExpired"`
	TimeStart          *string         `gorm:"column:time_start;type:time" json:"timeStart"`
	TimeEnd            *string         `gorm:"column:time_end;type:time" json:"timeEnd"`
	MaxSessionDuration uint            `gorm:"column:max_session_duration;type:int unsigned;default:0" json:"maxSessionDuration"`
	NeedApproval       bool            `gorm:"column:need_approval;type:tinyint(1);default:0" json:"needApproval"`
	ApproverIDs        JSONArray       `gorm:"column:approver_ids;type:json" json:"approverIds"`
	Remark             *string         `gorm:"column:remark;type:varchar(500)" json:"remark"`
	CreatedBy          *uint64         `gorm:"column:created_by;type:bigint unsigned" json:"createdBy"`
	CreatedAt          *time.Time      `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
	UpdatedAt          *time.Time      `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
	DeletedAt          gorm.DeletedAt  `gorm:"column:deleted_at;type:datetime(3);index" json:"-"`
}

func (*JumpserverAssetPermission) TableName() string { return TableNameJumpserverAssetPermission }

// ==================== 访问审批表 ====================

const TableNameJumpserverApproval = "jumpserver_approvals"

type JumpserverApproval struct {
	ID            uint64     `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	ApplicantID   uint64     `gorm:"column:applicant_id;type:bigint unsigned;not null;index" json:"applicantId"`
	ApplicantName string     `gorm:"column:applicant_name;type:varchar(100);not null" json:"applicantName"`
	HostID        uint64     `gorm:"column:host_id;type:bigint unsigned;not null;index" json:"hostId"`
	HostName      *string    `gorm:"column:host_name;type:varchar(200)" json:"hostName"`
	HostIP        *string    `gorm:"column:host_ip;type:varchar(50)" json:"hostIp"`
	CredentialID  *uint64    `gorm:"column:credential_id;type:bigint unsigned" json:"credentialId"`
	Reason        *string    `gorm:"column:reason;type:varchar(500)" json:"reason"`
	Duration      uint       `gorm:"column:duration;type:int unsigned;not null;default:3600" json:"duration"`
	Status        string     `gorm:"column:status;type:varchar(20);not null;default:pending;index" json:"status"`
	ApproverID    *uint64    `gorm:"column:approver_id;type:bigint unsigned;index" json:"approverId"`
	ApproverName  *string    `gorm:"column:approver_name;type:varchar(100)" json:"approverName"`
	ApprovedAt    *time.Time `gorm:"column:approved_at;type:datetime(3)" json:"approvedAt"`
	ExpiredAt     *time.Time `gorm:"column:expired_at;type:datetime(3)" json:"expiredAt"`
	Remark        *string    `gorm:"column:remark;type:varchar(500)" json:"remark"`
	CreatedAt     *time.Time `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
	UpdatedAt     *time.Time `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
}

func (*JumpserverApproval) TableName() string { return TableNameJumpserverApproval }

// ==================== 操作审计日志表 ====================

const TableNameJumpserverAuditLog = "jumpserver_audit_logs"

type JumpserverAuditLog struct {
	ID           uint64     `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	UserID       uint64     `gorm:"column:user_id;type:bigint unsigned;not null;index" json:"userId"`
	Username     string     `gorm:"column:username;type:varchar(100);not null" json:"username"`
	Action       string     `gorm:"column:action;type:varchar(50);not null;index" json:"action"`
	ResourceType *string    `gorm:"column:resource_type;type:varchar(50);index" json:"resourceType"`
	ResourceID   *string    `gorm:"column:resource_id;type:varchar(100)" json:"resourceId"`
	ResourceName *string    `gorm:"column:resource_name;type:varchar(200)" json:"resourceName"`
	Detail       *string    `gorm:"column:detail;type:json" json:"detail"`
	ClientIP     *string    `gorm:"column:client_ip;type:varchar(50)" json:"clientIp"`
	UserAgent    *string    `gorm:"column:user_agent;type:varchar(500)" json:"userAgent"`
	Status       string     `gorm:"column:status;type:varchar(20);not null;default:success" json:"status"`
	ErrorMsg     *string    `gorm:"column:error_msg;type:varchar(500)" json:"errorMsg"`
	CreatedAt    *time.Time `gorm:"column:created_at;type:datetime(3);index" json:"createdAt"`
}

func (*JumpserverAuditLog) TableName() string { return TableNameJumpserverAuditLog }

// ==================== 危险命令规则表 ====================

const TableNameJumpserverRiskRule = "jumpserver_risk_rules"

type JumpserverRiskRule struct {
	ID        uint64     `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement" json:"id"`
	Name      string     `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Pattern   string     `gorm:"column:pattern;type:varchar(500);not null" json:"pattern"`
	Level     string     `gorm:"column:level;type:varchar(10);not null;default:high" json:"level"`
	Action    string     `gorm:"column:action;type:varchar(20);not null;default:alert" json:"action"`
	IsActive  bool       `gorm:"column:is_active;type:tinyint(1);not null;default:1;index" json:"isActive"`
	Remark    *string    `gorm:"column:remark;type:varchar(500)" json:"remark"`
	CreatedAt *time.Time `gorm:"column:created_at;type:datetime(3)" json:"createdAt"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:datetime(3)" json:"updatedAt"`
}

func (*JumpserverRiskRule) TableName() string { return TableNameJumpserverRiskRule }