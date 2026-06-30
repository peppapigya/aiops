package middlewares

import (
	redisdal "devops-console-backend/internal/dal/redis"
	"devops-console-backend/pkg/utils/jwt"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
)

func TestRedisOperator_WhenRedisNotInitialized_ShouldNotPanic(t *testing.T) {
	originGetRedisClient := getRedisClient
	originNewRedisClient := newRedisClient
	originGetBlockedKey := getBlockedKey
	originDeleteBlockedKey := deleteBlockedKey
	defer func() {
		getRedisClient = originGetRedisClient
		newRedisClient = originNewRedisClient
		getBlockedKey = originGetBlockedKey
		deleteBlockedKey = originDeleteBlockedKey
	}()

	getRedisClient = func() *goredis.Client {
		return nil
	}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	claims := &jwt.Claims{ID: 1001, Username: "tester"}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("redisOperator 不应 panic，实际 panic: %v", r)
		}
	}()

	redisOperator(claims, ctx, "dummy-token")

	if ctx.IsAborted() {
		t.Fatalf("redis 未初始化时不应中断请求")
	}
}

func TestRedisOperator_WhenBlockedKeyExists_ShouldAbortWith401(t *testing.T) {
	originGetRedisClient := getRedisClient
	originNewRedisClient := newRedisClient
	originGetBlockedKey := getBlockedKey
	originDeleteBlockedKey := deleteBlockedKey
	defer func() {
		getRedisClient = originGetRedisClient
		newRedisClient = originNewRedisClient
		getBlockedKey = originGetBlockedKey
		deleteBlockedKey = originDeleteBlockedKey
	}()

	getRedisClient = func() *goredis.Client {
		return &goredis.Client{}
	}
	newRedisClient = func(client *goredis.Client) *redisdal.RedisClient {
		return redisdal.NewClient(client)
	}
	getBlockedKey = func(redisClient *redisdal.RedisClient, key string) string {
		return "force-offline"
	}
	deleteBlockedKey = func(redisClient *redisdal.RedisClient, key string) error {
		return nil
	}

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	claims := &jwt.Claims{ID: 1002, Username: "tester2"}
	redisOperator(claims, ctx, "dummy-token")

	if !ctx.IsAborted() {
		t.Fatalf("命中强制下线 key 时应中断请求")
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	status, ok := resp["status"].(float64)
	if !ok || int(status) != 401 {
		t.Fatalf("期望业务状态码 401，实际响应: %s", recorder.Body.String())
	}
}
