package ganzhi

import "testing"

func FuzzParseGender(f *testing.F) {
	f.Add("male")
	f.Add("female")
	f.Add("男")
	f.Add("女")
	f.Add("unknown")
	f.Add("")
	f.Add("MALE")
	f.Add("Male")

	f.Fuzz(func(t *testing.T, s string) {
		g, err := ParseGender(s)
		switch {
		case s == "male" || s == "男":
			if err != nil {
				t.Errorf("ParseGender(%q) unexpected error: %v", s, err)
			}
			if g != Male {
				t.Errorf("ParseGender(%q) = %v, want Male", s, g)
			}
		case s == "female" || s == "女":
			if err != nil {
				t.Errorf("ParseGender(%q) unexpected error: %v", s, err)
			}
			if g != Female {
				t.Errorf("ParseGender(%q) = %v, want Female", s, g)
			}
		default:
			if err == nil {
				t.Errorf("ParseGender(%q) = %v, want error", s, g)
			}
		}
	})
}
