package agent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"37.7749", 37.7749},
		{"0", 0},
		{"-122.4194", -122.4194},
		{"", 0},
		{"abc", 0},
		{"3.14", 3.14},
	}
	for _, tc := range tests {
		got := parseFloat(tc.input)
		if got != tc.want {
			t.Errorf("parseFloat(%q)=%f, want %f", tc.input, got, tc.want)
		}
	}
}

func TestHandleGetCityCoords_Valid(t *testing.T) {
	orig := geoClient
	geoClient = &http.Client{
		Transport: &mockGeoTransport{
			body: `[{"lat":"39.9042","lon":"116.4074","name":"Beijing","address":{"country":"China","country_code":"CN"}}]`,
		},
	}
	defer func() { geoClient = orig }()

	args := json.RawMessage(`{"city_name":"Beijing"}`)
	result, err := handleGetCityCoords(context.Background(), args)
	if err != nil {
		t.Fatalf("handleGetCityCoords: %v", err)
	}
	var r cityCoordsResult
	if err := json.Unmarshal(result, &r); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if r.Name != "Beijing" {
		t.Errorf("name = %q, want Beijing", r.Name)
	}
	if r.Longitude != 116.4074 {
		t.Errorf("longitude = %f, want 116.4074", r.Longitude)
	}
	if r.Latitude != 39.9042 {
		t.Errorf("latitude = %f, want 39.9042", r.Latitude)
	}
	if r.Country != "China" {
		t.Errorf("country = %q, want China", r.Country)
	}
}

func TestHandleGetCityCoords_EmptyCityName(t *testing.T) {
	args := json.RawMessage(`{"city_name":""}`)
	_, err := handleGetCityCoords(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for empty city_name")
	}
	if !strings.Contains(err.Error(), "city_name is required") {
		t.Errorf("error = %v, want 'city_name is required'", err)
	}
}

func TestHandleGetCityCoords_InvalidJSON(t *testing.T) {
	args := json.RawMessage(`not-json`)
	_, err := handleGetCityCoords(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestHandleGetCityCoords_GeocodeHTTPError(t *testing.T) {
	orig := geoClient
	geoClient = &http.Client{
		Transport: &mockGeoTransport{
			status: http.StatusInternalServerError,
			body:   "server error",
		},
	}
	defer func() { geoClient = orig }()

	args := json.RawMessage(`{"city_name":"Nowhere"}`)
	_, err := handleGetCityCoords(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for geocode failure")
	}
}

func TestHandleGetCityCoords_GeocodeEmptyResults(t *testing.T) {
	orig := geoClient
	geoClient = &http.Client{
		Transport: &mockGeoTransport{
			body: `[]`,
		},
	}
	defer func() { geoClient = orig }()

	args := json.RawMessage(`{"city_name":"Xyzzy"}`)
	_, err := handleGetCityCoords(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for empty results")
	}
}

func TestHandleGetCityCoords_GeocodeMalformedJSON(t *testing.T) {
	orig := geoClient
	geoClient = &http.Client{
		Transport: &mockGeoTransport{
			body: `not json`,
		},
	}
	defer func() { geoClient = orig }()

	args := json.RawMessage(`{"city_name":"X"}`)
	_, err := handleGetCityCoords(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for malformed response")
	}
}

type mockGeoTransport struct {
	status int
	body   string
}

func (m *mockGeoTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	if m.status == 0 {
		m.status = http.StatusOK
	}
	return &http.Response{
		StatusCode: m.status,
		Body:       io.NopCloser(strings.NewReader(m.body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil
}
