package agent

import (
	"context"
	"encoding/json"
	"fmt"
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
		City string `json:"city"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, fmt.Errorf("agent: city lookup: %w", err)
	}
	if args.City == "" {
		return nil, fmt.Errorf("city is required")
	}

	result, err := geocodeNominatim(ctx, args.City)
	if err != nil {
		return nil, fmt.Errorf("未找到城市 '%s'，请尝试附近大城市或直接提供经纬度和时区: %w", args.City, err)
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
	lon, err := parseFloat(r.Lon)
	if err != nil {
		return cityCoordsResult{}, fmt.Errorf("geocode: parse lon: %w", err)
	}
	lat, err := parseFloat(r.Lat)
	if err != nil {
		return cityCoordsResult{}, fmt.Errorf("geocode: parse lat: %w", err)
	}
	return cityCoordsResult{
		Name:      r.Name,
		Longitude: lon,
		Latitude:  lat,
		Country:   r.Address.Country,
	}, nil
}

func parseFloat(s string) (float64, error) {
	var f float64
	if n, err := fmt.Sscanf(s, "%f", &f); n != 1 || err != nil {
		return 0, fmt.Errorf("parseFloat: %q: %w", s, err)
	}
	return f, nil
}
