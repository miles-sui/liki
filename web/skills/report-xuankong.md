# 玄空风水报告模板

基于 chart 和 sanyuan API 返回数据生成玄空飞星风水报告。

## 数据来源

- chart（`xuankong.chart`）：`yun`（`year`/`yuan`/`yun_number`/`yun_name`/`start_year`/`end_year`）、`sit_mountain`/`face_mountain`（0-23 二十四山索引）、`palaces[]`（九宫各含 `palace_num`/`period_star`/`mountain_star`/`facing_star`，每星 `number`/`color`/`name`/`wuxing`/`auspicious`）、`wang_shan`/`wang_xiang`/`shan_xing`/`xia_shui`/`fan_yin`/`fu_yin`、`xing_jia_hui[]`（`shan_num`/`xiang_num`/`name`/`meaning`/`auspicious`）、`shou_shan_chu_sha`（`zheng_shen`/`ling_shen`/`shou_shan`/`chu_sha`/`assessment`）
- sanyuan（`xuankong.sanyuan`）：返回 `yun` 同上

只引用数据中实际存在的字段。若某字段数据中不存在，跳过该分析维度，不编造。

## 报告结构

### 一、三元九运与坐向

引述 `yun.yuan` + `yun.yun_number` + `yun.yun_name`（`yun.start_year` — `yun.end_year`），简述本运特点和当旺之星。列出 `sit_mountain` 和 `face_mountain` 对应的山名和五行，确认坐向关系。

### 二、九宫飞星盘

按九宫列出飞星分布表（宫位 → 运星/山星/向星），推荐用 3×3 表格格式。每宫三星逐个解读：运星为该宫时间能量，山星管丁（健康人丁），向星管财（财运事业）。重点展开：
- 当旺星（与运星同数）所在宫位，为最旺之位
- 山向飞星组合的五行生克关系
- 各宫 `auspicious` 标识，吉星位置宜布置重要功能区

### 三、格局判断

逐项引述 `wang_shan`/`wang_xiang`/`shan_xing`/`xia_shui`/`fan_yin`/`fu_yin`：
- 命中山向格局及含义
- 若为旺山旺向 → 重点展开，说明该宅丁财两旺的潜力
- 若有上山下水/反吟伏吟 → 如实说明但仍指出可用方位（吉星所在宫）

### 四、双星加会

逐宫引述 `xing_jia_hui`：
- 列出 `name` 组合名 + `meaning` 含义
- `auspicious` 为 true → 该宫组合吉利，可作为重要功能区
- `auspicious` 为 false → 该宫需化解，给出常见化解思路

### 五、收山出煞

引述 `shou_shan_chu_sha`：`zheng_shen`/`ling_shen` 方位 + `shou_shan`/`chu_sha` 是否得位 + `assessment` 综合评语。结合外部环境给出建议：正神方忌见水，零神方忌见山。

### 六、综合建议

基于全局分析：
- 最吉方位及最佳用途（如安床、书桌）
- 需化解方位及方法（如金属、水法）
- 若格局不理想，强调飞星随运而变，可择时再调整

## 边界处理

- `yun` 为空 → 告知当前元运数据未返回，跳过元运章节
- `palaces` 为空 → 飞星盘未生成，跳过
- `xing_jia_hui` 为空 → 跳过双星加会章节
- `shou_shan_chu_sha` 字段为空 → 跳过后半部分，只列出正神零神
- 格局各项全为 false → 说明该宅中规中矩，重点看吉星方位利用
- `wang_shan` 和 `wang_xiang` 同时为 true → 强调旺山旺向为最佳格局但不常见
- 坐向索引超出 0-23 → 提示参数可能有误
