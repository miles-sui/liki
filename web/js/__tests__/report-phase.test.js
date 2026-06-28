import { describe, it, expect } from 'vitest';

// Simulates the reportApp state machine from report.js.
// Actual states: loading | payment | timeout_payment | error | unlocking | ready

function createPhaseMachine() {
  let state = { phase: 'loading', error: '' };

  const VALID_TRANSITIONS = {
    loading:       ['payment', 'ready', 'error'],
    payment:       ['ready', 'timeout_payment', 'error'],
    timeout_payment: ['loading'],
    error:         ['loading'],
    unlocking:     ['ready'],
    ready:         [],
  };

  function transition(newPhase) {
    const allowed = VALID_TRANSITIONS[state.phase];
    if (!allowed || !allowed.includes(newPhase)) {
      throw new Error(`invalid transition: ${state.phase} → ${newPhase}`);
    }
    state.phase = newPhase;
  }

  return {
    get phase() { return state.phase; },
    get error() { return state.error; },

    // loadReport: status=pending → startPolling('payment')
    loadPending() {
      transition('payment');
    },

    // loadReport: status=paid (with or without llm_json) → showReport → ready
    loadReady() {
      transition('ready');
    },

    // loadReport: no orderID, unknown status, non-naming product, or catch
    loadError(msg) {
      state.error = msg;
      transition('error');
    },

    // Poll: status changes to paid with llm_json → showReport → ready
    pollToReady() {
      if (state.phase !== 'payment') throw new Error('not in payment');
      transition('ready');
    },

    // Poll: tries > MAX_TRIES or errors > MAX_ERRORS
    pollTimeout() {
      if (state.phase !== 'payment') throw new Error('not in payment');
      transition('timeout_payment');
    },

    // Poll: unexpected status
    pollError(msg) {
      if (state.phase !== 'payment') throw new Error('not in payment');
      state.error = msg;
      transition('error');
    },

    // showReport with transition: sets phase to unlocking
    enterUnlocking() {
      state.phase = 'unlocking';
    },

    // Transition animation complete → ready
    unlockReady() {
      if (state.phase !== 'unlocking') throw new Error('not in unlocking');
      transition('ready');
    },

    // retryPoll → loadReport from timeout
    retryFromTimeout() {
      if (state.phase !== 'timeout_payment') throw new Error('not in timeout_payment');
      transition('loading');
    },

    // retryPoll → loadReport from error
    retryFromError() {
      if (state.phase !== 'error') throw new Error('not in error');
      state.error = '';
      transition('loading');
    },
  };
}

describe('reportApp phase state machine (6-state)', () => {
  describe('initial state', () => {
    it('starts in loading', () => {
      const m = createPhaseMachine();
      expect(m.phase).toBe('loading');
    });
  });

  describe('loading transitions', () => {
    it('loading → payment when status is pending', () => {
      const m = createPhaseMachine();
      m.loadPending();
      expect(m.phase).toBe('payment');
    });

    it('loading → ready when status is paid (with or without llm_json)', () => {
      const m = createPhaseMachine();
      m.loadReady();
      expect(m.phase).toBe('ready');
    });

    it('loading → error on failure (no orderID, non-naming product, API error)', () => {
      const m = createPhaseMachine();
      m.loadError('not found');
      expect(m.phase).toBe('error');
      expect(m.error).toBe('not found');
    });
  });

  describe('payment polling transitions', () => {
    it('payment → ready when llm_json becomes available', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollToReady();
      expect(m.phase).toBe('ready');
    });

    it('payment → timeout_payment after max tries or max errors', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollTimeout();
      expect(m.phase).toBe('timeout_payment');
    });

    it('payment → error on unexpected status', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollError('unexpected status');
      expect(m.phase).toBe('error');
      expect(m.error).toBe('unexpected status');
    });
  });

  describe('timeout state', () => {
    it('timeout_payment → loading on retry', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollTimeout();
      m.retryFromTimeout();
      expect(m.phase).toBe('loading');
    });

    it('retry from timeout re-enters full loadReport flow', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollTimeout();
      m.retryFromTimeout();
      m.loadPending();
      expect(m.phase).toBe('payment');
    });
  });

  describe('error state', () => {
    it('error → loading on retry', () => {
      const m = createPhaseMachine();
      m.loadError('fail');
      m.retryFromError();
      expect(m.phase).toBe('loading');
      expect(m.error).toBe('');
    });
  });

  describe('unlocking state', () => {
    it('unlocking → ready after transition animation', () => {
      const m = createPhaseMachine();
      m.enterUnlocking();
      m.unlockReady();
      expect(m.phase).toBe('ready');
    });
  });

  describe('ready is terminal', () => {
    it('no transitions allowed from ready', () => {
      const m = createPhaseMachine();
      m.loadReady();
      expect(m.phase).toBe('ready');
      expect(() => m.loadPending()).toThrow('invalid transition');
      expect(() => m.loadReady()).toThrow('invalid transition');
      expect(() => m.loadError('x')).toThrow('invalid transition');
      expect(m.phase).toBe('ready');
    });
  });

  describe('invalid transitions', () => {
    it('loading → timeout_payment is invalid', () => {
      const m = createPhaseMachine();
      expect(() => m.pollTimeout()).toThrow();
    });

    it('payment → loading is invalid', () => {
      const m = createPhaseMachine();
      m.loadPending();
      expect(() => m.retryFromTimeout()).toThrow();
    });

    it('timeout_payment → payment is invalid (must go through loading)', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollTimeout();
      expect(() => m.pollToReady()).toThrow();
    });

    it('error → payment is invalid (must go through loading)', () => {
      const m = createPhaseMachine();
      m.loadError('x');
      expect(() => m.loadPending()).toThrow();
    });

    it('unlocking → payment is invalid', () => {
      const m = createPhaseMachine();
      m.enterUnlocking();
      expect(() => m.loadPending()).toThrow();
    });

    it('unlocking → error is invalid', () => {
      const m = createPhaseMachine();
      m.enterUnlocking();
      expect(() => m.loadError('x')).toThrow();
    });
  });

  describe('full lifecycle paths', () => {
    it('golden path: loading → payment → ready', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollToReady();
      expect(m.phase).toBe('ready');
    });

    it('direct ready: loading → ready (already paid)', () => {
      const m = createPhaseMachine();
      m.loadReady();
      expect(m.phase).toBe('ready');
    });

    it('payment timeout → retry → success', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollTimeout();
      m.retryFromTimeout();
      m.loadPending();
      m.pollToReady();
      expect(m.phase).toBe('ready');
    });

    it('error → retry → payment → ready', () => {
      const m = createPhaseMachine();
      m.loadError('network error');
      m.retryFromError();
      m.loadPending();
      m.pollToReady();
      expect(m.phase).toBe('ready');
    });

    it('loading → payment → unlocking → ready (transition animation)', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollToReady();
      m.enterUnlocking();
      m.unlockReady();
      expect(m.phase).toBe('ready');
    });
  });
});
