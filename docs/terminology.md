# 灵机命理术语表 · Liki Terminology

所有命理领域概念在代码中统一使用拼音标识符。英文译名参考 *BaZi — The Four Pillars of Destiny* (Joey Yap)、*The Complete Guide to Chinese Astrology* (Derek Walters) 等权威著作。

通用编程词保留英文，不在本表列出。

## 基础 · Fundamentals

| 符号 | 中文 | 英文 | Go 类型 | 说明 |
|------|------|------|--------|------|
| bazi | 八字 | Eight Characters / Four Pillars | `Bazi` | 四柱干支组成的命盘 |
| yinyang | 阴阳 | Yin-Yang | `YinYang` | 事物的二元属性，阳为奇数、阴为偶数 |
| Yang | 阳 | Yang | const | 天干地支的阳性 |
| Yin | 阴 | Yin | const | 天干地支的阴性 |

## 干支 · Gan-Zhi

| 符号 | 中文 | 英文 | Go 类型 | 说明 |
|------|------|------|--------|------|
| gan | 天干 | Heavenly Stem | `Gan` | 甲～癸，共十干（1=甲..10=癸） |
| zhi | 地支 | Earthly Branch | `Zhi` | 子～亥，共十二支（1=子..12=亥） |
| ganzhi | 干支 | Stem-Branch | — | 天干地支合称 |
| zhu | 柱 | Pillar | `Zhu` | 一个天干 + 一个地支的组合 |
| GanJia…GanGui | 甲…癸 | Jia…Gui | const | 十天干常量 |
| ZhiZi…ZhiHai | 子…亥 | Zi…Hai | const | 十二地支常量 |

辅助函数：`GanName`、`ZhiName`、`ParseGan`、`ParseZhi`、`GanWuxing`、`ZhiWuxing`、`GanYinYang`、`SixtyCycleName`、`SixtyToZhu`。

## 五行 · Wu Xing

| 符号 | 中文 | 英文 | Go 类型 | 说明 |
|------|------|------|--------|------|
| wuxing | 五行 | Five Elements | `Wuxing` | 木火土金水相生相克 |
| WxMu | 木 | Wood | const | |
| WxHuo | 火 | Fire | const | |
| WxTu | 土 | Earth | const | |
| WxJin | 金 | Metal | const | |
| WxShui | 水 | Water | const | |
| sheng | 生 | Produce | func | A 生 B，如木生火 |
| ke | 克 | Control | func | A 克 B，如木克土 |

辅助函数：`ParseWuxing`。

## 四柱 · Four Pillars

| 符号 | 中文 | 英文 | Go 字段 | 说明 |
|------|------|------|--------|------|
| nianzhu | 年柱 | Year Pillar | `NianZhu` | 出生年干支 |
| yuezhu | 月柱 | Month Pillar | `YueZhu` | 出生月干支 |
| rizhu | 日柱 | Day Pillar | `RiZhu` | 出生日干支 |
| shizhu | 时柱 | Hour Pillar | `ShiZhu` | 出生时干支 |

## 日主 · Day Master

| 符号 | 中文 | 英文 | Go 字段 | 说明 |
|------|------|------|--------|------|
| riyuan | 日元 | Day Master | `RiYuan` | 日柱天干，代表命主自身 |

> 日柱（rizhu）指整柱干支，日元（riyuan）仅指日干。

## 十神 · Shi Shen

| 符号 | 中文 | 英文 | Go 常量 | 说明 |
|------|------|------|--------|------|
| shishen | 十神 | Ten Gods | `ShiShen` | 天干相对于日主的十种关系类型 |
| ShiShenBiJian | 比肩 | Companion | const | 同五行、同阴阳 |
| ShiShenJieCai | 劫财 | Rob Wealth | const | 同五行、异阴阳 |
| ShiShenShiShen | 食神 | Eating God | const | 我生、同阴阳 |
| ShiShenShangGuan | 伤官 | Hurting Officer | const | 我生、异阴阳 |
| ShiShenPianCai | 偏财 | Indirect Wealth | const | 我克、同阴阳 |
| ShiShenZhengCai | 正财 | Direct Wealth | const | 我克、异阴阳 |
| ShiShenQiSha | 七杀 | Seven Killings | const | 克我、同阴阳 |
| ShiShenZhengGuan | 正官 | Direct Officer | const | 克我、异阴阳 |
| ShiShenPianYin | 偏印 | Indirect Resource | const | 生我、同阴阳 |
| ShiShenZhengYin | 正印 | Direct Resource | const | 生我、异阴阳 |

