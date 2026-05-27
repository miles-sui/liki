// Mingli formatters and helpers shared across all BaZi pages.
var STEM_CHARS = ['', '甲', '乙', '丙', '丁', '戊', '己', '庚', '辛', '壬', '癸'];
var BRANCH_CHARS = ['', '子', '丑', '寅', '卯', '辰', '巳', '午', '未', '申', '酉', '戌', '亥'];

function formatStem(s) { return STEM_CHARS[s] || '?'; }
function formatBranch(b) { return BRANCH_CHARS[b] || '?'; }
function formatPillar(s, b) { return formatStem(s) + formatBranch(b); }

function formatHiddenStems(hs) {
  if (!hs) return '';
  var parts = [formatStem(hs.main)];
  if (hs.mid) parts.push(formatStem(hs.mid));
  if (hs.minor) parts.push(formatStem(hs.minor));
  return parts.join(' ');
}

function fmtSolarTime(min) {
  if (min == null) return '';
  var h = Math.floor(min / 60);
  var m = Math.floor(min % 60);
  return String(h).padStart(2, '0') + ':' + String(m).padStart(2, '0');
}

function formatElement(el) {
  var names = { 1: '木', 2: '火', 3: '土', 4: '金', 5: '水' };
  return names[el] || String(el);
}

function mingliElemColor(el) {
  var store = Alpine.store('theme');
  return (store && store.chartColors && store.chartColors[el - 1]) || '#888';
}

function pad2(n) { return String(n).padStart(2, '0'); }
function fmtDatetimeLocal(y, m, d, h, min) {
  return [y, pad2(m), pad2(d)].join('-') + 'T' + [pad2(h), pad2(min)].join(':');
}

function adaptMingliChart(raw) {
  if (!raw) return null;
  var ec = raw.element_count || {};
  var maxEc = 0;
  for (var k in ec) { if (ec[k] > maxEc) maxEc = ec[k]; }
  return {
    pillars: raw.pillars || [raw.year_pillar, raw.month_pillar, raw.day_pillar, raw.hour_pillar],
    hidden_stems: raw.hidden_stems || [{}, {}, {}, {}],
    ten_gods: raw.ten_gods || [[], [], [], []],
    na_yin: raw.na_yin || ['', '', '', ''],
    life_stages: raw.life_stages || [],
    big_fortune: raw.big_fortune || { start_age: 0, direction: '', pillars: [] },
    day_master: raw.day_master || 0,
    element_count: ec,
    max_elem_count: maxEc,
    solar_time: raw.solar_time_minutes != null ? raw.solar_time_minutes : (raw.solar_time || 0),
    stem_branch: raw.stem_branch || 0,
  };
}

// --- shared utilities used by mingliPage, mingliMatchPage, and profilePage ---

function parseBirthDatetime(birthDate, birthTime, birth) {
  if (!birthDate) return;
  var dp = birthDate.split('-');
  var tp = (birthTime || '00:00').split(':');
  if (dp.length >= 3) {
    birth.year = parseInt(dp[0], 10);
    birth.month = parseInt(dp[1], 10);
    birth.day = parseInt(dp[2], 10);
    birth.hour = parseInt(tp[0], 10);
    birth.minute = parseInt(tp[1], 10);
  }
  birth.is_dst = isChinaDST(birth.year, birth.month, birth.day);
}

// isChinaDST checks whether the date falls in China's DST period (1986-1991).
function isChinaDST(year, month, day) {
  var s, e;
  switch (year) {
    case 1986: s = [5,4]; e = [9,14]; break;
    case 1987: s = [4,12]; e = [9,13]; break;
    case 1988: s = [4,10]; e = [9,11]; break;
    case 1989: s = [4,16]; e = [9,17]; break;
    case 1990: s = [4,15]; e = [9,16]; break;
    case 1991: s = [4,14]; e = [9,15]; break;
    default: return false;
  }
  var v = month * 100 + day;
  return v >= s[0] * 100 + s[1] && v <= e[0] * 100 + e[1];
}

