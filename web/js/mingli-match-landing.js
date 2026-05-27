// mingli-match-landing.js — BaZi match link landing page /b/{token}
// Depends on mingli.js globals: formatStem, formatBranch, formatPillar, parseBirthDatetime,
// cityDisplayName, applyCityToBirth, fetchFilteredCities, birthToPayload, normalizeTime,
// lunarToDate, isChinaDST, CHINESE_COUNTRIES.

function mingliMatchLandingPage() {
  return composeComponent(
    {
      token: '',
      loading: true,
      error: '',
      submitting: false,
      result: null,
      linkInfo: null,
      hasOwnBirthInfo: false,

      get matchResult() {
        if (!this.result) return null;
        return {
          total: this.result.total,
          level: this.result.level,
          scores: this.result.scores,
          details: this.result.details,
          chartA: this.result.chart_a,
          chartB: this.result.chart_b,
        };
      },

      get currentLocale() { return window.CURRENT_LOCALE || 'en'; },
      get auth() { return Alpine.store('auth'); },

      // Birth form state (for User B)
      birth: {
        year: 1990, month: 1, day: 1, hour: 12, minute: 0,
        longitude: 120.0, timezone: 120.0, is_dst: false, gender: 'male',
      },
      birthDate: '1990-01-01',
      birthTime: '12:00',
      cityQuery: '',
      filteredCities: [],
      showCityList: false,
      selectedCityName: '',
      showLunar: isLocaleChinese(),
      calendarType: isLocaleChinese() ? 'lunar' : 'solar',
      lunarYear: 1990,
      lunarMonth: 1,
      lunarDay: 1,

      filterCities() {
        var self = this;
        clearTimeout(this._cityTimer);
        this._cityTimer = setTimeout(async function () {
          self.filteredCities = await fetchFilteredCities(self.cityQuery);
          self.showCityList = self.filteredCities.length > 0;
        }, 200);
      },

      selectCity(c) {
        applyCityToBirth(c, this.birth);
        this.cityQuery = c.name_zh || c.name;
        this.selectedCityName = cityDisplayName(c);
        this.showCityList = false;
        if (CHINESE_COUNTRIES[c.country]) this.showLunar = true;
      },

      setLunarDate() {
        var m = this.lunarMonth;
        var leap = m > 100;
        if (leap) m -= 100;
        this.birthDate = lunarToDate(this.birth, this.lunarYear, m, this.lunarDay, leap);
        if (isChinaDST(this.birth.year, this.birth.month, this.birth.day)) {
          this.birth.is_dst = true;
        }
      },

      init() {
        var path = window.location.pathname;
        var match = path.match(/\/b\/([a-zA-Z0-9_-]+)/);
        this.token = match ? match[1] : '';

        if (!this.token) {
          this.error = this.$store.locale.t('error.invalidLink');
          this.loading = false;
          return;
        }

        var self = this;
        this.$nextTick(function () {
          self.loadLink();
        });
      },

      async loadLink() {
        try {
          var resp = await api('/api/m/' + this.token);
          this.linkInfo = resp.data || resp;

          // Check if logged-in user has own birth info.
          if (this.auth.id) {
            try {
              var profResp = await api('/api/profiles/' + encodeURIComponent(this.auth.name));
              if (profResp.data && profResp.data.birth_info) {
                this.hasOwnBirthInfo = true;
              }
            } catch (e) {}
          }
        } catch (e) {
          console.error(e);
          this.error = this.$store.locale.t('error.linkInvalid');
        }
        this.loading = false;
      },

      async matchExisting() {
        this.doSubmit({ use_existing: true });
      },

      async submitWithBirth() {
        parseBirthDatetime(this.birthDate, this.birthTime, this.birth);
        this.doSubmit({ birth_info: birthToPayload(this.birth) });
      },

      async doSubmit(body) {
        this.submitting = true;
        try {
          var resp = await api('/api/m/' + this.token, {
            method: 'POST',
            body: JSON.stringify(body),
          });
          this.result = resp.data || resp;
        } catch (e) {
          console.error(e);
          this.$store.toast.error(e.message || this.$store.locale.t('error.generic'));
        } finally {
          this.submitting = false;
        }
      },
    },
  );
}
