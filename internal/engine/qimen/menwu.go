package qimen

// doorEntry holds named door-palace data (door/palace set at runtime).
type doorEntry struct {
	DoorName    string
	PalaceName  string
	Name        string
	Meaning     string
}

// doorPalaceTable maps (door, palace) → doorEntry for key 克应.
var doorPalaceTable = map[[2]int]doorEntry{
	// 休门
	{1, 0}: {"休", "坎", "休门入坎", "安居之象，王相钓鱼，和合百事"},
	{1, 8}: {"休", "离", "休门加离", "和合得财，文书酒食"},
	{1, 1}: {"休", "坤", "休门加坤", "求财得利，土木之工"},
	{1, 5}: {"休", "乾", "休门加乾", "出门远行，谒贵得利"},
	{1, 7}: {"休", "艮", "休门加艮", "求财有财，出行有喜"},

	// 生门
	{2, 7}: {"生", "艮", "生门入艮", "万事如意，田宅进益"},
	{2, 1}: {"生", "坤", "生门加坤", "田宅之利，内财之喜"},
	{2, 0}: {"生", "坎", "生门加坎", "利水生之财，或酒食之利"},
	{2, 8}: {"生", "离", "生门加离", "文书之喜，或火食之喜"},
	{2, 5}: {"生", "乾", "生门加乾", "公门之利，谒贵之行"},

	// 伤门 (凶)
	{3, 2}: {"伤", "震", "伤门入震", "伤灾之象，宜捕猎追捕"},
	{3, 0}: {"伤", "坎", "伤门加坎", "道路之伤，水厄之灾"},
	{3, 8}: {"伤", "离", "伤门加离", "火伤之灾，文书受损"},

	// 杜门
	{4, 3}: {"杜", "巽", "杜门入巽", "藏身之方，修造之宜"},
	{4, 1}: {"杜", "坤", "杜门加坤", "土木之阻，田宅有损"},
	{4, 5}: {"杜", "乾", "杜门加乾", "公门不利，出行有阻"},

	// 景门
	{5, 8}: {"景", "离", "景门入离", "文书之喜，身荣贵显"},
	{5, 5}: {"景", "乾", "景门加乾", "谒贵文书有成"},
	{5, 0}: {"景", "坎", "景门加坎", "文书水火之灾"},

	// 死门 (大凶)
	{6, 1}: {"死", "坤", "死门入坤", "死亡之象，宜行刑吊丧"},
	{6, 5}: {"死", "乾", "死门加乾", "尊长有灾，出行不利"},
	{6, 8}: {"死", "离", "死门加离", "炎上之灾，文书不利"},

	// 惊门 (凶)
	{7, 6}: {"惊", "兑", "惊门入兑", "口舌之灾，宜捕盗赌博"},
	{7, 5}: {"惊", "乾", "惊门加乾", "尊长不安，口舌失财"},
	{7, 2}: {"惊", "震", "惊门加震", "争斗之象，出行惊恐"},

	// 开门 (大吉)
	{8, 5}: {"开", "乾", "开门入乾", "万事亨通，贵人相助"},
	{8, 0}: {"开", "坎", "开门加坎", "水路之吉，通行无碍"},
	{8, 1}: {"开", "坤", "开门加坤", "地脉之吉，田宅之利"},
	{8, 8}: {"开", "离", "开门加离", "光明之象，文书有成"},
	{8, 2}: {"开", "震", "开门加震", "长男之利，发科之吉"},
	{8, 3}: {"开", "巽", "开门加巽", "长女之利，入仕之吉"},
	{8, 7}: {"开", "艮", "开门加艮", "山门之利，求财可得"},
	{8, 6}: {"开", "兑", "开门加兑", "说合之利，口舌得财"},
}

// computeDoorInteractions returns door interactions for each palace.
func computeDoorInteractions(pan pan) [9]DoorInteraction {
	var result [9]DoorInteraction
	for i := 0; i < 9; i++ {
		p := pan.Palaces[i]
		if p.Door == 0 {
			continue
		}
		key := [2]int{int(p.Door), i}
		if entry, ok := doorPalaceTable[key]; ok {
			result[i] = DoorInteraction{
				Door:    p.Door,
				Palace:  PalaceIndex(i + 1),
				Name:    entry.Name,
				Meaning: entry.Meaning,
			}
		} else {
			// Generic: door name + palace name
			result[i] = DoorInteraction{
				Door:    p.Door,
				Palace:  PalaceIndex(i + 1),
				Name:    p.Door.String() + "加" + PalaceIndex(i+1).String(),
				Meaning: doorAuspicious(p.Door),
			}
		}
	}
	return result
}

// doorAuspicious returns a generic description based on whether the door is auspicious.
func doorAuspicious(d DoorIndex) string {
	switch d {
	case DoorXiu, DoorSheng, DoorKai:
		return "吉门得地，谋事可成"
	case DoorDu, DoorJing:
		return "中平之门，需择时而行"
	case DoorShang, DoorSi, DoorJingMen:
		return "凶门当位，行事多阻"
	}
	return ""
}

// menPo checks if a door is 门迫 (door overcomes palace) at the given palace.
func menPo(door DoorIndex, palace PalaceIndex) bool {
	de := doorWuxing(door)
	pe := palaceWuxing(palace)
	return de != 0 && pe != 0 && pe == (de+1)%5+1
}

// menZhi checks if a door is 门制 (palace overcomes door) at the given palace.
func menZhi(door DoorIndex, palace PalaceIndex) bool {
	de := doorWuxing(door)
	pe := palaceWuxing(palace)
	return de != 0 && pe != 0 && de == (pe+1)%5+1
}

// doorWuxing returns the element of a door.
func doorWuxing(d DoorIndex) int {
	switch d {
	case DoorXiu:
		return wxShui
	case DoorSheng, DoorSi:
		return wxTu
	case DoorShang, DoorDu:
		return wxMu
	case DoorJing:
		return wxHuo
	case DoorJingMen, DoorKai:
		return wxJin
	}
	return 0
}

// findMenPo returns palaces where the door is 门迫.
func findMenPo(pan pan) []PalaceIndex {
	var result []PalaceIndex
	for i, p := range pan.Palaces {
		if p.Door != 0 && menPo(p.Door, PalaceIndex(i+1)) {
			result = append(result, PalaceIndex(i+1))
		}
	}
	return result
}

// findMenZhi returns palaces where the door is 门制.
func findMenZhi(pan pan) []PalaceIndex {
	var result []PalaceIndex
	for i, p := range pan.Palaces {
		if p.Door != 0 && menZhi(p.Door, PalaceIndex(i+1)) {
			result = append(result, PalaceIndex(i+1))
		}
	}
	return result
}
