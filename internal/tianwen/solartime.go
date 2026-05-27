package tianwen

import (
	"math"

	"github.com/25types/25types/internal/ganzhi"
)

// ComputeSolarTime returns true solar time in minutes [0, 1440).
func ComputeSolarTime(year, month, day, hour, minute int, longitude, timezone float64, isDST bool) float64 {
	lst := float64(hour*60 + minute)
	if isDST {
		lst -= 60
	}
	lonOffset := 4.0 * (longitude - timezone)
	n := dayOfYear(year, month, day)
	B := 360.0 * float64(n-81) / 365.0
	BRad := B * math.Pi / 180.0
	eot := 9.87*math.Sin(2*BRad) - 7.53*math.Cos(BRad) - 1.5*math.Sin(BRad)
	ast := lst + lonOffset + eot
	ast = math.Mod(ast, 1440)
	if ast < 0 {
		ast += 1440
	}
	return ast
}

// HourBranchFromSolarTime converts solar time (minutes) to the earthly branch of the hour.
func HourBranchFromSolarTime(astMinutes float64) ganzhi.Branch {
	idx := (int(astMinutes+60) / 120) % 12
	return ganzhi.Branch(idx + 1)
}

// IsDST reports whether the given date falls in China's historical DST period (1986-1991).
func IsDST(year, month, day int) bool {
	switch year {
	case 1986:
		return inRange(month, day, 5, 4, 9, 14)
	case 1987:
		return inRange(month, day, 4, 12, 9, 13)
	case 1988:
		return inRange(month, day, 4, 10, 9, 11)
	case 1989:
		return inRange(month, day, 4, 16, 9, 17)
	case 1990:
		return inRange(month, day, 4, 15, 9, 16)
	case 1991:
		return inRange(month, day, 4, 14, 9, 15)
	}
	return false
}

func inRange(m, d, startM, startD, endM, endD int) bool {
	val := m*100 + d
	return val >= startM*100+startD && val <= endM*100+endD
}

func dayOfYear(year, month, day int) int {
	daysBefore := []int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}
	n := daysBefore[month-1] + day
	if month > 2 && isLeapYear(year) {
		n++
	}
	return n
}

func isLeapYear(y int) bool {
	return y%4 == 0 && (y%100 != 0 || y%400 == 0)
}
