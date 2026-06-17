package handler

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"
)

var locHTTPClient = &http.Client{Timeout: 5 * time.Second}

// locationResult holds IP geolocation data.
type locationResult struct {
	Country  string `json:"country"`
	City     string `json:"city"`
	Currency string `json:"currency"`
}

// handleLocation returns IP-based geolocation with city-level data from ip-api.com.
func handleLocation(w http.ResponseWriter, r *http.Request) {
	country := r.Header.Get("CF-IPCountry")
	city := ""

	ip := clientIP(r)
	if ip != "" && !isPrivateIP(ip) {
		resp, err := locHTTPClient.Get("https://ip-api.com/json/" + ip + "?fields=countryCode,city")
		if err == nil {
			defer resp.Body.Close()
			var raw struct {
				CountryCode string `json:"countryCode"`
				City        string `json:"city"`
			}
			if json.NewDecoder(resp.Body).Decode(&raw) == nil {
				if country == "" {
					country = raw.CountryCode
				}
				city = raw.City
			}
		}
	}

	if country == "" {
		country = "unknown"
	}

	respondJSON(w, http.StatusOK, locationResult{
		Country:  country,
		City:     city,
		Currency: detectCurrency(r),
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func isPrivateIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return true
	}
	return parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsUnspecified()
}
