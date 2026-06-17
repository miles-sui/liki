# 灵机命理术语表 · Liki Terminology

所有命理领域概念在代码中统一使用拼音标识符。英文译名参考 *BaZi — The Four Pillars of Destiny* (Joey Yap)、*The Complete Guide to Chinese Astrology* (Derek Walters) 等权威著作。

通用编程词保留英文，不在本表列出。

## 干支 · Gan-Zhi

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| gan | 天干 | Heavenly Stem | 甲～癸，共十干 |
| zhi | 地支 | Earthly Branch | 子～亥，共十二支 |
| ganzhi | 干支 | Stem-Branch | 天干地支合称 |
| zhu | 柱 | Pillar | 一个天干 + 一个地支的组合 |

Go 类型：Gan（天干，1=甲..10=癸）、Zhi（地支，1=子..12=亥）、Zhu（一柱）。

## 五行 · Wu Xing

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| wuxing | 五行 | Five Elements | 木火土金水相生相克 |
| WxMu | 木 | Wood | |
| WxHuo | 火 | Fire | |
| WxTu | 土 | Earth | |
| WxJin | 金 | Metal | |
| WxShui | 水 | Water | |
| sheng | 生 | Produce | A 生 B，如木生火 |
| ke | 克 | Control | A 克 B，如木克土 |

## 四柱 · Four Pillars

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| nianzhu | 年柱 | Year Pillar | 出生年干支 |
| yuezhu | 月柱 | Month Pillar | 出生月干支 |
| rizhu | 日柱 | Day Pillar | 出生日干支 |
| shizhu | 时柱 | Hour Pillar | 出生时干支 |

## 十神 · Ten Gods

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| shishen | 十神 | Ten Gods | 天干相对于日主的十种关系类型 |
| BiJian | 比肩 | Companion | 同五行、同阴阳 |
| JieCai | 劫财 | Rob Wealth | 同五行、异阴阳 |
| ShiShen | 食神 | Eating God | 我生、同阴阳 |
| ShangGuan | 伤官 | Hurting Officer | 我生、异阴阳 |
| PianCai | 偏财 | Indirect Wealth | 我克、同阴阳 |
| ZhengCai | 正财 | Direct Wealth | 我克、异阴阳 |
| QiSha | 七杀 | Seven Killings | 克我、同阴阳 |
| ZhengGuan | 正官 | Direct Officer | 克我、异阴阳 |
| PianYin | 偏印 | Indirect Resource | 生我、同阴阳 |
| ZhengYin | 正印 | Direct Resource | 生我、异阴阳 |

## 日主 · Day Master

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| riyuan | 日元 | Day Master | 日柱天干，代表命主自身 |

> 日柱（rizhu）指整柱干支，日元（riyuan）仅指日干。

## 藏干 · Hidden Stems

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| canggan | 藏干 | Hidden Stem | 地支中所藏的天干，分本气/中气/余气 |

## 长生十二宫 · Twelve Stages of Life

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| changsheng | 长生十二宫 | Twelve Stages of Life | 天干在地支的十二旺衰状态 |
| ChangSheng | 长生 | Birth | |
| MuYu | 沐浴 | Bath | |
| GuanDai | 冠带 | Adulthood | |
| LinGuan | 临官 | Career | |
| DiWang | 帝旺 | Peak | |
| Shuai | 衰 | Decline | |
| Bing | 病 | Sickness | |
| Si | 死 | Death | |
| Mu | 墓 | Tomb | |
| Jue | 绝 | Extinction | |
| Tai | 胎 | Conception | |
| Yang | 养 | Gestation | |

## 用神 · Yong Shen

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| yongshen | 用神 | Useful God | 对日主最有利的五行 |
| xishen | 喜神 | Favorable God | 辅助用神的五行 |
| jishen | 忌神 | Unfavorable God | 对日主不利的五行 |
| tiaohou | 调候 | Temperature Adjustment | 根据月令寒暖燥湿选用的调节用神 |
| fuyi | 扶抑 | Support & Restrain | 根据日主强弱扶弱抑强选取用神 |
| geju | 格局 | Structure | 命格分类，如正官格、从强格等 |
| qiangruo | 强弱 | Strength | 日主自身的强弱评定 |
| wangshuai | 旺衰 | Prosperity & Decline | 各五行在月令当令与否的状态 |

## 大运 · Da Yun

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| dayun | 大运 | Ten-Year Luck | 每十年一步的大运，月柱顺逆排出 |
| start_age | 起运岁数 | Starting Age | 从出生到第一步大运的岁数，保留英文 |
| direction | 顺逆 | Direction | 阳男阴女顺行、阴男阳女逆行，保留英文 |

