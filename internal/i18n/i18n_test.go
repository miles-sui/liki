package i18n

import (
	"net/http"
	"testing"
)

func TestT_KnownKey(t *testing.T) {
	if got := T(LangHans, "err.order_not_found"); got != "订单未找到" {
		t.Errorf("T(LangHans) = %q, want 订单未找到", got)
	}
	if got := T(LangHant, "err.order_not_found"); got != "訂單未找到" {
		t.Errorf("T(LangHant) = %q, want 訂單未找到", got)
	}
	if got := T(EN, "err.order_not_found"); got != "Order not found" {
		t.Errorf("T(EN) = %q, want Order not found", got)
	}
}

func TestT_UnknownKey(t *testing.T) {
	if got := T(LangHans, "nonexistent"); got != "nonexistent" {
		t.Errorf("T(unknown) = %q, want key itself", got)
	}
}

func TestT_FallbackToHK(t *testing.T) {
	// Lang("fr") is not in the messages map, so fallback to LangHant.
	if got := T("fr", "err.order_not_paid"); got != "訂單未支付" {
		t.Errorf("T(fr) = %q, want hk fallback 訂單未支付", got)
	}
}

func TestDetectLang_Empty(t *testing.T) {
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if got := DetectLang(r); got != LangHant {
		t.Errorf("DetectLang(empty) = %q, want LangHant", got)
	}
}

func TestDetectLang(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected Lang
	}{
		{"zh-CN", "zh-CN,zh;q=0.9,en;q=0.8", LangHans},
		{"zh-SG", "zh-SG,zh;q=0.9", LangHans},
		{"zh-TW", "zh-TW,zh;q=0.9", LangHant},
		{"zh-HK", "zh-HK,zh;q=0.9", LangHant},
		{"zh-Hant", "zh-Hant,zh;q=0.9", LangHant},
		{"zh", "zh", LangHant},
		{"en-US", "en-US,en;q=0.9", EN},
		{"en", "en", EN},
		{"zh_with_quality", "zh-CN;q=0.9,en;q=0.5", LangHans},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
			r.Header.Set("Accept-Language", tt.header)
			if got := DetectLang(r); got != tt.expected {
				t.Errorf("DetectLang(%q) = %q, want %q", tt.header, got, tt.expected)
			}
		})
	}
}
