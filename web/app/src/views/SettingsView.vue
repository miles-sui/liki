<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useLocaleStore } from '@/stores/locale'
import { api } from '@/api/client'
import { ElMessage, ElMessageBox } from 'element-plus'
import BirthInfoForm from '@/components/shared/BirthInfoForm.vue'
import type { BirthInfo } from '@shared/types'

const auth = useAuthStore()
const { t } = useLocaleStore()
const router = useRouter()

const name = ref('')
const email = ref('')
const isPublic = ref(false)
const birthInfo = ref<Partial<BirthInfo>>({ year: 2000, month: 1, day: 1, hour: 12, minute: 0, longitude: 120, timezone: 120, is_dst: false, gender: 'male' })
const savingProfile = ref(false)
const loading = ref(false)
const error = ref('')
const exporting = ref(false)
const deleting = ref(false)

onMounted(async () => {
  if (!auth.isLoggedIn) return
  loading.value = true
  try {
    await auth.fetchProfile()
    const u = auth.user
    if (u) {
      name.value = u.name
      email.value = u.email
      isPublic.value = (u as Record<string, unknown>).is_public as boolean ?? false
      if (u.birth_info) birthInfo.value = { ...u.birth_info }
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load profile'
  } finally { loading.value = false }
})

async function saveProfile() {
  savingProfile.value = true
  try {
    const body: Record<string, unknown> = { name: name.value, is_public: isPublic.value }
    if (email.value !== auth.user?.email) body.email = email.value
    body.birth_info = birthInfo.value
    await api('/users/me', { method: 'PATCH', body })
    await auth.fetchProfile()
    ElMessage.success(t('Profile saved', '已保存'))
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : 'Save failed')
  } finally { savingProfile.value = false }
}

async function exportData() {
  exporting.value = true
  try {
    const resp = await api<{ data: Record<string, unknown> }>('/users/me/export')
    const blob = new Blob([JSON.stringify(resp.data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url; a.download = '25types-data.json'; a.click()
    URL.revokeObjectURL(url)
    ElMessage.success(t('Data exported', '数据已导出'))
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : 'Export failed')
  } finally { exporting.value = false }
}

async function deleteAccount() {
  deleting.value = true
  try {
    await api('/users/me', { method: 'DELETE' })
    auth.logout()
    router.push('/ask')
    ElMessage.info(t('Account deactivated', '账户已注销'))
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : 'Delete failed')
    deleting.value = false
  }
}

async function doRetry() {
  loading.value = true; error.value = ''
  try {
    await auth.fetchProfile()
    const u = auth.user
    if (u) {
      name.value = u.name; email.value = u.email
      isPublic.value = (u as Record<string, unknown>).is_public as boolean ?? false
      if (u.birth_info) birthInfo.value = { ...u.birth_info }
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load profile'
  } finally { loading.value = false }
}
</script>

<template>
  <div class="settings-page stagger-in">
    <div class="page-header">
      <h1>{{ t('Settings', '设置') }}</h1>
    </div>

    <!-- Not logged in -->
    <el-result
      v-if="!auth.isLoggedIn"
      icon="warning"
      :title="t('Please login', '请先登录')"
      :sub-title="t('Login to manage your account.', '登录后管理你的账户。')"
    >
      <template #extra>
        <a href="/login"><el-button type="primary" size="large">{{ t('Login', '登录') }}</el-button></a>
      </template>
    </el-result>

    <el-skeleton v-else-if="loading" :rows="12" animated />

    <el-result
      v-else-if="error"
      icon="error" title="Error" :sub-title="error"
    >
      <template #extra>
        <el-button type="primary" @click="doRetry">{{ t('Retry', '重试') }}</el-button>
      </template>
    </el-result>

    <div v-else class="space-y-5">
      <!-- Plan -->
      <div class="section-card">
        <div class="section-header">
          <span class="w-2 h-2 rounded-full" style="background:var(--wuxing-wood)"></span>
          {{ t('Subscription', '订阅方案') }}
        </div>
        <div class="section-body">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <el-tag effect="dark" round>Free</el-tag>
              <span class="text-sm text-gray-500">{{ t('3 free reports per day', '每日3次免费生成') }}</span>
            </div>
            <el-button type="primary" round disabled>{{ t('Upgrade to Pro', '升级 Pro') }}</el-button>
          </div>
        </div>
      </div>

      <!-- Account info -->
      <div class="section-card">
        <div class="section-header">
          <span class="w-2 h-2 rounded-full" style="background:var(--wuxing-water)"></span>
          {{ t('Account', '账户信息') }}
        </div>
        <div class="section-body">
          <div class="grid grid-cols-2 gap-6">
            <div>
              <div class="text-xs text-gray-400 mb-1">{{ t('Username', '用户名') }}</div>
              <div class="font-semibold text-lg">{{ auth.userName }}</div>
            </div>
            <div>
              <div class="text-xs text-gray-400 mb-1">Email</div>
              <div class="font-semibold text-lg">{{ auth.user?.email || '-' }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Profile form -->
      <div class="section-card">
        <div class="section-header">
          <span class="w-2 h-2 rounded-full" style="background:var(--wuxing-fire)"></span>
          {{ t('Edit Profile', '编辑资料') }}
        </div>
        <div class="section-body">
          <el-form label-position="top" size="default">
            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="Name">
                  <el-input v-model="name" :disabled="savingProfile" size="large" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="Email">
                  <el-input v-model="email" :disabled="savingProfile" size="large" />
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item :label="t('Public Profile', '公开 Profile')">
              <el-switch v-model="isPublic" :disabled="savingProfile" />
              <span class="ml-2 text-sm text-gray-500">{{ t('Allow others to find your profile', '允许他人查看你的类型资料') }}</span>
            </el-form-item>
            <el-divider />
            <div class="text-sm font-medium mb-3">{{ t('Birth Info', '出生信息') }}</div>
            <BirthInfoForm v-model="birthInfo" :disabled="savingProfile" :show-d-s-t="true" />
            <div class="mt-4">
              <el-button type="primary" size="large" round :loading="savingProfile" @click="saveProfile">
                {{ t('Save Changes', '保存修改') }}
              </el-button>
            </div>
          </el-form>
        </div>
      </div>

      <!-- Data -->
      <div class="section-card">
        <div class="section-header">
          <span class="w-2 h-2 rounded-full" style="background:var(--wuxing-metal)"></span>
          {{ t('Data & Privacy', '数据与隐私') }}
        </div>
        <div class="section-body">
          <div class="flex items-center gap-4">
            <el-button :loading="exporting" round @click="exportData">
              {{ t('Export My Data', '导出我的数据') }}
            </el-button>
            <el-popconfirm
              :title="t('This will permanently deactivate your account. Continue?', '这将永久注销你的账户。确认继续？')"
              :confirm-button-text="t('Delete', '注销')"
              :cancel-button-text="t('Cancel', '取消')"
              @confirm="deleteAccount"
            >
              <template #reference>
                <el-button type="danger" round :loading="deleting">
                  {{ t('Delete Account', '注销账户') }}
                </el-button>
              </template>
            </el-popconfirm>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
