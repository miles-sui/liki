import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, getAuthToken } from '@/api/client'
import type { Report, DailySuggestionData } from '@shared/types'

export const useReportsStore = defineStore('reports', () => {
  const items = ref<Report[]>([])
  const current = ref<Report | null>(null)
  const loading = ref(false)
  const streaming = ref(false)
  const streamText = ref('')
  const dailySuggestion = ref<DailySuggestionData | null>(null)
  const dailyLoading = ref(false)

  async function fetchList() {
    loading.value = true
    try {
      const resp = await api<{ data: { items: Report[]; total: number } }>('/reports')
      items.value = resp.data.items
    } catch {
      items.value = []
    } finally {
      loading.value = false
    }
  }

  async function fetchDetail(id: number) {
    loading.value = true
    try {
      const resp = await api<{ data: Report }>(`/reports/${id}`)
      current.value = resp.data
    } finally {
      loading.value = false
    }
  }

  async function generate(scene: string, subScene: string, engineData: Record<string, unknown>) {
    streaming.value = true
    streamText.value = ''
    try {
      const token = getAuthToken()
      const resp = await fetch('/api/reports', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify({ scene, sub_scene: subScene, engine_data: engineData }),
      })
      if (!resp.ok) {
        const err = await resp.json().catch(() => ({}))
        throw new Error(err?.error?.message || `HTTP ${resp.status}`)
      }
      const reader = resp.body?.getReader()
      if (!reader) throw new Error('No response stream')
      const decoder = new TextDecoder()
      let buffer = ''
      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''
        for (const line of lines) {
          if (line.startsWith('data: ')) {
            try {
              const data = JSON.parse(line.slice(6))
              if (data.text) streamText.value += data.text
              if (data.report_id) {
                await fetchDetail(data.report_id)
              }
            } catch { /* skip parse errors */ }
          }
        }
      }
    } catch (e: unknown) {
      streamText.value += '\n\n[Error: ' + (e instanceof Error ? e.message : 'stream failed') + ']'
    } finally {
      streaming.value = false
    }
  }

  function clearStream() {
    streamText.value = ''
    streaming.value = false
    current.value = null
  }

  async function fetchDailySuggestion() {
    dailyLoading.value = true
    try {
      const resp = await api<{ data: DailySuggestionData }>('/daily/suggestion')
      dailySuggestion.value = resp.data
    } catch {
      dailySuggestion.value = null
    } finally {
      dailyLoading.value = false
    }
  }

  async function remove(id: number) {
    await api(`/reports/${id}`, { method: 'DELETE' })
    items.value = items.value.filter(r => r.id !== id)
    if (current.value?.id === id) current.value = null
  }

  return {
    items, current, loading, streaming, streamText,
    dailySuggestion, dailyLoading,
    fetchList, fetchDetail, generate, clearStream,
    fetchDailySuggestion, remove,
  }
})
