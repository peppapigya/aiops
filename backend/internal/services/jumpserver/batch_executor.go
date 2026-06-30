package jumpserver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BatchExecutor 批量命令执行器
type BatchExecutor struct {
	DB               *gorm.DB
	hostMapper       *mapper.AssetHostMapper
	credentialMapper *mapper.JumpserverCredentialMapper
	auditLogMapper   *mapper.JumpserverAuditLogMapper
	mu               sync.RWMutex
	tasks            map[string]*BatchTask
}

// BatchTask 批量执行任务
type BatchTask struct {
	TaskID    string
	UserID    uint64
	Username  string
	Command   string
	Timeout   time.Duration
	Status    string // pending/running/completed/failed
	Progress  int
	Results   []BatchExecHostResult
	CreatedAt time.Time
	mu        sync.Mutex
}

// BatchExecHostResult 单台主机执行结果
type BatchExecHostResult struct {
	HostID   uint64 `json:"hostId"`
	HostName string `json:"hostName"`
	HostIP   string `json:"hostIp"`
	Success  bool   `json:"success"`
	Output   string `json:"output"`
	ExitCode int    `json:"exitCode"`
	Duration int64  `json:"duration"`
	Error    string `json:"error,omitempty"`
}

// NewBatchExecutor 创建批量执行器
func NewBatchExecutor(db *gorm.DB) *BatchExecutor {
	return &BatchExecutor{
		DB:               db,
		hostMapper:       mapper.NewAssetHostMapper(db),
		credentialMapper: mapper.NewJumpserverCredentialMapper(db),
		auditLogMapper:   mapper.NewJumpserverAuditLogMapper(db),
		tasks:            make(map[string]*BatchTask),
	}
}

// Execute 批量执行命令
func (e *BatchExecutor) Execute(userID uint64, username string, hostIDs []uint64, credentialIDs []uint64, command string, timeout int) (string, error) {
	if timeout <= 0 {
		timeout = 30
	}

	taskID := uuid.New().String()
	task := &BatchTask{
		TaskID:    taskID,
		UserID:    userID,
		Username:  username,
		Command:   command,
		Timeout:   time.Duration(timeout) * time.Second,
		Status:    "running",
		CreatedAt: time.Now(),
	}

	e.mu.Lock()
	e.tasks[taskID] = task
	e.mu.Unlock()

	// 异步执行
	go e.runBatch(task, hostIDs, credentialIDs)

	// 审计日志
	now := time.Now()
	detailJSON := fmt.Sprintf(`{"command":"%s","hostCount":%d}`, command, len(hostIDs))
	_ = e.auditLogMapper.Create(&model.JumpserverAuditLog{
		UserID:       userID,
		Username:     username,
		Action:       "batch_exec",
		ResourceType: strPtr("host"),
		Detail:       &detailJSON,
		Status:       "success",
		CreatedAt:    &now,
	})

	return taskID, nil
}

// GetTask 获取任务状态
func (e *BatchExecutor) GetTask(taskID string) *BatchTask {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.tasks[taskID]
}

// runBatch 执行批量命令
func (e *BatchExecutor) runBatch(task *BatchTask, hostIDs, credentialIDs []uint64) {
	total := len(hostIDs)
	results := make([]BatchExecHostResult, 0, total)

	for i, hostID := range hostIDs {
		// 获取主机信息
		host, err := e.hostMapper.GetByID(hostID)
		if err != nil {
			results = append(results, BatchExecHostResult{
				HostID: hostID,
				Error:  fmt.Sprintf("主机不存在: %v", err),
			})
			continue
		}

		// 尝试每个凭证连接
		connected := false
		for _, credID := range credentialIDs {
			cred, err := e.credentialMapper.GetByID(credID)
			if err != nil {
				continue
			}

			password, err := e.credentialMapper.DecryptPassword(credID)
			if err != nil || password == "" {
				if cred.Password != nil {
					password = *cred.Password
				}
			}

			startTime := time.Now()
			ctx, cancel := context.WithTimeout(context.Background(), task.Timeout)
			result := ExecuteSSHCommand(ctx, host.IP, int(host.Port), cred.Username, password, task.Command)
			cancel()

			result.HostID = hostID
			result.HostName = host.Name
			result.HostIP = host.IP
			result.Duration = time.Since(startTime).Milliseconds()

			results = append(results, result)
			connected = true
			break
		}

		if !connected {
			results = append(results, BatchExecHostResult{
				HostID:   hostID,
				HostName: host.Name,
				HostIP:   host.IP,
				Success:  false,
				Error:    "无法连接主机，所有凭证均失败",
			})
		}

		// 更新进度
		task.mu.Lock()
		task.Results = results
		task.Progress = (i + 1) * 100 / total
		task.mu.Unlock()
	}

	task.mu.Lock()
	task.Status = "completed"
	task.Progress = 100
	task.mu.Unlock()
}

func strPtr(s string) *string {
	return &s
}