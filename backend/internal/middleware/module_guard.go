package middleware

import (
	"devops-console-backend/internal/dal/mapper"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ModuleGuardPath 根据请求路径自动匹配模块的中间件
// 无需在每个路由注册处单独加中间件，统一在入口处拦截
func ModuleGuardPath(db *gorm.DB) gin.HandlerFunc {
	mapper := mapper.NewSysModuleConfigMapper(db)
	// 路径前缀到模块key的映射
	pathModuleMap := map[string]string{
		"/api/v1/elasticsearch/": "elasticsearch",
		"/api/v1/instance/":      "elasticsearch",
		"/api/v1/node/":          "elasticsearch",
		"/api/v1/shard/":         "elasticsearch",
		"/api/v1/indices/":       "elasticsearch",
		"/api/v1/backup/":        "elasticsearch",
		"/api/v1/k8s/":           "kubernetes",
		"/api/v1/cluster/":       "kubernetes",
		"/api/v1/helm/":          "helm",
		"/api/v1/kafka/":         "kafka",
		"/api/v1/clusters/":      "kafka",
		"/api/v1/mysql/":         "mysql",
		"/api/v1/mongodb/":       "mongodb",
		"/api/v1/monitor/":       "monitor",
		"/api/v1/pipelines/":     "cicd",
		"/api/v1/pipeline-runs/": "cicd",
		"/api/v1/pipeline-steps/":"cicd",
		"/api/v1/projects/":      "cicd",
		"/api/v1/argo/":          "cicd",
		"/api/v1/asset/":         "asset",
		"/api/v1/hosts/":         "asset",
		"/api/v1/host-groups/":   "asset",
		"/api/v1/task-scheduler/":"task_scheduler",
		"/api/v1/workflows/":     "task_scheduler",
		"/api/v1/jumpserver/":    "jumpserver",
	}
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		for prefix, moduleKey := range pathModuleMap {
			if strings.HasPrefix(path, prefix) {
				cfg, err := mapper.GetByModuleKey(moduleKey)
				if err != nil {
					c.Next()
					return
				}
				if !cfg.IsEnabled {
					c.JSON(http.StatusOK, gin.H{
						"status":  403,
						"message": cfg.ModuleName + "模块已关闭，如需使用请联系管理员开启",
						"data":    nil,
					})
					c.Abort()
					return
				}
				break
			}
		}
		c.Next()
	}
}