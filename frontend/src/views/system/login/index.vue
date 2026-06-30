<template>
  <div class="login-container">
    <section class="login-shell">
      <aside class="brand-panel">
        <div class="brand-top">
          <div class="brand-mark">DC</div>
          <div>
            <h1>DevOps 控制台</h1>
            <p class="eyebrow">统一运维控制台</p>
          </div>
        </div>
        <p class="brand-desc">Kubernetes · CI/CD · Elasticsearch · Kafka · MySQL · MongoDB · 监控告警</p>
      </aside>

      <main class="login-card">
        <div class="login-header">
          <p class="eyebrow">安全访问</p>
          <h2 class="title">登录控制台</h2>
        </div>

        <el-form
          ref="loginFormRef"
          :model="loginForm"
          :rules="loginRules"
          class="login-form"
          @keyup.enter="handleLogin"
        >
          <el-form-item prop="username">
            <label class="field-label">用户名</label>
            <el-input
              v-model="loginForm.username"
              placeholder="输入用户名"
              :prefix-icon="User"
              size="large"
            />
          </el-form-item>

          <el-form-item prop="password">
            <label class="field-label">密码</label>
            <el-input
              v-model="loginForm.password"
              type="password"
              placeholder="输入密码"
              :prefix-icon="Lock"
              show-password
              size="large"
            />
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              :loading="loading"
              class="login-button"
              @click="handleLogin"
              size="large"
            >
              {{ loading ? '正在验证身份...' : '进入控制台' }}
            </el-button>
          </el-form-item>
        </el-form>

        <div class="security-strip">
          <span />
          <p>令牌会话 · RBAC权限控制 · 审计就绪</p>
        </div>

        <div class="login-footer">
          <p>© 2026 peppa-pig. All rights reserved.</p>
        </div>
      </main>
    </section>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { login } from '@/api/system/user.js'
import { SHA256 } from 'crypto-js'

const router = useRouter()
const loginFormRef = ref(null)
const loading = ref(false)

const loginForm = reactive({
  username: '',
  password: ''
})

const loginRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const handleLogin = async () => {
  if (!loginFormRef.value) return

  await loginFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        const loginData = {
          ...loginForm,
          password: SHA256(loginForm.password).toString()
        }
        const res = await login(loginData)
        console.log('Login Result:', res)
        let accessToken = res.accessToken
        let refreshToken = res.refreshToken

        if (res.data) {
          accessToken = accessToken || res.data.accessToken || res.data.access_token
          refreshToken = refreshToken || res.data.refreshToken || res.data.refresh_token

          if (res.data.data) {
            accessToken = accessToken || res.data.data.accessToken || res.data.data.access_token
            refreshToken = refreshToken || res.data.data.refreshToken || res.data.data.refresh_token
          }
        }

        console.log('Final extracted AccessToken:', accessToken)

        if (accessToken) {
          localStorage.setItem('access_token', accessToken)
          console.log('Token successfully saved to localStorage. Verification:', localStorage.getItem('access_token'))

          if (refreshToken) {
            localStorage.setItem('refresh_token', refreshToken)
          }
          ElMessage.success('登录成功')
          router.push('/')
        } else {
          console.error('FAILED to extract token. Response nested structure check failed.', res)
          ElMessage.error('登录失败：无法获取访问令牌')
          localStorage.removeItem('access_token')
          localStorage.removeItem('refresh_token')
        }
      } catch (error) {
        ElMessage.error(error.message || '登录失败')
      } finally {
        loading.value = false
      }
    }
  })
}
</script>

