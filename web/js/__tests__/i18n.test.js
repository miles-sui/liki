import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// Copy MSG translations (subset) from i18n JSON for test use
const MSG = {
  'zh-Hans': {
    'site.name': '灵机 Liki',
    'site.tagline': 'AI命理助手',
    'form.year': '出生年',
    'gender.male': '男',
    'chart.submit': '开始排盘分析',
    'lang.switch': 'English',
    'error.requestFailed': '请求失败',
    'naming.wuge': '五格: 天{0} 人{1} 地{2} 外{3} 总{4}',
  },
  'zh-Hant': {
    'site.name': '靈機 Liki',
    'site.tagline': 'AI命理助手',
    'form.year': '出生年',
    'gender.male': '男',
    'chart.submit': '開始排盤分析',
    'lang.switch': 'English',
    'error.requestFailed': '請求失敗',
    'naming.wuge': '五格: 天{0} 人{1} 地{2} 外{3} 總{4}',
  },
  en: {
    'site.name': 'Liki',
    'site.tagline': 'AI Chinese Metaphysics Assistant',
    'form.year': 'Year',
    'gender.male': 'Male',
    'chart.submit': 'Analyze Chart',
    'lang.switch': '中文',
    'error.requestFailed': 'Request failed',
    'naming.wuge': 'Wu Ge: Heaven{0} Person{1} Earth{2} Outer{3} Total{4}',
  },
};

// ---- Function copies from i18n.js ----

function detectLang(pathname, navigatorLanguage) {
  const m = pathname.match(/^\/(zh-Hans|zh-Hant|en)\b/);
  if (m) return m[1];
  const nav = (navigatorLanguage || 'en');
  if (nav === 'zh-HK' || nav === 'zh-TW' || nav === 'zh-MO') return 'zh-Hant';
  if (nav.startsWith('zh')) return 'zh-Hans';
  return 'en';
}

function t(lang, key, params) {
  let s = (MSG[lang] && MSG[lang][key]) || MSG['zh-Hans'][key] || key;
  if (params) {
    for (let i = 0; i < params.length; i++) {
      s = s.replace('{' + i + '}', params[i]);
    }
  }
  return s;
}

function setLang(l, pathname, localStorage) {
  localStorage.setItem('lingji-lang', l);
  const path = pathname.replace(/^\/(zh-Hans|zh-Hant|en)\/?/, '/');
  return '/' + l + (path === '/' ? '/' : path);
}

// ---- Tests ----

describe('detectLang', () => {
  it('returns "zh-Hans" for /zh-Hans/ paths', () => {
    expect(detectLang('/zh-Hans/')).toBe('zh-Hans');
    expect(detectLang('/zh-Hans/chat.html')).toBe('zh-Hans');
  });

  it('returns "zh-Hant" for /zh-Hant/ paths', () => {
    expect(detectLang('/zh-Hant/')).toBe('zh-Hant');
    expect(detectLang('/zh-Hant/chat.html')).toBe('zh-Hant');
  });

  it('returns "en" for /en/ paths', () => {
    expect(detectLang('/en/')).toBe('en');
    expect(detectLang('/en/chat.html')).toBe('en');
  });

  it('falls back to navigator.language when path has no locale prefix', () => {
    expect(detectLang('/chat.html', 'zh-CN')).toBe('zh-Hans');
    expect(detectLang('/chat.html', 'zh-TW')).toBe('zh-Hant');
    expect(detectLang('/chat.html', 'zh-HK')).toBe('zh-Hant');
    expect(detectLang('/chat.html', 'zh-MO')).toBe('zh-Hant');
    expect(detectLang('/chat.html', 'zh-SG')).toBe('zh-Hans');
    expect(detectLang('/chat.html', 'en-US')).toBe('en');
    expect(detectLang('/chat.html', 'fr')).toBe('en');
  });

  it('defaults to "en" when no navigator.language', () => {
    expect(detectLang('/chat.html', undefined)).toBe('en');
    expect(detectLang('/chat.html', '')).toBe('en');
  });

  it('handles root path', () => {
    expect(detectLang('/', 'en')).toBe('en');
    expect(detectLang('/', 'zh-CN')).toBe('zh-Hans');
    expect(detectLang('/', 'zh-HK')).toBe('zh-Hant');
  });
});

