package product

import "testing"

func TestProductNaming_EmailSubject(t *testing.T) {
	if got := ProductNaming.EmailSubject(); got != "您的起名报告" {
		t.Errorf("EmailSubject() = %q, want 您的起名报告", got)
	}
}

func TestProduct_UnknownEmailSubject(t *testing.T) {
	p := Product("unknown")
	if got := p.EmailSubject(); got != "您的命理报告" {
		t.Errorf("EmailSubject() = %q, want 您的命理报告", got)
	}
}

func TestNamingAmountCents(t *testing.T) {
	tests := []struct {
		name string
		c    Currency
		want int
	}{
		{"CNY", CNY, 9900},
		{"USD", USD, 2990},
		{"unknown", Currency("EUR"), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NamingAmountCents(tt.c); got != tt.want {
				t.Errorf("NamingAmountCents(%s) = %d, want %d", tt.c, got, tt.want)
			}
		})
	}
}

func TestNewOrderID(t *testing.T) {
	id := NewOrderID()
	if id == "" {
		t.Error("NewOrderID() returned empty string")
	}
}

func TestNewOrderID_Unique(t *testing.T) {
	ids := make(map[string]bool, 100)
	for i := 0; i < 100; i++ {
		id := NewOrderID()
		if ids[id] {
			t.Fatalf("duplicate UUID: %s", id)
		}
		ids[id] = true
	}
}

func TestCurrencyConstants(t *testing.T) {
	if CNY != "CNY" {
		t.Errorf("CNY = %q, want CNY", CNY)
	}
	if USD != "USD" {
		t.Errorf("USD = %q, want USD", USD)
	}
}
