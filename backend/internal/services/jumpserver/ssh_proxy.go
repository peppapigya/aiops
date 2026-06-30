package jumpserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

// SSHProxy  SSH 代理服务，管理 WebSocket 到 SSH 的桥接
type SSHProxy struct {
	DB                      *gorm.DB
	credentialMapper        *mapper.JumpserverCredentialMapper
	hostMapper              *mapper.AssetHostMapper
	sessionMapper           *mapper.JumpserverSessionMapper
	commandMapper           *mapper.JumpserverCommandMapper
	riskRuleMapper          *mapper.JumpserverRiskRuleMapper
	auditLogMapper          *mapper.JumpserverAuditLogMapper
	permissionMapper        *mapper.JumpserverAssetPermissionMapper
	approvalMapper          *mapper.JumpserverApprovalMapper
	recordDir               string
	activeSessions          map[string]*ActiveSession
	mu                      sync.RWMutex
}

// stdoutListener stdout 数据监听器
type stdoutListener struct {
	ch   chan []byte
	done chan struct{}
}

// ActiveSession 活跃的 SSH 会话
type ActiveSession struct {
	SessionID       string
	UserID          uint64
	Username        string
	HostID          uint64
	HostName        string
	HostIP          string
	SSHClient       *ssh.Client
	SSHSession      *ssh.Session
	Recorder        *SessionRecorder
	StartTime       time.Time
	Width           int
	Height          int
	RiskRules       []model.JumpserverRiskRule
	CommandCount    int
	MaxRiskLevel    string
	StdinPipe       io.WriteCloser
	StdoutPipe      io.Reader
	StderrPipe      io.Reader
	Cancel          context.CancelFunc
	stdoutListeners []*stdoutListener
	listenersMu     sync.RWMutex
	// 在 listener 注册前缓存的 stdout 数据
	stdoutBuffer [][]byte
	bufferMu     sync.Mutex
	// 命令记录
	commandMapper *mapper.JumpserverCommandMapper
}

// SessionRecorder 会话录像器，生成 asciicast v2 格式
type SessionRecorder struct {
	File      *os.File
	StartTime time.Time
	Width     int
	Height    int
	mu        sync.Mutex
}

// NewSSHProxy 创建 SSH 代理服务
func NewSSHProxy(db *gorm.DB, recordDir string) *SSHProxy {
	if recordDir == "" {
		recordDir = "./storage/replays"
	}
	os.MkdirAll(recordDir, 0755)
	return &SSHProxy{
		DB:               db,
		credentialMapper: mapper.NewJumpserverCredentialMapper(db),
		hostMapper:       mapper.NewAssetHostMapper(db),
		sessionMapper:    mapper.NewJumpserverSessionMapper(db),
		commandMapper:    mapper.NewJumpserverCommandMapper(db),
		riskRuleMapper:   mapper.NewJumpserverRiskRuleMapper(db),
		auditLogMapper:   mapper.NewJumpserverAuditLogMapper(db),
		permissionMapper: mapper.NewJumpserverAssetPermissionMapper(db),
		approvalMapper:   mapper.NewJumpserverApprovalMapper(db),
		recordDir:        recordDir,
		activeSessions:   make(map[string]*ActiveSession),
	}
}

