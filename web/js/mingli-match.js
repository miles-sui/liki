// Mingli Compatibility Match page component.
// Depends on mingli.js globals: STEM_CHARS, BRANCH_CHARS, formatStem, formatBranch, formatPillar,
// parseBirthDatetime, cityDisplayName, applyCityToBirth, fetchFilteredCities, birthToPayload.

function mingliMatchPage() {
  return composeComponent(
    {
      stateA: {
        birth: {
          year: 1984, month: 2, day: 5, hour: 12, minute: 0,
          longitude: 120.0, timezone: 120.0, is_dst: false, gender: 'male',
        },
        birthDate: '1984-02-05',
        birthTime: '12:00',
        cityQuery: '',
        filteredCities: [],
        showCityList: false,
        selectedCityName: '',
        showLunar: isLocaleChinese(),
        calendarType: isLocaleChinese() ? 'lunar' : 'solar',
        lunarYear: 1984,
        lunarMonth: 1,
        lunarDay: 1,

        setLunarDate() {
          var m = this.lunarMonth;
          var leap = m > 100;
          if (leap) m -= 100;
          this.birthDate = lunarToDate(this.birth, this.lunarYear, m, this.lunarDay, leap);
          if (isChinaDST(this.birth.year, this.birth.month, this.birth.day)) {
            this.birth.is_dst = true;
          }
        },
      },
      stateB: {
        birth: {
          year: 1990, month: 6, day: 15, hour: 12, minute: 0,
          longitude: 120.0, timezone: 120.0, is_dst: false, gender: 'female',
        },
        birthDate: '1990-06-15',
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

        setLunarDate() {
          var m = this.lunarMonth;
          var leap = m > 100;
          if (leap) m -= 100;
          this.birthDate = lunarToDate(this.birth, this.lunarYear, m, this.lunarDay, leap);
          if (isChinaDST(this.birth.year, this.birth.month, this.birth.day)) {
            this.birth.is_dst = true;
          }
        },
      },
      matchResult: null,
      loading: false,
      error: '',

      getState(which) {
        return which === 'A' ? this.stateA : this.stateB;
      },

      filterCities(which) {
        var self = this;
        var key = '_cityTimer' + which;
        clearTimeout(self[key]);
        self[key] = setTimeout(async function () {
          var s = self.getState(which);
          s.filteredCities = await fetchFilteredCities(s.cityQuery);
          s.showCityList = s.filteredCities.length > 0;
        }, 200);
      },

      selectCity(c, which) {
        var s = this.getState(which);
        applyCityToBirth(c, s.birth);
        s.cityQuery = c.name_zh || c.name;
        s.selectedCityName = cityDisplayName(c);
        s.showCityList = false;
        if (CHINESE_COUNTRIES[c.country]) s.showLunar = true;
      },

      async computeMatch() {
        parseBirthDatetime(this.stateA.birthDate, this.stateA.birthTime, this.stateA.birth);
        parseBirthDatetime(this.stateB.birthDate, this.stateB.birthTime, this.stateB.birth);
        this.loading = true;
        this.error = '';
        this.matchResult = null;
        try {
          var resp = await api('/api/mingli/bazi/match', {
            method: 'POST',
            body: JSON.stringify({
              a: birthToPayload(this.stateA.birth),
              b: birthToPayload(this.stateB.birth),
            }),
          });
          var data = resp.data || resp;
          this.matchResult = {
            chartA: data.chart_a,
            chartB: data.chart_b,
            scores: data.scores,
            total: data.total,
            level: data.level,
            details: data.details,
          };
        } catch (e) {
          this.error = e.message || 'Unknown error';
        } finally {
          this.loading = false;
        }
      },

    },
  );
}
