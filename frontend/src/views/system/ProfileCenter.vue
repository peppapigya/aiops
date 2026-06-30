<template>
  <div class="profile-page">
    <el-card class="profile-card" shadow="never">
      <template #header>
        <div class="card-header">
          <div>
            <h2>个人中心</h2>
            <p>修改个人信息、头像和登录密码</p>
          </div>
          <el-button :icon="Refresh" @click="reloadUserInfo">刷新</el-button>
        </div>
      </template>

      <div class="profile-content">
        <section class="avatar-section">
          <el-avatar :size="112" :src="profileForm.avatar || defaultAvatar" />
          <el-upload
            class="avatar-upload"
            :show-file-list="false"
            :auto-upload="false"
            accept="image/jpeg,image/png,image/gif,image/webp"
            :on-change="handleAvatarChange"
          >
            <el-button type="primary" :loading="avatarUploading">更换头像</el-button>
          </el-upload>
          <span class="avatar-tip">支持 jpg、png、gif、webp，最大 2MB</span>
        </section>

        <el-tabs class="profile-tabs" model-value="profile">
          <el-tab-pane label="基本资料" name="profile">
            <el-form ref="profileFormRef" :model="profileForm" :rules="profileRules" label-width="90px" class="profile-form">
              <el-form-item label="用户名">
                <el-input v-model="profileForm.username" disabled />
              </el-form-item>
              <el-form-item label="昵称" prop="nickname">
                <el-input v-model="profileForm.nickname" maxlength="50" placeholder="请输入昵称" />
              </el-form-item>
              <el-form-item label="邮箱" prop="email">
                <el-input v-model="profileForm.email" maxlength="100" placeholder="请输入邮箱" />
              </el-form-item>
              <el-form-item label="手机号" prop="phone">
                <el-input v-model="profileForm.phone" maxlength="20" placeholder="请输入手机号" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :loading="profileSaving" @click="saveProfile">保存资料</el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>

          <el-tab-pane label="修改密码" name="password">
            <el-form ref="passwordFormRef" :model="passwordForm" :rules="passwordRules" label-width="90px" class="profile-form">
              <el-form-item label="旧密码" prop="oldPassword">
                <el-input v-model="passwordForm.oldPassword" type="password" show-password placeholder="请输入旧密码" />
              </el-form-item>
              <el-form-item label="新密码" prop="newPassword">
                <el-input v-model="passwordForm.newPassword" type="password" show-password placeholder="至少6位" />
              </el-form-item>
              <el-form-item label="确认密码" prop="confirmPassword">
                <el-input v-model="passwordForm.confirmPassword" type="password" show-password placeholder="请再次输入新密码" />
              </el-form-item>
              <el-form-item>
                <el-button type="primary" :loading="passwordSaving" @click="savePassword">修改密码</el-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { SHA256 } from 'crypto-js'
import { changePassword, getAuthInfo, updateProfile, uploadAvatar } from '@/api/system/user.js'
import { usePermissionStore } from '@/stores/permissionStore.js'

const defaultAvatar = 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png'
const permStore = usePermissionStore()
const profileFormRef = ref(null)
const passwordFormRef = ref(null)
const profileSaving = ref(false)
const passwordSaving = ref(false)
const avatarUploading = ref(false)

const profileForm = reactive({ userId: null, username: '', nickname: '', email: '', phone: '', avatar: '' })
const passwordForm = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })

const profileRules = {
  email: [{ type: 'email', message: '邮箱格式不正确', trigger: 'blur' }],
  phone: [{ pattern: /^$|^1\d{10}$|^[0-9-+()\s]{6,20}$/, message: '手机号格式不正确', trigger: 'blur' }]
}
const passwordRules = {
  oldPassword: [{ required: true, message: '请输入旧密码', trigger: 'blur' }],
  newPassword: [{ required: true, message: '请输入新密码', trigger: 'blur' }, { min: 6, message: '至少6位', trigger: 'blur' }],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    { validator: (_, value, callback) => value === passwordForm.newPassword ? callback() : callback(new Error('两次输入的密码不一致')), trigger: 'blur' }
  ]
}

watch(() => permStore.userInfo, fillProfileForm, { immediate: true, deep: true })

function fillProfileForm(info) {
  if (!info) return
  Object.assign(profileForm, {
    userId: info.userId,
    username: info.username || '',
    nickname: info.nickname || '',
    email: info.email || '',
    phone: info.phone || '',
    avatar: info.avatar || ''
  })
}

async function reloadUserInfo() {
  const res = await getAuthInfo()
  const data = res?.data?.data || {}
  permStore.setUserInfo(data)
  ElMessage.success('用户信息已刷新')
}

async function saveProfile() {
  await profileFormRef.value?.validate()
  profileSaving.value = true
  try {
    const payload = {
      nickname: profileForm.nickname,
      email: profileForm.email,
      phone: profileForm.phone,
      avatar: profileForm.avatar
    }
    await updateProfile(payload)
    permStore.setUserInfo(payload)
    ElMessage.success('个人信息已更新')
  } catch (e) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    profileSaving.value = false
  }
}

async function handleAvatarChange(file) {
  const raw = file.raw
  if (!raw) return
  if (raw.size > 2 * 1024 * 1024) {
    ElMessage.warning('头像大小不能超过2MB')
    return
  }
  avatarUploading.value = true
  try {
    const res = await uploadAvatar(raw)
    const url = res?.data?.data?.url || res?.data?.url
    profileForm.avatar = url
    permStore.setUserInfo({ avatar: url })
    ElMessage.success('头像已更新')
  } catch (e) {
    ElMessage.error(e?.message || '头像上传失败')
  } finally {
    avatarUploading.value = false
  }
}

async function savePassword() {
  await passwordFormRef.value?.validate()
  passwordSaving.value = true
  try {
    await changePassword({
      oldPassword: SHA256(passwordForm.oldPassword).toString(),
      newPassword: SHA256(passwordForm.newPassword).toString()
    })
    Object.assign(passwordForm, { oldPassword: '', newPassword: '', confirmPassword: '' })
    passwordFormRef.value?.clearValidate()
    ElMessage.success('密码修改成功')
  } catch (e) {
    ElMessage.error(e?.message || '密码修改失败')
  } finally {
    passwordSaving.value = false
  }
}
</script>

<style scoped>
.profile-page { padding: 16px; }
.profile-card { max-width: 980px; margin: 0 auto; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.card-header h2 { margin: 0; font-size: 20px; }
.card-header p { margin: 6px 0 0; color: var(--el-text-color-secondary); font-size: 13px; }
.profile-content { display: grid; grid-template-columns: 220px 1fr; gap: 32px; }
.avatar-section { display: flex; flex-direction: column; align-items: center; gap: 14px; padding-top: 20px; }
.avatar-tip { color: var(--el-text-color-secondary); font-size: 12px; text-align: center; line-height: 1.5; }
.profile-tabs { min-width: 0; }
.profile-form { max-width: 520px; padding-top: 16px; }
@media (max-width: 768px) {
  .profile-content { grid-template-columns: 1fr; }
}
</style>