function cityDisplayName(c) {
  return c.name_zh ? c.name_zh + ' (' + c.name + ', ' + c.country + ')' : c.name + ', ' + c.country;
}

function cityTimezone(c) {
  return c.country === 'CN' ? 120 : Math.round(c.lng / 15) * 15;
}

function applyCityToBirth(c, birth) {
  birth.longitude = c.lng;
  birth.timezone = cityTimezone(c);
}

async function fetchFilteredCities(query) {
  if (!query || query.length < 2) return [];
  try {
    var resp = await api('/api/mingli/bazi/cities?q=' + encodeURIComponent(query), { quiet: true });
    var items = (resp.data && resp.data.items) ? resp.data.items : [];
    return items.slice(0, 20);
  } catch (_) {
    return [];
  }
}

function birthToPayload(birth) {
  return {
    year: birth.year, month: birth.month, day: birth.day,
    hour: birth.hour, minute: birth.minute,
    longitude: birth.longitude, timezone: birth.timezone,
    is_dst: birth.is_dst, gender: birth.gender,
  };
}

// Chinese-speaking country codes for showing lunar calendar option.
var CHINESE_COUNTRIES = { CN: true, HK: true, TW: true, SG: true, MO: true };

function isLocaleChinese() {
  return typeof CURRENT_LOCALE !== 'undefined' && CURRENT_LOCALE === 'zh-CN';
}

function shouldShowLunar(country) {
  return isLocaleChinese() || !!CHINESE_COUNTRIES[country];
}

function fmtElems(ec) {
  if (!ec) return '';
  var names = { 1: '木', 2: '火', 3: '土', 4: '金', 5: '水' };
  var parts = [];
  for (var k in ec) {
    if (ec[k] > 0) parts.push(names[k] + '×' + ec[k]);
  }
  return parts.join(' ');
}

// lunarToDate converts lunar date to Gregorian, sets birth, and returns YYYY-MM-DD.
function lunarToDate(birth, year, month, day, isLeap) {
  var d = toSolar(year, month, day, isLeap);
  if (!d) return '';
  birth.year = d.getFullYear();
  birth.month = d.getMonth() + 1;
  birth.day = d.getDate();
  return [birth.year, pad2(birth.month), pad2(birth.day)].join('-');
}

function mingliPage() {
  return composeComponent(
    {
      birth: {
        year: 1984, month: 2, day: 5, hour: 12, minute: 0,
        longitude: 120.0, timezone: 120.0, is_dst: false, gender: 'male',
      },
      birthDate: '1984-02-05',
      birthTime: '12:00',
      chart: null,
      loading: false,
      error: '',
      cityQuery: '',
      filteredCities: [],
      showCityList: false,
      selectedCityName: '',
      showLunar: isLocaleChinese(),
      calendarType: isLocaleChinese() ? 'lunar' : 'solar',
      lunarYear: 1984,
      lunarMonth: 1,
      lunarDay: 1,

      get auth() { return Alpine.store('auth'); },

      setLunarDate() {
        var m = this.lunarMonth;
        var leap = m > 100;
        if (leap) m -= 100;
        this.birthDate = lunarToDate(this.birth, this.lunarYear, m, this.lunarDay, leap);
        if (isChinaDST(this.birth.year, this.birth.month, this.birth.day)) {
          this.birth.is_dst = true;
        }
      },

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

      async compute() {
        parseBirthDatetime(this.birthDate, this.birthTime, this.birth);
        this.loading = true;
        this.error = '';
        this.chart = null;
        try {
          var self = this;
          var resp = await api('/api/mingli/bazi/chart', {
            method: 'POST',
            body: JSON.stringify(birthToPayload(this.birth)),
          });
          this.chart = adaptMingliChart(resp.data || resp);
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
    },
  );
}
