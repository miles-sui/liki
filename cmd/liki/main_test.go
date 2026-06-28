package main

import (
	"os"
	"testing"
)

func TestBuildTime(t *testing.T) {
	if BuildTime != "dev" {
		t.Logf("BuildTime = %q (set via -ldflags)", BuildTime)
	}
}

func TestEnvOr(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		def      string
		envValue string
		setEnv   bool
		want     string
	}{
		{name: "returns default when not set", key: "TEST_ENVOR_NOPE", def: "fallback", want: "fallback"},
		{name: "returns env when set", key: "TEST_ENVOR_SET", def: "fallback", envValue: "custom", setEnv: true, want: "custom"},
		{name: "returns empty default", key: "TEST_ENVOR_NOPE2", def: "", want: ""},
		{name: "returns empty string from env", key: "TEST_ENVOR_EMPTY", def: "fallback", envValue: "", setEnv: true, want: "fallback"},
		{name: "multiple chars in key", key: "REDIS_PORT", def: "6379", want: "6379"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}
			if got := envOr(tt.key, tt.def); got != tt.want {
				t.Errorf("envOr(%q, %q) = %q, want %q", tt.key, tt.def, got, tt.want)
			}
		})
	}
}

func TestEnvOrBool(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		def      bool
		envValue string
		setEnv   bool
		want     bool
	}{
		{name: "returns default when not set", key: "TEST_ENVBOOL_NOPE", def: true, want: true},
		{name: "true from env", key: "TEST_ENVBOOL_TRUE", def: false, envValue: "true", setEnv: true, want: true},
		{name: "false from env", key: "TEST_ENVBOOL_FALSE", def: true, envValue: "false", setEnv: true, want: false},
		{name: "1 means true", key: "TEST_ENVBOOL_1", def: false, envValue: "1", setEnv: true, want: true},
		{name: "0 means false", key: "TEST_ENVBOOL_0", def: true, envValue: "0", setEnv: true, want: false},
		{name: "returns default for invalid value", key: "TEST_ENVBOOL_INVALID", def: true, envValue: "notabool", setEnv: true, want: true},
		{name: "returns default for empty string", key: "TEST_ENVBOOL_EMPTY", def: true, envValue: "", setEnv: true, want: true},
		{name: "T means true", key: "TEST_ENVBOOL_T", def: false, envValue: "T", setEnv: true, want: true},
		{name: "F means false", key: "TEST_ENVBOOL_F", def: true, envValue: "F", setEnv: true, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}
			got := envOrBool(tt.key, tt.def)
			if got != tt.want {
				t.Errorf("envOrBool(%q, %v) = %v, want %v", tt.key, tt.def, got, tt.want)
			}
		})
	}
}
