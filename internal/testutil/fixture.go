// Package testutil provides shared test fixtures and helpers for all packages.
package testutil

import "fmt"

// Canonical test birth chart: 1984-02-15 08:00 Beijing (甲子 丙寅 己卯 戊辰, day master 己土).
const (
	BirthTime     = `"time":"1984-02-15T08:00:00+08:00"`
	BirthLong     = `"longitude":116.4`
	BirthFragment = `"birth":{` + BirthTime + `,` + BirthLong + `}`

	// AltBirthTime is an alternative birth time for tests needing a different chart.
	AltBirthTime     = `"time":"1984-02-04T06:00:00+08:00"`
	AltBirthFragment = `"birth":{` + AltBirthTime + `,` + BirthLong + `}`

	GenderMale   = `"gender":"male"`
	GenderFemale = `"gender":"female"`
)

// ChartReq builds a JSON body for chart/liunian/liuyue etc. requests.
// extra contains optional fields like "year":2025, "month":6, "gender":"female".
func ChartReq(birth, extra string) string {
	var tail string
	if extra != "" {
		tail = "," + extra
	}
	return `{` + birth + tail + `}`
}

// BondReq builds a JSON body for bond requests.
func BondReq(aBirth, aExtra, bBirth, bExtra string) string {
	a := `{"birth":{` + aBirth + `}` + extraField(aExtra) + `}`
	b := `{"birth":{` + bBirth + `}` + extraField(bExtra) + `}`
	return fmt.Sprintf(`{"a":%s,"b":%s}`, a, b)
}

func extraField(s string) string {
	if s == "" {
		return ""
	}
	return "," + s
}
