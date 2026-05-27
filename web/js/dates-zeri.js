// dates-zeri.js — Date selection (择日) page components.
// Used by /zeri/jiehun, /zeri/kaiye, /zeri/qianyue, /zeri/banjia.

function zeriWeddingPage() {
  return composeComponent(
    {
      birth: {
        year: 2000, month: 1, day: 1, hour: 12, minute: 0,
        longitude: 120.0, timezone: 120.0, is_dst: false, gender: 'male',
      },
      birthDate: '2000-01-01',
      birthTime: '12:00',
      yearMonth: new Date().getFullYear() + '-' + pad2(new Date().getMonth() + 1),
      eventType: 'wedding',
      loading: false,
      error: '',
      result: null,
      cityQuery: '',
      filteredCities: [],
      showCityList: false,
      selectedCityName: '',

      get auth() { return Alpine.store('auth'); },

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
      },

      async fetch() {
        parseBirthDatetime(this.birthDate, this.birthTime, this.birth);
        this.loading = true;
        this.error = '';
        this.result = null;
        try {
          if (this.auth.id) {
            await api('/api/users/me', {
              method: 'PATCH',
              body: JSON.stringify({ birth_info: this.birth }),
            });
          }
          var resp = await api('/api/mingli/huangli/bond', {
            method: 'POST',
            body: JSON.stringify({
              month: this.yearMonth,
              event_type: this.eventType,
              birth_info: birthToPayload(this.birth),
            }),
          });
          this.result = resp.data || resp;
        } catch (e) {
          this.error = e.message || '请先保存出生信息 / Save your birth info first';
        } finally {
          this.loading = false;
        }
      },
    },
  );
}
