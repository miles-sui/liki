<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useProfileStore } from '@/stores/profile'
import { useLocaleStore } from '@/stores/locale'
import { api } from '@/api/client'
import { ElMessage } from 'element-plus'
import BirthInfoForm from '@/components/shared/BirthInfoForm.vue'
import TypeIdentityCard from '@/components/overview/TypeIdentityCard.vue'
import BaziChartCard from '@/components/overview/BaziChartCard.vue'
import FlowRiver from '@/components/overview/FlowRiver.vue'
import DayunPreview from '@/components/overview/DayunPreview.vue'
import BondsPreview from '@/components/overview/BondsPreview.vue'
import type { BirthInfo } from '@shared/types'

const auth = useAuthStore()
const profile = useProfileStore()
const { t } = useLocaleStore()

const showBirthInfoDialog = ref(false)
const birthInfo = ref<Partial<BirthInfo>>({
  year: 2000, month: 1, day: 1, hour: 12, minute: 0,
  longitude: 120, timezone: 120, is_dst: false, gender: 'male',
})
const savingBirthInfo = ref(false)

const hasProfile = computed(() => !!profile.mingge)
const hasBazi = computed(() => !!profile.mingge?.chart)

onMounted(async () => {
  if (!auth.isLoggedIn) return
  if (auth.user?.birth_info) birthInfo.value = { ...auth.user.birth_info }
  await profile.fetchAll()
})

async function saveBirthInfo() {
  savingBirthInfo.value = true
  try {
    await api('/users/me', { method: 'PATCH', body: { birth_info: birthInfo.value } })
    await Promise.all([auth.fetchProfile(), profile.fetchAll()])
    showBirthInfoDialog.value = false
    ElMessage.success(t('Birth info saved', '出生信息已保存'))
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : 'Save failed')
  } finally {
    savingBirthInfo.value = false
  }
}
</script>

<template>
  <div class="overview-page stagger-in">
    <!-- Page header -->
    <div class="page-header">
      <h1>{{ hasProfile ? auth.userName : t('My Chart', '我的命盘') }}</h1>
      <p>{{ t('BaZi chart + 25types profile, dual-engine visualization.', '八字排盘 + 25types 类型，双引擎数据统一展示。') }}</p>
    </div>

    <!-- Not logged in -->
    <el-result
      v-if="!auth.isLoggedIn"
      icon="warning"
      :title="t('Please login', '请先登录')"
      :sub-title="t('Login to view your chart.', '登录后查看你的命盘。')"
    >
      <template #extra>
        <a href="/login"><el-button type="primary" size="large">{{ t('Login', '登录') }}</el-button></a>
      </template>
    </el-result>

    <!-- Loading -->
    <el-row v-else-if="profile.loading" :gutter="16">
      <el-col v-for="i in 4" :key="i" :xs="24" :md="12" class="mb-4">
        <el-card shadow="never" class="premium-card"><el-skeleton :rows="4" animated /></el-card>
      </el-col>
    </el-row>

    <!-- Error -->
    <el-result
      v-else-if="profile.error"
      icon="error" title="Error" :sub-title="profile.error"
    >
      <template #extra>
        <el-button type="primary" @click="profile.fetchAll()">{{ t('Retry', '重试') }}</el-button>
      </template>
    </el-result>

    <!-- Dashboard -->
    <template v-else>
      <!-- Hero stats -->
      <el-row :gutter="16" class="mb-4">
        <el-col :xs="12" :md="6" class="mb-3">
          <div class="bond-stat">
            <div class="count" :style="{color: 'var(--wuxing-water)'}">{{ profile.mingge?.day_master_name || '-' }}</div>
            <div class="label">{{ t('Day Master', '日主') }}</div>
          </div>
        </el-col>
        <el-col :xs="12" :md="6" class="mb-3">
          <div class="bond-stat">
            <div class="count" :style="{color: 'var(--wuxing-fire)'}">{{ profile.mingge?.yong_shen || '-' }}</div>
            <div class="label">{{ t('Yong Shen', '用神') }}</div>
          </div>
        </el-col>
        <el-col :xs="12" :md="6" class="mb-3">
          <div class="bond-stat">
            <div class="count" :style="{color: 'var(--wuxing-wood)'}">{{ profile.mingge?.pattern || '-' }}</div>
            <div class="label">{{ t('Pattern', '格局') }}</div>
          </div>
        </el-col>
        <el-col :xs="12" :md="6" class="mb-3">
          <div class="bond-stat">
            <div class="count" :style="{color: 'var(--wuxing-metal)'}">{{ profile.mingge?.strength || '-' }}</div>
            <div class="label">{{ t('Strength', '强弱') }}</div>
          </div>
        </el-col>
      </el-row>

      <!-- Main content grid -->
      <el-row :gutter="16">
        <el-col :xs="24" :md="12" class="mb-4">
          <TypeIdentityCard :mingge="profile.mingge" :loading="profile.loading" />
        </el-col>
        <el-col :xs="24" :md="12" class="mb-4">
          <BaziChartCard
            :mingge="profile.mingge"
            :loading="profile.loading"
            @add-birth-info="showBirthInfoDialog = true"
          />
        </el-col>
      </el-row>

      <el-row :gutter="16">
        <el-col :span="24" class="mb-4">
          <FlowRiver :data="profile.flow" />
        </el-col>
      </el-row>

      <el-row :gutter="16">
        <el-col :xs="24" :md="12" class="mb-4">
          <DayunPreview :dayun="profile.dayun" :loading="profile.loading" />
        </el-col>
        <el-col :xs="24" :md="12" class="mb-4">
          <BondsPreview />
        </el-col>
      </el-row>
    </template>

    <!-- Birth info dialog -->
    <el-dialog
      v-model="showBirthInfoDialog"
      :title="t('Add Birth Info', '录入出生信息')"
      width="600px"
      :close-on-click-modal="false"
    >
      <BirthInfoForm v-model="birthInfo" :show-d-s-t="true" />
      <template #footer>
        <el-button @click="showBirthInfoDialog = false">{{ t('Cancel', '取消') }}</el-button>
        <el-button type="primary" :loading="savingBirthInfo" @click="saveBirthInfo">
          {{ t('Save', '保存') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>
