<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useLocaleStore } from '@/stores/locale'
import { api } from '@/api/client'
import { ElMessage } from 'element-plus'
import { elementName, elementColor } from '@shared/elements'
import NamingFormDialog from '@/components/naming/NamingFormDialog.vue'
import type { BirthInfo, NameResult } from '@shared/types'

const auth = useAuthStore()
const locale = useLocaleStore()
const { t } = locale

interface NamingEntry {
  id: string
  surname: string
  result: NameResult
  created_at: string
}

const entries = ref<NamingEntry[]>([])
const loading = ref(false)
const error = ref('')
const showForm = ref(false)
const showDrawer = ref(false)
const selectedEntry = ref<NamingEntry | null>(null)
const analyzing = ref(false)

onMounted(async () => {
  if (!auth.isLoggedIn) return
  loading.value = true
  try {
    const resp = await api<{ data: { items: { id: number; scene: string; title: string; engine_data: Record<string, unknown>; created_at: string }[] } }>('/reports?scene=naming')
    entries.value = (resp.data.items || []).map(r => ({
      id: String(r.id),
      surname: (r.engine_data as Record<string, unknown>)?.surname as string || '',
      result: (r.engine_data as Record<string, unknown>)?.result as NameResult,
      created_at: r.created_at,
    }))
  } catch {
    error.value = 'Failed to load naming history'
  } finally {
    loading.value = false
  }
})

async function handleSubmit(data: { surname: string; birth_info: Partial<BirthInfo> }) {
  analyzing.value = true
  try {
    const resp = await api<{ data: NameResult }>('/naming/analyze', {
      method: 'POST',
      body: { surname: data.surname, birth_info: data.birth_info, locale: locale.current },
    })
    const entry: NamingEntry = {
      id: Date.now().toString(),
      surname: data.surname,
      result: resp.data,
      created_at: new Date().toISOString(),
    }
    entries.value.unshift(entry)
    showForm.value = false
    selectedEntry.value = entry
    showDrawer.value = true
    ElMessage.success(t('Analysis complete', '分析完成'))
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : 'Analysis failed')
  } finally {
    analyzing.value = false
  }
}

async function doRetry() {
  loading.value = true; error.value = ''
  try {
    const resp = await api<{ data: { items: { id: number; scene: string; title: string; engine_data: Record<string, unknown>; created_at: string }[] } }>('/reports?scene=naming')
    entries.value = (resp.data.items || []).map(r => ({
      id: String(r.id), surname: (r.engine_data as Record<string, unknown>)?.surname as string || '',
      result: (r.engine_data as Record<string, unknown>)?.result as NameResult, created_at: r.created_at,
    }))
  } catch { error.value = 'Failed to load naming history' }
  finally { loading.value = false }
}

function viewEntry(entry: NamingEntry) {
  selectedEntry.value = entry
  showDrawer.value = true
}

const elementEntries = (ec: Record<number, number>) =>
  Object.entries(ec).map(([k, v]) => [Number(k), v] as const)
</script>