// Connect 建立 SSH 连接并创建会话
func (p *SSHProxy) Connect(userID uint64, username string, hostID uint64, credentialID uint64, width, height int, clientIP string) (*ActiveSession, error) {
	// 1. 权限检查
	allowed, _, needApproval, err := p.permissionMapper.CheckPermission(userID, hostID)
	if err != nil {
		return nil, fmt.Errorf("权限检查失败: %w", err)
	}
	if !allowed {
		// 记录审计日志
		p.writeAuditLog(userID, username, "connect_denied", "host", fmt.Sprintf("%d", hostID), "", clientIP, "failure", "无权限访问该主机")
		return nil, fmt.Errorf("无权限访问该主机")
	}

	// 2. 审批检查：如果权限规则要求审批，检查是否有有效审批
	if needApproval {
		hasApproval, err := p.approvalMapper.HasValidApproval(userID, hostID)
		if err != nil {
			return nil, fmt.Errorf("审批检查失败: %w", err)
		}
		if !hasApproval {
			p.writeAuditLog(userID, username, "connect_denied", "host", fmt.Sprintf("%d", hostID), "", clientIP, "failure", "需要审批才能连接该主机，请先提交审批申请")
			return nil, fmt.Errorf("需要审批才能连接该主机，请先提交审批申请")
		}
	}

	// 2. 获取主机信息
	host, err := p.hostMapper.GetByID(hostID)
	if err != nil {
		return nil, fmt.Errorf("主机不存在: %w", err)
	}

	// 3. 获取凭证信息
	cred, err := p.credentialMapper.GetByID(credentialID)
	if err != nil {
		return nil, fmt.Errorf("凭证不存在: %w", err)
	}

	// 4. 解密密码
	password, err := p.credentialMapper.DecryptPassword(credentialID)
	if err != nil {
		return nil, fmt.Errorf("解密凭证失败: %w", err)
	}
	if password == "" && cred.Password != nil {
		password = *cred.Password // 如果解密失败，尝试使用原始值
	}

	// 5. 建立 SSH 连接
	sshConfig := &ssh.ClientConfig{
		User:            cred.Username,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	if password != "" {
		sshConfig.Auth = append(sshConfig.Auth, ssh.Password(password))
	}

	addr := fmt.Sprintf("%s:%d", host.IP, host.Port)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		// 记录审计日志
		p.writeAuditLog(userID, username, "connect", "host", fmt.Sprintf("%d", hostID), host.Name, clientIP, "failure", err.Error())
		return nil, fmt.Errorf("SSH连接失败: %w", err)
	}

	// 6. 创建 SSH 会话
	sshSession, err := sshClient.NewSession()
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("创建SSH会话失败: %w", err)
	}

	// 7. 获取 PTY
	modes := ssh.TerminalModes{
		ssh.ECHO:          1, // 开启SSH回显，shell正常回显输入
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
		ssh.ICRNL:         1, // 将输入 CR 转为 NL
		ssh.ONLCR:         1, // 将输出 NL 转为 CR+NL
		ssh.OPOST:         1, // 启用输出处理
		ssh.ISIG:          1, // 启用信号
		ssh.ICANON:        1, // 规范模式
		ssh.IEXTEN:        1, // 扩展输入处理
	}
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	if err := sshSession.RequestPty("xterm-256color", height, width, modes); err != nil {
		sshSession.Close()
		sshClient.Close()
		return nil, fmt.Errorf("请求PTY失败: %w", err)
	}

	// 8. 获取管道
	stdinPipe, err := sshSession.StdinPipe()
	if err != nil {
		sshSession.Close()
		sshClient.Close()
		return nil, fmt.Errorf("获取stdin管道失败: %w", err)
	}
	stdoutPipe, err := sshSession.StdoutPipe()
	if err != nil {
		sshSession.Close()
		sshClient.Close()
		return nil, fmt.Errorf("获取stdout管道失败: %w", err)
	}
	stderrPipe, err := sshSession.StderrPipe()
	if err != nil {
		sshSession.Close()
		sshClient.Close()
		return nil, fmt.Errorf("获取stderr管道失败: %w", err)
	}

	// 9. 启动 Shell
	if err := sshSession.Shell(); err != nil {
		sshSession.Close()
		sshClient.Close()
		return nil, fmt.Errorf("启动Shell失败: %w", err)
	}

	// 10. 创建会话录像
	sessionID := uuid.New().String()
	now := time.Now()
	recorder, err := NewSessionRecorder(p.recordDir, sessionID, width, height, now)
	if err != nil {
		sshSession.Close()
		sshClient.Close()
		return nil, fmt.Errorf("创建录像器失败: %w", err)
	}

	// 11. 加载危险命令规则
	rules, _ := p.riskRuleMapper.GetAllActive()

	// 12. 创建活跃会话
	ctx, cancel := context.WithCancel(context.Background())
	active := &ActiveSession{
		SessionID:     sessionID,
		UserID:        userID,
		Username:      username,
		HostID:        hostID,
		HostName:      host.Name,
		HostIP:        host.IP,
		SSHClient:     sshClient,
		SSHSession:    sshSession,
		Recorder:      recorder,
		StartTime:     now,
		Width:         width,
		Height:        height,
		RiskRules:     rules,
		MaxRiskLevel:  "low",
		StdinPipe:     stdinPipe,
		StdoutPipe:    stdoutPipe,
		StderrPipe:    stderrPipe,
		Cancel:        cancel,
		commandMapper: p.commandMapper,
	}

	// 启动 stdout 广播协程：从 stdoutPipe 读取，广播到所有注册的 listener
	// 每个 listener 有独立的 channel，确保每个消费者都能收到完整数据
	// 在 listener 注册前，数据会缓存到 buffer，等 listener 注册后批量发送
	go func() {
		buf := make([]byte, 4096)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			n, err := stdoutPipe.Read(buf)
			if err != nil {
				// 广播结束，关闭所有 listener
				active.listenersMu.Lock()
				for _, l := range active.stdoutListeners {
					close(l.ch)
				}
				active.listenersMu.Unlock()
				return
			}
			if n > 0 {
				data := make([]byte, n)
				copy(data, buf[:n])
				// 始终缓存数据（用于后续注册的 listener）
				active.bufferMu.Lock()
				active.stdoutBuffer = append(active.stdoutBuffer, data)
				// 限制缓存大小，防止内存泄漏（最多保留 256 个数据块）
				if len(active.stdoutBuffer) > 256 {
					active.stdoutBuffer = active.stdoutBuffer[len(active.stdoutBuffer)-256:]
				}
				active.bufferMu.Unlock()
				// 发送给现有 listener
				active.listenersMu.RLock()
				for _, l := range active.stdoutListeners {
					select {
					case l.ch <- data:
					default:
						// listener 消费太慢，跳过
					}
				}
				active.listenersMu.RUnlock()
			}
		}
	}()

	// 13. 保存到数据库
	hostName := host.Name
	hostIP := host.IP
	session := &model.JumpserverSession{
		SessionID:      sessionID,
		UserID:         userID,
		Username:       username,
		HostID:         hostID,
		HostName:       &hostName,
		HostIP:         &hostIP,
		CredentialID:   &credentialID,
		Protocol:       "ssh",
		LoginFrom:      "WT",
		RemoteAddr:     &clientIP,
		Status:         "active",
		TerminalWidth:  uint16(width),
		TerminalHeight: uint16(height),
		StartedAt:      now,
		RiskLevel:      "low",
		HasReplay:      true,
	}
	if err := p.sessionMapper.Create(session); err != nil {
		// 不中断连接，只记录错误
		os.Stderr.WriteString(fmt.Sprintf("保存会话失败: %v\n", err))
	}

	// 14. 注册活跃会话
	p.mu.Lock()
	p.activeSessions[sessionID] = active
	p.mu.Unlock()

	// 15. 记录审计日志
	p.writeAuditLog(userID, username, "connect", "host", fmt.Sprintf("%d", hostID), host.Name, clientIP, "success", "")

	return active, nil
}

