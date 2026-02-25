package currency_test

import (
	"testing"

	"github.com/mlhmz/finances/internal/currency"
)

func TestGet_known(t *testing.T) {
	c, ok := currency.Get("EUR")
	if !ok {
		t.Fatal("expected EUR to be found")
	}
	if c.Code != "EUR" || c.Name != "Euro" || c.Symbol != "€" || c.Exponent != 2 {
		t.Errorf("unexpected EUR entry: %+v", c)
	}
}

func TestGet_unknown(t *testing.T) {
	_, ok := currency.Get("XYZ")
	if ok {
		t.Fatal("expected XYZ to not be found")
	}
}

func TestSupported_containsEUR(t *testing.T) {
	list := currency.Supported()
	if len(list) == 0 {
		t.Fatal("Supported() returned empty slice")
	}
	for _, c := range list {
		if c.Code == "EUR" {
			return
		}
	}
	t.Error("EUR not found in Supported()")
}

func TestSupported_sortedByCode(t *testing.T) {
	list := currency.Supported()
	for i := 1; i < len(list); i++ {
		if list[i].Code < list[i-1].Code {
			t.Errorf("Supported() not sorted: %s before %s", list[i-1].Code, list[i].Code)
		}
	}
}

func TestCurrency_ValueAndScan(t *testing.T) {
	eur := currency.Registry["EUR"]

	// Value → should return code string
	v, err := eur.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	if v != "EUR" {
		t.Errorf("Value() = %v, want EUR", v)
	}

	// Scan from string
	var c currency.Currency
	if err := c.Scan("EUR"); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if c.Code != "EUR" || c.Exponent != 2 {
		t.Errorf("Scan() result unexpected: %+v", c)
	}

	// Scan from []byte
	var c2 currency.Currency
	if err := c2.Scan([]byte("EUR")); err != nil {
		t.Fatalf("Scan([]byte) error: %v", err)
	}
	if c2.Code != "EUR" {
		t.Errorf("Scan([]byte) result unexpected: %+v", c2)
	}

	// Scan unknown code → error
	var c3 currency.Currency
	if err := c3.Scan("XYZ"); err == nil {
		t.Error("Scan(unknown) should return error")
	}
}
