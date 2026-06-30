package jumpserver

// ==================== 凭证管理 ====================

type CredentialPageReq struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=200"`
	Name     string `form:"name"`
	Type     string `form:"type"`
	Username string `form:"username"`
}

type CredentialCreateReq struct {
	Name       string `json:"name" binding:"required,min=1,max=100"`
	Type       string `json:"type" binding:"required,oneof=password private_key token"`
	Username   string `json:"username" binding:"required,min=1,max=100"`
	Password   string `json:"password"`
	PrivateKey string `json:"privateKey"`
	Passphrase string `json:"passphrase"`
	Protocol   string `json:"protocol" binding:"required,oneof=ssh rdp vnc telnet"`
	Priority   int    `json:"priority"`
	IsGlobal   bool   `json:"isGlobal"`
	Remark     string `json:"remark"`
}

type CredentialUpdateReq struct {
	Name       string `json:"name" binding:"required,min=1,max=100"`
	Type       string `json:"type" binding:"required,oneof=password private_key token"`
	Username   string `json:"username" binding:"required,min=1,max=100"`
	Password   string `json:"password"`
	PrivateKey string `json:"privateKey"`
	Passphrase string `json:"passphrase"`
	Protocol   string `json:"protocol" binding:"required,oneof=ssh rdp vnc telnet"`
	Priority   int    `json:"priority"`
	IsGlobal   bool   `json:"isGlobal"`
	Remark     string `json:"remark"`
}

// ==================== 主机-凭证关联 ====================

type HostCredentialBindReq struct {
	HostID        uint64   `json:"hostId" binding:"required"`
	CredentialIDs []uint64 `json:"credentialIds" binding:"required,min=1"`
}

// ==================== 会话管理 ====================

type SessionPageReq struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=200"`
	UserID   uint64 `form:"userId"`
	HostID   uint64 `form:"hostId"`
	Status   string `form:"status"`
	RiskLevel string `form:"riskLevel"`
	DateFrom string `form:"dateFrom"`
	DateTo   string `form:"dateTo"`
	Keyword  string `form:"keyword"`
}

type ConnectReq struct {
	HostID       uint64 `json:"hostId" binding:"required"`
	CredentialID uint64 `json:"credentialId" binding:"required"`
	Width        uint16 `json:"width"`
	Height       uint16 `json:"height"`
}

type ConnectResp struct {
	SessionID string `json:"sessionId"`
	Token     string `json:"token"`
	HostName  string `json:"hostName"`
	HostIP    string `json:"hostIp"`
}

// ConnectResult 连接结果(用于返回)
type ConnectResult = ConnectResp

// ==================== 命令记录 ====================

type CommandPageReq struct {
	Page      int    `form:"page" binding:"required,min=1"`
	PageSize  int    `form:"pageSize" binding:"required,min=1,max=200"`
	SessionID string `form:"sessionId"`
	IsRisky   *bool  `form:"isRisky"`
	Keyword   string `form:"keyword"`
}

// ==================== 权限管理 ====================

type PermissionPageReq struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=200"`
	Name     string `form:"name"`
	IsActive *bool  `form:"isActive"`
}

type PermissionCreateReq struct {
	Name               string   `json:"name" binding:"required,min=1,max=100"`
	UserIDs            []uint64 `json:"userIds"`
	RoleIDs            []uint64 `json:"roleIds"`
	HostIDs            []uint64 `json:"hostIds"`
	HostGroupIDs       []uint64 `json:"hostGroupIds"`
	CredentialIDs      []uint64 `json:"credentialIds"`
	Protocols          []string `json:"protocols"`
	IsActive           bool     `json:"isActive"`
	DateStart          string   `json:"dateStart"`
	DateExpired        string   `json:"dateExpired"`
	TimeStart          string   `json:"timeStart"`
	TimeEnd            string   `json:"timeEnd"`
	MaxSessionDuration uint     `json:"maxSessionDuration"`
	NeedApproval       bool     `json:"needApproval"`
	ApproverIDs        []uint64 `json:"approverIds"`
	Remark             string   `json:"remark"`
}

