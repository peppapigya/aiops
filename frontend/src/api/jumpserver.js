import api from './index'

// ==================== 凭证管理 ====================
export const getCredentials = (params) => api.get('/jumpserver/credentials', { params })
export const getAllCredentials = () => api.get('/jumpserver/credentials/all')
export const getCredential = (id) => api.get(`/jumpserver/credentials/${id}`)
export const createCredential = (data) => api.post('/jumpserver/credentials', data)
export const updateCredential = (id, data) => api.put(`/jumpserver/credentials/${id}`, data)
export const deleteCredential = (id) => api.delete(`/jumpserver/credentials/${id}`)

// ==================== 主机-凭证关联 ====================
export const getHostCredentials = (hostId) => api.get(`/jumpserver/host-credentials/${hostId}`)
export const bindHostCredentials = (data) => api.post('/jumpserver/host-credentials/bind', data)

// ==================== 会话管理 ====================
export const getSessions = (params) => api.get('/jumpserver/sessions', { params })
export const getSessionStats = () => api.get('/jumpserver/sessions/stats')
export const getSession = (id) => api.get(`/jumpserver/sessions/${id}`)
export const getSessionCommands = (id, params) => api.get(`/jumpserver/sessions/${id}/commands`, { params })
export const deleteSession = (id) => api.delete(`/jumpserver/sessions/${id}`)
export const getRecordingUrl = (sessionId) => `/api/v1/jumpserver/recordings/${sessionId}`
export const getRecordingDownloadUrl = (sessionId) => `/api/v1/jumpserver/recordings/${sessionId}/download`

// ==================== 连接管理 ====================
export const connectHost = (data) => api.post('/jumpserver/connect', data)
export const disconnectSession = (sessionId) => api.post(`/jumpserver/disconnect/${sessionId}`)

// ==================== 权限管理 ====================
export const getPermissions = (params) => api.get('/jumpserver/permissions', { params })
export const getPermission = (id) => api.get(`/jumpserver/permissions/${id}`)
export const createPermission = (data) => api.post('/jumpserver/permissions', data)
export const updatePermission = (id, data) => api.put(`/jumpserver/permissions/${id}`, data)
export const deletePermission = (id) => api.delete(`/jumpserver/permissions/${id}`)
export const checkPermission = (hostId) => api.get('/jumpserver/permissions/check', { params: { hostId } })

// ==================== 审批管理 ====================
export const getApprovals = (params) => api.get('/jumpserver/approvals', { params })
export const getApproval = (id) => api.get(`/jumpserver/approvals/${id}`)
export const createApproval = (data) => api.post('/jumpserver/approvals', data)
export const approveApproval = (id, data) => api.put(`/jumpserver/approvals/${id}/approve`, data)
export const rejectApproval = (id, data) => api.put(`/jumpserver/approvals/${id}/reject`, data)

// ==================== 审计日志 ====================
export const getAuditLogs = (params) => api.get('/jumpserver/audit-logs', { params })

// ==================== 危险命令规则 ====================
export const getRiskRules = (params) => api.get('/jumpserver/risk-rules', { params })
export const getRiskRule = (id) => api.get(`/jumpserver/risk-rules/${id}`)
export const createRiskRule = (data) => api.post('/jumpserver/risk-rules', data)
export const updateRiskRule = (id, data) => api.put(`/jumpserver/risk-rules/${id}`, data)
export const deleteRiskRule = (id) => api.delete(`/jumpserver/risk-rules/${id}`)

// ==================== 批量执行 ====================
export const batchExec = (data) => api.post('/jumpserver/batch-exec', data)
export const getBatchTask = (taskId) => api.get(`/jumpserver/batch-exec/${taskId}`)

// ==================== 平台管理 ====================
export const getPlatforms = () => api.get('/jumpserver/platforms')