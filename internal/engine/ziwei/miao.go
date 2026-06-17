package ziwei

// brightness level: 0=庙 1=旺 2=利 3=平 4=陷.
type brightness int

const (
	Miao brightness = iota
	Wang
	Li
	Ping
	Xian
)

var brightnessNames = [5]string{"庙", "旺", "利", "平", "陷"}

func (b brightness) String() string { return brightnessNames[b] }

var miaoWangTable = [14][12]brightness{
	{/*紫微*/ Ping, Miao, Miao, Wang, Li, Wang, Miao, Ping, Wang, Miao, Wang, Ping},
	{/*天机*/ Ping, Xian, Miao, Wang, Li, Wang, Miao, Ping, Li, Xian, Li, Ping},
	{/*太阳*/ Xian, Ping, Wang, Miao, Wang, Wang, Miao, Li, Li, Ping, Ping, Xian},
	{/*武曲*/ Wang, Miao, Li, Ping, Miao, Li, Wang, Miao, Li, Miao, Li, Ping},
	{/*天同*/ Wang, Ping, Li, Miao, Li, Miao, Li, Ping, Li, Ping, Ping, Miao},
	{/*廉贞*/ Li, Xian, Miao, Ping, Miao, Li, Wang, Xian, Li, Ping, Wang, Ping},
	{/*天府*/ Miao, Miao, Miao, Li, Miao, Li, Miao, Li, Li, Wang, Miao, Li},
	{/*太阴*/ Miao, Miao, Ping, Xian, Xian, Ping, Li, Li, Miao, Miao, Miao, Miao},
	{/*贪狼*/ Wang, Miao, Li, Ping, Xian, Ping, Wang, Xian, Ping, Miao, Wang, Li},
	{/*巨门*/ Wang, Xian, Miao, Li, Xian, Wang, Li, Xian, Wang, Xian, Xian, Li},
	{/*天相*/ Wang, Li, Miao, Xian, Li, Xian, Miao, Xian, Li, Li, Li, Li},
	{/*天梁*/ Miao, Miao, Li, Wang, Miao, Xian, Miao, Li, Li, Li, Li, Ping},
	{/*七杀*/ Wang, Miao, Miao, Li, Xian, Li, Wang, Xian, Wang, Li, Miao, Li},
	{/*破军*/ Wang, Miao, Li, Ping, Xian, Ping, Wang, Li, Li, Xian, Li, Xian},
}

func miaoWang(star starIndex, zhi Zhi) brightness {
	if star < 0 || int(star) >= 14 {
		return Ping
	}
	z := int(zhi) - 1
	if z < 0 || z >= 12 {
		return Ping
	}
	return miaoWangTable[star][z]
}
