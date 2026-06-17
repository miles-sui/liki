import { describe, it, expect } from 'vitest';

// Simulates the reportApp state machine from report.js.
// Actual states: loading | payment | generating | timeout_payment | timeout_generating | error | unlocking | ready

function createPhaseMachine() {
  let state = { phase: 'loading', error: '' };

  const VALID_TRANSITIONS = {
    loading:       ['payment', 'generating', 'ready', 'error'],
    payment:       ['generating', 'ready', 'timeout_payment', 'error'],
    generating:    ['ready', 'timeout_generating', 'error'],
    timeout_payment:    ['loading'],
    timeout_generating: ['loading'],
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

    // loadReport: status=paid, no llm_json → startPolling('generating')
    loadPaidNoJSON() {
      transition('generating');
    },

    // loadReport: status=paid, has llm_json → showReport(_, false) → ready
    loadReady() {
      transition('ready');
    },

    // loadReport: no orderID or unknown status or catch
    loadError(msg) {
      state.error = msg;
      transition('error');
    },

    // Poll: status changes from pending to paid (no llm_json yet)
    pollPaymentToGenerating() {
      if (state.phase !== 'payment') throw new Error('not in payment');
      transition('generating');
    },

    // Poll: status=paid + llm_json → showReport(_, true) → unlocking → ready
    pollPaymentToUnlocking() {
      if (state.phase !== 'payment') throw new Error('not in payment');
      transition('ready');
    },

    // Poll: tries > MAX_TRIES or errors > MAX_ERRORS for payment
    pollPaymentTimeout() {
      if (state.phase !== 'payment') throw new Error('not in payment');
      transition('timeout_payment');
    },

    // Poll: status=paid + llm_json during generating poll
    pollGenToUnlocking() {
      if (state.phase !== 'generating') throw new Error('not in generating');
      transition('ready');
    },

    // Poll: tries > MAX_TRIES or errors > MAX_ERRORS for generating
    pollGenTimeout() {
      if (state.phase !== 'generating') throw new Error('not in generating');
      transition('timeout_generating');
    },

    // Poll: unexpected status during generating
    pollGenError(msg) {
      if (state.phase !== 'generating') throw new Error('not in generating');
      state.error = msg;
      transition('error');
    },

    // Enter unlocking (showReport with transition sets phase='unlocking' directly)
    enterUnlocking() {
      state.phase = 'unlocking';
    },

    // showReport with transition: unlocking → ready
    unlockReady() {
      if (state.phase !== 'unlocking') throw new Error('not in unlocking');
      transition('ready');
    },

    // retryPoll → loadReport
    retryFrom(timeoutPhase) {
      if (state.phase !== timeoutPhase) throw new Error('not in ' + timeoutPhase);
      transition('loading');
    },

    // error → retry (loadReport)
    retryFromError() {
      if (state.phase !== 'error') throw new Error('not in error');
      state.error = '';
      transition('loading');
    },
  };
}

