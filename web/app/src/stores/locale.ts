import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

function detectLocale(): 'en' | 'zh-CN' {
  const stored = localStorage.getItem('locale') as 'en' | 'zh-CN' | null
  if (stored === 'en' || stored === 'zh-CN') return stored
  const win = (window as any).CURRENT_LOCALE
  if (win === 'en' || win === 'zh-CN') return win
  return 'en'
}

export const useLocaleStore = defineStore('locale', () => {
  const current = ref<'en' | 'zh-CN'>(detectLocale())

  const isZh = computed(() => current.value === 'zh-CN')

  function setLocale(locale: 'en' | 'zh-CN') {
    current.value = locale
    localStorage.setItem('locale', locale)
    try { (window as any).CURRENT_LOCALE = locale } catch (_) { /* SPA runs standalone */ }
  }

  function toggle() {
    setLocale(current.value === 'en' ? 'zh-CN' : 'en')
  }

  function t(en: string, zh: string): string {
    return current.value === 'zh-CN' ? zh : en
  }

  return { current, isZh, setLocale, toggle, t }
})
