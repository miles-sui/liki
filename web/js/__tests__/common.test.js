/**
 * Unit tests for shared utility functions from common.js.
 * These functions are pure-ish — they only depend on a few globals that we mock.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest';

// ---- Mock browser globals that the functions need ----

// ELEMENT_CODES and ELEMENT_NAMES are injected at build time.
globalThis.window = globalThis;
globalThis.ELEMENT_CODES = ['W', 'F', 'E', 'M', 'R'];
globalThis.ELEMENT_NAMES = { W: 'Wood', F: 'Fire', E: 'Earth', M: 'Metal', R: 'Water' };
globalThis.ELEMENT_HTML_ORDER = [1, 0, 2, 3, 4]; // Fire, Wood, Earth, Metal, Water

// ---- Function definitions (copied from common.js to keep tests isolated) ----

function namedToArray(v) {
  if (!v) return [];
  if (Array.isArray(v)) return v;
  var keys = ['wood', 'fire', 'earth', 'metal', 'water'];
  return keys.map(function(k) { return v[k] || 0; });
}

function deviationToProportion(d) {
  var arr = Array.isArray(d) ? d : namedToArray(d);
  if (!arr.length) return [];
  return arr.map(function(v) { return 0.2 + v / 5; });
}

function computeBondShapes(bond) {
  if (!bond) return null;
  var ELEMENT_HTML_ORDER = [1, 0, 2, 3, 4];
  var dEffSelf = namedToArray(bond.self);
  var dEffOther = namedToArray(bond.other);
  var arrA = namedToArray(bond.delta_a);
  var arrB = namedToArray(bond.delta_b);
  var origSelf = deviationToProportion(dEffSelf.map(function(v, i) { return v - arrA[i]; }));
  var origOther = deviationToProportion(dEffOther.map(function(v, i) { return v - arrB[i]; }));
  var pSelf = deviationToProportion(dEffSelf);
  var pOther = deviationToProportion(dEffOther);
  return {
    origSelf: origSelf, origOther: origOther,
    pSelf: pSelf, pOther: pOther,
    selfDeltas: ELEMENT_HTML_ORDER.map(function(i) { return { idx: i, delta: pSelf[i] - origSelf[i] }; }),
    otherDeltas: ELEMENT_HTML_ORDER.map(function(i) { return { idx: i, delta: pOther[i] - origOther[i] }; }),
  };
}

// ---- namedToArray ----

describe('namedToArray', () => {
  it('converts named object to array', () => {
    const result = namedToArray({ wood: 0.5, fire: 0.3, earth: -0.2, metal: 0.1, water: -0.7 });
    expect(result).toEqual([0.5, 0.3, -0.2, 0.1, -0.7]);
  });

  it('passes through arrays unchanged', () => {
    expect(namedToArray([1, 2, 3, 4, 5])).toEqual([1, 2, 3, 4, 5]);
  });

  it('returns empty array for null', () => {
    expect(namedToArray(null)).toEqual([]);
  });

  it('returns empty array for undefined', () => {
    expect(namedToArray(undefined)).toEqual([]);
  });

  it('fills zeros for missing keys', () => {
    expect(namedToArray({ wood: 0.5 })).toEqual([0.5, 0, 0, 0, 0]);
  });

  it('returns zeros for empty object', () => {
    expect(namedToArray({})).toEqual([0, 0, 0, 0, 0]);
  });

  it('preserves wood, fire, earth, metal, water order', () => {
    const result = namedToArray({ water: 5, fire: 2, metal: 4, earth: 3, wood: 1 });
    expect(result).toEqual([1, 2, 3, 4, 5]);
  });
});

// ---- deviationToProportion ----

describe('deviationToProportion', () => {
  it('converts zero deviation to uniform 0.2', () => {
    const result = deviationToProportion([0, 0, 0, 0, 0]);
    expect(result).toEqual([0.2, 0.2, 0.2, 0.2, 0.2]);
  });

  it('sum of proportions is ~1', () => {
    const result = deviationToProportion([0.5, -0.3, 0.2, -0.1, -0.3]);
    const sum = result.reduce((a, b) => a + b, 0);
    expect(sum).toBeCloseTo(1, 5);
  });

  it('converts named object input', () => {
    const result = deviationToProportion({ wood: 0, fire: 0, earth: 0, metal: 0, water: 0 });
    expect(result).toEqual([0.2, 0.2, 0.2, 0.2, 0.2]);
  });

  it('returns empty array for empty input', () => {
    expect(deviationToProportion([])).toEqual([]);
  });

  it('d→p→d roundtrip', () => {
    // p = 0.2 + d/5, so d = 5*(p-0.2)
    const d = [0.5, -0.3, 0.1, -0.2, -0.1];
    const p = deviationToProportion(d);
    const d2 = p.map(v => 5 * (v - 0.2));
    d.forEach((v, i) => expect(d2[i]).toBeCloseTo(v, 10));
  });
});

// ---- computeBondShapes ----

describe('computeBondShapes', () => {
  const makeBond = () => ({
    self: { wood: 0.5, fire: 0.2, earth: 0.0, metal: -0.3, water: -0.4 },
    other: { wood: -0.2, fire: 0.3, earth: 0.1, metal: -0.1, water: -0.1 },
    delta_a: { wood: 0.1, fire: 0.05, earth: 0.0, metal: -0.05, water: -0.1 },
    delta_b: { wood: -0.05, fire: 0.1, earth: 0.05, metal: 0.0, water: -0.1 },
  });

  it('returns null for null input', () => {
    expect(computeBondShapes(null)).toBeNull();
  });

  it('returns all expected keys', () => {
    const result = computeBondShapes(makeBond());
    expect(result).toHaveProperty('origSelf');
    expect(result).toHaveProperty('origOther');
    expect(result).toHaveProperty('pSelf');
    expect(result).toHaveProperty('pOther');
    expect(result).toHaveProperty('selfDeltas');
    expect(result).toHaveProperty('otherDeltas');
  });

  it('output proportions are arrays of length 5', () => {
    const result = computeBondShapes(makeBond());
    expect(result.origSelf).toHaveLength(5);
    expect(result.pSelf).toHaveLength(5);
  });

  it('pSelf sum is ~1', () => {
    const result = computeBondShapes(makeBond());
    const sum = result.pSelf.reduce((a, b) => a + b, 0);
    expect(sum).toBeCloseTo(1, 5);
  });

  it('deltas have idx and delta fields', () => {
    const result = computeBondShapes(makeBond());
    expect(result.selfDeltas).toHaveLength(5);
    result.selfDeltas.forEach(d => {
      expect(d).toHaveProperty('idx');
      expect(d).toHaveProperty('delta');
      expect(typeof d.idx).toBe('number');
      expect(typeof d.delta).toBe('number');
    });
  });

  it('delta = pSelf - origSelf', () => {
    const result = computeBondShapes(makeBond());
    result.selfDeltas.forEach(d => {
      const expected = result.pSelf[d.idx] - result.origSelf[d.idx];
      expect(d.delta).toBeCloseTo(expected, 10);
    });
  });

  it('handles pure wood bond (extreme case)', () => {
    const bond = {
      self: { wood: 0.8, fire: -0.3, earth: -0.2, metal: -0.2, water: -0.1 },
      other: { wood: 0.9, fire: -0.4, earth: -0.2, metal: -0.2, water: -0.1 },
      delta_a: { wood: 0, fire: 0, earth: 0, metal: 0, water: 0 },
      delta_b: { wood: 0, fire: 0, earth: 0, metal: 0, water: 0 },
    };
    const result = computeBondShapes(bond);
    // With zero deltas, origSelf === pSelf, so deltas should all be ~0.
    result.selfDeltas.forEach(d => expect(d.delta).toBeCloseTo(0, 10));
  });
});

// ---- api() fetch wrapper ----

describe('api', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    globalThis.localStorage = {
      _store: {},
      getItem: vi.fn(k => globalThis.localStorage._store[k] || null),
      setItem: vi.fn((k, v) => { globalThis.localStorage._store[k] = v; }),
      removeItem: vi.fn(k => { delete globalThis.localStorage._store[k]; }),
    };
    // Minimal Alpine mock
    globalThis.Alpine = {
      store: vi.fn(() => ({ current: 'en' })),
    };
  });

  // Copy the actual api() implementation here (uses fetch, localStorage, Alpine, window)
  async function api(path, opts) {
    opts = opts || {};
    var quiet = opts.quiet;
    var token = localStorage.getItem('token');
    var headers = { 'Content-Type': 'application/json' };
    if (opts.headers) Object.assign(headers, opts.headers);
    if (token) headers['Authorization'] = 'Bearer ' + token;
    try { headers['X-Locale'] = Alpine.store('locale').current; } catch (_) {}

    var res;
    try {
      res = await fetch(path, { method: opts.method, body: opts.body, headers: headers });
    } catch (e) {
      console.error(e);
      if (!quiet) {
        try { Alpine.store('toast').error('Network error'); } catch (_) {}
      }
      throw new Error('Network error');
    }

    if (res.status === 401) {
      var body = {};
      try { body = await res.json(); } catch (e) { console.error(e); }
      localStorage.removeItem('token');
      try { Alpine.store('auth').id = null; } catch (_) {}
      try { Alpine.store('auth').token = ''; } catch (_) {}
      throw new Error(body.error ? body.error.message : 'Unauthorized');
    }

    var data = {};
    try { data = await res.json(); } catch (e) { console.error(e); }

    if (!res.ok) {
      var msg = (data && data.error && data.error.message) || 'Server error';
      if (!quiet) {
        try { Alpine.store('toast').error(msg); } catch (_) {}
      }
      throw new Error((data && data.error && data.error.message) || msg);
    }
    return data;
  }

  it('returns data on success', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ data: { id: 1 } }),
    });

    const result = await api('/api/test');
    expect(result.data.id).toBe(1);
  });

  it('throws Error on non-ok response', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      json: () => Promise.resolve({ error: { code: 'internal', message: 'boom' } }),
    });

    await expect(api('/api/test')).rejects.toThrow('boom');
  });

  it('throws Error on network failure', async () => {
    globalThis.fetch = vi.fn().mockRejectedValue(new Error('ECONNREFUSED'));

    await expect(api('/api/test')).rejects.toThrow('Network error');
  });

  it('sends Authorization header when token present', async () => {
    localStorage._store['token'] = 'test-jwt';
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ data: {} }),
    });

    await api('/api/protected');

    const call = fetch.mock.calls[0];
    expect(call[1].headers['Authorization']).toBe('Bearer test-jwt');
  });

  it('throws on 401 and removes token', async () => {
    localStorage._store['token'] = 'expired-token';
    let jsonCalls = 0;
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 401,
      json: () => {
        jsonCalls++;
        return Promise.resolve({ error: { code: 'token_expired', message: 'expired' } });
      },
    });

    await expect(api('/api/protected')).rejects.toThrow('expired');
    expect(localStorage.removeItem).toHaveBeenCalledWith('token');
  });
});

// ---- makePickAny ----

function makePickAny(getSelections) {
  return {
    isSelected(qid, element) {
      var s = getSelections.call(this)[qid];
      return s && s.includes(element);
    },
    toggleSelect(qid, element) {
      var sel = getSelections.call(this);
      var s = sel[qid];
      if (!s) s = sel[qid] = [];
      var idx = s.indexOf(element);
      if (idx >= 0) { s.splice(idx, 1); }
      else { s.push(element); }
      sel[qid] = s;
    },
  };
}

describe('makePickAny', () => {
  function makeContext(initial) {
    var store = initial || {};
    return { answers: store, getSelections: function() { return this.answers; } };
  }

  it('toggleSelect adds an element when not selected', () => {
    var ctx = makeContext({});
    var picker = makePickAny(ctx.getSelections.bind(ctx));
    picker.toggleSelect.call(ctx, 'Q01', 'W');
    expect(ctx.answers['Q01']).toEqual(['W']);
  });

  it('toggleSelect removes an element when already selected', () => {
    var ctx = makeContext({ Q01: ['W', 'F'] });
    var picker = makePickAny(ctx.getSelections.bind(ctx));
    picker.toggleSelect.call(ctx, 'Q01', 'W');
    expect(ctx.answers['Q01']).toEqual(['F']);
  });

  it('toggleSelect allows selecting all three options (no cap)', () => {
    var ctx = makeContext({ Q01: ['W', 'F'] });
    var picker = makePickAny(ctx.getSelections.bind(ctx));
    picker.toggleSelect.call(ctx, 'Q01', 'E');
    expect(ctx.answers['Q01']).toEqual(['W', 'F', 'E']);
  });

  it('toggleSelect allows selecting zero options', () => {
    var ctx = makeContext({ Q01: ['W'] });
    var picker = makePickAny(ctx.getSelections.bind(ctx));
    picker.toggleSelect.call(ctx, 'Q01', 'W');
    expect(ctx.answers['Q01']).toEqual([]);
  });

  it('isSelected returns true for selected element', () => {
    var ctx = makeContext({ Q01: ['W'] });
    var picker = makePickAny(ctx.getSelections.bind(ctx));
    expect(picker.isSelected.call(ctx, 'Q01', 'W')).toBe(true);
    expect(picker.isSelected.call(ctx, 'Q01', 'F')).toBe(false);
  });

  it('isSelected returns false for missing question', () => {
    var ctx = makeContext({});
    var picker = makePickAny(ctx.getSelections.bind(ctx));
    expect(!!picker.isSelected.call(ctx, 'Q01', 'W')).toBe(false);
  });

  it('selections are independent across questions', () => {
    var ctx = makeContext({});
    var picker = makePickAny(ctx.getSelections.bind(ctx));
    picker.toggleSelect.call(ctx, 'Q01', 'W');
    picker.toggleSelect.call(ctx, 'Q02', 'F');
    expect(ctx.answers['Q01']).toEqual(['W']);
    expect(ctx.answers['Q02']).toEqual(['F']);
  });

  it('toggleSelect toggles correctly with 3 selections (click all 3 then unclick one)', () => {
    var ctx = makeContext({ Q01: ['W', 'F', 'E'] });
    var picker = makePickAny(ctx.getSelections.bind(ctx));
    picker.toggleSelect.call(ctx, 'Q01', 'F');
    expect(ctx.answers['Q01']).toEqual(['W', 'E']);
  });
});

// ---- assessmentNavigation (answeredCount, answersInRound, firstUnanswered) ----

function assessmentNavigation() {
  return {
    get round() { return Math.floor(this.currentQIndex / 5) + 1; },
    get totalRounds() { return Math.ceil(this.allQuestions.length / 5) || 6; },
    get totalQuestions() { return this.allQuestions.length || 30; },
    get currentQuestion() { return this.allQuestions[this.currentQIndex]; },

    get answeredCount() {
      var n = 0;
      for (var k in this.answers) { if (this.answers[k]) n += this.answers[k].length; }
      return n;
    },

    roundStart() { return (this.round - 1) * 5; },
    roundEnd() { return Math.min(this.round * 5, this.totalQuestions) - 1; },

    answersInRound() {
      var n = 0;
      for (var i = this.roundStart(); i <= this.roundEnd(); i++) {
        var q = this.allQuestions[i];
        if (q && this.answers[q.qid]) n += this.answers[q.qid].length;
      }
      return n;
    },

    isRoundComplete() {
      return this.answersInRound() >= Math.min(5, this.totalQuestions - this.roundStart());
    },

    get firstUnanswered() {
      for (var i = 0; i < this.allQuestions.length; i++) {
        var q = this.allQuestions[i];
        if (!this.answers[q.qid] || this.answers[q.qid].length === 0) return i;
      }
      return -1;
    },
  };
}

function makeQuestions(n) {
  var qs = [];
  for (var i = 0; i < n; i++) {
    qs.push({ qid: 'Q' + String(i + 1).padStart(2, '0') });
  }
  return qs;
}

describe('assessmentNavigation', () => {
  describe('answeredCount (total selections)', () => {
    it('returns 0 when no answers', () => {
      var ctx = { answers: {}, allQuestions: makeQuestions(30), currentQIndex: 0 };
      Object.defineProperties(ctx, Object.getOwnPropertyDescriptors(assessmentNavigation()));
      expect(ctx.answeredCount).toBe(0);
    });

    it('sums selection lengths across all questions', () => {
      var ctx = { answers: { Q01: ['W'], Q02: ['F', 'E'], Q03: ['W', 'F', 'E'] }, allQuestions: makeQuestions(30), currentQIndex: 0 };
      Object.defineProperties(ctx, Object.getOwnPropertyDescriptors(assessmentNavigation()));
      expect(ctx.answeredCount).toBe(6); // 1 + 2 + 3
    });

    it('counts partial selections (1 or 2 instead of exactly 2)', () => {
      var ctx = { answers: { Q01: ['W'] }, allQuestions: makeQuestions(30), currentQIndex: 0 };
      Object.defineProperties(ctx, Object.getOwnPropertyDescriptors(assessmentNavigation()));
      expect(ctx.answeredCount).toBe(1);
    });

    it('counts all three selections', () => {
      var ctx = { answers: { Q01: ['W', 'F', 'E'] }, allQuestions: makeQuestions(30), currentQIndex: 0 };
      Object.defineProperties(ctx, Object.getOwnPropertyDescriptors(assessmentNavigation()));
      expect(ctx.answeredCount).toBe(3);
    });

    it('ignores empty arrays', () => {
      var ctx = { answers: { Q01: [] }, allQuestions: makeQuestions(30), currentQIndex: 0 };
      Object.defineProperties(ctx, Object.getOwnPropertyDescriptors(assessmentNavigation()));
      expect(ctx.answeredCount).toBe(0);
    });
  });

  describe('answersInRound', () => {
    it('returns total selections in current round', () => {
      var ctx = {
        answers: { Q01: ['W', 'F'], Q02: ['E'], Q03: [], Q04: ['W', 'F', 'E'], Q05: ['M'] },
        allQuestions: makeQuestions(30),
        currentQIndex: 0, // round 1
      };
      Object.defineProperties(ctx, Object.getOwnPropertyDescriptors(assessmentNavigation()));
      expect(ctx.answersInRound()).toBe(7); // 2+1+0+3+1
    });

    it('returns 0 for round with no selections', () => {
      var ctx = { answers: {}, allQuestions: makeQuestions(30), currentQIndex: 5 };
      Object.defineProperties(ctx, Object.getOwnPropertyDescriptors(assessmentNavigation()));
      expect(ctx.answersInRound()).toBe(0);
    });
  });

  describe('firstUnanswered', () => {
    it('returns index of first question with zero selections', () => {
      var ctx = {
        answers: { Q01: ['W'], Q02: [] },
        allQuestions: makeQuestions(30),
        currentQIndex: 0,
      };
      Object.defineProperties(ctx, Object.getOwnPropertyDescriptors(assessmentNavigation()));
      expect(ctx.firstUnanswered).toBe(1); // Q02 has 0 selections
    });

    it('returns index of first question with no entry', () => {
      var ctx = {
        answers: { Q01: ['W'] },
        allQuestions: makeQuestions(30),
        currentQIndex: 0,
      };
      Object.defineProperties(ctx, Object.getOwnPropertyDescriptors(assessmentNavigation()));
      expect(ctx.firstUnanswered).toBe(1); // Q02 missing
    });

    it('returns -1 when all questions have at least one selection', () => {
      var answers = {};
      for (var i = 0; i < 30; i++) {
        answers['Q' + String(i + 1).padStart(2, '0')] = ['W'];
      }
      var ctx = { answers: answers, allQuestions: makeQuestions(30), currentQIndex: 0 };
      Object.defineProperties(ctx, Object.getOwnPropertyDescriptors(assessmentNavigation()));
      expect(ctx.firstUnanswered).toBe(-1);
    });

    it('returns 0 when first question is empty', () => {
      var ctx = { answers: {}, allQuestions: makeQuestions(30), currentQIndex: 0 };
      Object.defineProperties(ctx, Object.getOwnPropertyDescriptors(assessmentNavigation()));
      expect(ctx.firstUnanswered).toBe(0);
    });
  });
});
