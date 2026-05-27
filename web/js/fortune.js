// fortune.js — Yearly/monthly fortune (流年/流月) page components.
// Used by /mingli/bazi/liunian and /mingli/bazi/liuyue.

function liunianPage() {
  return composeComponent(
    {
      birth: {
        year: 2000, month: 1, day: 1, hour: 12, minute: 0,
        longitude: 120.0, timezone: 120.0, is_dst: false, gender: 'male',
      },
      birthDate: '2000-01-01',
      birthTime: '12:00',
      year: new Date().getFullYear(),
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
          // Save birth_info if logged in, then fetch
          if (this.auth.id) {
            await api('/api/users/me', {
              method: 'PATCH',
              body: JSON.stringify({ birth_info: this.birth }),
            });
          }
          var resp = await api('/api/mingli/bazi/liunian?year=' + this.year, { quiet: true });
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

function liuyuePage() {
  return composeComponent(
    {
      birth: {
        year: 2000, month: 1, day: 1, hour: 12, minute: 0,
        longitude: 120.0, timezone: 120.0, is_dst: false, gender: 'male',
      },
      birthDate: '2000-01-01',
      birthTime: '12:00',
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
          var resp = await api('/api/mingli/bazi/liuyue', { quiet: true });
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
