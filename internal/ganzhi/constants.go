package ganzhi

// S (生) matrix: S[j][k] = 1 iff element j nourishes element k.
// Wood→Fire→Earth→Metal→Water→Wood
var S = [5][5]float64{
	{0, 1, 0, 0, 0},
	{0, 0, 1, 0, 0},
	{0, 0, 0, 1, 0},
	{0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0},
}

// C (克) matrix: C[j][k] = 1 iff element j controls element k.
// Wood→Earth→Water→Fire→Metal→Wood
var C = [5][5]float64{
	{0, 0, 1, 0, 0},
	{0, 0, 0, 1, 0},
	{0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0},
	{0, 1, 0, 0, 0},
}

// builtinPrototypes is the hardcoded fallback when no engine is initialized.
const alpha = 0.52

var BuiltinPrototypes = map[string][5]float64{
	"W":  {0.8, -0.2, -0.2, -0.2, -0.2},
	"F":  {-0.2, 0.8, -0.2, -0.2, -0.2},
	"E":  {-0.2, -0.2, 0.8, -0.2, -0.2},
	"M":  {-0.2, -0.2, -0.2, 0.8, -0.2},
	"R":  {-0.2, -0.2, -0.2, -0.2, 0.8},
	"WF": {alpha - 0.2, (1 - alpha) - 0.2, -0.2, -0.2, -0.2},
	"FW": {(1 - alpha) - 0.2, alpha - 0.2, -0.2, -0.2, -0.2},
	"FE": {-0.2, alpha - 0.2, (1 - alpha) - 0.2, -0.2, -0.2},
	"EF": {-0.2, (1 - alpha) - 0.2, alpha - 0.2, -0.2, -0.2},
	"EM": {-0.2, -0.2, alpha - 0.2, (1 - alpha) - 0.2, -0.2},
	"ME": {-0.2, -0.2, (1 - alpha) - 0.2, alpha - 0.2, -0.2},
	"MR": {-0.2, -0.2, -0.2, alpha - 0.2, (1 - alpha) - 0.2},
	"RM": {-0.2, -0.2, -0.2, (1 - alpha) - 0.2, alpha - 0.2},
	"RW": {(1 - alpha) - 0.2, -0.2, -0.2, -0.2, alpha - 0.2},
	"WR": {alpha - 0.2, -0.2, -0.2, -0.2, (1 - alpha) - 0.2},
	"WE": {alpha - 0.2, -0.2, (1 - alpha) - 0.2, -0.2, -0.2},
	"EW": {(1 - alpha) - 0.2, -0.2, alpha - 0.2, -0.2, -0.2},
	"FM": {-0.2, alpha - 0.2, -0.2, (1 - alpha) - 0.2, -0.2},
	"MF": {-0.2, (1 - alpha) - 0.2, -0.2, alpha - 0.2, -0.2},
	"ER": {-0.2, -0.2, alpha - 0.2, -0.2, (1 - alpha) - 0.2},
	"RE": {-0.2, -0.2, (1 - alpha) - 0.2, -0.2, alpha - 0.2},
	"MW": {(1 - alpha) - 0.2, -0.2, -0.2, alpha - 0.2, -0.2},
	"WM": {alpha - 0.2, -0.2, -0.2, (1 - alpha) - 0.2, -0.2},
	"RF": {-0.2, (1 - alpha) - 0.2, -0.2, -0.2, alpha - 0.2},
	"FR": {-0.2, alpha - 0.2, -0.2, -0.2, (1 - alpha) - 0.2},
}

// SSym is S + Sᵀ — symmetric sheng (nourishing) matrix for concord.
var SSym = func() [5][5]float64 {
	var m [5][5]float64
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			m[i][j] = S[i][j] + S[j][i]
		}
	}
	return m
}()

// CSym is C + Cᵀ — symmetric ke (controlling) matrix for concord.
var CSym = func() [5][5]float64 {
	var m [5][5]float64
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			m[i][j] = C[i][j] + C[j][i]
		}
	}
	return m
}()

// SolarMonthTable maps each solar month to a one-hot vector (Σ=1).
var SolarMonthTable = map[string][5]float64{
	"寅月": {1, 0, 0, 0, 0},
	"卯月": {1, 0, 0, 0, 0},
	"辰月": {0, 0, 1, 0, 0},
	"巳月": {0, 1, 0, 0, 0},
	"午月": {0, 1, 0, 0, 0},
	"未月": {0, 0, 1, 0, 0},
	"申月": {0, 0, 0, 1, 0},
	"酉月": {0, 0, 0, 1, 0},
	"戌月": {0, 0, 1, 0, 0},
	"亥月": {0, 0, 0, 0, 1},
	"子月": {0, 0, 0, 0, 1},
	"丑月": {0, 0, 1, 0, 0},
}

// SolarMonthOrder is the canonical order of solar months in a year.
var SolarMonthOrder = [12]string{
	"寅月", "卯月", "辰月",
	"巳月", "午月", "未月",
	"申月", "酉月", "戌月",
	"亥月", "子月", "丑月",
}

// MonthNamesEN maps solar month IDs to English names.
var MonthNamesEN = map[string]string{
	"寅月": "Early Spring",
	"卯月": "Mid Spring",
	"辰月": "Late Spring",
	"巳月": "Early Summer",
	"午月": "Mid Summer",
	"未月": "Late Summer",
	"申月": "Early Autumn",
	"酉月": "Mid Autumn",
	"戌月": "Late Autumn",
	"亥月": "Early Winter",
	"子月": "Mid Winter",
	"丑月": "Late Winter",
}

// ElementCodeIndex maps element code (W/F/E/M/R) to matrix index (0-4).
func ElementCodeIndex(code string) int {
	switch code {
	case "W":
		return 0
	case "F":
		return 1
	case "E":
		return 2
	case "M":
		return 3
	case "R":
		return 4
	}
	return -1
}
