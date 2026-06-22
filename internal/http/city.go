package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type cityCoordsResult struct {
	Name      string  `json:"name"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Country   string  `json:"country"`
}

var geoClient = &http.Client{Timeout: 5 * time.Second}

func handleCity(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		respondError(w, http.StatusUnprocessableEntity, "validation_error", "name is required")
		return
	}
	result, err := geocodeNominatim(r.Context(), name)
	if err != nil {
		respondError(w, http.StatusBadRequest, "not_found", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func geocodeNominatim(ctx context.Context, query string) (cityCoordsResult, error) {
	u := "https://nominatim.openstreetmap.org/search?q=" + query +
		"&format=json&limit=1&accept-language=zh&addressdetails=1"
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
			Country string `json:"country"`
		} `json:"address"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return cityCoordsResult{}, fmt.Errorf("geocode: decode: %w", err)
	}
	if len(results) == 0 {
		return cityCoordsResult{}, fmt.Errorf("geocode: no results for %s", query)
	}

	r := results[0]
	var lon, lat float64
	json.Unmarshal([]byte(r.Lon), &lon)
	json.Unmarshal([]byte(r.Lat), &lat)
	return cityCoordsResult{
		Name:      r.Name,
		Longitude: lon,
		Latitude:  lat,
		Country:   r.Address.Country,
	}, nil
}