辅助函数：`ShiShenType`、`ShiShenName`、`ShiShenFromGan`、`ParseShiShen`。

## 藏干 · Cang Gan

| 符号 | 中文 | 英文 | Go 类型 | 说明 |
|------|------|------|--------|------|
| canggan | 藏干 | Hidden Stem | `CangGan` | 地支中所藏的天干，分本气/中气/余气 |
| main | 本气 | Main Qi | 字段 | 地支藏干的主要天干 |
| mid | 中气 | Mid Qi | 字段 | 地支藏干的次要天干 |
| minor | 余气 | Minor Qi | 字段 | 地支藏干的残余天干 |

辅助函数：`CangGanForZhi`。

## 人元司令分野 · RenYuan SiLing FenYe

| 符号 | 中文 | 英文 | Go 类型 | 说明 |
|------|------|------|--------|------|
| RenYuanSiLingFenYe | 人元司令分野 | Hidden Stem Governing Phases | `RenYuanSiLingFenYe` | 月支藏干分时主事，按月令时间段划分 |
| renyuan | 人元 | Hidden Stem | — | 地支所藏天干，同藏干 |

辅助函数：`RenYuanSiLingFenYeForZhi`。

## 长生十二宫 · Chang Sheng

| 符号 | 中文 | 英文 | Go 常量 | 说明 |
|------|------|------|--------|------|
| changsheng | 长生十二宫 | Twelve Stages of Life | — | 天干在地支的十二旺衰状态 |
| ChangSheng | 长生 | Birth | stage 0 | |
| MuYu | 沐浴 | Bath | stage 1 | |
| GuanDai | 冠带 | Adulthood | stage 2 | |
| LinGuan | 临官 | Career | stage 3 | |
| DiWang | 帝旺 | Peak | stage 4 | |
| Shuai | 衰 | Decline | stage 5 | |
| Bing | 病 | Sickness | stage 6 | |
| Si | 死 | Death | stage 7 | |
| Mu | 墓 | Tomb | stage 8 | |
| Jue | 绝 | Extinction | stage 9 | |
| Tai | 胎 | Conception | stage 10 | |
| Yang | 养 | Gestation | stage 11 | |

## 纳音 · Na Yin

| 符号 | 中文 | 英文 | Go 函数 | 说明 |
|------|------|------|--------|------|
| nayin | 纳音 | Na Yin | — | 六十甲子配五行音律，共三十种纳音 |

辅助函数：`NaYinLabel`、`NaYinWuxing`。

## 用神 · Yong Shen

| 符号 | 中文 | 英文 | Go 类型 | 说明 |
|------|------|------|--------|------|
| yongshen | 用神 | Useful God | `YongShen` | 对日主最有利的五行 |
| xishen | 喜神 | Favorable God | const | 辅助用神的五行 |
| jishen | 忌神 | Unfavorable God | const | 对日主不利的五行 |
| tiaohou | 调候 | Temperature Adjustment | `TiaoHou` | 根据月令寒暖燥湿选用的调节用神 |
| fuyi | 扶抑 | Support & Restrain | `FuYi` | 根据日主强弱扶弱抑强选取用神 |
| geju | 格局 | Structure | — | 命格分类，如正官格、从强格等 |
| qiangruo | 强弱 | Strength | — | 日主自身的强弱评定 |
| wangshuai | 旺衰 | Prosperity & Decline | `WangShuai` | 各五行在月令当令与否的状态 |

辅助函数：`ParseYongShen`、`WangShuaiOf`。

## 大运 · Da Yun

| 符号 | 中文 | 英文 | Go 类型 | 说明 |
|------|------|------|--------|------|
| dayun | 大运 | Ten-Year Luck | `DaYun` | 每十年一步的大运，月柱顺逆排出 |
| dayunzhu | 大运柱 | Luck Pillar | `DaYunZhu` | 大运中的一柱干支 |
| start_age | 起运岁数 | Starting Age | 字段 | 从出生到第一步大运的岁数，保留英文 |
| direction | 顺逆 | Direction | 字段 | 阳男阴女顺行、阴男阳女逆行，保留英文 |

## 流年流月流日流时 · Liu Fortune

| 符号 | 中文 | 英文 | Go 类型 | 说明 |
|------|------|------|--------|------|
| liunian | 流年 | Annual Luck | `LiuNian` | 当年干支及运势 |
| liuyue | 流月 | Monthly Luck | `LiuYue` | 当月干支及运势 |
| liuri | 流日 | Daily Luck | `LiuRi` | 当日干支及运势 |
| liushi | 流时 | Hourly Luck | `LiuShi` | 当令时辰干支及运势 |

