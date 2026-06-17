package qimen

import "liki/internal/engine/ganzhi"

// stemEntry holds the named interaction data (without stem fields set at runtime).
type stemEntry struct {
	Name        string
	PatternName string
	Meaning     string
	Auspicious  bool
}

// stemInteractionTable maps (earth, heaven) → stemEntry for named 克应.
var stemInteractionTable = map[[2]ganzhi.Gan]stemEntry{
	// 戊 combinations
	{ganzhi.GanWu, ganzhi.GanBing}: {"戊+丙", "青龙返首", "大吉大利，百事顺遂", true},
	{ganzhi.GanWu, ganzhi.GanDing}: {"戊+丁", "青龙耀明", "谒贵求名，多见喜事", true},
	{ganzhi.GanWu, ganzhi.GanYi}:  {"戊+乙", "青龙和会", "门吉事吉，门凶事凶", false},
	{ganzhi.GanWu, ganzhi.GanJi}:  {"戊+己", "贵人入狱", "公私皆不利，谋为多阻", false},
	{ganzhi.GanWu, ganzhi.GanGeng}: {"戊+庚", "值符飞宫", "吉事不吉，凶事更凶", false},
	{ganzhi.GanWu, ganzhi.GanXin}: {"戊+辛", "青龙折足", "招灾失财，需防破败", false},
	{ganzhi.GanWu, ganzhi.GanRen}: {"戊+壬", "青龙入牢", "公私皆碍，动用不吉", false},
	{ganzhi.GanWu, ganzhi.GanGui}: {"戊+癸", "青龙华盖", "门吉招福，门凶多乖", false},
	{ganzhi.GanWu, ganzhi.GanWu}:  {"戊+戊", "伏吟", "凡事闭塞，静守为吉", false},

	// 乙 combinations
	{ganzhi.GanYi, ganzhi.GanWu}:  {"乙+戊", "阴害阳门", "利阴不利阳，门凶迫制", false},
	{ganzhi.GanYi, ganzhi.GanBing}: {"乙+丙", "奇仪顺遂", "吉星临门，谋事可就", true},
	{ganzhi.GanYi, ganzhi.GanDing}: {"乙+丁", "奇仪相佐", "文书喜事，求婚可成", true},
	{ganzhi.GanYi, ganzhi.GanXin}: {"乙+辛", "青龙逃走", "奴仆拐带，测婚离散", false},
	{ganzhi.GanYi, ganzhi.GanJi}:  {"乙+己", "日奇入雾", "遁迹藏形，暗昧不明", false},
	{ganzhi.GanYi, ganzhi.GanGeng}: {"乙+庚", "日奇被刑", "争讼财产，夫妻不和", false},
	{ganzhi.GanYi, ganzhi.GanRen}: {"乙+壬", "日奇入地", "尊卑悖乱，官讼是非", false},
	{ganzhi.GanYi, ganzhi.GanGui}: {"乙+癸", "花开遇雨", "遁迹修道，隐遁为宜", false},
	{ganzhi.GanYi, ganzhi.GanYi}:  {"乙+乙", "日奇伏吟", "不宜谒贵求名，安分守己", false},

	// 丙 combinations
	{ganzhi.GanBing, ganzhi.GanWu}:  {"丙+戊", "飞鸟跌穴", "百事洞彻，谋为最吉", true},
	{ganzhi.GanBing, ganzhi.GanYi}:  {"丙+乙", "日月并行", "公私皆吉，合作可成", true},
	{ganzhi.GanBing, ganzhi.GanDing}: {"丙+丁", "星奇朱雀", "贵人文书，常人安乐", true},
	{ganzhi.GanBing, ganzhi.GanJi}:  {"丙+己", "火悖入刑", "囚人刑杖，文书不行", false},
	{ganzhi.GanBing, ganzhi.GanGeng}: {"丙+庚", "荧入太白", "门户破财，盗贼必来", false},
	{ganzhi.GanBing, ganzhi.GanXin}: {"丙+辛", "日月相会", "谋事成就，病人不凶", true},
	{ganzhi.GanBing, ganzhi.GanRen}: {"丙+壬", "火入天罗", "为客不利，是非颇多", false},
	{ganzhi.GanBing, ganzhi.GanGui}: {"丙+癸", "月奇地网", "阴人害事，暗昧不明", false},
	{ganzhi.GanBing, ganzhi.GanBing}: {"丙+丙", "月奇悖师", "文书逼迫，急迫之象", false},

	// 丁 combinations
	{ganzhi.GanDing, ganzhi.GanWu}:  {"丁+戊", "青龙转光", "官人升迁，常人威昌", true},
	{ganzhi.GanDing, ganzhi.GanYi}:  {"丁+乙", "人遁吉格", "加官进禄，常人婚喜", true},
	{ganzhi.GanDing, ganzhi.GanBing}: {"丁+丙", "星随月转", "贵人升进，常人康乐", true},
	{ganzhi.GanDing, ganzhi.GanJi}:  {"丁+己", "火入勾陈", "奸私仇冤，事因女人", false},
	{ganzhi.GanDing, ganzhi.GanGeng}: {"丁+庚", "行人必归", "音信通达，行人必至", true},
	{ganzhi.GanDing, ganzhi.GanXin}: {"丁+辛", "罪人释囚", "官人失位，罪人得赦", false},
	{ganzhi.GanDing, ganzhi.GanRen}: {"丁+壬", "五神互合", "贵人恩诏，讼狱公平", true},
	{ganzhi.GanDing, ganzhi.GanGui}: {"丁+癸", "朱雀投江", "词讼不利，音信沈溺", false},
	{ganzhi.GanDing, ganzhi.GanDing}: {"丁+丁", "星奇入太阴", "文书即至，喜事遂心", true},

	// 己 combinations
	{ganzhi.GanJi, ganzhi.GanWu}:  {"己+戊", "犬遇青龙", "门吉谋遂，门凶枉费", false},
	{ganzhi.GanJi, ganzhi.GanBing}: {"己+丙", "火悖地户", "阳人冤冤，阴人必辱", false},
	{ganzhi.GanJi, ganzhi.GanRen}: {"己+壬", "地网高张", "狡童佚女，奸情杀伤", false},
	{ganzhi.GanJi, ganzhi.GanGui}: {"己+癸", "地刑玄武", "男女疾病垂危，词讼有狱", false},
	{ganzhi.GanJi, ganzhi.GanJi}:  {"己+己", "地户逢鬼", "病者必死，百事不遂", false},

	// 庚 combinations
	{ganzhi.GanGeng, ganzhi.GanWu}:  {"庚+戊", "天乙伏宫", "百事不可谋为，大凶", false},
	{ganzhi.GanGeng, ganzhi.GanBing}: {"庚+丙", "太白入荧", "为客进利，为主破财", false},
	{ganzhi.GanGeng, ganzhi.GanDing}: {"庚+丁", "亭亭之格", "因私匿起官司，门吉有救", false},
	{ganzhi.GanGeng, ganzhi.GanRen}: {"庚+壬", "移荡格", "道路移动，音信阻隔", false},
	{ganzhi.GanGeng, ganzhi.GanGui}: {"庚+癸", "大格", "行人不至，官事不止", false},
	{ganzhi.GanGeng, ganzhi.GanGeng}: {"庚+庚", "战格", "官灾横祸，兄弟争财", false},

	// 辛 combinations
	{ganzhi.GanXin, ganzhi.GanWu}:  {"辛+戊", "困龙被伤", "官司破财，屈抑安分", false},
	{ganzhi.GanXin, ganzhi.GanYi}:  {"辛+乙", "白虎猖狂", "家败人亡，远行多殃", false},
	{ganzhi.GanXin, ganzhi.GanBing}: {"辛+丙", "干合悖师", "因财致讼，门吉则安", false},
	{ganzhi.GanXin, ganzhi.GanDing}: {"辛+丁", "狱神得奇", "经商获倍利，囚人逢赦", true},
	{ganzhi.GanXin, ganzhi.GanGeng}: {"辛+庚", "白虎出力", "刀刃相交，主客皆伤", false},
	{ganzhi.GanXin, ganzhi.GanXin}: {"辛+辛", "伏吟天庭", "公废私就，讼狱自刑", false},

	// 壬 combinations
	{ganzhi.GanRen, ganzhi.GanWu}:  {"壬+戊", "小蛇化龙", "男人发达，事业亨通", true},
	{ganzhi.GanRen, ganzhi.GanBing}: {"壬+丙", "水蛇入火", "官灾刑禁，络绎不绝", false},
	{ganzhi.GanRen, ganzhi.GanDing}: {"壬+丁", "干合蛇刑", "文书牵连，贵人顺遂", false},
	{ganzhi.GanRen, ganzhi.GanGui}: {"壬+癸", "幼女奸淫", "家有丑声，门吉不凶", false},
	{ganzhi.GanRen, ganzhi.GanRen}: {"壬+壬", "天狱自刑", "求谋无成，祸患自招", false},

	// 癸 combinations
	{ganzhi.GanGui, ganzhi.GanWu}:  {"癸+戊", "天乙会合", "吉门宜求财，婚姻喜美", true},
	{ganzhi.GanGui, ganzhi.GanYi}:  {"癸+乙", "华盖逢星", "贵人禄位，常人平安", true},
	{ganzhi.GanGui, ganzhi.GanBing}: {"癸+丙", "华盖悖师", "贱者遇贵，贵者遇贱", false},
	{ganzhi.GanGui, ganzhi.GanDing}: {"癸+丁", "寅蛇夭矫", "文书官司，火焚莫逃", false},
	{ganzhi.GanGui, ganzhi.GanGeng}: {"癸+庚", "太白入网", "以暴争讼，自罹罪责", false},
	{ganzhi.GanGui, ganzhi.GanXin}: {"癸+辛", "网盖天牢", "占病占讼，死罪莫逃", false},
	{ganzhi.GanGui, ganzhi.GanGui}: {"癸+癸", "天网四张", "行人失伴，病讼皆伤", false},
}

