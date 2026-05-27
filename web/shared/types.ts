// Shared types between SSG and SPA.

export type ElementCode = 'W' | 'F' | 'E' | 'M' | 'R'

export interface ElementInfo {
  code: ElementCode
  name: string
  color: string
  darkColor: string
}

export interface PillarData {
  stem: number
  branch: number
}

export interface TenGodEntry {
  stem: number
  ten_god: string
  source: string   // "stem"|"main_qi"|"mid_qi"|"minor_qi"
}

export interface LifeStageEntry {
  stem: number
  branch: number
  stage: string     // "长生"..."养"
}

export interface HiddenStemsData {
  main: number
  mid?: number
  minor?: number
}

export interface ShenShaEntry {
  name: string
  category: string   // "吉"|"凶"|"中性"
  description: string
}

export interface PillarOutputData {
  index: number
  stem: number
  branch: number
  nayin: string
  hidden_stems: HiddenStemsData
  ten_gods: TenGodEntry[]
  life_stages: LifeStageEntry[]
  shensha: ShenShaEntry[]
  is_void: boolean
  is_self_he: boolean
  self_he_name?: string
  is_kui_gang: boolean
}

export interface ChartData {
  year_pillar: PillarData
  month_pillar: PillarData
  day_pillar: PillarData
  hour_pillar: PillarData
  hidden_stems: { main: number; mid?: number; minor?: number }[]
  ten_gods: [string, string][]
  na_yin: string[]
  life_stages: string[]
  big_fortune: {
    start_age: number
    direction: string
    pillars: { stem: number; branch: number; age_start: number; age_end: number }[]
  }
  day_master: number
  element_count: Record<number, number>
  stem_branch: number
  solar_time_minutes?: number
  pillars?: PillarOutputData[]
  nayin_relations?: NaYinRelation[]
  sanqi?: string
  sanqi_name?: string
  full_he_hui?: TripleHeFull[]
  half_he?: string[]
  gong_jia?: GongJiaEntry[]
  tai_yuan_ming_gong?: TaiYuanMingGong
  zodiac?: string
  season?: string
  lunar_month?: string
  hour_range?: string
  xun_name?: string
  xun_index?: number
  wang_shuai?: Record<string, string>
  day_mansion?: MansionEntry
  pillar_bagua?: TrigramData[]
  strength?: string
  yong_shen?: string
  xi_shen?: string
  ji_shen?: string
  pattern?: string
}

export interface BirthInfo {
  year: number
  month: number
  day: number
  hour: number
  minute: number
  longitude: number
  timezone: number
  is_dst: boolean
  gender: 'male' | 'female'
}

export interface UserProfile {
  id: number
  name: string
  email: string
  email_verified: boolean
  birth_info?: BirthInfo
  created_at: string
}

export interface Report {
  id: number
  scene: string
  sub_scene: string
  title: string
  question?: string
  content: string
  engine_data: Record<string, unknown>
  status?: string
  created_at: string
  updated_at?: string
}

export interface DayMark {
  date: string
  day_pillar: { gan: string; zhi: string; na_yin: string }
  jian_chu: string
  suitable: boolean
  gan_relation: string
  zhi_relation: string
  tai_sui_relation: string
  shen_sha: string[]
  marks: string[]
  warnings: string[]
}

export interface CalendarData {
  chart: ChartData
  day_master: number
  tai_sui: number
  year_month: string
  days: DayMark[]
}

export interface MinggeData {
  day_master: number
  day_master_name: string
  strength: string
  yong_shen: string
  xi_shen: string
  pattern: string
  element_count: Record<number, number>
  content?: string
  chart?: ChartData
}

export interface DayunPillar {
  stem: number
  branch: number
  age_start: number
  age_end: number
  name: string
  element: string
  ten_god: string
}

export interface DayunData {
  start_age: number
  direction: string
  pillars: DayunPillar[]
  current_pillar_index: number
  dayun_interactions: PillarInteraction[]
  dayun_interactions_text: string
  current_shensha: ShenShaEntry[]
}

export interface StemRelation {
  stem_a: number
  stem_b: number
  type: string
  relation: string
}

export interface BranchRelation {
  branch_a: number
  branch_b: number
  type: string
  detail: string
}

export interface PillarInteraction {
  pillar_label: string
  stem_rels: StemRelation[]
  branch_rels: BranchRelation[]
}

export interface FuYinFanYinEntry {
  natal_index: number
  type: string
  detail: string
}

export interface GongJiaEntry {
  pillar_a: number
  pillar_b: number
  type: string
  branch: number
  branch_name: string
}

export interface TaiYuanMingGong {
  tai_yuan: PillarData
  ming_gong: PillarData
  shen_gong: PillarData
}

