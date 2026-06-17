// Package tianwen provides astronomical computation (天文) shared by
// all Chinese metaphysics packages.
//
// Public API:
//
// Types
//   BirthTime     — 三套日历输出（公历/太阳时/农历）
//   SolarTime     — 真太阳时
//   GregorianTime — 公历时间
//   LunarTime     — 农历时间（含时辰地支）
//
// Core functions
//   ComputeBirthTime(y,m,d,h,min,lon,tz) → BirthTime
//   ComputeSolarTime(y,m,d,h,min,lon,tz) → SolarTime
//   ComputeBazi(SolarTime) → Bazi
//
// Pillar computations
//   DayPillar(year, month, day) → Zhu
//   YearPillar(year, month, day) → Zhu
//   MonthPillar(birthTime, yearGan) → Zhu
//   HourPillar(solarMinutes, dayGan) → Zhu
//
// Solar terms
//   SolarTermTime(year, targetLon) → time.Time
//   SolarTermIndex(year, month, day) → int
//   SolarMonthIndex(time) → int
//   JieQiLongitudes — [24]float64
package tianwen