// Disconnect 断开会话
func (p *SSHProxy) Disconnect(sessionID string) error {
	p.mu.Lock()
	active, ok := p.activeSessions[sessionID]
	if ok {
		delete(p.activeSessions, sessionID)
	}
	p.mu.Unlock()

	if !ok {
		return fmt.Errorf("会话不存在")
	}

	// 取消命令解析
	active.Cancel()

	// 关闭 SSH 会话和连接
	active.SSHSession.Close()
	active.SSHClient.Close()

	// 关闭录像
	recordingPath := active.Recorder.Close()

	// 更新数据库
	now := time.Now()
	duration := uint(now.Sub(active.StartTime).Seconds())
	recordingSize := uint64(0)
	if info, err := os.Stat(recordingPath); err == nil {
		recordingSize = uint64(info.Size())
	}
	session, err := p.sessionMapper.GetBySessionID(sessionID)
	if err == nil {
		session.Status = "closed"
		session.EndedAt = &now
		session.Duration = duration
		session.RecordingPath = &recordingPath
		session.RecordingSize = recordingSize
		session.CommandCount = uint(active.CommandCount)
		session.RiskLevel = active.MaxRiskLevel
		session.HasReplay = true
		p.sessionMapper.Update(session)
	}

	// 审计日志
	p.writeAuditLog(active.UserID, active.Username, "disconnect", "host", fmt.Sprintf("%d", active.HostID), active.HostName, "", "success", "")
	return nil
}

