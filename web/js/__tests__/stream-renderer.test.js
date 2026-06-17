import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// Duplicated from chat.js createStreamRenderer for unit testing the invariants.
// The real function lives inside Vue setup() and is not directly importable.
// These tests catch regressions of the flush/accept/cancel invariants.
function createStreamRenderer(asst, renderMD, escapeHTML, scrollDown, t) {
  let timer = null;
  let lastRender = 0;

  const _thinkingHTML = () => {
    if (!asst.thinking) return '';
    const open = (!asst.content && !asst._thinkingDismissed) ? ' open' : '';
    return '<details class="thinking-block"' + open + '><summary>' + t('chat.thinking') + '</summary><div>' + escapeHTML(asst.thinking) + '</div></details>';
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

function fakeRenderMD(text) { return '<p>' + text + '</p>'; }
function fakeEscapeHTML(text) { return text; }
function fakeT(key) { return key; }

describe('StreamRenderer', () => {
  let asst, scrollCount, renderer;

  beforeEach(() => {
    asst = { content: '', thinking: '', html: '' };
    scrollCount = 0;
    renderer = createStreamRenderer(asst, fakeRenderMD, fakeEscapeHTML, () => { scrollCount++; }, fakeT);
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('initial state: content empty, html empty', () => {
    expect(asst.content).toBe('');
    expect(asst.html).toBe('');
  });

  it('first accept renders immediately (lastRender=0)', () => {
    asst.content = 'first token';
    renderer.accept();
    // First token always renders immediately because lastRender=0 makes elapsed huge.
    expect(asst.html).toContain('first token');
    expect(scrollCount).toBe(1);
  });

  it('second accept within 80ms is throttled', () => {
    // First render
    asst.content = 'a';
    renderer.accept();
    const afterFirst = asst.html;
    expect(afterFirst).toContain('a');

    // Second call right away — should be throttled
    asst.content += 'b';
    renderer.accept();
    // html unchanged because timer is pending
    expect(asst.html).toBe(afterFirst);

    // Third call within throttle window — skipped (timer already pending)
    asst.content += 'c';
    renderer.accept();
    expect(asst.html).toBe(afterFirst);

    // Advance past throttle — single render with all content
    vi.advanceTimersByTime(80);
    expect(asst.html).toContain('abc');
    expect(scrollCount).toBe(2); // initial + one deferred
  });

  it('render immediately if 80ms elapsed since last render', () => {
    asst.content = 'first';
    renderer.accept();
    expect(asst.html).toContain('first');

    // Advance past 80ms
    vi.advanceTimersByTime(100);

    asst.content += ' second';
    renderer.accept(); // elapsed > 80ms → immediate render
    expect(asst.html).toContain('first second');
    expect(scrollCount).toBe(2);
  });

  it('flush() always renders — primary invariant', () => {
    asst.content = 'a';
    renderer.accept(); // immediate (first token)
    asst.content += 'b';
    renderer.accept(); // throttled, timer pending
    // Timer not yet fired, so html has only 'a'
    expect(asst.html).toContain('a');
    expect(asst.html).not.toContain('ab');

    renderer.flush();
    // After flush, content MUST be fully reflected in html
    expect(asst.html).toContain('ab');
  });

  it('flush() clears pending timer then renders', () => {
    asst.content = 'first';
    renderer.accept(); // immediate
    asst.content += ' more';
    renderer.accept(); // throttled — timer pending

    renderer.flush();
    expect(asst.html).toContain('first more');
    // No double render from expired timer
    const afterFlush = scrollCount;
    vi.advanceTimersByTime(100);
    expect(scrollCount).toBe(afterFlush);
  });

  it('cancel() clears timer but does not render pending content', () => {
    asst.content = 'rendered';
    renderer.accept(); // immediate
    const afterFirst = asst.html;
    expect(afterFirst).toContain('rendered');

    asst.content += ' not rendered';
    renderer.accept(); // throttled — timer pending
    renderer.cancel();

    // Still shows old content (not the pending "not rendered" part)
    expect(asst.html).toBe(afterFirst);

    // Timer won't fire
    vi.advanceTimersByTime(100);
    expect(asst.html).toBe(afterFirst);
  });

  it('multiple flush calls are idempotent', () => {
    asst.content = 'same';
    renderer.flush();
    const firstHTML = asst.html;
    renderer.flush();
    expect(asst.html).toBe(firstHTML);
  });

  it('thinking content renders in details block', () => {
    asst.thinking = 'step 1\nstep 2';
    renderer.flush();
    expect(asst.html).toContain('step 1');
    expect(asst.html).toContain('<details class="thinking-block" open>');
  });

  it('thinking block collapses after first text-delta', () => {
    asst.thinking = 'reasoning';
    asst._thinkingDismissed = true;
    asst.content = 'actual output';
    renderer.flush();
    expect(asst.html).not.toContain('<details class="thinking-block" open>');
    expect(asst.html).toContain('<details class="thinking-block">');
  });
});
