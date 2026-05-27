// naming.js — Naming (起名) page components.
// Shared across /qiming/baobao, /qiming/gongsi, /qiming/gaichen, etc.

function formatBranchName(b) {
  var names = ['', '子', '丑', '寅', '卯', '辰', '巳', '午', '未', '申', '酉', '戌', '亥'];
  return names[b] || '?';
}

function formatZodiacAnimal(hint) {
  if (hint && hint.animal) return hint.animal;
  return '';
}

function namingBaobaoPage() {
  return composeComponent(
    {
      surname: '',
      birth: {
        year: 2026, month: 5, day: 24, hour: 12, minute: 0,
        longitude: 120.0, timezone: 120.0, is_dst: false, gender: 'male',
      },
      birthDate: '2026-05-24',
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

      async analyze() {
        parseBirthDatetime(this.birthDate, this.birthTime, this.birth);
        if (!this.surname) {
          this.error = '请输入姓氏 / Please enter a surname';
          return;
        }
        this.loading = true;
        this.error = '';
        this.result = null;
        try {
          var resp = await api('/api/qiming/analyze', {
            method: 'POST',
            body: JSON.stringify({
              surname: this.surname,
              birth_info: birthToPayload(this.birth),
              locale: CURRENT_LOCALE || 'zh-CN',
            }),
          });
          this.result = resp.data || resp;
          // Save birth_info if logged in
          if (this.auth.id) {
            await api('/api/users/me', {
              method: 'PATCH',
              body: JSON.stringify({ birth_info: this.birth }),
            });
          }
        } catch (e) {
          this.error = e.message || 'Unknown error';
        } finally {
          this.loading = false;
        }
      },

      reset() {
        this.result = null;
        this.error = '';
      },
    },
  );
}

function formatElementChinese(el) {
  var names = { Wood: '木', Fire: '火', Earth: '土', Metal: '金', Water: '水' };
  return names[el] || el || '';
}

function elementColor(el) {
  var colors = { Wood: '#1A7A6F', Fire: '#CB3B2D', Earth: '#9A7410', Metal: '#8E8880', Water: '#1C2238' };
  return colors[el] || '#888';
}