## 流年流月流日流时 · Annual Fortune

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| liunian | 流年 | Annual Luck | 当年干支及运势 |
| liuyue | 流月 | Monthly Luck | 当月干支及运势 |
| liuri | 流日 | Daily Luck | 当日干支及运势 |
| liushi | 流时 | Hourly Luck | 当令时辰干支及运势 |

## 小运小限 · Minor Luck

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| xiaoyun | 小运 | Minor Luck | 年柱起小运，每五年一步 |
| xiaoxian | 小限 | Minor Boundary | 命宫起小限，每年一换 |

## 伏吟反吟 · Fu Yin & Fan Yin

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| fuyin | 伏吟 | Fu Yin | 干支与某柱完全相同 |
| fanyin | 反吟 | Fan Yin | 天干相克、地支相冲 |

## 合会冲刑害 · Combinations & Clashes

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| hehui | 合会 | Combinations | 天干合 + 地支三合六合三会的总集 |
| he | 合 | Combination | 干支相合，如甲己合、子丑合 |
| sanhe | 三合 | Triple Combination | 三个地支合成一局，如申子辰合水 |
| sanhui | 三会 | Triple Meeting | 三个地支同气一方，如寅卯辰会木 |
| liuhe | 六合 | Six Combination | 两两地支相合，共六组 |
| chong | 冲 | Clash | 地支相冲，如子午冲 |
| xing | 刑 | Punishment | 地支相刑，如寅巳申三刑 |
| hai | 害 | Harm | 地支相害，如子未害 |
| gongjia | 拱夹 | Gong Jia | 三合缺中或两柱夹出之格 |

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

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| taiyuan | 胎元 | Fetal Origin | 受胎之月，月柱天干前一位、地支前三位 |
| minggong | 命宫 | Destiny Palace | 安命之宫，从月柱气深逆推 |
| shengong | 身宫 | Self Palace | 安身之宫，从月柱气深顺推 |

## 纳音 · Na Yin

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| nayin | 纳音 | Na Yin | 六十甲子配五行音律，共三十种纳音 |

## 二十八宿 · 28 Mansions

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| xiu | 值日星宿 | Daily Mansion | 当日所值二十八宿之一 |

## 合盘 · Bond

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| hepan | 合盘 | Bond | 两人八字综合分析 |
| gan_rel | 天干关系 | Stem Relation | 双方天干之间的合冲关系 |
| zhi_rel | 地支关系 | Branch Relation | 双方地支之间的合冲刑害关系 |
| zhuzhu_rel | 柱柱关系 | Pillar Cross-Relation | 四柱两两之间的十六组交互 |
| shishen_rel | 十神互看 | Ten God Cross-View | 从一方日主看对方十神 |
| nayin_rel | 纳音关系 | Na Yin Relation | 纳音五行之间的生克关系 |
| shensha_rel | 神煞共现 | Star Co-occurrence | 双方神煞的共有/互补/冲突 |

## 起名 · Naming

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| wuge | 五格 | Five Patterns | 天格/人格/地格/外格/总格 |
| tiange | 天格 | Heaven Pattern | 姓氏笔画 + 1 |
| renge | 人格 | Person Pattern | 姓尾 + 名首，主运 |
| dige | 地格 | Earth Pattern | 名字笔画，前运 |
| waige | 外格 | Outer Pattern | 总格 − 人格 + 1，副运 |
| zongge | 总格 | Total Pattern | 全名总笔画数，后运 |
| sancai | 三才 | Three Talents | 天格·人格·地格三者的五行配置 |
| sanqi | 三奇 | Three Wonders | 天上/人中/地下三奇贵人 |

## 风水 · Feng Shui

| 符号 | 中文 | 英文 | 说明 |
|------|------|------|------|
| minggua | 命卦 | Life Trigram | 东西四命，基于出生年份 |
| shengqi | 生气 | Life Energy | 八宅吉方·贪狼 |
| tianyi | 天医 | Heavenly Doctor | 八宅吉方·巨门 |
| yannian | 延年 | Longevity | 八宅吉方·武曲 |
| fuwei | 伏位 | Resting Position | 八宅吉方·辅弼 |
| huohai | 祸害 | Disaster | 八宅凶方·禄存 |
| wugui | 五鬼 | Five Ghosts | 八宅凶方·廉贞 |
| liusha | 六煞 | Six Evils | 八宅凶方·文曲 |
| jueming | 绝命 | Death | 八宅凶方·破军 |
| feixing | 飞星 | Flying Star | 玄空飞星九宫 |
| sanyuan | 三元 | Three Cycles | 三元九运，上元/中元/下元 |
| ershisi_shan | 二十四山 | 24 Mountains | 罗经二十四方位 |