// AddStdoutListener 注册一个新的 stdout 监听器，返回独立的 channel
// 每个消费者（parseCommands、WebSocket handler）都应该调用此方法获取自己的 channel
// 注册后会在后台 goroutine 中发送之前缓存的数据（如果存在）
func (s *ActiveSession) AddStdoutListener() chan []byte {
	ch := make(chan []byte, 1024)
	l := &stdoutListener{ch: ch}

	// 先获取缓存数据
	s.bufferMu.Lock()
	buffered := make([][]byte, len(s.stdoutBuffer))
	copy(buffered, s.stdoutBuffer)
	s.bufferMu.Unlock()

	// 注册 listener
	s.listenersMu.Lock()
	s.stdoutListeners = append(s.stdoutListeners, l)
	s.listenersMu.Unlock()

	// 异步发送缓存数据，避免阻塞调用方
	if len(buffered) > 0 {
		go func() {
			for _, data := range buffered {
				select {
				case ch <- data:
				case <-time.After(5 * time.Second):
					return
				}
			}
		}()
	}

	return ch
}

// RecordCommand 保存用户输入的命令（前端已过滤，只发送纯命令文本）
func (s *ActiveSession) RecordCommand(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" || len(cmd) > 200 {
		return
	}

	s.CommandCount++
	now := time.Now()
	elapsed := now.Sub(s.StartTime).Seconds()

	var riskLevel string
	var riskRule *string
	for _, rule := range s.RiskRules {
		matched, _ := regexp.MatchString(rule.Pattern, cmd)
		if matched {
			level := rule.Level
			name := rule.Name
			riskLevel = level
			riskRule = &name
			break
		}
	}

	if riskLevel == "critical" {
		s.MaxRiskLevel = "critical"
	} else if riskLevel == "high" && s.MaxRiskLevel != "critical" {
		s.MaxRiskLevel = "high"
	} else if riskLevel == "medium" && s.MaxRiskLevel == "low" {
		s.MaxRiskLevel = "medium"
	}

	cmdRecord := &model.JumpserverCommand{
		SessionID: s.SessionID,
		UserID:    s.UserID,
		HostID:    s.HostID,
		Command:   cmd,
		Timestamp: elapsed,
		IsRisky:   riskLevel != "",
		RiskLevel: &riskLevel,
		RiskRule:  riskRule,
	}
	if riskLevel == "" {
		rl := "low"
		cmdRecord.RiskLevel = &rl
	}
	if err := s.commandMapper.Create(cmdRecord); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("保存命令失败: %v\n", err))
	} else {
		os.Stderr.WriteString(fmt.Sprintf("命令已保存: session=%s cmd=%s\n", s.SessionID, cmd))
	}
}

func (p *SSHProxy) GetActiveSession(sessionID string) *ActiveSession {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.activeSessions[sessionID]
}