## 小运小限 · Xiao Yun & Xiao Xian

| 符号 | 中文 | 英文 | Go 类型 | 说明 |
|------|------|------|--------|------|
| xiaoyun | 小运 | Minor Luck | `XiaoYun` | 年柱起小运，每五年一步 |
| xiaoxian | 小限 | Minor Boundary | `XiaoXian` | 命宫起小限，每年一换 |

## 伏吟反吟 · Fu Yin & Fan Yin

| 符号 | 中文 | 英文 | Go 类型 | 说明 |
|------|------|------|--------|------|
| fuyin | 伏吟 | Fu Yin | `FuYinFanYin` | 干支与某柱完全相同 |
| fanyin | 反吟 | Fan Yin | `FuYinFanYin` | 天干相克、地支相冲 |

## 合会冲刑害 · Combinations & Clashes

| 符号 | 中文 | 英文 | Go 类型 | 说明 |
|------|------|------|--------|------|
| hehui | 合会 | Combinations | — | 天干合 + 地支三合六合三会的总集 |
| he | 合 | Combination | — | 干支相合，如甲己合、子丑合 |
| ganhe | 天干合 | Stem Combination | `GanHe` | 天干五合，甲己合土等五组 |
| zhihe | 地支合 | Branch Combination | `ZhiHe` | 地支六合，子丑合土等六组 |
| sanhe | 三合 | Triple Combination | `SanHeHui` | 三个地支合成一局，如申子辰合水 |
| sanhui | 三会 | Triple Meeting | `SanHeHui` | 三个地支同气一方，如寅卯辰会木 |
| liuhe | 六合 | Six Combination | 同上 zhihe | 两两地支相合，共六组 |
| chong | 冲 | Clash | `BranchPair` | 地支相冲，如子午冲 |
| xing | 刑 | Punishment | `Xing` | 地支相刑，如寅巳申三刑 |
| hai | 害 | Harm | `BranchPair` | 地支相害，如子未害 |
| gongjia | 拱夹 | Gong Jia | `GongJia` | 三合缺中或两柱夹出之格 |

辅助函数：`IsGanHe`、`IsZhiHe`、`IsTripleHe`、`IsTripleHui`、`IsLiuChong`、`IsXing`、`IsHai`、`IsAnHe`、`IsPo`。

## 神煞 · Shen Sha

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| shensha | 神煞 | Stars | 吉凶神煞的总称 |
| tianyi | 天乙贵人 | Heavenly Noble | 吉神，解厄之星 |
| taohua | 桃花 | Peach Blossom | 姻缘桃花，子午卯酉 |
| yima | 驿马 | Traveling Horse | 奔走变动，寅申巳亥 |
| kongwang | 空亡 | Emptiness | 旬空亡，虚无落空 |
| kuigang | 魁罡 | Kui Gang | 辰为魁、戌为罡，刚烈之星 |
| ride | 日德 | Sun Virtue | 甲寅丙辰等六日 |
| rigui | 日贵 | Sun Noble | 丁酉癸卯等四日 |
| lu | 禄 | Emolument | 十干临官之位 |

## 胎元命宫身宫 · Supplementary Palaces

| 符号 | 中文 | 英文 | Go 类型 | 说明 |
|------|------|------|--------|------|
| taiyuan | 胎元 | Fetal Origin | `TaiYuanMingGong` | 受胎之月，月柱天干前一位、地支前三位 |
| minggong | 命宫 | Destiny Palace | `TaiYuanMingGong` | 安命之宫，从月柱气深逆推 |
| shengong | 身宫 | Self Palace | `TaiYuanMingGong` | 安身之宫，从月柱气深顺推 |

## 合盘 · Bond

| 符号 | 中文 | 英文 | Go 类型/字段 | 说明 |
|------|------|------|--------|------|
| hepan | 合盘 | Bond | `Bond` | 两人八字综合分析 |
| gan_rel | 天干关系 | Stem Relation | `GanRelation` | 双方天干之间的合冲关系 |
| zhi_rel | 地支关系 | Branch Relation | `ZhiRelation` | 双方地支之间的合冲刑害关系 |
| zhuzhu_rel | 柱柱关系 | Pillar Cross-Relation | `DayRelation` | 四柱两两之间的十六组交互 |
| shishen_rel | 十神互看 | Ten God Cross-View | — | 从一方日主看对方十神 |
| nayin_rel | 纳音关系 | Na Yin Relation | `naYinGuanXi` | 纳音五行之间的生克关系 |
| shensha_rel | 神煞共现 | Star Co-occurrence | — | 双方神煞的共有/互补/冲突 |

