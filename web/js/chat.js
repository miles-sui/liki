const { createApp, ref, reactive, computed, watch, nextTick, onMounted, onUnmounted } = Vue;

// ── composables ──

function useSessionStorage(key, fallback) {
  const stored = sessionStorage.getItem(key);
  let initial = fallback;
  if (stored != null) {
    try { initial = JSON.parse(stored); } catch (_) { initial = stored; }
  }
  const v = ref(initial);
  watch(v, (val) => {
    if (val == null || val === '') sessionStorage.removeItem(key);
    else sessionStorage.setItem(key, JSON.stringify(val));
  });
  return v;
}

function useSSE(t) {
  let ctrl = null;

  function parseEvents(buf, onEvent) {
    for (const event of buf.split('\n\n')) {
      for (const line of event.split('\n')) {
        if (!line.startsWith('data: ')) continue;
        const raw = line.slice(6).trim();
        if (!raw || raw === '[DONE]') continue;
        try { onEvent(JSON.parse(raw)); } catch (_) {}
      }
    }
  }

  async function readStream(reader, onEvent) {
    const decoder = new TextDecoder();
    let buf = '';
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buf += decoder.decode(value, { stream: true });
      const events = buf.split('\n\n');
      buf = events.pop();
      parseEvents(events.join('\n\n'), onEvent);
    }
    buf += decoder.decode();
    parseEvents(buf, onEvent);
  }

  async function send(url, body, onEvent) {
    ctrl = new AbortController();
    const resp = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      signal: ctrl.signal,
    });
    if (!resp.ok) {
      const err = await resp.json().catch(() => ({}));
      const e = new Error(err.error?.message || t('error.requestFailed'));
      e.code = err.error?.code || '';
      throw e;
    }
    await readStream(resp.body.getReader(), onEvent);
    return resp;
  }

  function abort() {
    if (ctrl) { ctrl.abort(); ctrl = null; }
  }

  onUnmounted(() => abort());

  return { send, abort };
}

// ── app ──

