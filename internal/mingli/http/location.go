package minglihttp

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/25types/25types/internal/httputil"
)

// LocationResult holds a geolocation result.
type LocationResult struct {
	City    string  `json:"city"`
	Country string  `json:"country"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

var defaultLocation = LocationResult{City: "Beijing", Country: "China", Lat: 39.9042, Lng: 116.4074}

var locHTTPClient = &http.Client{Timeout: 5 * time.Second}

// LocationHandler serves geolocation endpoints.
type LocationHandler struct{}

// GET /api/location — IP-based geolocation using ip-api.com (free, no key).
func (h *LocationHandler) GetLocation(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)

	if ip == "" || isPrivateIP(ip) {
		httputil.RespondJSON(w, http.StatusOK, defaultLocation)
		return
	}

	resp, err := locHTTPClient.Get("http://ip-api.com/json/" + ip + "?fields=city,country,lat,lon")
	if err != nil {
		httputil.RespondJSON(w, http.StatusOK, defaultLocation)
		return
	}
	defer resp.Body.Close()

	var raw struct {
		City    string  `json:"city"`
		Country string  `json:"country"`
		Lat     float64 `json:"lat"`
		Lon     float64 `json:"lon"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil || raw.City == "" {
		httputil.RespondJSON(w, http.StatusOK, defaultLocation)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, LocationResult{City: raw.City, Country: raw.Country, Lat: raw.Lat, Lng: raw.Lon})
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
