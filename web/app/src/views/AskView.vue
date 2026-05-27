<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useReportsStore } from '@/stores/reports'
import { useAuthStore } from '@/stores/auth'
import { useProfileStore } from '@/stores/profile'
import { useLocaleStore } from '@/stores/locale'
import SceneChips from '@/components/ask/SceneChips.vue'
import DailySuggestion from '@/components/ask/DailySuggestion.vue'
import ReportList from '@/components/ask/ReportList.vue'
import ReportReader from '@/components/ask/ReportReader.vue'
import ReportInput from '@/components/ask/ReportInput.vue'
import { ElMessage } from 'element-plus'

const reports = useReportsStore()
const auth = useAuthStore()
const profile = useProfileStore()
const locale = useLocaleStore()
const { t } = locale

const question = ref('')
const activeScene = ref<string | null>(null)
const generating = ref(false)

const hasActiveReport = computed(() => !!reports.current || reports.streaming)

onMounted(async () => {
  if (auth.isLoggedIn) {
    await Promise.all([
      reports.fetchList(),
      reports.fetchDailySuggestion(),
      profile.fetchAll(),
    ])
  }
})

function selectScene(scene: string) {
  activeScene.value = scene
}

function askToday() {
  question.value = t('What does today hold for me?', '今天有什么需要注意的？')
  submitQuestion()
}

async function submitQuestion() {
  const q = question.value.trim()
  if (!q) return
  generating.value = true
  reports.clearStream()

  const scene = activeScene.value || 'general'

  // Build enriched engine_data for BaZi scenes.
  const engineData: Record<string, unknown> = {
    question: q,
    locale: locale.current,
  }

  if (profile.mingge) {
    engineData.day_master = profile.mingge.day_master_name || String(profile.mingge.day_master)
    engineData.day_master_element = profile.mingge.day_master_name
    engineData.strength = profile.mingge.strength
    engineData.yong_shen = profile.mingge.yong_shen
    engineData.pattern = profile.mingge.pattern
    engineData.element_distribution = formatElementCount(profile.mingge.element_count)
  }

  if (scene === 'dayun' && profile.dayun) {
    const d = profile.dayun
    engineData.dayun_table = d.dayun_interactions_text || formatDayunTable(d)
    engineData.current_dayun = d.pillars[d.current_pillar_index]?.ten_god + '(' + d.pillars[d.current_pillar_index]?.name + ')'
    engineData.dayun_interactions = d.dayun_interactions_text || ''
  }

  if (scene === 'liunian' && profile.liunian) {
    const l = profile.liunian
    engineData.year = String(l.year)
    engineData.year_name = l.year_name
    engineData.year_element = l.element
    engineData.ten_god = l.ten_god
    engineData.generates_restrains = l.generates ? '生我' : (l.restrains ? '克我' : '中和')
    engineData.current_dayun_info = l.current_dayun_info || ''
    engineData.natal_interactions = formatInteractions(l.natal_interactions)
    engineData.dayun_interactions = formatInteractions(l.dayun_interactions)
    engineData.three_layer_summary = l.three_layer_summary || ''
  }

  try {
    await reports.generate(scene, 'ask', engineData)
    if (reports.current) await reports.fetchList()
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : 'Generation failed')
  } finally {
    generating.value = false
    question.value = ''
  }
}

function formatElementCount(ec: Record<number, number>): string {
  const elNames: Record<number, string> = { 1: '木', 2: '火', 3: '土', 4: '金', 5: '水' }
  return Object.entries(ec).map(([k, v]) => (elNames[Number(k)] || k) + String(v)).join(' ')
}

function formatDayunTable(d: { start_age: number; pillars: { name: string; element: string; ten_god: string; age_start: number; age_end: number }[] }): string {
  return d.pillars.map(p => `${p.name}（${p.element}）${p.ten_god} ${p.age_start}-${p.age_end}岁`).join('\n')
}

function formatInteractions(interactions: { pillar_label: string; stem_rels: { relation: string }[]; branch_rels: { detail: string; type: string }[] }[]): string {
  if (!interactions || interactions.length === 0) return ''
  return interactions.map(p => {
    const lines = [p.pillar_label + ':']
    p.stem_rels.filter(s => s.relation && s.relation !== '无特殊关系').forEach(s => lines.push('  天干: ' + s.relation))
    p.branch_rels.filter(b => b.type !== '无').forEach(b => lines.push('  地支: ' + b.detail + ' (' + b.type + ')'))
    return lines.join('\n')
  }).join('\n\n')
}

function selectReport(r: { id: number }) {
  reports.fetchDetail(r.id)
}

function deleteReport(r: { id: number }) {
  reports.remove(r.id)
  ElMessage.success(t('Report deleted', '报告已删除'))
}

function reAsk() {
  if (reports.current?.question) {
    question.value = (reports.current.question as string) || ''
    activeScene.value = reports.current.scene
  }
  reports.clearStream()
}

function shareReport() {
  if (!reports.current) return
  const url = `${window.location.origin}/app#/ask`
  navigator.clipboard?.writeText(url).then(() => {
    ElMessage.success(t('Link copied', '链接已复制'))
  }).catch(() => {
    ElMessage.info(url)
  })
}
</script>

<template>
  <div class="ask-page stagger-in">
    <!-- Page header -->
    <div class="page-header">
      <h1>{{ t('Reports', '智能报告') }}</h1>
      <p>{{ t('AI-powered insights from your BaZi chart and 25types profile.', 'AI驱动的八字命盘 + 人格类型解读。引擎数据免费看，完整解读 ¥9.9。') }}</p>
    </div>

    <!-- Not logged in -->
    <el-result
      v-if="!auth.isLoggedIn"
      icon="warning"
      :title="t('Please login', '请先登录')"
      :sub-title="t('Login to generate personalized reports.', '登录后获取个性化解读报告。')"
    >
      <template #extra>
        <a href="/login"><el-button type="primary" size="large">{{ t('Login', '登录') }}</el-button></a>
      </template>
    </el-result>

    <!-- Two-panel layout -->
    <div v-else class="ask-layout">
      <!-- Main content -->
      <div class="ask-main">
        <!-- Default state: scene chips + daily suggestion -->
        <div v-if="!hasActiveReport" class="mb-6">
          <div class="mb-4">
            <SceneChips :active-scene="activeScene" @select="selectScene" />
          </div>
          <DailySuggestion
            :suggestion="reports.dailySuggestion"
            :loading="reports.dailyLoading"
            @ask-today="askToday"
          />
        </div>

        <!-- Report reader -->
        <div v-if="hasActiveReport" class="mb-4">
          <ReportReader
            :report="reports.current"
            :streaming="reports.streaming"
            :stream-text="reports.streamText"
            @re-ask="reAsk"
            @share="shareReport"
          />
        </div>

        <!-- Input bar -->
        <ReportInput
          :generating="generating"
          @submit="(q: string) => { question = q; submitQuestion() }"
        />
      </div>

      <!-- Sidebar: report list -->
      <div class="ask-sidebar">
        <ReportList
          :items="reports.items"
          :active-id="reports.current?.id ?? null"
          :loading="reports.loading"
          @select="selectReport"
          @delete="deleteReport"
        />
      </div>
    </div>
  </div>
</template>
