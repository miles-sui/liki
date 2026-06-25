# 八字报告模板

本模板是八字解读的权威规范。若在合盘（https://liki.hk/skills/report-bond.md）或起名（https://liki.hk/skills/report-naming.md）中需要解读单方八字，应参考本模板的解读方式，但只引述结论不做完整展开。

## 数据结构

- `nian` — 年柱
- `yue` — 月柱
- `ri` — 日柱（日主为 `ri.gan`，顶层便捷字段 `ri_yuan` 与之等价）
- `shi` — 时柱
  每柱含：
  - `gan`: 天干（如 `"甲"`）
  - `zhi`: 地支（如 `"子"`）
  - `na_yin`: 纳音（如 `"海中金"`）
  - `cang_gan`: `{main, mid, minor}` — 藏干，值为天干名或空
  - `shi_shens`: [{shi_shen, name, source, gan}] — 十神列表。`shi_shen` 为十神名，`name` 为人事含义
  - `chang_sheng`: [{stage, name, gan}] — 长生十二宫
  - `shen_sha`: [{name, category, description}] — 神煞。`category` 为 `"吉"`/`"凶"`
  - `is_void`: 是否空亡
  - `is_self_he`: 是否自合
  - `is_kui_gang`: 是否魁罡
  - `self_he_name`: 自合名称（若空则为空串）
- `solar_time`: 真太阳时（RFC3339 字符串）
- `fu_yi`: `{qiangruo, geju, yong, xi, ji}` — 扶抑用神
  - `qiangruo`: 身强/身弱/中和
  - `geju`: 格局名称
  - `yong`: 用神
  - `xi`: 喜神
  - `ji`: 忌神
- `tiao_hou`: `{season, yong, xi, ji, detail?}` — 调候用神。`detail` 可省略
- `da_yun`: `{start_age, direction, zhu, current_zhu_index}` — 大运
  - `start_age`: 起运年龄
  - `direction`: 排运方向
  - `zhu`: [{gan, zhi, age_start, age_end, name, element, shi_shen}] — 大运周期
  - `current_zhu_index`: 当前所在大运索引
- `chang_sheng`: [{name, index}] — 十二长生全局表（12 项）
- `wuxing_count`: `{木,火,土,金,水}` — 全局五行计数
- `tai_yuan_ming_gong`: `{tai_yuan:{gan,zhi}, ming_gong:{gan,zhi}, shen_gong:{gan,zhi}}` — 胎元/命宫/身宫
- `he_hui`: [{type, name, element}] — 合会（六合/三合/三会等）
- `gong_jia`: [{pillar_a, pillar_b, type, branch}] — 拱夹
- `san_qi_name`: 三奇名（若空则为空串）
- `wang_shuai`: `{木,火,土,金,水}` — 五行旺衰，值为旺/相/休/囚/死
- `nayin_rel`: [{A, B, Relation}] — 纳音关系
- `liunian` — 流年（可选）
  - `year`: 流年年份
  - `year_stem`: 流年天干
  - `year_branch`: 流年地支
  - `year_name`: 流年干支名
  - `wuxing`: 纳音五行
  - `nayin`: 纳音名
  - `shi_shen`: 流年十神
  - `generates`: 生数
  - `restrains`: 克数
  - `natal_interactions`: [{pillar_label, gan_rels, zhi_rels}] — 与四柱互动
  - `dayun_interactions`: [{...}] — 与大运互动
  - `shensha`: [{name, category, description}] — 流年神煞
  - `fuyin_fanyin`: [{natal_index, type, detail}] — 伏吟反吟

只引用数据中实际存在的字段。若某字段数据中不存在，跳过该分析维度，不要编造。

## 报告结构

### 一、格局总论

列出八字四柱表格（年/月/日/时，`gan` + `zhi` + `cang_gan` + `na_yin`），然后依次分析：
- 日主定位：`ri.gan` 是何天干、阴阳属性、在何月令（`yue.zhi`）、得令/失令
- 身强身弱：引述 `fu_yi.qiangruo`，结合得令、得地、得势说明理由
- 格局：引述 `fu_yi.geju`，简述此格局的特征
- 全局五行态势：引述 `wuxing_count`，指出最旺和最弱五行。引述 `wang_shuai`，点明各五行状态。若有 `he_hui`，说明合会对五行的影响

### 二、用神详解

- 扶抑用神：引述 `fu_yi.yong` + `fu_yi.xi` + `fu_yi.ji`，引用取用法则说明理由
- 调候用神：引述 `tiao_hou.yong` + `tiao_hou.xi` + `tiao_hou.ji`，说明调候需求（`tiao_hou.season` 为季节依据）。若 `tiao_hou.detail` 存在则引述
- 扶抑与调候的关系：一致则加强，冲突则说明取舍理由
- 五行流通：日主 → `fu_yi.yong` → `fu_yi.xi` 之间的生克链条。结合 `wuxing_count` 判断流通是否顺畅

### 三、四柱十神分析

每柱逐一解读（年/月/日/时）：
- `gan` + `zhi` 各自的十神（引述 `shi_shens`），及其人事含义
- 柱内天干与地支的关系（盖头/截脚/相生/相克）
- `cang_gan` 的藏干及透出情况
- `shen_sha` 的神煞及含义。`category` 为 `"吉"` 则正面解读，`"凶"` 则如实说明但不渲染
- `chang_sheng` 的长生状态
- 柱间互动：相邻柱的地支关系（合/冲/害/刑），天干合化
- 重点展开月柱（父母/事业宫）和时柱（子女/晚年宫）对日主的影响

若四柱全展开篇幅过长，可将年柱和日柱概述，月柱和时柱详述。

### 四、大运提示

- 引述 `da_yun.start_age` 起运年龄 + `da_yun.direction` 排运方向
- 列出 `da_yun.zhu` 全部大运周期（起止年龄 + `gan/zhi` + `name` + `shi_shen`）
- 标注关键转折运：
  - 用神运：`zhu[].element` 为用神或喜神，运势上升
  - 忌神运：`zhu[].element` 为忌神，挑战较多
  - 冲日支运：大运 `zhi` 冲 `ri.zhi`（夫妻宫），主变动
  - 伏吟运：大运干支与四柱某柱相同，主该柱事项应验
- 当前运（`current_zhu_index`）和未来 1-2 运提供具体提示（事业/财运/感情/健康）

### 五、流年分析

若数据含 `liunian` 字段：
- 流年 `year_name` + `shi_shen` + 与 `ri.gan` 关系
- 引述 `natal_interactions`：流年与各柱的互动（`gan_rels` 天干关系 + `zhi_rels` 地支关系）。重点关注冲日支（变动年）和 `fuyin_fanyin`
- 引述 `liunian.shensha`：流年神煞解读
- 综合评定该年事业/财运/感情/健康各方面的趋势
- 具体建议：应抓住的机会、应规避的风险

若数据不含 `liunian` → 跳过本章，不要用生肖或其他不基于数据的方法替代。

## 边界处理

- 某柱 `shen_sha` 为空 → 不提该柱神煞
- `cang_gan` 某层为空 → 只分析有值的藏干层
- `fu_yi.geju` 为空 → 只描述日主和月令关系，不强行归入常见格局
- `liunian` 为空 → 跳过流年章节
- `da_yun.zhu` 为空 → 说明大运数据未返回，跳过
- `he_hui` 为空 → 不提合会
- `san_qi_name` 为空 → 不提三奇
- 用户询问未排盘的问题 → 解释需先排盘，或基于已有数据给出大运/流年的宫位提示，不过度断言
