// Shared element constants and helpers.

export const ELEMENTS: Record<number, { code: string; name: string; nameZh: string; color: string }> = {
  1: { code: 'W', name: 'Wood',  nameZh: '木', color: '#1A7A6F' },
  2: { code: 'F', name: 'Fire',  nameZh: '火', color: '#CB3B2D' },
  3: { code: 'E', name: 'Earth', nameZh: '土', color: '#9A7410' },
  4: { code: 'M', name: 'Metal', nameZh: '金', color: '#8E8880' },
  5: { code: 'R', name: 'Water', nameZh: '水', color: '#1C2238' },
}

export const STEM_NAMES = ['', '甲', '乙', '丙', '丁', '戊', '己', '庚', '辛', '壬', '癸']
export const BRANCH_NAMES = ['', '子', '丑', '寅', '卯', '辰', '巳', '午', '未', '申', '酉', '戌', '亥']

export function formatStem(s: number): string {
  return STEM_NAMES[s] || '?'
}

export function formatBranch(b: number): string {
  return BRANCH_NAMES[b] || '?'
}

export function formatPillar(s: number, b: number): string {
  return formatStem(s) + formatBranch(b)
}

export function elementName(code: number, locale: string): string {
  const el = ELEMENTS[code]
  if (!el) return '?'
  return locale === 'zh-CN' ? el.nameZh : el.name
}

export function elementColor(code: number): string {
  return ELEMENTS[code]?.color || '#888'
}

const elementNameToIndex: Record<string, number> = {
  wood: 1, fire: 2, earth: 3, metal: 4, water: 5,
}

export function elementIndex(name: string): number {
  return elementNameToIndex[name.toLowerCase()] || 0
}