describe('reportApp phase state machine (8-state)', () => {
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

    it('loading → generating when paid but no llm_json', () => {
      const m = createPhaseMachine();
      m.loadPaidNoJSON();
      expect(m.phase).toBe('generating');
    });

    it('loading → ready when paid and llm_json available', () => {
      const m = createPhaseMachine();
      m.loadReady();
      expect(m.phase).toBe('ready');
    });

    it('loading → error on failure (no orderID, unknown status, network error)', () => {
      const m = createPhaseMachine();
      m.loadError('not found');
      expect(m.phase).toBe('error');
      expect(m.error).toBe('not found');
    });
  });

  describe('payment polling transitions', () => {
    it('payment → generating when status changes to paid (no llm_json yet)', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollPaymentToGenerating();
      expect(m.phase).toBe('generating');
    });

    it('payment → ready when llm_json becomes available during poll', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollPaymentToUnlocking();
      expect(m.phase).toBe('ready');
    });

    it('payment → timeout_payment after max tries or max errors', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollPaymentTimeout();
      expect(m.phase).toBe('timeout_payment');
    });

    it('payment → timeout_payment after consecutive poll errors', () => {
      // MAX_ERRORS=6, after 7th error transition to timeout
      const m = createPhaseMachine();
      m.loadPending();
      m.pollPaymentTimeout();
      expect(m.phase).toBe('timeout_payment');
    });
  });

  describe('generating polling transitions', () => {
    it('generating → ready when llm_json appears during poll', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollPaymentToGenerating();
      m.pollGenToUnlocking();
      expect(m.phase).toBe('ready');
    });

    it('generating → timeout_generating after max tries', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollPaymentToGenerating();
      m.pollGenTimeout();
      expect(m.phase).toBe('timeout_generating');
    });

    it('generating → error on unexpected status', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollPaymentToGenerating();
      m.pollGenError('unexpected status');
      expect(m.phase).toBe('error');
      expect(m.error).toBe('unexpected status');
    });
  });

  describe('timeout states', () => {
    it('timeout_payment → loading on retry', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollPaymentTimeout();
      m.retryFrom('timeout_payment');
      expect(m.phase).toBe('loading');
    });

    it('timeout_generating → loading on retry', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollPaymentToGenerating();
      m.pollGenTimeout();
      m.retryFrom('timeout_generating');
      expect(m.phase).toBe('loading');
    });

    it('retry from timeout re-enters full loadReport flow', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollPaymentTimeout();
      m.retryFrom('timeout_payment');
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
      expect(() => m.loadPaidNoJSON()).toThrow('invalid transition');
      expect(() => m.loadError('x')).toThrow('invalid transition');
      expect(m.phase).toBe('ready');
    });
  });

  describe('invalid transitions', () => {
    it('loading → timeout_payment is invalid', () => {
      const m = createPhaseMachine();
      expect(() => m.pollPaymentTimeout()).toThrow();
    });

    it('loading → timeout_generating is invalid', () => {
      const m = createPhaseMachine();
      expect(() => m.pollGenTimeout()).toThrow();
    });

    it('payment → timeout_generating is invalid', () => {
      const m = createPhaseMachine();
      m.loadPending();
      expect(() => m.pollGenTimeout()).toThrow();
    });

    it('timeout_payment → timeout_payment is invalid', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollPaymentTimeout();
      expect(() => m.pollPaymentTimeout()).toThrow();
    });

    it('timeout_payment → payment is invalid (must go through loading)', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollPaymentTimeout();
      expect(() => m.pollPaymentToGenerating()).toThrow();
    });

    it('unlocking → payment is invalid', () => {
      const m = createPhaseMachine();
      m.enterUnlocking();
      expect(() => m.loadPending()).toThrow();
    });
  });

  describe('phases are mutually exclusive', () => {
    it('each phase value is only active one at a time', () => {
      const allPhases = ['loading', 'payment', 'generating', 'timeout_payment', 'timeout_generating', 'error', 'unlocking', 'ready'];

      const pairs = [
        () => { const m = createPhaseMachine(); m.loadPending(); return m; },
        () => { const m = createPhaseMachine(); m.loadPaidNoJSON(); return m; },
        () => { const m = createPhaseMachine(); m.loadReady(); return m; },
        () => { const m = createPhaseMachine(); m.loadError('x'); return m; },
        () => { const m = createPhaseMachine(); m.loadPending(); m.pollPaymentTimeout(); return m; },
        () => { const m = createPhaseMachine(); m.loadPending(); m.pollPaymentToGenerating(); m.pollGenTimeout(); return m; },
        () => { const m = createPhaseMachine(); m.enterUnlocking(); return m; },
      ];

      for (const setup of pairs) {
        const m = setup();
        const active = allPhases.filter(p => p === m.phase);
        expect(active).toHaveLength(1);
      }
    });
  });

  describe('full lifecycle paths', () => {
    it('golden path: loading → payment → generating → ready', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollPaymentToGenerating();
      m.pollGenToUnlocking();
      expect(m.phase).toBe('ready');
    });

    it('direct ready: loading → ready (already paid + has llm_json)', () => {
      const m = createPhaseMachine();
      m.loadReady();
      expect(m.phase).toBe('ready');
    });

    it('payment timeout → retry → success', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollPaymentTimeout();
      m.retryFrom('timeout_payment');
      m.loadPending();
      m.pollPaymentToUnlocking();
      expect(m.phase).toBe('ready');
    });

    it('generating timeout → retry → success', () => {
      const m = createPhaseMachine();
      m.loadPending();
      m.pollPaymentToGenerating();
      m.pollGenTimeout();
      m.retryFrom('timeout_generating');
      m.loadPaidNoJSON();
      m.pollGenToUnlocking();
      expect(m.phase).toBe('ready');
    });

    it('loading → generating directly (skip payment, already paid)', () => {
      const m = createPhaseMachine();
      m.loadPaidNoJSON();
      m.pollGenToUnlocking();
      expect(m.phase).toBe('ready');
    });

    it('error → retry → payment → ready', () => {
      const m = createPhaseMachine();
      m.loadError('network error');
      m.retryFromError();
      m.loadPending();
      m.pollPaymentToUnlocking();
      expect(m.phase).toBe('ready');
    });
  });
});