// GetRecordingPath 获取录像文件路径
func (p *SSHProxy) GetRecordingPath(sessionID string) string {
	session, err := p.sessionMapper.GetBySessionID(sessionID)
	if err != nil || session.RecordingPath == nil {
		return ""
	}
	return *session.RecordingPath
}

// checkRiskCommand 检查命令是否危险
func (p *SSHProxy) checkRiskCommand(command string, rules []model.JumpserverRiskRule) (string, *string) {
	for _, rule := range rules {
		matched, _ := regexp.MatchString(rule.Pattern, command)
		if matched {
			level := rule.Level
			name := rule.Name
			return level, &name
		}
	}
	return "", nil
}

// writeAuditLog 写审计日志
func (p *SSHProxy) writeAuditLog(userID uint64, username, action, resourceType, resourceID, resourceName, clientIP, status, errMsg string) {
	now := time.Now()
	detail := map[string]interface{}{
		"action":      action,
		"resourceType": resourceType,
	}
	detailJSON, _ := json.Marshal(detail)
	detailStr := string(detailJSON)

	var errMsgPtr *string
	if errMsg != "" {
		errMsgPtr = &errMsg
	}
	var resourceNamePtr *string
	if resourceName != "" {
		resourceNamePtr = &resourceName
	}
	var resourceIDPtr *string
	if resourceID != "" {
		resourceIDPtr = &resourceID
	}
	var resourceTypePtr *string
	if resourceType != "" {
		resourceTypePtr = &resourceType
	}
	var clientIPPtr *string
	if clientIP != "" {
		clientIPPtr = &clientIP
	}

	log := &model.JumpserverAuditLog{
		UserID:       userID,
		Username:     username,
		Action:       action,
		ResourceType: resourceTypePtr,
		ResourceID:   resourceIDPtr,
		ResourceName: resourceNamePtr,
		Detail:       &detailStr,
		ClientIP:     clientIPPtr,
		Status:       status,
		ErrorMsg:     errMsgPtr,
		CreatedAt:    &now,
	}
	p.auditLogMapper.Create(log)
}

// ==================== SessionRecorder ====================

// NewSessionRecorder 创建会话录像器
func NewSessionRecorder(recordDir, sessionID string, width, height int, startTime time.Time) (*SessionRecorder, error) {
	dateDir := startTime.Format("2006/01/02")
	fullDir := filepath.Join(recordDir, dateDir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(fullDir, sessionID+".cast")
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	recorder := &SessionRecorder{
		File:      file,
		StartTime: startTime,
		Width:     width,
		Height:    height,
	}

	// 写入 asciicast v2 头部
	header := map[string]interface{}{
		"version":   2,
		"width":     width,
		"height":    height,
		"timestamp": startTime.Unix(),
		"title":     fmt.Sprintf("Session %s", sessionID),
		"env": map[string]string{
			"SHELL": "/bin/bash",
			"TERM":  "xterm-256color",
		},
	}
	headerJSON, _ := json.Marshal(header)
	file.Write(headerJSON)
	file.Write([]byte("\n"))

	return recorder, nil
}

// RecordOutput 记录输出数据
func (r *SessionRecorder) RecordOutput(data []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

	elapsed := time.Since(r.StartTime).Seconds()
	encoded := base64.StdEncoding.EncodeToString(data)
	line := fmt.Sprintf("[%.6f, \"o\", \"%s\"]\n", elapsed, encoded)
	r.File.WriteString(line)
}

// RecordInput 记录输入数据
func (r *SessionRecorder) RecordInput(data []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

	elapsed := time.Since(r.StartTime).Seconds()
	encoded := base64.StdEncoding.EncodeToString(data)
	line := fmt.Sprintf("[%.6f, \"i\", \"%s\"]\n", elapsed, encoded)
	r.File.WriteString(line)
}

// Resize 更新终端尺寸
func (r *SessionRecorder) Resize(width, height int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Width = width
	r.Height = height
}

// Close 关闭录像文件
func (r *SessionRecorder) Close() string {
	r.File.Close()
	return r.File.Name()
}