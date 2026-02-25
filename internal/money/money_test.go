package money_test

import (
	"encoding/json"
	"testing"

	"github.com/mlhmz/finances/internal/currency"
	"github.com/mlhmz/finances/internal/money"
)

var EUR = currency.Registry["EUR"]

func TestNew(t *testing.T) {
	m := money.New(1099, EUR)
	if m.Amount != 1099 || m.Currency.Code != "EUR" {
		t.Errorf("New() = %+v", m)
	}
}

func TestAdd_sameCurrency(t *testing.T) {
	m, err := money.New(1099, EUR).Add(money.New(50, EUR))
	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}
	if m.Amount != 1149 || m.Currency.Code != "EUR" {
		t.Errorf("Add() = %+v, want {1149, EUR}", m)
	}
}

func TestAdd_currencyMismatch(t *testing.T) {
	usd := currency.Currency{Code: "USD", Name: "US Dollar", Symbol: "$", Exponent: 2}
	_, err := money.New(1099, EUR).Add(money.New(50, usd))
	if err == nil {
		t.Error("Add() with different currencies should return error")
	}
}

func TestSubtract_sameCurrency(t *testing.T) {
	m, err := money.New(1099, EUR).Subtract(money.New(99, EUR))
	if err != nil {
		t.Fatalf("Subtract() error: %v", err)
	}
	if m.Amount != 1000 || m.Currency.Code != "EUR" {
		t.Errorf("Subtract() = %+v, want {1000, EUR}", m)
	}
}

func TestSubtract_currencyMismatch(t *testing.T) {
	usd := currency.Currency{Code: "USD", Name: "US Dollar", Symbol: "$", Exponent: 2}
	_, err := money.New(1099, EUR).Subtract(money.New(50, usd))
	if err == nil {
		t.Error("Subtract() with different currencies should return error")
	}
}

func TestFormat(t *testing.T) {
	got := money.New(1099, EUR).Format()
	// The x/text/currency formatter for EUR in English locale produces "€10.99".
	if got != "€10.99" {
		t.Errorf("Format() = %q, want %q", got, "€10.99")
	}
}

func TestIsZero(t *testing.T) {
	if !money.New(0, EUR).IsZero() {
		t.Error("IsZero() should be true for zero amount")
	}
	if money.New(1, EUR).IsZero() {
		t.Error("IsZero() should be false for non-zero amount")
	}
}

func TestMarshalJSON(t *testing.T) {
	b, err := json.Marshal(money.New(1099, EUR))
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	want := `{"amount":"10.99","currency":"EUR"}`
	if string(b) != want {
		t.Errorf("MarshalJSON() = %s, want %s", b, want)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	var m money.Money
	if err := json.Unmarshal([]byte(`{"amount":"10.99","currency":"EUR"}`), &m); err != nil {
		t.Fatalf("UnmarshalJSON error: %v", err)
	}
	if m.Amount != 1099 || m.Currency.Code != "EUR" {
		t.Errorf("UnmarshalJSON() = %+v, want {1099, EUR}", m)
	}
}

func TestUnmarshalJSON_unknownCurrency(t *testing.T) {
	var m money.Money
	if err := json.Unmarshal([]byte(`{"amount":"10.99","currency":"XYZ"}`), &m); err == nil {
		t.Error("UnmarshalJSON with unknown currency should return error")
	}
}

func TestMarshalUnmarshalRoundtrip(t *testing.T) {
	original := money.New(1099, EUR)
	b, _ := json.Marshal(original)
	var m money.Money
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("roundtrip UnmarshalJSON error: %v", err)
	}
	if m.Amount != original.Amount || m.Currency.Code != original.Currency.Code {
		t.Errorf("roundtrip = %+v, want %+v", m, original)
	}
}
