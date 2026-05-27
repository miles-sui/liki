import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/api/client'
import type { ChartData, MinggeData, DayunData, LiunianData, DeficiencyData, FlowYearlyData } from '@shared/types'

export const useProfileStore = defineStore('profile', () => {
  const chart = ref<ChartData | null>(null)
  const mingge = ref<MinggeData | null>(null)
  const dayun = ref<DayunData | null>(null)
  const liunian = ref<LiunianData | null>(null)
  const deficiency = ref<DeficiencyData | null>(null)
  const flow = ref<FlowYearlyData | null>(null)
  const loading = ref(false)
  const error = ref('')

  const dayMaster = computed(() => chart.value?.day_master ?? mingge.value?.day_master ?? 0)
  const elementCount = computed(() => chart.value?.element_count ?? mingge.value?.element_count ?? {})

  async function computeChart(birthInfo: Record<string, unknown>) {
    loading.value = true
    error.value = ''
    try {
      const resp = await api<{ data: ChartData }>('/bazi/chart', {
        method: 'POST',
        body: birthInfo,
      })
      chart.value = resp.data
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Failed to compute chart'
    } finally {
      loading.value = false
    }
  }

  async function fetchMingge() {
    try {
      const resp = await api<{ data: MinggeData }>('/bazi/mingge')
      mingge.value = resp.data
    } catch { /* no birth info saved */ }
  }

  async function fetchDayun() {
    try {
      const resp = await api<{ data: DayunData }>('/bazi/dayun')
      dayun.value = resp.data
    } catch { /* no birth info saved */ }
  }

  async function fetchLiunian() {
    try {
      const resp = await api<{ data: LiunianData }>('/bazi/liunian')
      liunian.value = resp.data
    } catch { /* no birth info saved */ }
  }

  async function fetchDeficiency() {
    try {
      const resp = await api<{ data: DeficiencyData }>('/bazi/deficiency')
      deficiency.value = resp.data
    } catch { /* no birth info saved */ }
  }

  async function fetchFlow() {
    try {
      const resp = await api<{ data: FlowYearlyData }>('/flow/yearly')
      flow.value = resp.data
    } catch { /* optional */ }
  }

  async function fetchAll() {
    loading.value = true
    error.value = ''
    try {
      await Promise.all([
        fetchMingge(),
        fetchDayun(),
        fetchLiunian(),
        fetchDeficiency(),
        fetchFlow(),
      ])
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Failed to load profile'
    } finally {
      loading.value = false
    }
  }

  return {
    chart, mingge, dayun, liunian, deficiency, flow,
    loading, error, dayMaster, elementCount,
    computeChart, fetchMingge, fetchDayun, fetchLiunian,
    fetchDeficiency, fetchFlow, fetchAll,
  }
})