type PermissionUpdateReq struct {
	Name               string   `json:"name" binding:"required,min=1,max=100"`
	UserIDs            []uint64 `json:"userIds"`
	RoleIDs            []uint64 `json:"roleIds"`
	HostIDs            []uint64 `json:"hostIds"`
	HostGroupIDs       []uint64 `json:"hostGroupIds"`
	CredentialIDs      []uint64 `json:"credentialIds"`
	Protocols          []string `json:"protocols"`
	IsActive           bool     `json:"isActive"`
	DateStart          string   `json:"dateStart"`
	DateExpired        string   `json:"dateExpired"`
	TimeStart          string   `json:"timeStart"`
	TimeEnd            string   `json:"timeEnd"`
	MaxSessionDuration uint     `json:"maxSessionDuration"`
	NeedApproval       bool     `json:"needApproval"`
	ApproverIDs        []uint64 `json:"approverIds"`
	Remark             string   `json:"remark"`
}

// ==================== 审批管理 ====================

type ApprovalPageReq struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=200"`
	Status   string `form:"status"`
	MyApplies *bool `form:"myApplies"` // true=我的申请, false=待我审批
}

type ApprovalCreateReq struct {
	HostID       uint64 `json:"hostId" binding:"required"`
	CredentialID uint64 `json:"credentialId" binding:"required"`
	Reason       string `json:"reason"`
	Duration     uint   `json:"duration" binding:"required,min=60,max=86400"`
}

type ApprovalHandleReq struct {
	Remark string `json:"remark"`
}

// ==================== 批量执行 ====================

type BatchExecReq struct {
	HostIDs       []uint64 `json:"hostIds" binding:"required,min=1"`
	CredentialIDs []uint64 `json:"credentialIds" binding:"required,min=1"`
	Command       string   `json:"command" binding:"required"`
	Timeout       int      `json:"timeout"` // 超时时间(秒)
}

type BatchExecResult struct {
	TaskID   string               `json:"taskId"`
	Status   string               `json:"status"` // pending/running/completed
	Results  []BatchExecHostResult `json:"results"`
	Progress int                  `json:"progress"` // 0-100
}

type BatchExecHostResult struct {
	HostID   uint64 `json:"hostId"`
	HostName string `json:"hostName"`
	HostIP   string `json:"hostIp"`
	Success  bool   `json:"success"`
	Output   string `json:"output"`
	ExitCode int    `json:"exitCode"`
	Duration int64  `json:"duration"` // ms
	Error    string `json:"error,omitempty"`
}

// ==================== 危险命令规则 ====================

type RiskRulePageReq struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=200"`
	Name     string `form:"name"`
	Level    string `form:"level"`
	IsActive *bool  `form:"isActive"`
}

type RiskRuleCreateReq struct {
	Name    string `json:"name" binding:"required,min=1,max=100"`
	Pattern string `json:"pattern" binding:"required,min=1,max=500"`
	Level   string `json:"level" binding:"required,oneof=low medium high critical"`
	Action  string `json:"action" binding:"required,oneof=alert block approve"`
	IsActive bool  `json:"isActive"`
	Remark  string `json:"remark"`
}

type RiskRuleUpdateReq struct {
	Name    string `json:"name" binding:"required,min=1,max=100"`
	Pattern string `json:"pattern" binding:"required,min=1,max=500"`
	Level   string `json:"level" binding:"required,oneof=low medium high critical"`
	Action  string `json:"action" binding:"required,oneof=alert block approve"`
	IsActive bool  `json:"isActive"`
	Remark  string `json:"remark"`
}

// ==================== 审计日志 ====================

type AuditLogPageReq struct {
	Page         int    `form:"page" binding:"required,min=1"`
	PageSize     int    `form:"pageSize" binding:"required,min=1,max=200"`
	UserID       uint64 `form:"userId"`
	Action       string `form:"action"`
	ResourceType string `form:"resourceType"`
	DateFrom     string `form:"dateFrom"`
	DateTo       string `form:"dateTo"`
}

// ==================== 平台管理 ====================

type PlatformPageReq struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=200"`
	Name     string `form:"name"`
	Category string `form:"category"`
}

type PlatformCreateReq struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Category    string `json:"category" binding:"required,oneof=host network database"`
	Type        string `json:"type" binding:"required"`
	Protocol    string `json:"protocol" binding:"required"`
	DefaultPort uint16 `json:"defaultPort" binding:"required,min=1,max=65535"`
	Charset     string `json:"charset"`
}

type PlatformUpdateReq struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Category    string `json:"category" binding:"required,oneof=host network database"`
	Type        string `json:"type" binding:"required"`
	Protocol    string `json:"protocol" binding:"required"`
	DefaultPort uint16 `json:"defaultPort" binding:"required,min=1,max=65535"`
	Charset     string `json:"charset"`
}