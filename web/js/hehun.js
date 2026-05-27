// hehun.js — Relationship matching (合婚/恋爱匹配) page components.
// Used by /hehun/lianai and other hehun pages.

function hehunLianaiPage() {
  function defaultBirth() {
    return {
      year: 2000, month: 1, day: 1, hour: 12, minute: 0,
      longitude: 120.0, timezone: 120.0, is_dst: false, gender: 'male',
    };
  }

  return composeComponent(
    {
      birthA: defaultBirth(),
      birthADate: '2000-01-01',
      birthATime: '12:00',
      birthB: defaultBirth(),
      birthBDate: '2000-01-02',
      birthBTime: '12:00',
      loading: false,
      error: '',
      result: null,

      get auth() { return Alpine.store('auth'); },

      async compute() {
        parseBirthDatetime(this.birthADate, this.birthATime, this.birthA);
        parseBirthDatetime(this.birthBDate, this.birthBTime, this.birthB);
        this.loading = true;
        this.error = '';
        this.result = null;
        try {
          var resp = await api('/api/mingli/bazi/match', {
            method: 'POST',
            body: JSON.stringify({
              a: birthToPayload(this.birthA),
              b: birthToPayload(this.birthB),
            }),
          });
          this.result = resp.data || resp;
        } catch (e) {
          this.error = e.message || '计算出错';
        } finally {
          this.loading = false;
        }
      },
    },
  );
}
