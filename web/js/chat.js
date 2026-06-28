const { createApp, ref, reactive, computed, watch, nextTick, onMounted, onUnmounted } = Vue;

// ── SSE helper ──

function useSSE(t) {
  let ctrl = null;
  const MAX_RETRIES = 3;

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
    let lastErr;
    for (let attempt = 0; attempt <= MAX_RETRIES; attempt++) {
      ctrl = new AbortController();
      try {
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
          e.status = resp.status;
          throw e;
        }
        await readStream(resp.body.getReader(), onEvent);
        return resp;
      } catch (e) {
        lastErr = e;
        if (e.name === 'AbortError' || e.status) break; // user abort or HTTP error
        if (attempt < MAX_RETRIES) {
          await new Promise(r => setTimeout(r, Math.pow(2, attempt) * 1000));
        }
      }
    }
    throw lastErr;
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
    const t = window.Liki.t;
    const sse = useSSE(t);

    // ── state ──
    const state = ref('login');     // login | select | chat
    const orderID = ref('');
    const email = ref('');
    const loginLoading = ref(false);
    const loginError = ref('');
    const orderList = ref([]);
    const currency = ref('CNY');
    const currencies = ref(['CNY', 'USD']);
    const buyLoading = ref(false);

    const messages = ref([]);
    const ui = reactive({
      phase: 'welcome',
      substate: 'idle',
      error: '',
    });
    const input = ref('');
    const phaseStatus = ref('');
    const greeting = ref('');
    const countdown = ref('');
    let chatExpiresAt = null;
    const expired = ref(false);
    const chatMessagesEl = ref(null);
    const chatInputEl = ref(null);

    let pending = false;
    let isComposing = false;
    let renderer = null;
    let countdownTimer = null;

    const lang = i18next.language;

    // ── cleanup ──
    onUnmounted(() => {
      if (renderer) { renderer.cancel(); renderer = null; }
      if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null; }
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

    const formatExpiry = (expiresAt) => {
      if (!expiresAt) return '';
      try {
        const d = new Date(expiresAt.replace(' ', 'T') + 'Z');
        const now = new Date();
        const diff = d - now;
        if (diff <= 0) return t('chat.expired');
        const days = Math.floor(diff / 86400000);
        const hours = Math.floor((diff % 86400000) / 3600000);
        if (days > 0) return t('chat.remainingDays').replace('{n}', days);
        return t('chat.remainingHours').replace('{n}', hours);
      } catch (_) { return ''; }
    };

    const finishStream = () => {
      renderer.flush();
      phaseStatus.value = '';
      ui.phase = 'chatting';
      ui.substate = 'idle';
    };

    const tickCountdown = () => {
      if (!chatExpiresAt) return;
      countdown.value = formatExpiry(chatExpiresAt);
      if (countdown.value === t('chat.expired')) {
        expired.value = true;
        countdownTimer = null;
        return;
      }
      countdownTimer = setTimeout(tickCountdown, 60000);
    };

    const startCountdown = () => {
      tickCountdown();
    };

    const DRAFT_KEY = 'likiChatDraft';

    const saveDraft = () => {
      try { sessionStorage.setItem(DRAFT_KEY, input.value); } catch (_) {}
    };

    watch(input, saveDraft);

    const clearDraft = () => {
      try { sessionStorage.removeItem(DRAFT_KEY); } catch (_) {}
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
        if (!el) return;
        const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 40;
        if (atBottom) el.scrollTop = el.scrollHeight;
      });
    };

    // ── Stream renderer ──
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
        accept() {
          if (timer) return;
          const elapsed = Date.now() - lastRender;
          if (elapsed >= 80) doRender();
          else timer = setTimeout(doRender, 80 - elapsed);
        },
        flush() {
          if (timer) { clearTimeout(timer); timer = null; }
          doRender();
        },
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

        case 'report-ready':
          finishStream();
          setTimeout(() => { location.href = evt.content; }, 1500);
          break;

        case 'error':
          renderer.flush();
          phaseStatus.value = '';
          handleError(evt.content);
          break;
      }
    };

    // ── send message ──
    const sendMessage = async (msg) => {
      if (pending) return;
      if (isComposing) return;
      if (expired.value) return;
      const text = (msg || input.value).trim();
      if (!text) return;

      pending = true;
      input.value = '';
      clearDraft();
      ui.error = '';
      ui.phase = 'chatting';
      ui.substate = 'loading';

      messages.value.push({ role: 'user', content: text, time: new Date().toISOString() });

      const asst = reactive({ role: 'advisor', content: '', thinking: '', html: '', time: new Date().toISOString() });
      messages.value.push(asst);
      renderer = createStreamRenderer(asst);
      scrollDown();

      try {
        await sse.send('/api/agent/naming', { message: text, lang: lang },
          (evt) => handleEvent(evt, asst, renderer),
        );
      } catch (e) {
        if (e.name !== 'AbortError') {
          if (e.status === 401) {
            // JWT expired — back to login
            state.value = 'login';
            loginError.value = t('chat.sessionExpired');
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
          phaseStatus.value = '';
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
      messages.value = [];
      ui.phase = 'welcome';
      ui.substate = 'idle';
      ui.error = '';
      phaseStatus.value = '';
      pending = false;
      focusInput();
    };

    // ── login ──
    const doLogin = async () => {
      if (!email.value || loginLoading.value) return;
      loginLoading.value = true;
      loginError.value = '';

      try {
        const resp = await fetch('/api/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email: email.value }),
        });
        const data = await resp.json();
        if (!resp.ok) throw new Error(data.error?.message || t('error.requestFailed'));

        const result = data.data;
        if (result.orders) {
          orderList.value = result.orders;
          state.value = 'select';
        } else if (result.order_id) {
          orderID.value = result.order_id;
          sessionStorage.setItem('likiOrderID', result.order_id);
          await enterChat();
        }
      } catch (e) {
        loginError.value = e.message;
      } finally {
        loginLoading.value = false;
      }
    };

    const selectOrder = async (oid) => {
      try {
        const resp = await fetch('/api/orders/select', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ order_id: oid, email: email.value }),
        });
        const data = await resp.json();
        if (!resp.ok) throw new Error(data.error?.message || t('error.requestFailed'));

        orderID.value = oid;
        sessionStorage.setItem('likiOrderID', oid);
        await enterChat();
      } catch (e) {
        loginError.value = e.message;
        state.value = 'login';
      }
    };

    // ── buy ──
    const doBuy = async () => {
      if (!email.value || buyLoading.value) return;
      buyLoading.value = true;
      loginError.value = '';
      try {
        // 1. Create order
        let resp = await fetch('/api/orders', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email: email.value, product: 'naming', currency: currency.value }),
        });
        let data = await resp.json();
        if (!resp.ok) throw new Error(data.error?.message || t('error.requestFailed'));
        const oid = data.data?.order_id;
        if (!oid) throw new Error('Missing order_id');

        // 2. Checkout
        resp = await fetch('/api/payments/checkout', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ order_id: oid, email: email.value }),
        });
        data = await resp.json();
        if (!resp.ok) throw new Error(data.error?.message || t('error.requestFailed'));
        const url = data.data?.checkout_url;
        if (!url) throw new Error('Missing checkout_url');

        sessionStorage.setItem('likiOrderID', oid);
        location.href = url;
      } catch (e) {
        loginError.value = e.message;
      } finally {
        buyLoading.value = false;
      }
    };

    // ── enter chat ──
    const enterChat = async (prefetched) => {
      try {
        let o;
        if (prefetched) {
          o = prefetched;
        } else {
          const resp = await fetch('/api/orders/' + orderID.value + '/status');
          const data = await resp.json();
          if (!resp.ok) throw new Error('order not found');
          o = data.data;
        }

        if (o.status !== 'paid') {
          loginError.value = t('chat.orderNotPaid');
          state.value = 'login';
          return;
        }

        chatExpiresAt = o.chat_expires_at || null;
        countdown.value = formatExpiry(o.chat_expires_at);
        if (countdown.value === t('chat.expired')) expired.value = true;

        state.value = 'chat';
        startCountdown();
        try { const d = sessionStorage.getItem(DRAFT_KEY); if (d) input.value = d; } catch (_) {}
        focusInput();

        greeting.value = t('chat.greeting');
        messages.value.push({
          role: 'advisor',
          content: greeting.value,
          html: renderMD(greeting.value),
          time: new Date().toISOString(),
        });
      } catch (e) {
        loginError.value = e.message;
        state.value = 'login';
      }
    };

    // ── init ──
    onMounted(async () => {
      const appEl = document.getElementById('app');
      if (appEl) appEl.removeAttribute('v-cloak');

      // Check for order_id from URL (payment callback) or sessionStorage.
      const qp = new URLSearchParams(location.search);
      let oid = qp.get('order_id');
      if (oid) {
        sessionStorage.setItem('likiOrderID', oid);
        // Clean URL
        history.replaceState(null, '', '/chat');
      } else {
        oid = sessionStorage.getItem('likiOrderID');
      }

      if (oid) {
        orderID.value = oid;
        try {
          const resp = await fetch('/api/orders/' + oid + '/status');
          if (resp.status === 401) {
            // JWT expired
            state.value = 'login';
            return;
          }
          if (resp.ok) {
            const data = await resp.json();
            if (data.data?.status === 'paid') {
              await enterChat(data.data);
              return;
            }
          }
        } catch (_) {}
        // If we get here, the order is invalid — fall through to login
        sessionStorage.removeItem('likiOrderID');
      }

      state.value = 'login';
    });

    // ── expose ──
    return {
      state, orderID, email, loginLoading, loginError, orderList,
      currency, currencies, buyLoading,
      messages, ui, input, phaseStatus, greeting, countdown, expired,
      chatMessagesEl, chatInputEl, lang,
      formatTime, formatExpiry, onCompositionStart, onCompositionEnd,
      sendMessage, stopStream, newChat,
      doLogin, selectOrder, doBuy,
      t,
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