// computeStemInteractions returns the 十干克应 for each palace.
func computeStemInteractions(pan pan) [9]StemInteraction {
	var result [9]StemInteraction
	for i := 0; i < 9; i++ {
		p := pan.Palaces[i]
		key := [2]ganzhi.Gan{p.EarthStem, p.HeavenStem}
		if entry, ok := stemInteractionTable[key]; ok {
			result[i] = StemInteraction{
				EarthStem:  p.EarthStem,
				HeavenStem: p.HeavenStem,
				Name:       entry.Name,
				Meaning:    entry.PatternName + "：" + entry.Meaning,
				Auspicious: entry.Auspicious,
			}
		} else {
			result[i] = genericStemInteraction(p.EarthStem, p.HeavenStem)
		}
	}
	return result
}

// genericStemInteraction generates a five-element-based description for unnamed combinations.
func genericStemInteraction(earth, heaven ganzhi.Gan) StemInteraction {
	eWuxing := ganzhi.GanWuxing(earth)
	hWuxing := ganzhi.GanWuxing(heaven)
	name := ganzhi.GanName(earth) + "+" + ganzhi.GanName(heaven)

	var meaning string
	var auspicious bool
	if eWuxing == hWuxing {
		meaning = "比和，静守为宜"
		auspicious = false
	} else if eWuxing == (hWuxing%5)+1 { // heaven generates earth → 上生下
		meaning = "上生下，谋事可成"
		auspicious = true
	} else if hWuxing == (eWuxing%5)+1 { // earth generates heaven → 下生上
		meaning = "下生上，耗损有忧"
		auspicious = false
	} else if eWuxing == (hWuxing+1)%5+1 { // heaven overcomes earth → 上克下
		meaning = "上克下，主胜于客"
		auspicious = false
	} else { // earth overcomes heaven → 下克上
		meaning = "下克上，客胜于主"
		auspicious = true
	}

	return StemInteraction{
		EarthStem:  earth,
		HeavenStem: heaven,
		Name:       name,
		Meaning:    meaning,
		Auspicious: auspicious,
	}
}