<style scoped>
.login-container {
  display: flex;
  min-height: 100vh;
  width: 100%;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background:
    radial-gradient(circle at 18% 20%, rgba(59, 130, 246, 0.14), transparent 34%),
    radial-gradient(circle at 82% 70%, rgba(34, 197, 94, 0.08), transparent 30%),
    var(--ds-bg-app, #0d1117);
  color: var(--ds-text-primary, #f0f6fc);
}

.login-shell {
  display: grid;
  width: min(1120px, calc(100vw - 48px));
  min-height: 640px;
  grid-template-columns: minmax(0, 1.35fr) 440px;
  overflow: hidden;
  border: 1px solid var(--ds-border-default, rgba(255, 255, 255, 0.1));
  border-radius: 8px;
  background: rgba(22, 27, 34, 0.92);
}

.brand-panel,
.login-card {
  min-width: 0;
  padding: 24px;
}

.brand-panel {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  text-align: center;
  border-right: 1px solid var(--ds-border-default, rgba(255, 255, 255, 0.1));
  background: linear-gradient(180deg, rgba(59, 130, 246, 0.06), transparent 60%),
    var(--ds-bg-surface, #161b22);
}

.brand-top {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.brand-mark {
  display: inline-flex;
  width: 44px;
  height: 44px;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(59, 130, 246, 0.45);
  border-radius: 10px;
  background: rgba(59, 130, 246, 0.10);
  color: var(--ds-accent, #3b82f6);
  font-size: 16px;
  font-weight: 800;
  letter-spacing: 0.06em;
}

.brand-top h1 {
  margin: 0;
  color: var(--ds-text-primary, #f0f6fc);
  font-size: 18px;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.brand-panel .eyebrow {
  margin: 2px 0 0;
  color: var(--ds-text-tertiary, #8b949e);
  font-size: 12px;
  font-weight: 500;
}

.login-header .eyebrow {
  margin: 0 0 8px;
  color: var(--ds-accent, #3b82f6);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.brand-desc {
  margin: 16px 0 0;
  max-width: 280px;
  color: var(--ds-text-muted, #6e7681);
  font-size: 11px;
  line-height: 1.5;
  text-align: center;
}

.title {
  margin: 0;
  color: var(--ds-text-primary, #f0f6fc);
  font-weight: 700;
  letter-spacing: -0.02em;
}

.login-card {
  display: flex;
  flex-direction: column;
  justify-content: center;
  background: var(--ds-bg-surface, #161b22);
}

.login-header {
  margin-bottom: 18px;
}

.title {
  font-size: 24px;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

:deep(.el-form-item) {
  margin-bottom: 6px;
}

:deep(.el-form-item__content) {
  display: block;
}

.field-label {
  display: block;
  margin-bottom: 5px;
  color: var(--ds-text-secondary, #c9d1d9);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

:deep(.el-input__wrapper) {
  min-height: 38px;
  border-radius: 6px;
  background: var(--ds-bg-app, #0d1117) !important;
  box-shadow: 0 0 0 1px var(--ds-border-default, rgba(255,255,255,0.1)) inset !important;
}

:deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--ds-border-strong, rgba(255,255,255,0.16)) inset !important;
}

:deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--ds-border-focus, rgba(59,130,246,0.72)) inset !important;
}

:deep(.el-input__inner) {
  color: var(--ds-text-primary, #f0f6fc);
}

:deep(.el-input__prefix),
:deep(.el-input__suffix) {
  color: var(--ds-text-tertiary, #8b949e);
}

.login-button {
  width: 100%;
  height: 38px;
  border: 1px solid var(--ds-accent, #3b82f6) !important;
  border-radius: 6px;
  background: var(--ds-accent, #3b82f6) !important;
  color: #ffffff !important;
  font-size: 13px;
  font-weight: 700;
  box-shadow: none !important;
}

.login-button:hover,
.login-button:focus {
  border-color: var(--ds-accent-hover, #60a5fa) !important;
  background: var(--ds-accent-hover, #60a5fa) !important;
}

.security-strip {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 6px;
  padding: 8px 10px;
  border: 1px solid var(--ds-border-default, rgba(255,255,255,0.1));
  border-radius: 6px;
  background: rgba(13, 17, 23, 0.48);
}

.security-strip span {
  width: 7px;
  height: 7px;
  flex: 0 0 auto;
  border-radius: 50%;
  background: var(--ds-success, #22c55e);
}

.security-strip p,
.login-footer p {
  margin: 0;
  font-size: 12px;
}

.login-footer {
  margin-top: 18px;
  text-align: center;
}

@media (max-width: 960px) {
  .login-shell {
    grid-template-columns: 1fr;
    width: min(520px, calc(100vw - 32px));
  }

  .brand-panel {
    display: none;
  }

  .login-card {
    min-height: 560px;
  }
}

@media (max-width: 520px) {
  .login-shell {
    width: calc(100vw - 24px);
  }

  .login-card {
    padding: 26px;
  }
}
</style>
