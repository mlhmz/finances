package currency

import (
	"database/sql/driver"
	"fmt"
	"sort"
)

// Currency represents an ISO 4217 currency with display metadata.
type Currency struct {
	Code     string // ISO 4217, e.g. "EUR"
	Name     string // e.g. "Euro"
	Symbol   string // e.g. "€"
	Exponent int    // decimal places, e.g. 2
}

// Registry is the single source of truth for supported currencies.
var Registry = map[string]Currency{
	"EUR": {Code: "EUR", Name: "Euro", Symbol: "€", Exponent: 2},
}

// Get returns the Currency for the given ISO 4217 code and whether it was found.
func Get(code string) (Currency, bool) {
	c, ok := Registry[code]
	return c, ok
}

// Supported returns all currencies in the registry, sorted by code.
func Supported() []Currency {
	currencies := make([]Currency, 0, len(Registry))
	for _, c := range Registry {
		currencies = append(currencies, c)
	}
	sort.Slice(currencies, func(i, j int) bool {
		return currencies[i].Code < currencies[j].Code
	})
	return currencies
}

// Value implements driver.Valuer so GORM stores Currency as a TEXT code column.
func (c Currency) Value() (driver.Value, error) {
	return c.Code, nil
}

// Scan implements sql.Scanner so GORM can reconstruct Currency from a TEXT code.
func (c *Currency) Scan(value interface{}) error {
	var code string
	switch v := value.(type) {
	case string:
		code = v
	case []byte:
		code = string(v)
	default:
		return fmt.Errorf("currency: unsupported scan type %T", value)
	}
	found, ok := Get(code)
	if !ok {
		return fmt.Errorf("currency: unsupported currency code %q", code)
	}
	*c = found
	return nil
}
