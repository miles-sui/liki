package tianwen

import (
	_ "embed"
	"encoding/json"
	"strings"
	"sync"
)

// City holds a pre-loaded city entry.
type City struct {
	Name    string  `json:"name"`
	NameZh  string  `json:"name_zh,omitempty"`
	Country string  `json:"country"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

var (
	citiesOnce sync.Once
	citiesData []City
	citiesZh   []City

	//go:embed data/cities.json
	citiesJSON []byte

	//go:embed data/cities_zh.json
	citiesZhJSON []byte
)

// InitCities loads the embedded city data. Safe to call multiple times.
func InitCities() {
	citiesOnce.Do(func() {
		json.Unmarshal(citiesJSON, &citiesData)
		json.Unmarshal(citiesZhJSON, &citiesZh)
		for i := range citiesZh {
			citiesZh[i].Country = "CN"
		}
	})
}

// SearchCities returns up to 20 cities matching the query prefix.
func SearchCities(q string) []City {
	InitCities()

	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return nil
	}
	var result []City

	for _, c := range citiesZh {
		if len(result) >= 20 {
			return result
		}
		if strings.HasPrefix(strings.ToLower(c.NameZh), q) ||
			strings.HasPrefix(strings.ToLower(c.Name), q) {
			result = append(result, c)
		}
	}

	for _, c := range citiesData {
		if len(result) >= 20 {
			return result
		}
		dup := false
		for _, zh := range result {
			if abs(c.Lat-zh.Lat) < 0.5 && abs(c.Lng-zh.Lng) < 0.5 && c.Country == "CN" {
				dup = true
				break
			}
		}
		if dup {
			continue
		}
		if strings.HasPrefix(strings.ToLower(c.Name), q) ||
			(c.NameZh != "" && strings.HasPrefix(strings.ToLower(c.NameZh), q)) {
			result = append(result, c)
		}
	}
	return result
}

// LoadedCities returns the pre-loaded worldwide city list.
func LoadedCities() []City {
	InitCities()
	return citiesData
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
