// Liki shared UI components — Web Components (light DOM, Tailwind compatible)
(function(){
  var LOCALES = ['zh', 'hk', 'en'];
  var LANG_NAME = { zh: '简体中文', hk: '繁體中文', en: 'English' };

  var SVG = '<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><title>Language</title><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>';

  function esc(s) { return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;'); }

  function setLang(loc) {
    if (!loc || !i18next || loc === i18next.language) return;
    var path = location.pathname.replace(/^\/(zh|hk|en)\/?/, '/');
    location.href = '/' + loc + (path === '/' ? '/' : path);
  }
  window.__setLang = setLang;

  // ── <lang-switcher> ──

  class LangSwitcher extends HTMLElement {
    connectedCallback() {
      var self = this;
      this.style.position = 'relative';
      build();

      function build() {
        var cur = (i18next && i18next.language) || 'zh';
        self.innerHTML =
          '<button data-lang-toggle class="flex items-center gap-1.5 text-amber-400 hover:text-amber-100 transition-colors bg-transparent border-0 cursor-pointer px-3 py-1.5 rounded" aria-label="Switch language">' + SVG + '</button>' +
          '<div data-lang-dropdown class="hidden absolute right-0 top-full mt-1 bg-white border border-stone-200 rounded-lg shadow-lg py-1 z-50 min-w-[130px]">' +
          LOCALES.map(function(loc) {
            var cls = loc === cur ? 'bg-amber-50 text-amber-700 font-medium block px-4 py-2 text-sm transition-colors' : 'text-stone-600 hover:bg-stone-50 block px-4 py-2 text-sm transition-colors';
            return '<a href="#" data-lang-option="' + loc + '" class="' + cls + '">' + LANG_NAME[loc] + '</a>';
          }).join('') + '</div>';

        var toggle = self.querySelector('[data-lang-toggle]');
        var dropdown = self.querySelector('[data-lang-dropdown]');
        if (!toggle || !dropdown) return;

        toggle.addEventListener('click', function(e) {
          e.preventDefault(); e.stopPropagation();
          dropdown.classList.toggle('hidden');
        });
        dropdown.querySelectorAll('[data-lang-option]').forEach(function(opt) {
          opt.addEventListener('click', function(e) {
            e.preventDefault();
            dropdown.classList.add('hidden');
            setLang(opt.dataset.langOption);
          });
        });
      }

      // Close dropdown on outside click
      document.addEventListener('click', function(e) {
        if (!self.contains(e.target)) {
          var dd = self.querySelector('[data-lang-dropdown]');
          if (dd) dd.classList.add('hidden');
        }
      });
    }
  }
  customElements.define('lang-switcher', LangSwitcher);

  // ── <liki-header> ──

  class LikiHeader extends HTMLElement {
    connectedCallback() {
      var type = this.getAttribute('type') || 'back';
      var titleKey = this.getAttribute('title-key') || '';
      var titleFallback = this.getAttribute('title') || '';
      var size = this.getAttribute('size') || 'xl';
      var sizeCls = size === 'lg' ? 'text-lg' : 'text-xl';

      var titleHtml;
      if (titleKey) {
        titleHtml = '<h1 class="' + sizeCls + ' font-bold text-amber-100" data-i18n="' + titleKey + '">' + esc(titleFallback) + '</h1>';
      } else {
        titleHtml = '<h1 class="' + sizeCls + ' font-bold text-amber-100" id="report-header-title">' + esc(titleFallback) + '</h1>';
      }

      if (type === 'brand') {
        this.innerHTML =
          '<header class="header-dark border-b border-stone-700">' +
          '<div class="max-w-4xl mx-auto px-4 py-6 flex items-center justify-between">' +
          '<div><h1 class="text-3xl font-brand text-amber-100 tracking-tight" data-i18n="site.name">' + esc(titleFallback) + '</h1></div>' +
          '<lang-switcher></lang-switcher>' +
          '</div></header>';
      } else {
        this.innerHTML =
          '<header class="header-dark border-b border-stone-700">' +
          '<div class="max-w-4xl mx-auto px-4 py-4 flex items-center justify-between">' +
          '<a href="/" class="back-link" data-i18n="site.backHome">← 返回首页</a>' +
          titleHtml +
          '<lang-switcher></lang-switcher>' +
          '</div></header>';
      }
    }
  }
  customElements.define('liki-header', LikiHeader);

  // ── <liki-footer> ──

  class LikiFooter extends HTMLElement {
    connectedCallback() {
      var cls = this.hasAttribute('hide-mobile') ? 'border-t border-stone-100 mt-6 py-8 hide-mobile' : 'border-t border-stone-100 mt-6 py-8';
      this.innerHTML =
        '<footer class="' + cls + '">' +
        '<div class="max-w-4xl mx-auto px-4 flex flex-wrap items-center justify-between gap-4">' +
        '<div class="text-left">' +
        '<p class="text-stone-600 text-sm" data-i18n="index.footer">灵机 Liki · AI命理助手</p>' +
        '<p class="text-stone-500 text-xs mt-0.5" data-i18n="site.footer">&copy; 2026 Liki. All rights reserved.</p>' +
        '</div>' +
        '<p class="text-xs text-stone-500 flex flex-wrap gap-3">' +
        '<a href="about.html" class="hover:text-stone-600 transition-colors" data-i18n="footer.about">关于我们</a>' +
        '<a href="privacy.html" class="hover:text-stone-600 transition-colors" data-i18n="footer.privacy">隐私政策</a>' +
        '<a href="terms.html" class="hover:text-stone-600 transition-colors" data-i18n="footer.terms">服务条款</a>' +
        '<a href="disclaimer.html" class="hover:text-stone-600 transition-colors" data-i18n="footer.disclaimer">免责声明</a>' +
        '<a href="contact.html" class="hover:text-stone-600 transition-colors" data-i18n="footer.contact">联系我们</a>' +
        '</p></div></footer>';
    }
  }
  customElements.define('liki-footer', LikiFooter);

  // ── <print-cover> ──

  class PrintCover extends HTMLElement {
    connectedCallback() {
      var product = this.getAttribute('product') || '';
      this.innerHTML =
        '<div class="print-cover no-print" style="display:none;">' +
        '<h1 class="font-brand">灵机 Liki</h1>' +
        '<div class="pc-sub" id="print-cover-sub">' + esc(product) + '</div>' +
        '<div class="pc-date" id="print-date"></div></div>';
    }
  }
  customElements.define('print-cover', PrintCover);

  // ── <print-brand> ──

  class PrintBrand extends HTMLElement {
    connectedCallback() {
      this.innerHTML = '<div class="print-brand no-print" style="display:none;" id="print-brand" data-i18n="report.printBrand">灵机 Liki · 命理報告</div>';
    }
  }
  customElements.define('print-brand', PrintBrand);

  // ── <print-bar> ──

  class PrintBar extends HTMLElement {
    connectedCallback() {
      var variant = this.getAttribute('variant') || 'static';
      var pb = variant === 'dynamic'
        ? '<button id="btn-print" class="btn-print">🖨 打印 / 保存 PDF</button>'
        : '<button class="btn-print" onclick="window.print()" data-i18n="report.print">🖨 打印 / 保存 PDF</button>';
      var sb = variant === 'dynamic'
        ? '<button id="btn-share" class="btn-share">↗ 分享</button>'
        : '<button class="btn-share" onclick="shareDemo()" data-i18n="report.share">↗ 分享</button>';
      this.innerHTML = '<div class="print-bar no-print" style="display:flex;gap:.5rem;justify-content:center;">' + pb + sb + '</div>';
    }
  }
  customElements.define('print-bar', PrintBar);

  // ── <sample-banner> ──

  class SampleBanner extends HTMLElement {
    connectedCallback() {
      this.innerHTML = '<div class="max-w-4xl mx-auto px-4 pt-2"><p class="text-xs text-stone-500 text-center" data-i18n="demo.sampleNote">此为示例报告，实际内容根据您的命盘生成</p></div>';
    }
  }
  customElements.define('sample-banner', SampleBanner);

  // ── <security-section> ──

  class SecuritySection extends HTMLElement {
    connectedCallback() {
      this.innerHTML =
        '<section class="mt-12">' +
        '<h2 class="text-xl font-bold text-stone-800 mb-6 text-center">安全与隐私</h2>' +
        '<div class="trust-badges">' +
        '<div class="trust-badge">🔒 SSL/TLS 加密传输</div>' +
        '<div class="trust-badge">🛡 隐私数据保护</div>' +
        '<div class="trust-badge">🔄 支付后自动清除敏感数据</div>' +
        '</div></section>';
    }
  }
  customElements.define('security-section', SecuritySection);

  // ── <cta-section> ──

  class CtaSection extends HTMLElement {
    connectedCallback() {
      var p = this.getAttribute('product') || '';
      var fallbacks = {
        chart:  ['获取您的专属八字报告', 'AI 深度解读您的命理格局、用神喜忌、大运流年，仅需 $9.90'],
        bond:   ['获取您和 TA 的专属合盘报告', 'AI 深度解读你们的合盘关系，仅需 $19.90'],
        naming: ['获取您的专属起名报告', 'AI 精选吉祥名字，仅需 $29.90'],
      };
      var fb = fallbacks[p] || ['', ''];
      this.innerHTML =
        '<div class="cta-bar no-print">' +
        '<h2 data-i18n="demo.ctaHeading.' + p + '">' + esc(fb[0]) + '</h2>' +
        '<p data-i18n="demo.ctaDesc.' + p + '">' + esc(fb[1]) + '</p>' +
        '<a href="/chat.html" class="btn btn-primary" data-i18n="demo.ctaButton">立即获取 →</a></div>';
    }
  }
  customElements.define('cta-section', CtaSection);
})();