<template>
  <div class="naming-page stagger-in">
    <div class="page-header">
      <div class="flex items-center justify-between">
        <div>
          <h1>{{ t('Naming', '起名分析') }}</h1>
          <p>{{ t('BaZi-element-based name analysis with transparent reasoning.', '基于八字五行的透明推理起名分析，非黑箱推荐。') }}</p>
        </div>
        <el-button v-if="auth.isLoggedIn" type="primary" size="large" round @click="showForm = true">
          + {{ t('New Analysis', '新增起名') }}
        </el-button>
      </div>
    </div>

    <!-- Not logged in -->
    <el-result
      v-if="!auth.isLoggedIn"
      icon="warning"
      :title="t('Please login', '请先登录')"
      :sub-title="t('Login to use the naming analysis tool.', '登录后使用起名分析工具。')"
    >
      <template #extra>
        <a href="/login"><el-button type="primary" size="large">{{ t('Login', '登录') }}</el-button></a>
      </template>
    </el-result>

    <el-skeleton v-else-if="loading" :rows="6" animated />

    <el-result
      v-else-if="error"
      icon="error" title="Error" :sub-title="error"
    >
      <template #extra>
        <el-button type="primary" @click="doRetry">{{ t('Retry', '重试') }}</el-button>
      </template>
    </el-result>

    <el-empty
      v-else-if="entries.length === 0"
      description=""
    >
      <template #image>
        <div class="premium-empty">
          <div class="icon-wrap">📛</div>
          <div class="text-sm text-gray-500 mt-2">{{ t('No naming history yet.', '还没有起名方案。') }}</div>
        </div>
      </template>
      <template #default>
        <el-button type="primary" size="large" round @click="showForm = true">{{ t('Start Naming', '开始起名') }}</el-button>
      </template>
    </el-empty>

    <!-- Card grid -->
    <el-row v-else :gutter="16">
      <el-col v-for="e in entries" :key="e.id" :xs="24" :sm="12" :md="8" class="mb-4">
        <div class="naming-card" @click="viewEntry(e)">
          <div class="flex items-center justify-between mb-3">
            <span class="text-lg font-bold">{{ e.surname }}{{ t(' Family', '氏') }}</span>
            <span class="text-xs text-gray-400">{{ new Date(e.created_at).toLocaleDateString() }}</span>
          </div>
          <div class="flex gap-2">
            <el-tag size="small" effect="dark" round>{{ e.result?.yong_shen || '-' }}</el-tag>
            <el-tag size="small" effect="plain" round type="info">{{ e.result?.strategy || '-' }}</el-tag>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- Form dialog -->
    <NamingFormDialog v-model="showForm" @submit="handleSubmit" />

    <!-- Detail drawer -->
    <el-drawer
      v-model="showDrawer"
      :title="t('Naming Analysis', '起名分析')"
      size="480px"
    >
      <div v-if="selectedEntry?.result">
        <!-- Key metrics -->
        <div class="grid grid-cols-2 gap-3 mb-4">
          <div class="bond-stat">
            <div class="count text-base">{{ selectedEntry.result.yong_shen }}</div>
            <div class="label">{{ t('Yong Shen', '用神') }}</div>
          </div>
          <div class="bond-stat">
            <div class="count text-base">{{ selectedEntry.result.strategy }}</div>
            <div class="label">{{ t('Strategy', '策略') }}</div>
          </div>
        </div>

        <!-- Element distribution -->
        <div v-if="selectedEntry.result.element_count" class="mb-4">
          <div class="text-sm font-medium mb-2">{{ t('Elements', '五行分布') }}</div>
          <div v-for="[code, count] in elementEntries(selectedEntry.result.element_count)" :key="code" class="element-bar-wrap">
            <span class="w-8 text-xs text-right">{{ elementName(code, locale.current) }}</span>
            <div class="element-bar-track">
              <div class="element-bar-fill" :style="{ width: count * 20 + '%', backgroundColor: elementColor(code) }"></div>
            </div>
            <span class="text-xs w-4 text-gray-400">{{ count }}</span>
          </div>
        </div>

        <!-- Zodiac hints -->
        <div v-if="selectedEntry.result.zodiac_hint" class="mb-4">
          <div class="text-sm font-medium mb-2">{{ t('Zodiac', '生肖') }}: {{ selectedEntry.result.zodiac_hint.animal }}</div>
          <div class="flex flex-wrap gap-1">
            <span class="text-xs text-gray-400 mr-1">{{ t('Preferred', '宜用') }}:</span>
            <el-tag v-for="r in selectedEntry.result.zodiac_hint.preferred_radicals" :key="r" size="small" effect="dark" type="success" round>{{ r }}</el-tag>
          </div>
        </div>

        <!-- CTA -->
        <el-card shadow="never" class="text-center mt-4 premium-card">
          <div class="py-2">
            <p class="font-semibold mb-1">{{ t('Get Full Naming Report', '获取完整起名方案') }}</p>
            <p class="text-sm text-gray-400 mb-3">{{ t('Top 20 candidate names with full element analysis.', 'Top 20 候选名字含完整五行分析。') }}</p>
            <el-button type="primary" round @click="showDrawer = false; showForm = true">
              {{ t('Generate Report', '生成完整报告') }}
            </el-button>
          </div>
        </el-card>
      </div>
    </el-drawer>
  </div>
</template>
