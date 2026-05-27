const locale = process.env.LOCALE || 'en';

const config = {
  en: {
    siteName: '25types',
    defaultTitle: '25types — Your Life Answer Engine',
    defaultDescription: 'Computable personality & relationship analysis. Who are you? Are you compatible? Which direction? Free assessment.',
    locale: 'en',
    alternateLocale: 'zh-CN',
    mingliApiHost: process.env.MINGLI_API_HOST || '',
  },
  'zh-CN': {
    siteName: '25types',
    defaultTitle: '25types — 你的人生答案引擎',
    defaultDescription: '可计算的性格与关系分析。你是什么样的人？你们合不合适？选哪个方向？免费评估。',
    locale: 'zh-CN',
    alternateLocale: 'en',
    mingliApiHost: process.env.MINGLI_API_HOST || '',
  },
};

module.exports = config[locale] || config.en;
