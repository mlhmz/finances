package models

import "testing"

func TestDeriveInitials(t *testing.T) {
	tests := []struct {
		name     string
		fullName string
		want     string
	}{
		{"empty string", "", ""},
		{"single word", "Ada", "A"},
		{"two words", "Ada Lovelace", "AL"},
		{"three words", "Ada Betty Lovelace", "AL"},
		{"lowercase input", "malek mustafa", "MM"},
		{"extra spaces", "  Ada   Lovelace  ", "AL"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DeriveInitials(tc.fullName)
			if got != tc.want {
				t.Errorf("DeriveInitials(%q) = %q; want %q", tc.fullName, got, tc.want)
			}
		})
	}
}