describe('t', () => {
  it('returns the translated string for a valid key in current lang', () => {
    expect(t('en', 'site.name', null)).toBe('Liki');
    expect(t('zh-Hans', 'site.name', null)).toBe('灵机 Liki');
    expect(t('zh-Hant', 'site.name', null)).toBe('靈機 Liki');
  });

  it('falls back to MSG.zh when current lang lacks the key', () => {
    MSG['zh-Hans']['test.zhOnly'] = '仅中文测试';
    expect(t('en', 'test.zhOnly', null)).toBe('仅中文测试');
    expect(t('zh-Hant', 'test.zhOnly', null)).toBe('仅中文测试');
    delete MSG['zh-Hans']['test.zhOnly'];
  });

  it('falls back to the key itself when no translation exists', () => {
    expect(t('en', 'nonexistent.key', null)).toBe('nonexistent.key');
    expect(t('zh-Hans', 'nonexistent.key', null)).toBe('nonexistent.key');
    expect(t('zh-Hant', 'nonexistent.key', null)).toBe('nonexistent.key');
  });

  it('replaces {0}, {1}, {2} etc. with provided params', () => {
    expect(t('zh-Hans', 'naming.wuge', ['5', '10', '15', '8', '12'])).toBe('五格: 天5 人10 地15 外8 总12');
    expect(t('zh-Hant', 'naming.wuge', ['5', '10', '15', '8', '12'])).toBe('五格: 天5 人10 地15 外8 總12');
    expect(t('en', 'naming.wuge', ['5', '10', '15', '8', '12'])).toBe('Wu Ge: Heaven5 Person10 Earth15 Outer8 Total12');
  });

  it('handles partial params gracefully (missing index in template)', () => {
    const result = t('en', 'site.name', ['extra']);
    expect(result).toBe('Liki');
  });

  it('handles undefined params', () => {
    expect(t('en', 'site.name', undefined)).toBe('Liki');
    expect(t('en', 'site.name', null)).toBe('Liki');
  });

  it('handles empty array params', () => {
    expect(t('en', 'site.name', [])).toBe('Liki');
  });
});

describe('setLang', () => {
  let storage;

  beforeEach(() => {
    storage = { _data: {}, getItem(k) { return this._data[k] || null; }, setItem(k, v) { this._data[k] = v; } };
  });

  it('stores the language preference in localStorage', () => {
    setLang('en', '/zh-Hans/chat.html', storage);
    expect(storage._data['lingji-lang']).toBe('en');
  });

  it('returns redirect URL for switching to en from zh', () => {
    expect(setLang('en', '/zh-Hans/chat.html', storage)).toBe('/en/chat.html');
  });

  it('returns redirect URL for switching to zh from en', () => {
    expect(setLang('zh-Hans', '/en/chat.html', storage)).toBe('/zh-Hans/chat.html');
  });

  it('returns redirect URL for switching to hk from zh', () => {
    expect(setLang('zh-Hant', '/zh-Hans/chat.html', storage)).toBe('/zh-Hant/chat.html');
  });

  it('returns redirect URL for switching to zh from hk', () => {
    expect(setLang('zh-Hans', '/zh-Hant/chat.html', storage)).toBe('/zh-Hans/chat.html');
  });

  it('handles root path', () => {
    expect(setLang('zh-Hans', '/en/', storage)).toBe('/zh-Hans/');
    expect(setLang('zh-Hant', '/', storage)).toBe('/zh-Hant/');
  });
});
