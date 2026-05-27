package bazi

var xunNames = [6]string{
	"甲子旬", "甲戌旬", "甲申旬", "甲午旬", "甲辰旬", "甲寅旬",
}

// XunName returns the 旬 name for the day pillar (e.g. "甲子旬").
func XunName(dayPillar Pillar) string {
	sbIdx := sixtyCycleName(dayPillar.Stem, dayPillar.Branch)
	return xunNames[sbIdx/10]
}

// XunIndex returns the xun index (0-5) for a day pillar.
func XunIndex(dayPillar Pillar) int {
	return sixtyCycleName(dayPillar.Stem, dayPillar.Branch) / 10
}
