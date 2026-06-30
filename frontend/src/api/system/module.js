import api from '../index.js'

// 获取所有模块配置
export const getModuleConfigs = () => api.get('/system/modules')

// 切换模块启用状态
export const toggleModule = (id) => api.put(`/system/modules/${id}/toggle`)