## 起名 · Naming

| 符号 | 中文 | 英文 | Go 类型 | 说明 |
|------|------|------|--------|------|
| wuge | 五格 | Five Patterns | `WuGe` | 天格/人格/地格/外格/总格 |
| tiange | 天格 | Heaven Pattern | 字段 | 姓氏笔画 + 1 |
| renge | 人格 | Person Pattern | 字段 | 姓尾 + 名首，主运 |
| dige | 地格 | Earth Pattern | 字段 | 名字笔画，前运 |
| waige | 外格 | Outer Pattern | 字段 | 总格 − 人格 + 1，副运 |
| zongge | 总格 | Total Pattern | 字段 | 全名总笔画数，后运 |
| sancai | 三才 | Three Talents | `SanCai` | 天格·人格·地格三者的五行配置 |
| sanqi | 三奇 | Three Wonders | — | 天上/人中/地下三奇贵人 |

## 风水 · Feng Shui

| 符号 | 中文 | 英文 | Go 类型 | 说明 |
|------|------|------|--------|------|
| minggua | 命卦 | Life Trigram | `MingGua` | 东西四命，基于出生年份 |
| shengqi | 生气 | Life Energy | — | 八宅吉方·贪狼 |
| tianyi | 天医 | Heavenly Doctor | — | 八宅吉方·巨门 |
| yannian | 延年 | Longevity | — | 八宅吉方·武曲 |
| fuwei | 伏位 | Resting Position | — | 八宅吉方·辅弼 |
| huohai | 祸害 | Disaster | — | 八宅凶方·禄存 |
| wugui | 五鬼 | Five Ghosts | — | 八宅凶方·廉贞 |
| liusha | 六煞 | Six Evils | — | 八宅凶方·文曲 |
| jueming | 绝命 | Death | — | 八宅凶方·破军 |
| feixing | 飞星 | Flying Star | `FlyingStar` | 玄空飞星九宫 |
| sanyuan | 三元 | Three Cycles | `SanYuanYun` | 三元九运，上元/中元/下元 |
| ershisi_shan | 二十四山 | 24 Mountains | `Mountain24` | 罗经二十四方位 |

## 六爻 · Liu Yao

| 符号 | 中文 | 英文 | Go 类型 | 说明 |
|------|------|------|--------|------|
| liuyao | 六爻 | Six Lines | — | 以三枚铜钱起卦，六爻断事 |
| yongshen | 用神 | Useful God | — | 六爻中代表所测事物的爻 |
| liuqin | 六亲 | Six Relations | `LiuQin` | 父母/兄弟/妻财/官鬼/子孙/世应 |
| liushou | 六兽 | Six Animals | `LiuShou` | 青龙/朱雀/勾陈/螣蛇/白虎/玄武 |
| fuchen | 伏神 | Hidden Line | `FuShen` | 本宫六爻中未现之爻 |
| xungong | 旬空 | Xun Kong | — | 旬空之爻，甲子至癸亥循环 |

## 奇门 · Qi Men

| 符号 | 中文 | 英文 | Go 类型 | 说明 |
|------|------|------|--------|------|
| qimen | 奇门 | Qi Men | — | 奇门遁甲排盘 |
| shipan | 式盘 | Divination Board | — | 天地人神四盘布局 |
| tianpan | 天盘 | Heaven Board | — | 九星加临八门之上 |
| renpan | 人盘 | Man Board | — | 八门排布 |
| shenpan | 神盘 | Spirit Board | — | 八神排布 |
| yingqi | 营气 | Ying Qi | `YingQi` | 三元气运，阳遁/阴遁之别 |
| jushu | 局数 | Bureau Number | — | 阳遁一～九局、阴遁一～九局 |
| angans | 暗干 | Hidden Stem | — | 天盘暗藏之干 |

## 黄历 · Huang Li

| 符号 | 中文 | 英文 | Go 类型 | 说明 |
|------|------|------|--------|------|
| huangli | 黄历 | Chinese Almanac | — | 择日宜忌查询 |
| jieqi | 节气 | Solar Term | — | 二十四节气 |
| huangdao | 黄道 | Ecliptic Path | — | 十二值位：建除满平定执破危成收开闭 |
| xiu | 值日星宿 | Daily Mansion | — | 当日所值二十八宿之一 |