export interface XiaoYunPillar {
  age: number
  stem: number
  branch: number
  name: string
  ten_god: string
}

export interface LiunianData {
  year: number
  year_stem: number
  year_branch: number
  year_name: string
  element: string
  nayin?: string
  ten_god: string
  generates: number
  restrains: number
  current_dayun_info: string
  natal_interactions: PillarInteraction[]
  dayun_interactions: PillarInteraction[]
  three_layer_summary: string
  shensha: ShenShaEntry[]
  fuyin_fanyin: FuYinFanYinEntry[]
}

export interface LiushiData {
  time: string
  hour_stem: number
  hour_branch: number
  hour_name: string
  ten_god: string
  stem_rels: StemRelation[]
  branch_rels: BranchRelation[]
}

export interface LiuriData {
  date: string
  day_stem: number
  day_branch: number
  day_name: string
  day_nayin: string
  ten_god: string
  stem_rels: StemRelation[]
  branch_rels: BranchRelation[]
  dayun_rels: BranchRelation[]
  liunian_rels: BranchRelation[]
  shensha: ShenShaEntry[]
}

export interface DeficiencyData {
  element_count: Record<number, number>
  deficient: string[]
  excess: string[]
  suggestion: string
  recommended_scenes: string[]
}

export interface LiuyueData {
  year: number
  month: number
  month_stem: number
  month_branch: number
  month_name: string
  element: string
  ten_god: string
  generates: number
  restrains: number
  stem_rels: StemRelation[]
  branch_rels: BranchRelation[]
  shensha: ShenShaEntry[]
  tip: string
}

export interface FlowMonth {
  id: string
  name_en: string
  generates: number
  restrains: number
}

export interface FlowYearlyData {
  months: FlowMonth[]
  current: string
}

export interface DailySuggestionData {
  date: string
  day_stem: number
  day_branch: number
  day_name: string
  element: string
  suggestion_type: string
  suggestion: string
  color: string
  direction: string
}

export interface BondItem {
  other_user: {
    name: string
    identity_label: string
    identity_id: string
  }
  bond: {
    self: Record<string, number>
    other: Record<string, number>
    delta_a: Record<string, number>
    delta_b: Record<string, number>
    concord: string
  }
  source: string
  created_at: string
}

export interface MatchLink {
  id: number
  type: string
  token: string
  match_count: number
  created_at: string
}

export interface TripleHeFull {
  type: string
  name: string
  element: string
}

export interface NaYinRelation {
  from_pillar: number
  to_pillar: number
  relation: string
  detail: string
}

export interface NameResult {
  yong_shen: string
  strategy: string
  element_count: Record<number, number>
  zodiac_hint: {
    animal: string
    preferred_radicals: string[]
    avoid_radicals?: string[]
  } | null
}

export interface XiaoXianEntry {
  age: number
  branch: number
}

export interface JieQiDepth {
  term_name: string
  days_in: number
  next_term_name: string
  days_to_next: number
}

export interface RenYuanPhase {
  stem: number
  stem_name: string
  days: number
}

export interface RenYuanSiLing {
  month_branch: number
  phases: RenYuanPhase[]
  current?: RenYuanPhase
}

export interface JieQiData {
  jieqi_depth: JieQiDepth
  ren_yuan: RenYuanSiLing
  wang_shuai: Record<string, string>
}

export interface MansionEntry {
  index: number
  name: string
  animal: string
  element: string
  group: string
  group_idx: number
}

export interface TrigramData {
  index: number
  name: string
  element: string
  direction: string
}

export interface MingGuaResult {
  gua: TrigramData
  gua_number: number
  group: string
}

export interface MansionData {
  day_mansion: MansionEntry
  all_mansions: MansionEntry[]
}

export interface MingGuaData {
  ming_gua: MingGuaResult
  all_trigrams: TrigramData[]
}

export interface DayStemDirections {
  day_stem: number
  xi_shen: string
  cai_shen: string
  fu_shen: string
}

export interface DayTaboos {
  day_stem: number
  day_branch: number
  stem_taboo: string
  branch_taboo: string
}

export interface ZeRiData {
  directions: DayStemDirections
  taboos: DayTaboos
  huangdao: HuangDaoStar
  all_stars: HuangDaoStar[]
}

export interface HuangDaoStar {
  index: number
  name: string
  path: string
  sequence: number
}

export interface Mountain24 {
  index: number
  name: string
  angle: number
  element: string
  yin_yang: string
  trigram: string
  yuan_long: string
}

export interface SanYuanYun {
  year: number
  yuan: string
  yun_number: number
  yun_name: string
  start_year: number
  end_year: number
}

export interface SanYuanData {
  current: SanYuanYun
  all_periods: SanYuanYun[]
}
