package money

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/mlhmz/finances/internal/currency"
	xtextcurrency "golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Money is a lossless monetary value: an integer minor-unit amount plus its currency.
// Use New to construct, never set fields directly.
//
// GORM embedding:
//
//	type Model struct {
//	    Amount Money `gorm:"embedded;embeddedPrefix:amount_"`
//	}
//
// produces columns: amount_amount (INTEGER), amount_currency_code (TEXT).
type Money struct {
	Amount   int64
	Currency currency.Currency `gorm:"column:currency_code"`
}

// New creates a Money value from minor units and a currency.
func New(amount int64, c currency.Currency) Money {
	return Money{Amount: amount, Currency: c}
}

// Add returns the sum of m and other. Both must share the same currency.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency.Code != other.Currency.Code {
		return Money{}, fmt.Errorf("money: cannot add %s and %s", m.Currency.Code, other.Currency.Code)
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

// Subtract returns m minus other. Both must share the same currency.
func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency.Code != other.Currency.Code {
		return Money{}, fmt.Errorf("money: cannot subtract %s and %s", m.Currency.Code, other.Currency.Code)
	}
	return Money{Amount: m.Amount - other.Amount, Currency: m.Currency}, nil
}

// Format returns a compact formatted string, e.g. "€10.99".
// Uses golang.org/x/text/currency for symbol lookup and number formatting,
// then strips any locale-inserted spaces between symbol and amount.
func (m Money) Format() string {
	unit := xtextcurrency.MustParseISO(m.Currency.Code)
	divisor := math.Pow10(m.Currency.Exponent)
	p := message.NewPrinter(language.English)
	formatted := p.Sprint(xtextcurrency.Symbol(unit.Amount(float64(m.Amount) / divisor)))
	// Normalize: remove any space/no-break-space the locale inserts between symbol and digits.
	formatted = strings.ReplaceAll(formatted, "\u00a0", "") // U+00A0 no-break space
	formatted = strings.ReplaceAll(formatted, "\u202f", "") // U+202F narrow no-break space
	return strings.ReplaceAll(formatted, " ", "")
}

// IsZero reports whether the amount is zero.
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// moneyJSON is the wire format: {"amount":"10.99","currency":"EUR"}.
type moneyJSON struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// MarshalJSON encodes Money as {"amount":"10.99","currency":"EUR"}.
func (m Money) MarshalJSON() ([]byte, error) {
	sign := ""
	abs := m.Amount
	if abs < 0 {
		sign = "-"
		abs = -abs
	}
	divisor := int64(math.Pow10(m.Currency.Exponent))
	major := abs / divisor
	minor := abs % divisor
	decStr := fmt.Sprintf("%s%d.%0*d", sign, major, m.Currency.Exponent, minor)
	return json.Marshal(moneyJSON{Amount: decStr, Currency: m.Currency.Code})
}

// UnmarshalJSON decodes {"amount":"10.99","currency":"EUR"} into Money.
func (m *Money) UnmarshalJSON(data []byte) error {
	var raw moneyJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	c, ok := currency.Get(raw.Currency)
	if !ok {
		return fmt.Errorf("money: unsupported currency %q", raw.Currency)
	}
	amount, err := parseDecimal(raw.Amount, c.Exponent)
	if err != nil {
		return fmt.Errorf("money: invalid amount %q: %w", raw.Amount, err)
	}
	m.Amount = amount
	m.Currency = c
	return nil
}

// parseDecimal converts a decimal string (e.g. "10.99") to minor units using
// the given exponent (e.g. 2 for EUR → result 1099).
func parseDecimal(s string, exponent int) (int64, error) {
	negative := false
	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	}
	parts := strings.SplitN(s, ".", 2)
	majorStr := parts[0]
	minorStr := ""
	if len(parts) == 2 {
		minorStr = parts[1]
	}
	for len(minorStr) < exponent {
		minorStr += "0"
	}
	minorStr = minorStr[:exponent]

	major, err := strconv.ParseInt(majorStr, 10, 64)
	if err != nil {
		return 0, err
	}
	minor, err := strconv.ParseInt(minorStr, 10, 64)
	if err != nil {
		return 0, err
	}
	result := major*int64(math.Pow10(exponent)) + minor
	if negative {
		result = -result
	}
	return result, nil
}
