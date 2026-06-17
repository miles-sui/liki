package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

var geoClient = &http.Client{Timeout: 5 * time.Second}

type cityCoordsResult struct {
	Name      string  `json:"name"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Country   string  `json:"country"`
}

func handleGetCityCoords(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		CityName string `json:"city_name"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, fmt.Errorf("agent: city lookup: %w", err)
	}
	if args.CityName == "" {
		return nil, fmt.Errorf("city_name is required")
	}

	result, err := geocodeNominatim(ctx, args.CityName)
	if err != nil {
		return nil, fmt.Errorf("未找到城市 '%s'，请尝试附近大城市或直接提供经纬度和时区: %w", args.CityName, err)
	}
	return json.Marshal(result)
}

func geocodeNominatim(ctx context.Context, query string) (cityCoordsResult, error) {
	u := "https://nominatim.openstreetmap.org/search?" + url.Values{
		"q":               {query},
		"format":          {"json"},
		"limit":           {"1"},
		"accept-language": {"zh"},
		"addressdetails":  {"1"},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return cityCoordsResult{}, fmt.Errorf("geocode: new request: %w", err)
	}
	req.Header.Set("User-Agent", "Liki/1.0 (lingji.app)")

	resp, err := geoClient.Do(req)
	if err != nil {
		return cityCoordsResult{}, fmt.Errorf("geocode: get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return cityCoordsResult{}, fmt.Errorf("geocode: status %d", resp.StatusCode)
	}

	var results []struct {
		Lat     string `json:"lat"`
		Lon     string `json:"lon"`
		Name    string `json:"name"`
		Address struct {
			Country     string `json:"country"`
			CountryCode string `json:"country_code"`
		} `json:"address"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return cityCoordsResult{}, fmt.Errorf("geocode: decode: %w", err)
	}
	if len(results) == 0 {
		return cityCoordsResult{}, fmt.Errorf("geocode: no results for %s", query)
	}

	r := results[0]
	return cityCoordsResult{
		Name:      r.Name,
		Longitude: parseFloat(r.Lon),
		Latitude:  parseFloat(r.Lat),
		Country:   r.Address.Country,
	}, nil
}

func parseFloat(s string) float64 {
	var f float64
	if n, err := fmt.Sscanf(s, "%f", &f); n != 1 || err != nil {
		slog.Warn("city_lookup: parseFloat failed", "value", s)
		return 0
	}
	return f
}