function mountApp() {
  createApp({
    setup() {
    const sessionID = useSessionStorage('chatSessionID', '');

    // ── i18n ──
    const t = window.Liki.t;

    const sse = useSSE(t);

    const messages = ref([]);
    const ui = reactive({
      phase: 'welcome',     // welcome | chatting | streaming | closed
      substate: 'idle',     // idle | loading
      error: '',
    });
    const input = ref('');
    const orderID = ref('');
    const amount = ref(0);
    const phaseStatus = ref('');
    const greeting = ref('');
    let locationInfo = null; // {country, city} from /api/location
    const chatMessagesEl = ref(null);
    const chatInputEl = ref(null);

    let pending = false;
    let isComposing = false;
    let renderer = null;  // active StreamRenderer, for stopStream/cancel

    // ── derived ──
    const lang = i18next.language;

    const welcomeChips = computed(() => [
      { label: t('chat.chipChartLabel'), msg: t('chat.chipChartMsg') },
      { label: t('chat.chipBondLabel'), msg: t('chat.chipBondMsg') },
      { label: t('chat.chipNamingLabel'), msg: t('chat.chipNamingMsg') },
    ]);

    const showInput = computed(() => ui.phase !== 'closed');

    const inputPlaceholder = computed(() => {
      const hasUserMsg = messages.value.some(m => m.role === 'user');
      return hasUserMsg ? t('chat.placeholderReply') : t('chat.placeholderWelcome');
    });

    const chipsDisabled = ref(false);

    const showChips = computed(() => {
      const userMsgs = messages.value.filter(m => m.role === 'user');
      return !chipsDisabled.value && userMsgs.length === 0 && ui.phase !== 'closed';
    });

    // ── cleanup ──
    onUnmounted(() => {
      if (renderer) { renderer.cancel(); renderer = null; }
    });

    // ── utilities ──
    const formatTime = (ts) => {
      if (!ts) return '';
      const d = new Date(ts);
      const now = new Date();
      const pad = n => String(n).padStart(2, '0');
      const hm = pad(d.getHours()) + ':' + pad(d.getMinutes());
      if (d.toDateString() === now.toDateString()) return hm;
      return pad(d.getMonth() + 1) + '-' + pad(d.getDate()) + ' ' + hm;
    };

    const focusInput = () => {
      nextTick(() => {
        const el = chatInputEl.value;
        if (el) el.focus();
      });
    };

    const onCompositionStart = () => { isComposing = true; };
    const onCompositionEnd = () => { isComposing = false; };

    const scrollDown = () => {
      nextTick(() => {
        const el = chatMessagesEl.value;
        if (el) el.scrollTop = el.scrollHeight;
      });
    };

    // ── throttled rendering ──
    // StreamRenderer encapsulates the throttled streaming render pipeline.
    //
    // Model: content (append-only) → html (derived, idempotent).
    // Throttle defers rendering; flush guarantees content ≡ html.
    function createStreamRenderer(asst) {
      let timer = null;
      let lastRender = 0;

      const _thinkingHTML = () => {
        if (!asst.thinking) return '';
        const open = (!asst.content && !asst._thinkingDismissed) ? ' open' : '';
        return '<details class="thinking-block"' + open + '><summary>' + t('chat.thinking') + '</summary><div>' + window.Liki.escapeHTML(asst.thinking) + '</div></details>';
      };

      const doRender = () => {
        timer = null;
        lastRender = Date.now();
        asst.html = _thinkingHTML() + renderMD(asst.content);
        scrollDown();
      };

      return {
        // Called on each new token — throttles to ≤80ms intervals.
        accept() {
          if (timer) return;
          const elapsed = Date.now() - lastRender;
          if (elapsed >= 80) doRender();
          else timer = setTimeout(doRender, 80 - elapsed);
        },

        // Called on stream end / error — guarantees content → html.
        flush() {
          if (timer) { clearTimeout(timer); timer = null; }
          doRender();
        },

        // Called on user abort — clears timer, keeps rendered content as-is.
        cancel() {
          if (timer) { clearTimeout(timer); timer = null; }
        },
      };
    }

    // ── error ──
    const handleError = (msg) => {
      pending = false;
      ui.error = msg;
      ui.phase = 'chatting';
      ui.substate = 'idle';
    };

    // ── SSE event dispatch ──
    const handleEvent = (evt, asst, renderer) => {
      switch (evt.type) {
        case 'thinking':
          phaseStatus.value = t('chat.thinkingStatus');
          break;

        case 'phase':
          phaseStatus.value = evt.content || '';
          break;

        case 'thinking-delta':
          asst.thinking += evt.content;
          if (!asst.content) renderer.accept();
          break;

        case 'text-delta':
          if (!asst.content) {
            asst._thinkingDismissed = true;
            phaseStatus.value = '';
          }
          ui.phase = 'streaming';
          asst.content += evt.content;
          renderer.accept();
          break;

        case 'done':
          renderer.flush();
          phaseStatus.value = '';
          ui.phase = 'closed';
          ui.substate = 'idle';
          const data = evt.data || {};
          orderID.value = data.order_id || '';
          amount.value = data.amount || 0;
          var currency = (data.currency === 'CNY') ? '¥' : '$';
          messages.value.push({ role: 'buy', amount: amount.value, displayAmount: amount.value, currency: currency, time: new Date().toISOString() });
          scrollDown();
          break;

        case 'error':
          renderer.flush();
          phaseStatus.value = '';
          handleError(evt.content);
          break;
      }
    };

    // ── send ──
    const sendMessage = async (msg) => {
      if (pending) return;
      if (isComposing) return;      // IME input in progress
      const text = (msg || input.value).trim();
      if (!text || ui.phase === 'closed') return;

      pending = true;
      chipsDisabled.value = true;
      input.value = '';
      ui.error = '';
      ui.phase = 'chatting';
      ui.substate = 'loading';

      messages.value.push({ role: 'user', content: text, time: new Date().toISOString() });

      const asst = reactive({ role: 'assistant', content: '', thinking: '', html: '', time: new Date().toISOString() });
      messages.value.push(asst);
      renderer = createStreamRenderer(asst);
      scrollDown();

      try {
        const body = { session_id: sessionID.value, message: text, lang: lang };
        if (!sessionID.value && locationInfo) {
          body.country = locationInfo.country;
          body.city = locationInfo.city;
        }
        const resp = await sse.send('/api/agent/chat', body,
          (evt) => handleEvent(evt, asst, renderer),
        );
        if (resp) {
          const sid = resp.headers.get('X-Session-ID');
          if (sid) sessionID.value = sid;
        }
      } catch (e) {
        if (e.name !== 'AbortError') {
          // Server restart clears sessions — retry once without stale ID.
          if (sessionID.value && e.code === 'not_found') {
            sessionID.value = '';
            messages.value.pop(); // remove empty assistant bubble
            messages.value.pop(); // remove user message (will re-add)
            setTimeout(() => sendMessage(text), 0);
            return;
          }
          handleError(e.message);
          if (!asst.content && !asst.thinking && !asst.html) messages.value.pop();
        }
      } finally {
        pending = false;
        renderer.flush();
        if (ui.phase === 'streaming') {
          ui.phase = 'chatting';
          ui.substate = 'idle';
          focusInput();
        }
        renderer = null;
      }
    };

    const stopStream = () => {
      sse.abort();
      if (renderer) renderer.flush();
      ui.phase = 'chatting';
      ui.substate = 'idle';
      pending = false;
      focusInput();
    };

    const newChat = () => {
      if (renderer) { renderer.cancel(); renderer = null; }
      sessionID.value = '';
      messages.value = greeting.value
        ? [{ role: 'assistant', content: greeting.value, html: renderMD(greeting.value) }]
        : [];
      ui.phase = 'welcome';
      ui.substate = 'idle';
      ui.error = '';
      phaseStatus.value = '';
      orderID.value = '';
      amount.value = 0;
      pending = false;
      chipsDisabled.value = false;
      focusInput();
    };

    onMounted(async () => {
      const appEl = document.getElementById('app');
      if (appEl) appEl.removeAttribute('v-cloak');
      // Fetch greeting and location in parallel.
      const [greetResp] = await Promise.allSettled([
        fetch('/api/agent/greeting', { signal: AbortSignal.timeout(10000) }),
        fetch('/api/location', { signal: AbortSignal.timeout(5000) }).then(r => r.ok ? r.json() : null).then(d => { locationInfo = d && d.data ? { country: d.data.country, city: d.data.city } : null; }).catch(() => {}),
      ]);
      try {
        const resp = greetResp.status === 'fulfilled' ? greetResp.value : null;
        if (resp && resp.ok) {
          const data = await resp.json();
          greeting.value = data.data?.greeting || data.greeting || '';
        }
      } catch (_) {}
      if (greeting.value) {
        messages.value.push({
          role: 'assistant',
          content: greeting.value,
          html: renderMD(greeting.value),
          time: new Date().toISOString(),
        });
      }
      // Auto-select product from query param (e.g. ?product=chart)
      const qp = new URLSearchParams(location.search);
      const prod = qp.get('product');
      if (prod && (prod === 'chart' || prod === 'bond' || prod === 'naming')) {
        const msgMap = {
          chart: t('chat.chipChartMsg'),
          bond: t('chat.chipBondMsg'),
          naming: t('chat.chipNamingMsg'),
        };
        await nextTick();
        sendMessage(msgMap[prod]);
      } else {
        focusInput();
      }
    });

    const buyLoading = ref(false);

    const goPayment = async () => {
      if (!orderID.value || buyLoading.value) return;
      buyLoading.value = true;
      try {
        await goPay(orderID.value);
      } catch (e) {
        ui.error = e.message;
        buyLoading.value = false;
      }
    };

    const dismissBuy = (buyMsg) => {
      buyMsg._dismissed = true;
      ui.phase = 'chatting';
      ui.substate = 'idle';
      focusInput();
    };

    // ── expose ──
    return {
      sessionID, messages, ui, input, orderID,
      amount,
      lang, welcomeChips, showChips, chipsDisabled, showInput, chatMessagesEl, chatInputEl,
      phaseStatus,
      sendMessage, stopStream, newChat, goPayment, dismissBuy, buyLoading, formatTime,
      onCompositionStart, onCompositionEnd,
      t, inputPlaceholder,
    };
  },
  render: window.__chatAppRender,
  }).mount('#app');
}

if (i18next.isInitialized) {
  mountApp();
} else {
  i18next.on('initialized', mountApp);
}
