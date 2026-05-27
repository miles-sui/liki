import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  const dark = ref(localStorage.getItem('theme') === 'dark')

  function toggle() {
    dark.value = !dark.value
    localStorage.setItem('theme', dark.value ? 'dark' : 'light')
    applyTheme()
  }

  function applyTheme() {
    document.documentElement.setAttribute('data-theme', dark.value ? 'wuxing-dark' : 'wuxing')
    document.documentElement.classList.toggle('dark', dark.value)
  }

  // Apply on init
  applyTheme()

  return { dark, toggle }
})
