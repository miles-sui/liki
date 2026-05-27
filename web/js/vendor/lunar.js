// Lunar calendar conversion (Gregorian ↔ Chinese lunar), 1900–2100.
// Based on the Larson algorithm. MIT-licensed adaptation from @kabeep/lunar-date-fns.
var LUNAR_DATA = [
  19416, 19168, 42352, 21717, 53856, 55632, 91476, 22176, 39632, 21970,
  19168, 42422, 42192, 53840, 119381, 46400, 54944, 44450, 38320, 84343,
  18800, 42160, 46261, 27216, 27968, 109396, 11104, 38256, 21234, 18800,
  25958, 54432, 59984, 28309, 23248, 11104, 100067, 37600, 116951, 51536,
  54432, 120998, 46416, 22176, 107956, 9680, 37584, 53938, 43344, 46423,
  27808, 46416, 86869, 19872, 42416, 83315, 21168, 43432, 59728, 27296,
  44710, 43856, 19296, 43748, 42352, 21088, 62051, 55632, 23383, 22176,
  38608, 19925, 19152, 42192, 54484, 53840, 54616, 46400, 46752, 103846,
  38320, 18864, 43380, 42160, 45690, 27216, 27968, 44870, 43872, 38256,
  19189, 18800, 25776, 29859, 59984, 27480, 21952, 43872, 38613, 37600,
  51552, 55636, 54432, 55888, 30034, 22176, 43959, 9680, 37584, 51893,
  43344, 46240, 47780, 44368, 21977, 19360, 42416, 86390, 21168, 43312,
  31060, 27296, 44368, 23378, 19296, 42726, 42208, 53856, 60005, 54576,
  23200, 30371, 38608, 19195, 19152, 42192, 118966, 53840, 54560, 56645,
  46496, 22224, 21938, 18864, 42359, 42160, 43600, 111189, 27936, 44448,
  84835, 37744, 18936, 18800, 25776, 92326, 59984, 27424, 108228, 43744,
  41696, 53987, 51552, 54615, 54432, 55888, 23893, 22176, 42704, 21972,
  21200, 43448, 43344, 46240, 46758, 44368, 21920, 43940, 42416, 21168,
  45683, 26928, 29495, 27296, 44368, 84821, 19296, 42352, 21732, 53600,
  59752, 54560, 55968, 92838, 22224, 19168, 43476, 41680, 53584, 62034,
  54560
];

var LUNAR_BASE_YEAR = 1900;
var LUNAR_BASE_DATE = Date.UTC(1900, 0, 31);

function lunarGetLeapMonth(year) {
  return LUNAR_DATA[year - LUNAR_BASE_YEAR] & 15;
}

function lunarGetLeapMonthDays(year) {
  if (!lunarGetLeapMonth(year)) return 0;
  return LUNAR_DATA[year - LUNAR_BASE_YEAR] & 65536 ? 30 : 29;
}

function lunarGetMonthDays(year, month) {
  return LUNAR_DATA[year - LUNAR_BASE_YEAR] & (65536 >> month) ? 30 : 29;
}

function lunarGetYearDays(year) {
  var sum = 348;
  var d = LUNAR_DATA[year - LUNAR_BASE_YEAR];
  for (var i = 32768; i > 8; i >>= 1) {
    sum += d & i ? 1 : 0;
  }
  return sum + lunarGetLeapMonthDays(year);
}

// toLunar converts a JS Date to {year, month, day, isLeapMonth}.
function toLunar(date) {
  var target = Date.UTC(date.getFullYear(), date.getMonth(), date.getDate());
  var offset = Math.floor((target - LUNAR_BASE_DATE) / 86400000);
  if (offset < 0) return null;

  var year, yearDays;
  for (year = LUNAR_BASE_YEAR; year <= 2100 && offset > 0; year++) {
    yearDays = lunarGetYearDays(year);
    offset -= yearDays;
  }
  if (offset < 0) { year--; offset += yearDays; }
  if (year > 2100) return null;

  var leap = lunarGetLeapMonth(year);
  var month, monthDays;
  var isLeap = false;
  for (month = 1; month < 13 && offset > 0; month++) {
    if (leap > 0 && month === leap + 1 && !isLeap) {
      monthDays = lunarGetLeapMonthDays(year);
      isLeap = true;
      month--;
    } else {
      monthDays = lunarGetMonthDays(year, month);
    }
    if (isLeap && month === leap + 1) isLeap = false;
    offset -= monthDays;
  }
  if (offset === 0 && leap > 0 && month === leap + 1) {
    isLeap = !isLeap;
    if (isLeap) month--;
  }
  if (offset < 0) { month--; offset += monthDays; }
  return { year: year, month: month, day: offset + 1, isLeapMonth: isLeap };
}

// toSolar converts {year, month, day, isLeapMonth} to a JS Date (local).
function toSolar(year, month, day, isLeapMonth) {
  if (year < 1900 || year > 2100 || month < 1 || month > 12 || day < 1) return null;
  var leap = lunarGetLeapMonth(year);
  if (isLeapMonth && leap !== month) return null;
  if (day > (isLeapMonth ? lunarGetLeapMonthDays(year) : lunarGetMonthDays(year, month))) return null;

  var offset = 0;
  var leapDone = false;
  for (var y = LUNAR_BASE_YEAR; y < year; y++) offset += lunarGetYearDays(y);
  for (var m = 1; m < month; m++) {
    offset += lunarGetMonthDays(year, m);
    if (!leapDone && leap > 0 && leap <= m) {
      offset += lunarGetLeapMonthDays(year);
      leapDone = true;
    }
  }
  if (isLeapMonth) offset += lunarGetMonthDays(year, month);
  offset += day - 1;

  var tzOffset = new Date().getTimezoneOffset() * 60000;
  return new Date(LUNAR_BASE_DATE + offset * 86400000 + tzOffset);
}

// --- Chinese display helpers ---
var LUNAR_MONTH_NAMES = ['', '正月', '二月', '三月', '四月', '五月', '六月', '七月', '八月', '九月', '十月', '十一月', '十二月'];
var LUNAR_DAY_NAMES = [
  '', '初一', '初二', '初三', '初四', '初五', '初六', '初七', '初八', '初九', '初十',
  '十一', '十二', '十三', '十四', '十五', '十六', '十七', '十八', '十九', '二十',
  '廿一', '廿二', '廿三', '廿四', '廿五', '廿六', '廿七', '廿八', '廿九', '三十'
];
var CHINESE_DIGITS = '零一二三四五六七八九';

function lunarMonthName(month) { return LUNAR_MONTH_NAMES[month] || '' + month; }
function lunarDayName(day) { return LUNAR_DAY_NAMES[day] || '' + day; }

function lunarYearName(year) {
  var s = '' + year;
  var out = '';
  for (var i = 0; i < s.length; i++) {
    var d = parseInt(s[i]);
    out += CHINESE_DIGITS[d];
  }
  return out;
}

function lunarDateDisplay(year, month, day, isLeap) {
  var s = lunarYearName(year) + '年';
  if (isLeap) s += '闰';
  s += lunarMonthName(month);
  s += lunarDayName(day);
  return s;
}
